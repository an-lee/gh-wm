package output

import (
	"context"
	"fmt"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/ghclient"
	"github.com/an-lee/gh-wm/internal/types"
)

func normalizeReviewCommentSide(s string) (string, error) {
	x := strings.TrimSpace(strings.ToUpper(s))
	switch x {
	case "LEFT", "RIGHT":
		return x, nil
	default:
		return "", fmt.Errorf("side must be LEFT or RIGHT")
	}
}

func runCreatePullRequestReviewComment(ctx context.Context, task *config.Task, tc *types.TaskContext, item ItemCreatePullRequestReviewComment) error {
	n := resolvePRTarget(tc, item.Target)
	if n <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("create_pull_request_review_comment: no PR number or repository")
	}
	side, err := normalizeReviewCommentSide(item.Side)
	if err != nil {
		return fmt.Errorf("create_pull_request_review_comment: %w", err)
	}
	body := strings.TrimSpace(item.Body)
	if body == "" {
		return fmt.Errorf("create_pull_request_review_comment: empty body")
	}
	body = AppendMessagesFooter(task, tc, body)
	commit := strings.TrimSpace(item.CommitID)
	path := strings.TrimSpace(item.Path)
	if commit == "" || path == "" {
		return fmt.Errorf("create_pull_request_review_comment: commit_id and path required")
	}
	if item.Line <= 0 {
		return fmt.Errorf("create_pull_request_review_comment: invalid line")
	}
	start := item.StartLine
	if start > 0 && start > item.Line {
		return fmt.Errorf("create_pull_request_review_comment: start_line must be <= line")
	}
	return ghclient.CreatePullRequestReviewComment(ctx, tc.Repo, n, body, commit, path, item.Line, side, start)
}

func runReplyToPullRequestReviewComment(ctx context.Context, tc *types.TaskContext, item ItemReplyToPullRequestReviewComment) error {
	if resolvePRTarget(tc, item.Target) <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("reply_to_pull_request_review_comment: no PR number or repository")
	}
	body := strings.TrimSpace(item.Body)
	if body == "" {
		return fmt.Errorf("reply_to_pull_request_review_comment: empty body")
	}
	if item.CommentID <= 0 {
		return fmt.Errorf("reply_to_pull_request_review_comment: invalid comment_id")
	}
	return ghclient.ReplyToPullRequestReviewComment(ctx, tc.Repo, item.CommentID, body)
}

func runResolvePullRequestReviewThread(ctx context.Context, tc *types.TaskContext, item ItemResolvePullRequestReviewThread) error {
	if resolvePRTarget(tc, item.Target) <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("resolve_pull_request_review_thread: no PR number or repository")
	}
	tid := strings.TrimSpace(item.ThreadID)
	if tid == "" {
		return fmt.Errorf("resolve_pull_request_review_thread: empty thread_id")
	}
	return ghclient.ResolveReviewThread(ctx, tid)
}
