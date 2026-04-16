package output

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	return postComment(tc, n, body)
}

func postComment(tc *types.TaskContext, n int, body string) error {
	if len(body) > maxCommentBody {
		body = body[:maxCommentBody] + "\n\n…(truncated)"
	}
	var cmd *exec.Cmd
	if tc.PRNumber > 0 && n == tc.PRNumber {
		cmd = exec.Command("gh", "pr", "comment", fmt.Sprintf("%d", n), "--body", body)
	} else {
		cmd = exec.Command("gh", "issue", "comment", fmt.Sprintf("%d", n), "--body", body)
	}
	cmd.Dir = tc.RepoPath
	cmd.Env = os.Environ()
	if tc.Repo != "" {
		cmd.Env = append(cmd.Env, "GITHUB_REPOSITORY="+tc.Repo)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}

func commentTargetNumber(tc *types.TaskContext) int {
	if tc.PRNumber > 0 {
		return tc.PRNumber
	}
	return tc.IssueNumber
}
