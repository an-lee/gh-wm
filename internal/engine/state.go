package engine

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

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
func ApplyStateWorking(tc *types.TaskContext, wm config.WMExtension) error {
	n := issueOrPRNumber(tc)
	if n <= 0 || tc.Repo == "" {
		return nil
	}
	l := wm.StateLabels["working"]
	if l == "" {
		return nil
	}
	if err := ghclient.AddIssueLabel(tc.Repo, n, l); err != nil {
		slog.Info("wm: ApplyStateWorking", "err", err)
		return fmt.Errorf("add working label %q: %w", l, err)
	}
	return nil
}

// ApplyStateDone removes "working" and adds "done".
func ApplyStateDone(tc *types.TaskContext, wm config.WMExtension) error {
	return transition(tc, wm, "working", "done")
}

// ApplyStateFailed removes "working" and adds "failed".
func ApplyStateFailed(tc *types.TaskContext, wm config.WMExtension) error {
	return transition(tc, wm, "working", "failed")
}

func transition(tc *types.TaskContext, wm config.WMExtension, fromKey, toKey string) error {
	n := issueOrPRNumber(tc)
	if n <= 0 || tc.Repo == "" {
		return nil
	}
	from := wm.StateLabels[fromKey]
	to := wm.StateLabels[toKey]
	var errs []error
	if from != "" {
		if err := ghclient.RemoveIssueLabel(tc.Repo, n, from); err != nil {
			if isLabelRemoveNotFound(err) {
				slog.Info("wm: transition remove label: label not present (ignored)", "label", from)
			} else {
				slog.Info("wm: transition remove label", "label", from, "err", err)
				errs = append(errs, fmt.Errorf("remove label %q: %w", from, err))
			}
		}
	}
	if to != "" {
		if err := ghclient.AddIssueLabel(tc.Repo, n, to); err != nil {
			slog.Info("wm: transition add label", "label", to, "err", err)
			errs = append(errs, fmt.Errorf("add label %q: %w", to, err))
		}
	}
	return errors.Join(errs...)
}

// isLabelRemoveNotFound treats missing labels (404 / Not Found) as non-errors for races and manual edits.
func isLabelRemoveNotFound(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "404") || strings.Contains(s, "not found")
}
