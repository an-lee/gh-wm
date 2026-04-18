package ghclient

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/an-lee/gh-wm/internal/gh"
)

// CreatePullRequestReviewComment adds an inline review comment on a PR diff.
func CreatePullRequestReviewComment(ctx context.Context, repo string, pullNumber int, body, commitID, path string, line int, side string, startLine int) error {
	if useREST() {
		return gh.CreatePullRequestReviewComment(ctx, repo, pullNumber, body, commitID, path, line, side, startLine)
	}
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo: %s", repo)
	}
	owner, name := parts[0], parts[1]
	apiPath := fmt.Sprintf("/repos/%s/%s/pulls/%d/comments", owner, name, pullNumber)
	args := []string{"api", "-X", "POST", apiPath,
		"-f", "body=" + body,
		"-f", "commit_id=" + strings.TrimSpace(commitID),
		"-f", "path=" + path,
		"-f", "line=" + strconv.Itoa(line),
		"-f", "side=" + side,
	}
	if startLine > 0 {
		args = append(args, "-f", "start_line="+strconv.Itoa(startLine))
	}
	cmd := exec.CommandContext(ctx, "gh", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh api: %w: %s", err, stderr.String())
	}
	return nil
}

// ReplyToPullRequestReviewComment replies to a pull request review comment.
func ReplyToPullRequestReviewComment(ctx context.Context, repo string, commentID int, body string) error {
	if useREST() {
		return gh.ReplyToPullRequestReviewComment(ctx, repo, commentID, body)
	}
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo: %s", repo)
	}
	owner, name := parts[0], parts[1]
	apiPath := fmt.Sprintf("/repos/%s/%s/pulls/comments/%d/replies", owner, name, commentID)
	cmd := exec.CommandContext(ctx, "gh", "api", "-X", "POST", apiPath, "-f", "body="+body)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh api: %w: %s", err, stderr.String())
	}
	return nil
}

// ResolveReviewThread resolves a review thread by GraphQL node id.
func ResolveReviewThread(ctx context.Context, threadID string) error {
	if useREST() {
		return gh.ResolveReviewThread(ctx, threadID)
	}
	q := `mutation($id:ID!){resolveReviewThread(input:{threadId:$id}){thread{isResolved}}}`
	cmd := exec.CommandContext(ctx, "gh", "api", "graphql",
		"-f", "query="+q,
		"-f", "id="+threadID,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh api graphql: %w: %s", err, stderr.String())
	}
	return nil
}
