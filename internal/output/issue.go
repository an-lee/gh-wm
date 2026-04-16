package output

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/an-lee/gh-wm/internal/types"
)

func runCreateIssue(ctx context.Context, tc *types.TaskContext, item ItemCreateIssue) error {
	t := strings.TrimSpace(item.Title)
	body := strings.TrimSpace(item.Body)
	if t == "" {
		return fmt.Errorf("create_issue: empty title")
	}
	if tc.Repo == "" {
		return fmt.Errorf("create_issue: GITHUB_REPOSITORY not set")
	}
	args := []string{"issue", "create", "--title", t, "--body", body}
	for _, l := range item.Labels {
		if l != "" {
			args = append(args, "--label", l)
		}
	}
	for _, a := range item.Assignees {
		if a != "" {
			args = append(args, "--assignee", a)
		}
	}
	if tc.Repo != "" {
		args = append(args, "--repo", tc.Repo)
	}
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = tc.RepoPath
	env := os.Environ()
	if tc.Repo != "" {
		env = append(env, "GITHUB_REPOSITORY="+tc.Repo)
	}
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gh issue create: %w: %s", err, string(out))
	}
	return nil
}
