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

func addRunErr(r *types.RunResult, err error) {
	if r == nil || err == nil {
		return
	}
	r.Errors = append(r.Errors, err)
}

// RunTask executes one task: activation (validation, state, branch prep), agent, validation, safe-outputs, and deferred conclusion (labels, checkpoint, branch rollback).
func RunTask(ctx context.Context, repoRoot string, taskName string, event *types.GitHubEvent, opts *RunOptions) (*types.RunResult, error) {
	start := time.Now()
	result := &types.RunResult{Phase: types.PhaseActivation}

	var tc *types.TaskContext
	var glob *config.GlobalConfig
	var task *config.Task
	var wm config.WMExtension
	var branchCreated bool
	var prevBranch string
	runSucceeded := false

	defer func() {
		result.Duration = time.Since(start)
		concludeRun(result, &concludeArgs{
			runSucceeded:  runSucceeded,
			tc:            tc,
			task:          task,
			wm:            wm,
			repoRoot:      repoRoot,
			branchCreated: branchCreated,
			prevBranch:    prevBranch,
		})
	}()

	glob, tasks, err := config.Load(repoRoot)
	if err != nil {
		addRunErr(result, err)
		return result, err
	}
	glob = config.DefaultGlobal(glob)

	for _, t := range tasks {
		if t.Name == taskName {
			task = t
			break
		}
	}
	if task == nil {
		err := fmt.Errorf("task not found: %s", taskName)
		addRunErr(result, err)
		return result, err
	}

	if err := validateEventContext(event); err != nil {
		addRunErr(result, err)
		return result, err
	}
	if err := validateTaskConfig(task, glob); err != nil {
		addRunErr(result, err)
		return result, err
	}

	tc = &types.TaskContext{
		TaskName: taskName,
		Repo:     os.Getenv("GITHUB_REPOSITORY"),
		RepoPath: repoRoot,
		Event:    event,
	}
	if event != nil {
		extractNumbers(event.Payload, tc)
	}

	wm = task.WM()
	loadCheckpointHint(tc)

	if err := ApplyStateWorking(tc, wm); err != nil {
		addRunErr(result, err)
	}

	if task.HasSafeOutputKey("create-pull-request") {
		prev, _, created, prepErr := gitbranch.PrepareFeatureForPR(repoRoot, taskName)
		if prepErr != nil {
			addRunErr(result, prepErr)
			return result, prepErr
		}
		prevBranch = prev
		branchCreated = created
	}

	result.Phase = types.PhaseAgent
	res, agentErr := runAgent(ctx, glob, task, tc, opts)
	result.AgentResult = res
	if agentErr != nil {
		addRunErr(result, agentErr)
		return result, agentErr
	}

	result.Phase = types.PhaseValidation
	if err := validateAgentOutputErr(res); err != nil {
		addRunErr(result, err)
		return result, err
	}

	result.Phase = types.PhaseOutputs
	if outErr := output.RunSuccessOutputs(ctx, glob, task, tc, res); outErr != nil {
		addRunErr(result, outErr)
		return result, outErr
	}

	runSucceeded = true
	result.Success = true
	return result, nil
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

// postCheckpoint is kept for tests; errors are ignored (see postCheckpointWithErr for surfaced errors).
func postCheckpoint(tc *types.TaskContext, res *types.AgentResult) {
	_ = postCheckpointWithErr(tc, res)
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
