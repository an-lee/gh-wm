package output

import (
	"context"
	"fmt"
	"strings"

	"github.com/an-lee/gh-wm/internal/ghclient"
	"github.com/an-lee/gh-wm/internal/types"
)

func runUpdateIssue(ctx context.Context, tc *types.TaskContext, item ItemUpdateIssue) error {
	n := resolveIssueTarget(tc, item.Target)
	if n <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("update_issue: no issue number or repository")
	}
	title := strings.TrimSpace(item.Title)
	body := strings.TrimSpace(item.Body)
	if title == "" && body == "" {
		return fmt.Errorf("update_issue: empty title and body")
	}
	return ghclient.UpdateIssue(ctx, tc.Repo, n, title, body)
}

func runUpdatePullRequest(ctx context.Context, tc *types.TaskContext, item ItemUpdatePullRequest) error {
	n := resolvePRTarget(tc, item.Target)
	if n <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("update_pull_request: no pull request number or repository")
	}
	title := strings.TrimSpace(item.Title)
	body := strings.TrimSpace(item.Body)
	if title == "" && body == "" {
		return fmt.Errorf("update_pull_request: empty title and body")
	}
	return ghclient.UpdatePullRequest(ctx, tc.Repo, n, title, body)
}

func runCloseIssue(ctx context.Context, tc *types.TaskContext, item ItemCloseIssue) error {
	n := resolveIssueTarget(tc, item.Target)
	if n <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("close_issue: no issue number or repository")
	}
	return ghclient.CloseIssue(ctx, tc.Repo, n, item.Comment, item.StateReason)
}

func runClosePullRequest(ctx context.Context, tc *types.TaskContext, item ItemClosePullRequest) error {
	n := resolvePRTarget(tc, item.Target)
	if n <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("close_pull_request: no pull request number or repository")
	}
	return ghclient.ClosePullRequest(ctx, tc.Repo, n, item.Comment)
}

func runAddReviewers(ctx context.Context, tc *types.TaskContext, item ItemAddReviewer) error {
	n := resolvePRTarget(tc, item.Target)
	if n <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("add_reviewer: no pull request number or repository")
	}
	if len(item.Reviewers) == 0 {
		return fmt.Errorf("add_reviewer: empty reviewers")
	}
	return ghclient.RequestPullReviewers(ctx, tc.Repo, n, item.Reviewers)
}
