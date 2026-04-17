package ghclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// AddIssueReaction adds an emoji reaction to an issue or pull request (issue number).
func AddIssueReaction(repo string, issue int, content string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/reactions", owner, name, issue)
	return postReaction(path, content)
}

// AddIssueCommentReaction adds an emoji reaction to an issue comment.
func AddIssueCommentReaction(repo string, commentID int64, content string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/repos/%s/%s/issues/comments/%d/reactions", owner, name, commentID)
	return postReaction(path, content)
}

func postReaction(apiPath, content string) error {
	payload, err := json.Marshal(map[string]string{"content": content})
	if err != nil {
		return err
	}
	cmd := exec.Command("gh", "api", "-X", "POST", apiPath, "--input", "-")
	cmd.Stdin = bytes.NewReader(payload)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	combined := string(out)
	if reactionAlreadyExists(combined) {
		return nil
	}
	return fmt.Errorf("gh api reaction: %w: %s", err, strings.TrimSpace(combined))
}

func reactionAlreadyExists(msg string) bool {
	return strings.Contains(msg, "already_exists") || strings.Contains(msg, `"code":"already_exists"`)
}
