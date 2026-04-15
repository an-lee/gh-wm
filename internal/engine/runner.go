package engine

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gh-wm/gh-wm/internal/checkpoint"
	"github.com/gh-wm/gh-wm/internal/config"
	"github.com/gh-wm/gh-wm/internal/ghclient"
	"github.com/gh-wm/gh-wm/internal/output"
	"github.com/gh-wm/gh-wm/internal/types"
)

// RunTask executes one task by name: load task, build context, optional state labels, run agent, outputs.
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

	wm := task.WM()
	loadCheckpointHint(tc)

	ApplyStateWorking(tc, wm)

	res, err := runAgent(ctx, glob, task, tc)
	if err != nil {
		ApplyStateFailed(tc, wm)
		return res, err
	}
	if outErr := output.RunSuccessOutputs(ctx, glob, task, tc, res); outErr != nil {
		ApplyStateFailed(tc, wm)
		return res, outErr
	}
	postCheckpoint(tc, res)
	ApplyStateDone(tc, wm)
	return res, nil
}

func loadCheckpointHint(tc *types.TaskContext) {
	if os.Getenv("WM_CHECKPOINT") != "1" {
		return
	}
	n := issueOrPRNumber(tc)
	if n <= 0 || tc.Repo == "" {
		return
	}
	bodies, err := ghclient.ListIssueCommentBodies(tc.Repo, n)
	if err != nil {
		return
	}
	c, err := checkpoint.ParseLatest(bodies)
	if err != nil || c == nil || strings.TrimSpace(c.Summary) == "" {
		return
	}
	tc.CheckpointHint = strings.TrimSpace(c.Summary)
}

func postCheckpoint(tc *types.TaskContext, res *types.AgentResult) {
	if os.Getenv("WM_CHECKPOINT") != "1" {
		return
	}
	n := issueOrPRNumber(tc)
	if n <= 0 || tc.Repo == "" {
		return
	}
	summary := strings.TrimSpace(res.Summary)
	if summary == "" {
		summary = strings.TrimSpace(res.Stdout)
	}
	if len(summary) > 2000 {
		summary = summary[:2000] + "…"
	}
	cp := checkpoint.Checkpoint{
		Summary:   summary,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	_ = ghclient.PostIssueComment(tc.Repo, n, checkpoint.Encode(cp))
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
		if tc.IssueNumber == 0 && tc.PRNumber > 0 {
			tc.IssueNumber = tc.PRNumber
		}
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
