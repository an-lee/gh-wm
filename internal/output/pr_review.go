package output

import (
	"context"
	"fmt"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/ghclient"
	"github.com/an-lee/gh-wm/internal/types"
)

func execKindSubmitPullRequestReview(ctx context.Context, _ *config.GlobalConfig, task *config.Task, tc *types.TaskContext, _ *Policy, raw map[string]any) error {
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
