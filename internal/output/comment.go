package output

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gh-wm/gh-wm/internal/config"
	"github.com/gh-wm/gh-wm/internal/types"
)

const maxCommentBody = 60000

func runCommentOutput(_ context.Context, _ *config.GlobalConfig, task *config.Task, tc *types.TaskContext, res *types.AgentResult) error {
	n := commentTargetNumber(tc)
	if n <= 0 {
		return fmt.Errorf("no issue or PR number in event context for add-comment")
	}
	body := strings.TrimSpace(res.Summary)
	if body == "" {
		body = strings.TrimSpace(res.Stdout)
	}
	if body == "" {
		body = fmt.Sprintf("Task %q completed (no agent output).", task.Name)
	}
	if len(body) > maxCommentBody {
		body = body[:maxCommentBody] + "\n\n…(truncated)"
	}

	var cmd *exec.Cmd
	if tc.PRNumber > 0 {
		cmd = exec.Command("gh", "pr", "comment", fmt.Sprintf("%d", tc.PRNumber), "--body", body)
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
