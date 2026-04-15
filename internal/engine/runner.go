package engine

import (
	"context"
	"fmt"
	"os"

	"github.com/gh-wm/gh-wm/internal/config"
	"github.com/gh-wm/gh-wm/internal/types"
)

// RunTask executes one task by name: load task, build context, run agent, outputs.
func RunTask(ctx context.Context, repoRoot string, taskName string, event *types.GitHubEvent) (*types.AgentResult, error) {
	glob, tasks, err := config.Load(repoRoot)
	if err != nil {
		return nil, err
	}
	glob = config.DefaultGlobal(glob)
	var task *config.Task
	for _, t := range tasks {
		if t.Name == taskName {
			task = t
			break
		}
	}
	if task == nil {
		return nil, fmt.Errorf("task not found: %s", taskName)
	}

	tc := &types.TaskContext{
		TaskName: taskName,
		Repo:     os.Getenv("GITHUB_REPOSITORY"),
		RepoPath: repoRoot,
		Event:    event,
	}
	extractNumbers(event.Payload, tc)

	return runAgent(ctx, glob, task, tc)
}

func extractNumbers(payload map[string]any, tc *types.TaskContext) {
	if payload == nil {
		return
	}
	if iss, ok := payload["issue"].(map[string]any); ok {
		tc.IssueNumber = num(iss["number"])
	}
	if pr, ok := payload["pull_request"].(map[string]any); ok {
		tc.PRNumber = num(pr["number"])
	}
}

func num(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	default:
		return 0
	}
}
