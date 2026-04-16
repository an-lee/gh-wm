package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/an-lee/gh-wm/internal/checkpoint"
	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/ghclient"
	"github.com/an-lee/gh-wm/internal/gitbranch"
	"github.com/an-lee/gh-wm/internal/output"
	"github.com/an-lee/gh-wm/internal/types"
)

// RunOptions configures optional behavior for RunTask (e.g. CLI streaming).
type RunOptions struct {
	// LogWriter receives a live copy of the agent subprocess combined stdout+stderr.
	// When nil, output is buffered until the process exits.
	LogWriter io.Writer
}

// RunTask executes one task by name: load task, build context, optional state labels, run agent, outputs.
func RunTask(ctx context.Context, repoRoot string, taskName string, event *types.GitHubEvent, opts *RunOptions) (*types.AgentResult, error) {
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

	var prevBranch string
	var branchCreated bool
	if task.HasSafeOutputKey("create-pull-request") {
		prev, _, created, prepErr := gitbranch.PrepareFeatureForPR(repoRoot, taskName)
		if prepErr != nil {
			ApplyStateFailed(tc, wm)
			return nil, prepErr
		}
		prevBranch = prev
		branchCreated = created
	}

	res, err := runAgent(ctx, glob, task, tc, opts)
	if err != nil {
		if branchCreated && prevBranch != "HEAD" {
			_ = gitbranch.Checkout(repoRoot, prevBranch)
		}
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
