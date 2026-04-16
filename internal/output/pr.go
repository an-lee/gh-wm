package output

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/gitbranch"
	"github.com/an-lee/gh-wm/internal/types"
)

func runPROutput(ctx context.Context, glob *config.GlobalConfig, task *config.Task, tc *types.TaskContext, _ *types.AgentResult) error {
	dir := tc.RepoPath
	base := gitbranch.DefaultBaseBranch(dir)
	cur, err := gitbranch.CurrentBranch(dir)
	if err != nil {
		return fmt.Errorf("git current branch: %w", err)
	}
	if cur == base {
		return nil
	}

	if has, err := headHasOpenPR(ctx, dir, tc.Repo, cur); err != nil {
		return fmt.Errorf("gh pr list: %w", err)
	} else if has {
		return nil
	}

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
	var titlePrefix string
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
			if p, ok := m["title-prefix"].(string); ok && strings.TrimSpace(p) != "" {
				titlePrefix = strings.TrimSpace(p)
			}
		}
	}

	title := fmt.Sprintf("[%s] wm task", task.Name)
	if titlePrefix != "" {
		title = titlePrefix + title
	}
	body := "Opened by **gh-wm** task `" + task.Name + "`."

	args := []string{"pr", "create", "--base", base, "--title", title, "--body", body}
	if draft {
		args = append(args, "--draft")
	}
	for _, l := range labels {
		args = append(args, "--label", l)
	}
	if tc.Repo != "" {
		args = append(args, "--repo", tc.Repo)
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

func headHasOpenPR(ctx context.Context, dir, repo, headBranch string) (bool, error) {
	args := []string{"pr", "list", "--head", headBranch, "--json", "number"}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	if repo != "" {
		cmd.Env = append(cmd.Env, "GITHUB_REPOSITORY="+repo)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	var list []struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal(out, &list); err != nil {
		return false, err
	}
	return len(list) > 0, nil
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
