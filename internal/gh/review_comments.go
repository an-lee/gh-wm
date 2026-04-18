package gh

import (
	"context"
	"fmt"
	"strings"
)

// CreatePullRequestReviewComment adds an inline review comment on a pull request diff.
func CreatePullRequestReviewComment(ctx context.Context, repo string, pullNumber int, body, commitID, path string, line int, side string, startLine int) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	pathAPI := fmt.Sprintf("repos/%s/%s/pulls/%d/comments", owner, name, pullNumber)
	payload := map[string]any{
		"body":      body,
		"commit_id": strings.TrimSpace(commitID),
		"path":      path,
		"line":      line,
		"side":      side,
	}
	if startLine > 0 {
		payload["start_line"] = startLine
	}
	return PostJSONChecked(ctx, pathAPI, payload)
}

// ReplyToPullRequestReviewComment replies to an existing pull request review comment.
func ReplyToPullRequestReviewComment(ctx context.Context, repo string, commentID int, body string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	pathAPI := fmt.Sprintf("repos/%s/%s/pulls/comments/%d/replies", owner, name, commentID)
	return PostJSONChecked(ctx, pathAPI, map[string]any{"body": body})
}
