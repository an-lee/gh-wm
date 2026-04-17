package output

import (
	"context"
	"fmt"
	"strings"

	"github.com/an-lee/gh-wm/internal/antiloop"
	"github.com/an-lee/gh-wm/internal/ghclient"
	"github.com/an-lee/gh-wm/internal/types"
)

const maxCommentBody = 60000

func resolveCommentTarget(tc *types.TaskContext, target int) int {
	if target > 0 {
		return target
	}
	return commentTargetNumber(tc)
}

// runCommentFromItem posts add_comment from structured output.
func runCommentFromItem(_ context.Context, tc *types.TaskContext, item ItemAddComment) error {
	n := resolveCommentTarget(tc, item.Target)
	if n <= 0 {
		return fmt.Errorf("add_comment: no issue or PR number (set target or use a triggering event)")
	}
	body := strings.TrimSpace(item.Body)
	if body == "" {
		return fmt.Errorf("add_comment: empty body")
	}
	body = body + WMAgentCommentMarkerFooter(tc.TaskName)
	return postComment(tc, n, body)
}

func postComment(tc *types.TaskContext, n int, body string) error {
	if len(body) > maxCommentBody {
		body = body[:maxCommentBody] + "\n\n…(truncated)"
	}
	if tc.Repo == "" {
		return fmt.Errorf("add_comment: GITHUB_REPOSITORY not set")
	}
	return ghclient.PostIssueComment(tc.Repo, n, body)
}

func commentTargetNumber(tc *types.TaskContext) int {
	if tc.PRNumber > 0 {
		return tc.PRNumber
	}
	return tc.IssueNumber
}

// WMAgentCommentMarkerFooter appends a hidden HTML marker so resolve can ignore wm-authored comments (loop guard).
func WMAgentCommentMarkerFooter(taskName string) string {
	return antiloop.WMAgentCommentMarkerFooter(taskName)
}
