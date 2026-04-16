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

// runCreatePullRequestItem runs create_pull_request from agent output.json.
func runCreatePullRequestItem(ctx context.Context, glob *config.GlobalConfig, task *config.Task, tc *types.TaskContext, p *Policy, item ItemCreatePullRequest) error {
	if p == nil {
		p = newPolicy(task)
	}
	title := strings.TrimSpace(item.Title)
	if title == "" {
		title = fmt.Sprintf("[%s] wm task", task.Name)
	}
	title = p.ApplyTitlePrefix(KindCreatePullRequest, title)
	body := strings.TrimSpace(item.Body)
	if body == "" {
		body = "Opened by **gh-wm** task `" + task.Name + "`."
	}
	draft := p.ResolveDraft(glob, KindCreatePullRequest, item.Draft)
	labels := p.MergeLabels(KindCreatePullRequest, item.Labels)
	return tryCreatePullRequest(ctx, task, tc, title, body, draft, labels)
}

func tryCreatePullRequest(ctx context.Context, task *config.Task, tc *types.TaskContext, title, body string, draft bool, labels []string) error {
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

	args := []string{"pr", "create", "--base", base, "--title", title, "--body", body}
	if draft {
		args = append(args, "--draft")
	}
	for _, l := range labels {
		if l != "" {
			args = append(args, "--label", l)
		}
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
