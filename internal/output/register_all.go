package output

import (
	"context"
	"fmt"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

func init() {
	registerKind(KindCreatePullRequest, execKindCreatePullRequest)
	registerKind(KindAddComment, execKindAddComment)
	registerKind(KindAddLabels, execKindAddLabels)
	registerKind(KindRemoveLabels, execKindRemoveLabels)
	registerKind(KindCreateIssue, execKindCreateIssue)
}

func execKindCreatePullRequest(ctx context.Context, glob *config.GlobalConfig, task *config.Task, tc *types.TaskContext, p *Policy, raw map[string]any) error {
	item := mapToCreatePR(raw)
	return runCreatePullRequestItem(ctx, glob, task, tc, p, item)
}

func execKindAddComment(ctx context.Context, _ *config.GlobalConfig, _ *config.Task, tc *types.TaskContext, _ *Policy, raw map[string]any) error {
	item := mapToAddComment(raw)
	return runCommentFromItem(ctx, tc, item)
}

func execKindAddLabels(ctx context.Context, _ *config.GlobalConfig, _ *config.Task, tc *types.TaskContext, p *Policy, raw map[string]any) error {
	item := mapToLabels(raw)
	return runAddLabelsFromItem(ctx, tc, p, item)
}

func execKindRemoveLabels(ctx context.Context, _ *config.GlobalConfig, _ *config.Task, tc *types.TaskContext, p *Policy, raw map[string]any) error {
	item := mapToLabels(raw)
	if len(item.Labels) == 0 {
		return fmt.Errorf("remove_labels: empty labels")
	}
	return runRemoveLabelsFromItemWithPolicy(ctx, tc, p, item)
}

func execKindCreateIssue(ctx context.Context, _ *config.GlobalConfig, _ *config.Task, tc *types.TaskContext, p *Policy, raw map[string]any) error {
	item := mapToCreateIssue(raw)
	item.Title = p.ApplyTitlePrefix(KindCreateIssue, strings.TrimSpace(item.Title))
	if strings.TrimSpace(item.Title) == "" {
		return fmt.Errorf("create_issue: empty title")
	}
	item.Labels = p.MergeLabels(KindCreateIssue, item.Labels)
	return runCreateIssue(ctx, tc, item)
}
