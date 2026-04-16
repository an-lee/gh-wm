package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

// runAgent invokes the configured AI CLI with task body + optional context files.
func runAgent(ctx context.Context, glob *config.GlobalConfig, task *config.Task, tc *types.TaskContext, opts *RunOptions) (*types.AgentResult, error) {
	prompt := strings.TrimSpace(task.Body)
	if prompt == "" {
		prompt = task.Name + ": process repository event."
	}
	for _, f := range glob.Context.Files {
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
	if engineName == "" {
		engineName = glob.Engine
	}

	cmdLine := os.Getenv("WM_AGENT_CMD")
	var cmd *exec.Cmd
	if cmdLine != "" {
		args := strings.Fields(cmdLine)
		cmd = exec.CommandContext(ctx, args[0], append(args[1:], prompt)...)
	} else {
		switch strings.ToLower(strings.TrimSpace(engineName)) {
		case "codex":
			if alt := strings.TrimSpace(os.Getenv("WM_ENGINE_CODEX_CMD")); alt != "" {
				a := strings.Fields(alt)
				cmd = exec.CommandContext(ctx, a[0], append(a[1:], prompt)...)
			} else {
				cmd = exec.CommandContext(ctx, "codex", "-p", prompt)
			}
		case "copilot":
			err := fmt.Errorf("engine copilot: set WM_AGENT_CMD to invoke your Copilot-compatible CLI")
			return &types.AgentResult{Success: false, ExitCode: -1, Stderr: err.Error()}, err
		default:
			cmd = exec.CommandContext(ctx, "claude", "-p", prompt)
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

	var buf bytes.Buffer
	var err error
	if opts != nil && opts.LogWriter != nil {
		w := io.MultiWriter(opts.LogWriter, &buf)
		cmd.Stdout = w
		cmd.Stderr = w
		err = cmd.Run()
	} else {
		var out []byte
		out, err = cmd.CombinedOutput()
		buf.Write(out)
	}
	combined := buf.String()

	res := &types.AgentResult{
		Stdout:   combined,
		Summary:  combined,
		Success:  err == nil,
		ExitCode: 0,
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.ExitCode = -1
		}
		res.Stderr = err.Error()
		return res, fmt.Errorf("agent: %w", err)
	}
	return res, nil
}
