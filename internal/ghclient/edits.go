package ghclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/an-lee/gh-wm/internal/gh"
)

// GetIssueSnapshot returns the current title and body (issue or PR number).
func GetIssueSnapshot(ctx context.Context, repo string, number int) (title, body string, err error) {
	if useREST() {
		return gh.GetIssueSnapshot(ctx, repo, number)
	}
	args := []string{"issue", "view", strconv.Itoa(number), "--json", "title,body"}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	cmd := exec.CommandContext(ctx, "gh", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("gh issue view: %w: %s", err, stderr.String())
	}
	var v struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.Unmarshal(out, &v); err != nil {
		return "", "", fmt.Errorf("parse gh issue view json: %w", err)
	}
	return v.Title, v.Body, nil
}

// UpdateIssue edits an issue title/body (at least one of title, body must be non-empty).
func UpdateIssue(ctx context.Context, repo string, number int, title, body string) error {
	if useREST() {
		return gh.UpdateIssue(ctx, repo, number, title, body)
	}
	args := []string{"issue", "edit", strconv.Itoa(number)}
	if t := strings.TrimSpace(title); t != "" {
		args = append(args, "--title", t)
	}
	if strings.TrimSpace(body) != "" {
		args = append(args, "--body", body)
	}
	if len(args) <= 3 {
		return fmt.Errorf("update issue: nothing to update")
	}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	cmd := exec.CommandContext(ctx, "gh", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh issue edit: %w: %s", err, stderr.String())
	}
	return nil
}

// CloseIssue closes an issue with optional comment and reason (completed, not_planned, duplicate, or gh forms).
func CloseIssue(ctx context.Context, repo string, number int, comment, stateReason string) error {
	if useREST() {
		return gh.CloseIssue(ctx, repo, number, comment, stateReason)
	}
	args := []string{"issue", "close", strconv.Itoa(number)}
	if c := strings.TrimSpace(comment); c != "" {
		args = append(args, "--comment", c)
	}
	if r := issueCloseReasonForGHCLI(stateReason); r != "" {
		args = append(args, "--reason", r)
	}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	cmd := exec.CommandContext(ctx, "gh", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh issue close: %w: %s", err, stderr.String())
	}
	return nil
}

func issueCloseReasonForGHCLI(stateReason string) string {
	s := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(stateReason, "-", "_")))
	s = strings.ReplaceAll(s, " ", "_")
	switch s {
	case "":
		return ""
	case "completed":
		return "completed"
	case "not_planned":
		return "not planned"
	case "duplicate":
		return "duplicate"
	default:
		return "completed"
	}
}

// UpdatePullRequest edits a PR title/body.
func UpdatePullRequest(ctx context.Context, repo string, number int, title, body string) error {
	if useREST() {
		return gh.UpdatePullRequest(ctx, repo, number, title, body)
	}
	args := []string{"pr", "edit", strconv.Itoa(number)}
	if t := strings.TrimSpace(title); t != "" {
		args = append(args, "--title", t)
	}
	if strings.TrimSpace(body) != "" {
		args = append(args, "--body", body)
	}
	if len(args) <= 3 {
		return fmt.Errorf("update pull request: nothing to update")
	}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	cmd := exec.CommandContext(ctx, "gh", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh pr edit: %w: %s", err, stderr.String())
	}
	return nil
}

// ClosePullRequest closes a pull request without merging.
func ClosePullRequest(ctx context.Context, repo string, number int, comment string) error {
	if useREST() {
		return gh.ClosePullRequest(ctx, repo, number, comment)
	}
	args := []string{"pr", "close", strconv.Itoa(number)}
	if c := strings.TrimSpace(comment); c != "" {
		args = append(args, "--comment", c)
	}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	cmd := exec.CommandContext(ctx, "gh", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh pr close: %w: %s", err, stderr.String())
	}
	return nil
}

// RequestPullReviewers add reviewers to a pull request.
func RequestPullReviewers(ctx context.Context, repo string, number int, reviewers []string) error {
	if useREST() {
		return gh.RequestPullReviewers(ctx, repo, number, reviewers)
	}
	args := []string{"pr", "edit", strconv.Itoa(number)}
	for _, r := range reviewers {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		args = append(args, "--add-reviewer", r)
	}
	if len(args) <= 3 {
		return fmt.Errorf("request reviewers: empty reviewers")
	}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	cmd := exec.CommandContext(ctx, "gh", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh pr edit: %w: %s", err, stderr.String())
	}
	return nil
}
