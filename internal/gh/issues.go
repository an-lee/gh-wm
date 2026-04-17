package gh

import (
	"context"
	"fmt"
	"net/url"
)

// AddIssueLabel adds a label to an issue or PR (same issue API).
func AddIssueLabel(ctx context.Context, repo string, issue int, label string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("repos/%s/%s/issues/%d/labels", owner, name, issue)
	return PostJSON(ctx, path, map[string][]string{"labels": {label}})
}

// RemoveIssueLabel removes a label from an issue or PR.
func RemoveIssueLabel(ctx context.Context, repo string, issue int, label string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	enc := url.PathEscape(label)
	path := fmt.Sprintf("repos/%s/%s/issues/%d/labels/%s", owner, name, issue, enc)
	return DeletePath(ctx, path)
}

// IssueComment is a minimal issue comment JSON for listing bodies.
type IssueComment struct {
	Body string `json:"body"`
}

// ListIssueCommentBodies returns comment bodies in order (oldest first).
func ListIssueCommentBodies(ctx context.Context, repo string, issue int) ([]string, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("repos/%s/%s/issues/%d/comments", owner, name, issue)
	var comments []IssueComment
	if err := GetJSON(ctx, path, &comments); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(comments))
	for _, c := range comments {
		out = append(out, c.Body)
	}
	return out, nil
}

// PostIssueComment adds a comment to an issue or PR.
func PostIssueComment(ctx context.Context, repo string, issue int, body string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("repos/%s/%s/issues/%d/comments", owner, name, issue)
	return PostJSON(ctx, path, map[string]string{"body": body})
}

// CreateIssue opens a new issue in the repository.
func CreateIssue(ctx context.Context, repo, title, body string, labels, assignees []string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("repos/%s/%s/issues", owner, name)
	payload := map[string]any{
		"title": title,
		"body":  body,
	}
	if len(labels) > 0 {
		payload["labels"] = labels
	}
	if len(assignees) > 0 {
		payload["assignees"] = assignees
	}
	return PostJSON(ctx, path, payload)
}
