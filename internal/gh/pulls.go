package gh

import (
	"context"
	"fmt"
	"strings"
)

// UpdatePullRequest PATCHes title and/or body on a pull request.
func UpdatePullRequest(ctx context.Context, repo string, number int, title, body string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("repos/%s/%s/pulls/%d", owner, name, number)
	payload := map[string]any{}
	if t := strings.TrimSpace(title); t != "" {
		payload["title"] = t
	}
	if strings.TrimSpace(body) != "" {
		payload["body"] = body
	}
	if len(payload) == 0 {
		return fmt.Errorf("update pull request: nothing to update")
	}
	return PatchJSON(ctx, path, payload)
}

// ClosePullRequest closes a pull request without merging; optional comment is posted first (issue comments API).
func ClosePullRequest(ctx context.Context, repo string, number int, comment string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	if strings.TrimSpace(comment) != "" {
		if err := PostIssueComment(ctx, repo, number, comment); err != nil {
			return err
		}
	}
	path := fmt.Sprintf("repos/%s/%s/pulls/%d", owner, name, number)
	return PatchJSON(ctx, path, map[string]any{"state": "closed"})
}

// RequestPullReviewers requests reviewers on a pull request.
func RequestPullReviewers(ctx context.Context, repo string, number int, reviewers []string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	var logins []string
	for _, r := range reviewers {
		r = strings.TrimSpace(r)
		if r != "" {
			logins = append(logins, r)
		}
	}
	if len(logins) == 0 {
		return fmt.Errorf("request reviewers: empty reviewers")
	}
	path := fmt.Sprintf("repos/%s/%s/pulls/%d/requested_reviewers", owner, name, number)
	return PostJSONChecked(ctx, path, map[string][]string{"reviewers": logins})
}
