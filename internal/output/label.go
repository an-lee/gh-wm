package output

import (
	"context"
	"fmt"

	"github.com/an-lee/gh-wm/internal/ghclient"
	"github.com/an-lee/gh-wm/internal/types"
)

func resolveLabelTarget(tc *types.TaskContext, target int) int {
	if target > 0 {
		return target
	}
	n := tc.IssueNumber
	if n == 0 {
		n = tc.PRNumber
	}
	return n
}

// runAddLabelsFromItem applies labels from structured output with policy checks.
func runAddLabelsFromItem(_ context.Context, tc *types.TaskContext, p *Policy, item ItemLabels) error {
	if len(item.Labels) == 0 {
		return fmt.Errorf("add_labels: empty labels")
	}
	n := resolveLabelTarget(tc, item.Target)
	if n <= 0 || tc.Repo == "" {
		return fmt.Errorf("add_labels: no issue/PR number or repository")
	}
	for _, label := range item.Labels {
		if label == "" {
			continue
		}
		if !p.LabelAllowed(KindAddLabels, label) {
			return fmt.Errorf("add_labels: label %q not allowed by policy", label)
		}
		if err := ghclient.AddIssueLabel(tc.Repo, n, label); err != nil {
			return err
		}
	}
	return nil
}

// runRemoveLabelsFromItemWithPolicy validates allowed/blocked before removal.
func runRemoveLabelsFromItemWithPolicy(_ context.Context, tc *types.TaskContext, p *Policy, item ItemLabels) error {
	n := resolveLabelTarget(tc, item.Target)
	if n <= 0 || tc.Repo == "" {
		return fmt.Errorf("remove_labels: no issue/PR number or repository")
	}
	for _, label := range item.Labels {
		if label == "" {
			continue
		}
		if !p.RemoveLabelAllowed(label) {
			return fmt.Errorf("remove_labels: label %q not allowed by policy", label)
		}
		if err := ghclient.RemoveIssueLabel(tc.Repo, n, label); err != nil {
			return err
		}
	}
	return nil
}
