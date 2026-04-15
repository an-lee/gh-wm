package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gh-wm/gh-wm/internal/config"
	"github.com/gh-wm/gh-wm/internal/types"
)

// runAgent invokes the configured AI CLI with task body + optional context files.
func runAgent(ctx context.Context, glob *config.GlobalConfig, task *config.Task, tc *types.TaskContext) (*types.AgentResult, error) {
	prompt := strings.TrimSpace(task.Body)
	if prompt == "" {
		prompt = task.Name + ": process repository event."
	}
	// Append context files
	for _, f := range glob.Context.Files {
		p := filepath.Join(tc.RepoPath, f)
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		prompt += "\n\n---\n## " + f + "\n\n" + string(b)
	}
	engine := task.Engine()
	if engine == "" {
		engine = glob.Engine
	}
	_ = engine // future: switch copilot/codex

	cmdLine := os.Getenv("WM_AGENT_CMD")
	var cmd *exec.Cmd
	if cmdLine != "" {
		args := strings.Fields(cmdLine)
		cmd = exec.CommandContext(ctx, args[0], append(args[1:], prompt)...)
	} else {
		// Default: claude -p (Claude Code CLI)
		cmd = exec.CommandContext(ctx, "claude", "-p", prompt)
	}
	cmd.Dir = tc.RepoPath
	cmd.Env = append(os.Environ(),
		"GITHUB_REPOSITORY="+tc.Repo,
		fmt.Sprintf("WM_TASK=%s", tc.TaskName),
	)
	out, err := cmd.CombinedOutput()
	res := &types.AgentResult{
		Stdout:   string(out),
		Summary:  string(out),
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
