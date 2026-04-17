package output

import (
	"context"
	"fmt"
	"strings"

	"github.com/an-lee/gh-wm/internal/ghclient"
	"github.com/an-lee/gh-wm/internal/types"
)

func runCreateIssue(ctx context.Context, tc *types.TaskContext, item ItemCreateIssue) error {
	t := strings.TrimSpace(item.Title)
	body := strings.TrimSpace(item.Body)
	if t == "" {
		return fmt.Errorf("create_issue: empty title")
	}
	if tc.Repo == "" {
		return fmt.Errorf("create_issue: GITHUB_REPOSITORY not set")
	}
	var labels []string
	for _, l := range item.Labels {
		if l != "" {
			labels = append(labels, l)
		}
	}
	var assignees []string
	for _, a := range item.Assignees {
		if a != "" {
			assignees = append(assignees, a)
		}
	}
	return ghclient.CreateIssue(ctx, tc.Repo, t, body, labels, assignees)
}
