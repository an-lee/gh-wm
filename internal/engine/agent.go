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
	"strings"
	"syscall"
	"time"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/engine/engines"
	"github.com/an-lee/gh-wm/internal/output"
	"github.com/an-lee/gh-wm/internal/types"
)

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
	prompt += output.AvailableOutputsSection(glob, task)

	engineName := task.Engine()
	if engineName == "" && glob != nil {
		engineName = glob.Engine
	}

	cmdLine := os.Getenv("WM_AGENT_CMD")
	forceStream := opts != nil && opts.LogWriter != nil && engines.IsBuiltinClaude(cmdLine, engineName)
	appendSys := ""
	if engines.IsBuiltinClaude(cmdLine, engineName) {
		appendSys = output.SafeOutputsSystemPromptAppend(task)
	}
	cmd, stdin, artifactFormat, errBuild := engines.BuildAgentCommand(ctx, glob, engineName, cmdLine, prompt, forceStream, appendSys)
	if errBuild != nil {
		return &types.AgentResult{Success: false, ExitCode: -1, Stderr: errBuild.Error()}, errBuild
	}

	if rd != nil {
		if err := rd.WritePrompt(prompt); err != nil {
			return &types.AgentResult{Success: false, ExitCode: -1, Stderr: err.Error()}, err
		}
	}
	cmd.Dir = tc.RepoPath
	outputPath := ""
	safeOutPath := ""
	if rd != nil {
		outputPath = rd.OutputJSONPath()
		safeOutPath = rd.SafeOutputJSONLPath()
	}
	env := append(os.Environ(),
		"GITHUB_REPOSITORY="+tc.Repo,
		fmt.Sprintf("WM_TASK=%s", tc.TaskName),
		fmt.Sprintf("WM_REPO_ROOT=%s", tc.RepoPath),
	)
	if outputPath != "" {
		env = append(env, "WM_OUTPUT_FILE="+outputPath)
	}
	if safeOutPath != "" {
		env = append(env, "WM_SAFE_OUTPUT_FILE="+safeOutPath)
	}
	if tc.IssueNumber > 0 {
		env = append(env, fmt.Sprintf("WM_ISSUE_NUMBER=%d", tc.IssueNumber))
	}
	if tc.PRNumber > 0 {
		env = append(env, fmt.Sprintf("WM_PR_NUMBER=%d", tc.PRNumber))
	}
	if tools := task.ToolsYAML(); tools != "" {
		env = append(env, "WM_TASK_TOOLS="+tools)
	}
	cmd.Env = env
	if stdin != nil {
		cmd.Stdin = stdin
	}

	SetGracefulAgentCancel(cmd)

	var logW io.Writer
	if opts != nil && opts.LogWriter != nil {
		logW = opts.LogWriter
		if engines.IsBuiltinClaude(cmdLine, engineName) {
			logW = newLogStreamWriter(opts.LogWriter)
		}
	}

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
		if logW != nil {
			stdoutWriter = io.MultiWriter(agentLog, &tail, logW)
		}
	} else {
		if logW != nil {
			stdoutWriter = io.MultiWriter(&buf, logW)
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
		Stdout:             combined,
		Summary:            combined,
		Success:            err == nil,
		ExitCode:           0,
		AgentStdoutPath:    agentPath,
		OutputFilePath:     outputPath,
		SafeOutputFilePath: safeOutPath,
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
	if rd != nil {
		if agentLog != nil {
			_ = agentLog.Sync()
		}
		if stats, statErr := parseClaudeConversationArtifacts(rd.Path); statErr == nil && stats != nil && stats.LastResult != nil {
			res.LastResponseText = strings.TrimSpace(stats.LastResult.ResultText)
		}
	}
	return res, nil
}

// SetGracefulAgentCancel configures SIGTERM on context cancel (exported for tests).
func SetGracefulAgentCancel(cmd *exec.Cmd) {
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
