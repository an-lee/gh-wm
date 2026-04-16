package engine

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/an-lee/gh-wm/internal/checkpoint"
	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/ghclient"
	"github.com/an-lee/gh-wm/internal/gitbranch"
	"github.com/an-lee/gh-wm/internal/types"
)

type concludeArgs struct {
	runSucceeded  bool
	tc            *types.TaskContext
	glob          *config.GlobalConfig
	task          *config.Task
	wm            config.WMExtension
	repoRoot      string
	branchCreated bool
	prevBranch    string
	rd            *RunDir
}

// concludeRun runs phase-5 cleanup: state labels, optional branch rollback, checkpoint on success.
// It appends non-fatal errors to result.Errors.
func concludeRun(result *types.RunResult, a *concludeArgs) {
	if result == nil || a == nil {
		return
	}
	if a.tc == nil || a.task == nil {
		return
	}

	if a.runSucceeded {
		if err := postCheckpointWithErr(a.tc, result.AgentResult); err != nil {
			addRunErr(result, fmt.Errorf("checkpoint: %w", err))
		}
		if err := ApplyStateDone(a.tc, a.wm); err != nil {
			addRunErr(result, fmt.Errorf("state done: %w", err))
		}
	} else {
		if a.branchCreated && a.prevBranch != "" && a.prevBranch != "HEAD" {
			if err := gitbranch.Checkout(a.repoRoot, a.prevBranch); err != nil {
				addRunErr(result, fmt.Errorf("git checkout previous branch: %w", err))
			}
		}
		if err := ApplyStateFailed(a.tc, a.wm); err != nil {
			addRunErr(result, fmt.Errorf("state failed: %w", err))
		}
	}

	if a.rd != nil {
		_ = a.rd.UpdateMeta(types.PhaseConclusion, a.runSucceeded)
		if err := a.rd.WriteResult(result); err != nil {
			addRunErr(result, fmt.Errorf("write run result: %w", err))
		}
		appendClaudeGitHubStepSummary(result, a)
	}
}

func postCheckpointWithErr(tc *types.TaskContext, res *types.AgentResult) error {
	if res == nil {
		return nil
	}
	if os.Getenv("WM_CHECKPOINT") != "1" {
		return nil
	}
	n := issueOrPRNumber(tc)
	if n <= 0 || tc.Repo == "" {
		return nil
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
	return ghclient.PostIssueComment(tc.Repo, n, checkpoint.Encode(cp))
}
