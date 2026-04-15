package output

import (
	"context"
	"fmt"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/ghclient"
	"github.com/an-lee/gh-wm/internal/types"
)

func runLabelOutput(_ context.Context, _ *config.GlobalConfig, task *config.Task, tc *types.TaskContext, _ *types.AgentResult) error {
	n := tc.IssueNumber
	if n == 0 {
		n = tc.PRNumber
	}
	if n <= 0 || tc.Repo == "" {
		return fmt.Errorf("add-labels: no issue/PR number or repository")
	}
	so := task.SafeOutputsMap()
	raw, ok := so["add-labels"]
	if !ok {
		return nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("add-labels: expected mapping")
	}
	list, ok := m["labels"].([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	for _, x := range list {
		s, ok := x.(string)
		if !ok || s == "" {
			continue
		}
		if err := ghclient.AddIssueLabel(tc.Repo, n, s); err != nil {
			return err
		}
	}
	return nil
}
