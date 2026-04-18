package gh

import (
	"context"
	"fmt"
	"net/url"
	"strings"
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

// issueSnapshot is the subset of GET /repos/{owner}/{repo}/issues/{number} we need for edits.
type issueSnapshot struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// GetIssueSnapshot returns the current title and body for an issue or PR (same issue number).
func GetIssueSnapshot(ctx context.Context, repo string, number int) (title, body string, err error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return "", "", err
	}
	path := fmt.Sprintf("repos/%s/%s/issues/%d", owner, name, number)
	var v issueSnapshot
	if err := GetJSON(ctx, path, &v); err != nil {
		return "", "", err
	}
	return v.Title, v.Body, nil
}

// UpdateIssue PATCHes title and/or body. Empty strings are omitted from the payload.
func UpdateIssue(ctx context.Context, repo string, number int, title, body string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("repos/%s/%s/issues/%d", owner, name, number)
	payload := map[string]any{}
	if t := strings.TrimSpace(title); t != "" {
		payload["title"] = t
	}
	if strings.TrimSpace(body) != "" {
		payload["body"] = body
	}
	if len(payload) == 0 {
		return fmt.Errorf("update issue: nothing to update")
	}
	return PatchJSON(ctx, path, payload)
}

// CloseIssue closes an issue; optional comment is posted first. stateReason is REST form (completed, not_planned, duplicate) or empty for completed.
func CloseIssue(ctx context.Context, repo string, number int, comment, stateReason string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	if strings.TrimSpace(comment) != "" {
		if err := PostIssueComment(ctx, repo, number, comment); err != nil {
			return err
		}
	}
	path := fmt.Sprintf("repos/%s/%s/issues/%d", owner, name, number)
	reason := normalizeIssueStateReasonREST(stateReason)
	payload := map[string]any{
		"state":        "closed",
		"state_reason": reason,
	}
	return PatchJSON(ctx, path, payload)
}

func normalizeIssueStateReasonREST(s string) string {
	s = strings.TrimSpace(strings.ToLower(strings.ReplaceAll(s, "-", "_")))
	s = strings.ReplaceAll(s, " ", "_")
	switch s {
	case "", "completed":
		return "completed"
	case "not_planned":
		return "not_planned"
	case "duplicate":
		return "duplicate"
	default:
		return "completed"
	}
}
