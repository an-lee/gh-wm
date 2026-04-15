package engine

import (
	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/ghclient"
	"github.com/an-lee/gh-wm/internal/types"
)

func issueOrPRNumber(tc *types.TaskContext) int {
	if tc == nil {
		return 0
	}
	if tc.IssueNumber > 0 {
		return tc.IssueNumber
	}
	return tc.PRNumber
}

// ApplyStateWorking adds the "working" label if wm.state_labels is configured.
func ApplyStateWorking(tc *types.TaskContext, wm config.WMExtension) {
	n := issueOrPRNumber(tc)
	if n <= 0 || tc.Repo == "" {
		return
	}
	l := wm.StateLabels["working"]
	if l == "" {
		return
	}
	_ = ghclient.AddIssueLabel(tc.Repo, n, l)
}

// ApplyStateDone removes "working" and adds "done".
func ApplyStateDone(tc *types.TaskContext, wm config.WMExtension) {
	transition(tc, wm, "working", "done")
}

// ApplyStateFailed removes "working" and adds "failed".
func ApplyStateFailed(tc *types.TaskContext, wm config.WMExtension) {
	transition(tc, wm, "working", "failed")
}

func transition(tc *types.TaskContext, wm config.WMExtension, fromKey, toKey string) {
	n := issueOrPRNumber(tc)
	if n <= 0 || tc.Repo == "" {
		return
	}
	from := wm.StateLabels[fromKey]
	to := wm.StateLabels[toKey]
	if from != "" {
		_ = ghclient.RemoveIssueLabel(tc.Repo, n, from)
	}
	if to != "" {
		_ = ghclient.AddIssueLabel(tc.Repo, n, to)
	}
}
