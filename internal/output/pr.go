package output

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gh-wm/gh-wm/internal/config"
	"github.com/gh-wm/gh-wm/internal/types"
)

func runPROutput(ctx context.Context, glob *config.GlobalConfig, task *config.Task, tc *types.TaskContext, _ *types.AgentResult) error {
	dir := tc.RepoPath
	base := detectDefaultBaseBranch(dir)
	ahead, err := commitsAheadOfBase(dir, base)
	if err != nil || ahead == 0 {
		return nil
	}

	push := exec.CommandContext(ctx, "git", "-C", dir, "push", "-u", "origin", "HEAD")
	push.Env = os.Environ()
	if out, err := push.CombinedOutput(); err != nil {
		return fmt.Errorf("git push: %w: %s", err, string(out))
	}

	draft := glob.PR.Draft
	var labels []string
	if so := task.SafeOutputsMap(); so != nil {
		if m, ok := so["create-pull-request"].(map[string]any); ok {
			if d, ok := m["draft"].(bool); ok {
				draft = d
			}
			if raw, ok := m["labels"].([]any); ok {
				for _, x := range raw {
					if s, ok := x.(string); ok && s != "" {
						labels = append(labels, s)
					}
				}
			}
		}
	}

	title := fmt.Sprintf("[%s] wm task", task.Name)
	body := "Opened by **gh-wm** task `" + task.Name + "`."

	args := []string{"pr", "create", "--title", title, "--body", body}
	if draft {
		args = append(args, "--draft")
	}
	for _, l := range labels {
		args = append(args, "--label", l)
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = dir
	env := os.Environ()
	if tc.Repo != "" {
		env = append(env, "GITHUB_REPOSITORY="+tc.Repo)
	}
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gh pr create: %w: %s", err, string(out))
	}
	return nil
}

func detectDefaultBaseBranch(dir string) string {
	cmd := exec.Command("git", "-C", dir, "symbolic-ref", "refs/remotes/origin/HEAD")
	out, err := cmd.Output()
	if err == nil {
		s := strings.TrimSpace(string(out))
		if i := strings.LastIndex(s, "/"); i >= 0 {
			return strings.TrimSpace(s[i+1:])
		}
	}
	for _, b := range []string{"main", "master"} {
		c := exec.Command("git", "-C", dir, "rev-parse", "--verify", "origin/"+b)
		if c.Run() == nil {
			return b
		}
	}
	return "main"
}

func commitsAheadOfBase(dir, base string) (int, error) {
	ref := "origin/" + base
	cmd := exec.Command("git", "-C", dir, "rev-list", "--count", ref+"..HEAD")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, err
	}
	return n, nil
}
