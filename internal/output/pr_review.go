package output

import (
	"context"
	"fmt"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/config/scalar"
	"github.com/an-lee/gh-wm/internal/ghclient"
	"github.com/an-lee/gh-wm/internal/types"
)

func resolvePRTarget(tc *types.TaskContext, target int) int {
	if target > 0 {
		return target
	}
	if tc.PRNumber > 0 {
		return tc.PRNumber
	}
	return tc.IssueNumber
}

func defaultReviewCommentSide(p *Policy, kind OutputKind) string {
	block := p.fmBlock(kind)
	s := strings.TrimSpace(strings.ToUpper(scalar.StringFromMap(block, "side")))
	if s == "LEFT" || s == "RIGHT" {
		return s
	}
	return "RIGHT"
}

func execKindCreatePullRequestReviewComment(ctx context.Context, _ *config.GlobalConfig, task *config.Task, tc *types.TaskContext, p *Policy, raw map[string]any) error {
	_ = ctx
	item := mapToCreatePullRequestReviewComment(raw)
	pr := resolvePRTarget(tc, item.Target)
	if pr <= 0 {
		return fmt.Errorf("create_pull_request_review_comment: no PR number (set target or use PR context)")
	}
	if strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("create_pull_request_review_comment: GITHUB_REPOSITORY not set")
	}
	path := strings.TrimSpace(item.Path)
	if path == "" {
		return fmt.Errorf("create_pull_request_review_comment: empty path")
	}
	if item.Line <= 0 {
		return fmt.Errorf("create_pull_request_review_comment: line must be positive")
	}
	body := strings.TrimSpace(item.Body)
	if body == "" {
		return fmt.Errorf("create_pull_request_review_comment: empty body")
	}
	body = AppendMessagesFooter(task, tc, body)
	side := strings.TrimSpace(strings.ToUpper(item.Side))
	if side == "" {
		side = defaultReviewCommentSide(p, KindCreatePullRequestReviewComment)
	}
	if side != "LEFT" && side != "RIGHT" {
		side = "RIGHT"
	}
	commitID := strings.TrimSpace(item.CommitID)
	if commitID == "" {
		sha, err := ghclient.PullRequestHeadSHA(tc.Repo, pr)
		if err != nil {
			return fmt.Errorf("create_pull_request_review_comment: resolve head sha: %w", err)
		}
		commitID = sha
	}
	return ghclient.CreatePullRequestReviewComment(tc.Repo, pr, body, path, commitID, item.Line, side)
}

func execKindSubmitPullRequestReview(ctx context.Context, _ *config.GlobalConfig, task *config.Task, tc *types.TaskContext, p *Policy, raw map[string]any) error {
	_ = ctx
	_ = p
	item := mapToSubmitPullRequestReview(raw)
	pr := resolvePRTarget(tc, item.Target)
	if pr <= 0 {
		return fmt.Errorf("submit_pull_request_review: no PR number (set target or use PR context)")
	}
	if strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("submit_pull_request_review: GITHUB_REPOSITORY not set")
	}
	ev := strings.TrimSpace(strings.ToUpper(item.Event))
	switch ev {
	case "APPROVE", "REQUEST_CHANGES", "COMMENT":
	default:
		return fmt.Errorf("submit_pull_request_review: event must be APPROVE, REQUEST_CHANGES, or COMMENT")
	}
	body := strings.TrimSpace(item.Body)
	body = AppendMessagesFooter(task, tc, body)
	commitID := strings.TrimSpace(item.CommitID)
	if commitID == "" {
		sha, err := ghclient.PullRequestHeadSHA(tc.Repo, pr)
		if err != nil {
			return fmt.Errorf("submit_pull_request_review: resolve head sha: %w", err)
		}
		commitID = sha
	}
	return ghclient.SubmitPullRequestReview(tc.Repo, pr, commitID, ev, body)
}
