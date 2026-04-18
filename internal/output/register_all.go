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
	registerKind(KindUpdateIssue, execKindUpdateIssue)
	registerKind(KindUpdatePullRequest, execKindUpdatePullRequest)
	registerKind(KindCloseIssue, execKindCloseIssue)
	registerKind(KindClosePullRequest, execKindClosePullRequest)
	registerKind(KindAddReviewer, execKindAddReviewer)
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

func execKindUpdateIssue(ctx context.Context, _ *config.GlobalConfig, _ *config.Task, tc *types.TaskContext, p *Policy, raw map[string]any) error {
	item := mapToUpdateIssue(raw)
	if t := strings.TrimSpace(item.Title); t != "" {
		item.Title = p.ApplyTitlePrefix(KindUpdateIssue, t)
	}
	return runUpdateIssue(ctx, tc, item)
}

func execKindUpdatePullRequest(ctx context.Context, _ *config.GlobalConfig, _ *config.Task, tc *types.TaskContext, p *Policy, raw map[string]any) error {
	item := mapToUpdatePullRequest(raw)
	if t := strings.TrimSpace(item.Title); t != "" {
		item.Title = p.ApplyTitlePrefix(KindUpdatePullRequest, t)
	}
	return runUpdatePullRequest(ctx, tc, item)
}

func execKindCloseIssue(ctx context.Context, _ *config.GlobalConfig, _ *config.Task, tc *types.TaskContext, _ *Policy, raw map[string]any) error {
	item := mapToCloseIssue(raw)
	return runCloseIssue(ctx, tc, item)
}

func execKindClosePullRequest(ctx context.Context, _ *config.GlobalConfig, _ *config.Task, tc *types.TaskContext, _ *Policy, raw map[string]any) error {
	item := mapToClosePullRequest(raw)
	return runClosePullRequest(ctx, tc, item)
}

func execKindAddReviewer(ctx context.Context, _ *config.GlobalConfig, _ *config.Task, tc *types.TaskContext, _ *Policy, raw map[string]any) error {
	item := mapToAddReviewer(raw)
	return runAddReviewers(ctx, tc, item)
}
