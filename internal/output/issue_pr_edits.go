package output

import (
	"context"
	"fmt"
	"strings"

	"github.com/an-lee/gh-wm/internal/ghclient"
	"github.com/an-lee/gh-wm/internal/types"
)

func runUpdateIssue(ctx context.Context, tc *types.TaskContext, item ItemUpdateIssue) error {
	n := resolveIssueTarget(tc, item.Target)
	if n <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("update_issue: no issue number or repository")
	}
	titleIn := strings.TrimSpace(item.Title)
	bodyIn := item.Body
	op := normalizeUpdateOperation(item.Operation)
	if titleIn == "" && strings.TrimSpace(bodyIn) == "" {
		return fmt.Errorf("update_issue: empty title and body")
	}
	if strings.TrimSpace(bodyIn) == "" {
		return ghclient.UpdateIssue(ctx, tc.Repo, n, titleIn, "")
	}
	outTitle, outBody, err := computeUpdatedBody(ctx, tc.Repo, n, titleIn, bodyIn, op, ghclient.GetIssueSnapshot)
	if err != nil {
		return fmt.Errorf("update_issue: %w", err)
	}
	return ghclient.UpdateIssue(ctx, tc.Repo, n, outTitle, outBody)
}

func runUpdatePullRequest(ctx context.Context, tc *types.TaskContext, item ItemUpdatePullRequest) error {
	n := resolvePRTarget(tc, item.Target)
	if n <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("update_pull_request: no pull request number or repository")
	}
	titleIn := strings.TrimSpace(item.Title)
	bodyIn := item.Body
	op := normalizeUpdateOperation(item.Operation)
	if titleIn == "" && strings.TrimSpace(bodyIn) == "" {
		return fmt.Errorf("update_pull_request: empty title and body")
	}
	if strings.TrimSpace(bodyIn) == "" {
		return ghclient.UpdatePullRequest(ctx, tc.Repo, n, titleIn, "")
	}
	outTitle, outBody, err := computeUpdatedBody(ctx, tc.Repo, n, titleIn, bodyIn, op, ghclient.GetIssueSnapshot)
	if err != nil {
		return fmt.Errorf("update_pull_request: %w", err)
	}
	return ghclient.UpdatePullRequest(ctx, tc.Repo, n, outTitle, outBody)
}

// snapshotFn fetches current title and body (issue and PR share the same issue API for PR numbers).
type snapshotFn func(ctx context.Context, repo string, number int) (title, body string, err error)

func computeUpdatedBody(ctx context.Context, repo string, number int, titleIn, bodyIn, op string, fetch snapshotFn) (outTitle, outBody string, err error) {
	switch op {
	case "replace", "":
		return titleIn, strings.TrimSpace(bodyIn), nil
	case "append", "prepend", "replace_island":
		curTitle, curBody, err := fetch(ctx, repo, number)
		if err != nil {
			return "", "", fmt.Errorf("fetch current: %w", err)
		}
		outTitle = titleIn
		if outTitle == "" {
			outTitle = curTitle
		}
		trim := strings.TrimSpace(bodyIn)
		switch op {
		case "append":
			outBody = strings.TrimRight(curBody, "\n\r") + "\n\n" + trim
		case "prepend":
			outBody = trim + "\n\n" + strings.TrimLeft(curBody, "\n\r")
		case "replace_island":
			outBody, err = replaceGhWMIsland(curBody, trim)
			if err != nil {
				return "", "", err
			}
		}
		return outTitle, outBody, nil
	default:
		return "", "", fmt.Errorf("unknown operation %q (want replace, append, prepend, replace-island)", op)
	}
}

func runCloseIssue(ctx context.Context, tc *types.TaskContext, item ItemCloseIssue) error {
	n := resolveIssueTarget(tc, item.Target)
	if n <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("close_issue: no issue number or repository")
	}
	return ghclient.CloseIssue(ctx, tc.Repo, n, item.Comment, item.StateReason)
}

func runClosePullRequest(ctx context.Context, tc *types.TaskContext, item ItemClosePullRequest) error {
	n := resolvePRTarget(tc, item.Target)
	if n <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("close_pull_request: no pull request number or repository")
	}
	return ghclient.ClosePullRequest(ctx, tc.Repo, n, item.Comment)
}

func runAddReviewers(ctx context.Context, tc *types.TaskContext, item ItemAddReviewer) error {
	n := resolvePRTarget(tc, item.Target)
	if n <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("add_reviewer: no pull request number or repository")
	}
	if len(item.Reviewers) == 0 {
		return fmt.Errorf("add_reviewer: empty reviewers")
	}
	return ghclient.RequestPullReviewers(ctx, tc.Repo, n, item.Reviewers)
}
