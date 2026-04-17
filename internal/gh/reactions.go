package gh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func postReaction(ctx context.Context, path, content string) error {
	c, err := REST()
	if err != nil {
		return err
	}
	b, err := json.Marshal(map[string]string{"content": content})
	if err != nil {
		return err
	}
	resp, err := c.RequestWithContext(ctx, "POST", path, bytes.NewReader(b))
	if err != nil {
		if reactionAlreadyExists(err) {
			return nil
		}
		return err
	}
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return nil
}

func reactionAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "already_exists") ||
		strings.Contains(s, `"code":"already_exists"`) ||
		strings.Contains(s, "Resource already exists")
}

// AddIssueReaction adds a reaction to an issue or PR.
func AddIssueReaction(ctx context.Context, repo string, issue int, content string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("repos/%s/%s/issues/%d/reactions", owner, name, issue)
	return postReaction(ctx, path, content)
}

// AddIssueCommentReaction adds a reaction to an issue comment.
func AddIssueCommentReaction(ctx context.Context, repo string, commentID int64, content string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("repos/%s/%s/issues/comments/%d/reactions", owner, name, commentID)
	return postReaction(ctx, path, content)
}
