package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

const wmAgentPromptPlaceholder = "{prompt}"

const agentGracefulShutdownWait = 30 * time.Second

// runAgent invokes the configured AI CLI with task body + optional context files.
func runAgent(ctx context.Context, glob *config.GlobalConfig, task *config.Task, tc *types.TaskContext, opts *RunOptions, rd *RunDir) (*types.AgentResult, error) {
	prompt := strings.TrimSpace(task.Body)
	if prompt == "" {
		prompt = task.Name + ": process repository event."
	}
	var contextFiles []string
	if glob != nil {
		contextFiles = glob.Context.Files
	}
	for _, f := range contextFiles {
		p := filepath.Join(tc.RepoPath, f)
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		prompt += "\n\n---\n## " + f + "\n\n" + string(b)
	}
	if strings.TrimSpace(tc.CheckpointHint) != "" {
		prompt += "\n\n---\n## Previous checkpoint\n\n" + strings.TrimSpace(tc.CheckpointHint)
	}

	engineName := task.Engine()
	if engineName == "" && glob != nil {
		engineName = glob.Engine
	}

	cmdLine := os.Getenv("WM_AGENT_CMD")
	artifactFormat := agentOutputFormatForRun(cmdLine, engineName, glob)

	if rd != nil {
		if err := rd.WritePrompt(prompt); err != nil {
			return &types.AgentResult{Success: false, ExitCode: -1, Stderr: err.Error()}, err
		}
	}

	var cmd *exec.Cmd
	var stdin io.Reader
	if cmdLine != "" {
		name, args, r, err := parseWM_AGENT_CMD(cmdLine, prompt)
		if err != nil {
			return &types.AgentResult{Success: false, ExitCode: -1, Stderr: err.Error()}, err
		}
		cmd = exec.CommandContext(ctx, name, args...)
		stdin = r
	} else {
		switch strings.ToLower(strings.TrimSpace(engineName)) {
		case "codex":
			if alt := strings.TrimSpace(os.Getenv("WM_ENGINE_CODEX_CMD")); alt != "" {
				name, args, r, err := parseWM_AGENT_CMD(alt, prompt)
				if err != nil {
					return &types.AgentResult{Success: false, ExitCode: -1, Stderr: err.Error()}, err
				}
				cmd = exec.CommandContext(ctx, name, args...)
				stdin = r
			} else {
				cmd = exec.CommandContext(ctx, "codex", agentCLIArgs(glob, config.ClaudeOutputFormatText)...)
				stdin = strings.NewReader(prompt)
			}
		case "copilot":
			err := fmt.Errorf("engine copilot: set WM_AGENT_CMD to invoke your Copilot-compatible CLI")
			return &types.AgentResult{Success: false, ExitCode: -1, Stderr: err.Error()}, err
		default:
			cmd = exec.CommandContext(ctx, "claude", agentCLIArgs(glob, artifactFormat)...)
			stdin = strings.NewReader(prompt)
		}
	}
	cmd.Dir = tc.RepoPath
	env := append(os.Environ(),
		"GITHUB_REPOSITORY="+tc.Repo,
		fmt.Sprintf("WM_TASK=%s", tc.TaskName),
	)
	if tools := task.ToolsYAML(); tools != "" {
		env = append(env, "WM_TASK_TOOLS="+tools)
	}
	cmd.Env = env
	if stdin != nil {
		cmd.Stdin = stdin
	}

	setGracefulAgentCancel(cmd)

	var tail tailBuffer
	var buf bytes.Buffer
	var agentLog *os.File
	var err error

	var stdoutWriter io.Writer
	if rd != nil {
		agentLog, err = rd.OpenAgentOutput(artifactFormat)
		if err != nil {
			return &types.AgentResult{Success: false, ExitCode: -1, Stderr: err.Error()}, err
		}
		defer func() { _ = agentLog.Close() }()
		stdoutWriter = io.MultiWriter(agentLog, &tail)
		if opts != nil && opts.LogWriter != nil {
			stdoutWriter = io.MultiWriter(agentLog, &tail, opts.LogWriter)
		}
	} else {
		if opts != nil && opts.LogWriter != nil {
			stdoutWriter = io.MultiWriter(&buf, opts.LogWriter)
		} else {
			stdoutWriter = &buf
		}
	}

	cmd.Stdout = stdoutWriter
	cmd.Stderr = stdoutWriter

	err = cmd.Run()

	var combined string
	var agentPath string
	if rd != nil {
		combined = tail.String()
		agentPath = rd.AgentOutputPath(artifactFormat)
	} else {
		combined = buf.String()
	}

	res := &types.AgentResult{
		Stdout:          combined,
		Summary:         combined,
		Success:         err == nil,
		ExitCode:        0,
		AgentStdoutPath: agentPath,
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.ExitCode = -1
		}
		res.Stderr = err.Error()
		if ctx.Err() == context.DeadlineExceeded {
			res.TimedOut = true
		}
		return res, fmt.Errorf("agent: %w", err)
	}
	return res, nil
}

func setGracefulAgentCancel(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	cmd.WaitDelay = agentGracefulShutdownWait
	if runtime.GOOS == "windows" {
		return
	}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return cmd.Process.Signal(syscall.SIGTERM)
	}
}

// parseWM_AGENT_CMD builds exec name and args from WM_AGENT_CMD / WM_ENGINE_CODEX_CMD.
// If the line contains "{prompt}", it is replaced by the prompt as a single argument.
// Otherwise the prompt is appended as the last argument (backward compatible).
// When stdin is non-nil, the caller should set cmd.Stdin.
func parseWM_AGENT_CMD(cmdLine, prompt string) (name string, args []string, stdin io.Reader, err error) {
	if strings.Contains(cmdLine, wmAgentPromptPlaceholder) {
		parts := strings.SplitN(cmdLine, wmAgentPromptPlaceholder, 2)
		pre := strings.Fields(strings.TrimSpace(parts[0]))
		suf := strings.Fields(strings.TrimSpace(parts[1]))
		if len(pre) == 0 {
			return "", nil, nil, fmt.Errorf("WM_AGENT_CMD: missing executable before {prompt}")
		}
		name = pre[0]
		args = append(append(append([]string{}, pre[1:]...), prompt), suf...)
		return name, args, nil, nil
	}
	fields := strings.Fields(cmdLine)
	if len(fields) == 0 {
		return "", nil, nil, fmt.Errorf("WM_AGENT_CMD: empty")
	}
	return fields[0], append(fields[1:], prompt), nil, nil
}

// agentOutputFormatForRun selects the claude -p --output-format and run-dir filename base.
// Custom agent commands (WM_AGENT_CMD), codex, and copilot always use plain text capture.
func agentOutputFormatForRun(wmAgentCmd, engineName string, glob *config.GlobalConfig) string {
	if strings.TrimSpace(wmAgentCmd) != "" {
		return config.ClaudeOutputFormatText
	}
	switch strings.ToLower(strings.TrimSpace(engineName)) {
	case "codex", "copilot":
		return config.ClaudeOutputFormatText
	default:
		return config.EffectiveClaudeOutputFormat(glob)
	}
}

// agentCLIArgs returns argv for claude/codex non-interactive runs (prompt via stdin).
// claudeOutputFormat is only applied for the built-in claude binary; use ClaudeOutputFormatText for codex.
func agentCLIArgs(glob *config.GlobalConfig, claudeOutputFormat string) []string {
	args := []string{"-p", "--dangerously-skip-permissions"}
	if glob == nil {
		if claudeOutputFormat == config.ClaudeOutputFormatJSON || claudeOutputFormat == config.ClaudeOutputFormatStreamJSON {
			args = append(args, "--output-format", claudeOutputFormat)
		}
		return args
	}
	if glob.Model != "" {
		args = append(args, "--model", glob.Model)
	}
	if glob.MaxTurns > 0 {
		args = append(args, "--max-turns", strconv.Itoa(glob.MaxTurns))
	}
	if claudeOutputFormat == config.ClaudeOutputFormatJSON || claudeOutputFormat == config.ClaudeOutputFormatStreamJSON {
		args = append(args, "--output-format", claudeOutputFormat)
	}
	return args
}
