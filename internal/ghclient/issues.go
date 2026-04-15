package ghclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

func splitRepo(repo string) (owner, name string, err error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo: %s", repo)
	}
	return parts[0], parts[1], nil
}

// RemoveIssueLabel removes a label from an issue or PR (same number).
func RemoveIssueLabel(repo string, issue int, label string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	enc := url.PathEscape(label)
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/labels/%s", owner, name, issue, enc)
	cmd := exec.Command("gh", "api", "-X", "DELETE", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh api DELETE label: %w: %s", err, stderr.String())
	}
	return nil
}

// ListIssueCommentBodies returns comment bodies in order (oldest first).
func ListIssueCommentBodies(repo string, issue int) ([]string, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, name, issue)
	cmd := exec.Command("gh", "api", path)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh api comments: %w", err)
	}
	var comments []struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(out, &comments); err != nil {
		return nil, fmt.Errorf("parse comments: %w", err)
	}
	var bodies []string
	for _, c := range comments {
		bodies = append(bodies, c.Body)
	}
	return bodies, nil
}

// PostIssueComment adds a comment to an issue or PR.
func PostIssueComment(repo string, issue int, body string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, name, issue)
	payload, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		return err
	}
	cmd := exec.Command("gh", "api", "-X", "POST", path, "--input", "-")
	cmd.Stdin = bytes.NewReader(payload)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh api comment: %w: %s", err, stderr.String())
	}
	return nil
}
