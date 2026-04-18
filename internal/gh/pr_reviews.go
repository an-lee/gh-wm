package gh

import (
	"context"
	"fmt"
	"strings"
)

// PullRequestHeadSHA returns the head commit SHA for a pull request.
func PullRequestHeadSHA(ctx context.Context, repo string, pr int) (string, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("repos/%s/%s/pulls/%d", owner, name, pr)
	var out struct {
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := GetJSON(ctx, path, &out); err != nil {
		return "", err
	}
	s := strings.TrimSpace(out.Head.SHA)
	if s == "" {
		return "", fmt.Errorf("empty head sha for PR %d", pr)
	}
	return s, nil
}

// CreatePullRequestReviewComment creates an inline review comment on a PR diff.
func CreatePullRequestReviewComment(ctx context.Context, repo string, pr int, body, path, commitID string, line int, side string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	apiPath := fmt.Sprintf("repos/%s/%s/pulls/%d/comments", owner, name, pr)
	payload := map[string]any{
		"body":      body,
		"commit_id": commitID,
		"path":      path,
		"line":      line,
		"side":      side,
	}
	return PostJSON(ctx, apiPath, payload)
}

// SubmitPullRequestReview submits a pull request review (APPROVE, REQUEST_CHANGES, or COMMENT).
func SubmitPullRequestReview(ctx context.Context, repo string, pr int, commitID, event, body string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	apiPath := fmt.Sprintf("repos/%s/%s/pulls/%d/reviews", owner, name, pr)
	payload := map[string]any{
		"event": event,
	}
	if strings.TrimSpace(commitID) != "" {
		payload["commit_id"] = commitID
	}
	if strings.TrimSpace(body) != "" {
		payload["body"] = body
	}
	return PostJSON(ctx, apiPath, payload)
}
