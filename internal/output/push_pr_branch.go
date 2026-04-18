package output

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/gitbranch"
	"github.com/an-lee/gh-wm/internal/types"
)

type prViewForPush struct {
	HeadRefName string `json:"headRefName"`
	Title       string `json:"title"`
	Labels      []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

func runPushToPullRequestBranch(ctx context.Context, _ *config.GlobalConfig, task *config.Task, tc *types.TaskContext, p *Policy, raw map[string]any) error {
	item := mapToPushToPullRequestBranch(raw)
	n := resolvePRTarget(tc, item.Target)
	if n <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
		return fmt.Errorf("push_to_pull_request_branch: no PR number or repository")
	}
	dir := strings.TrimSpace(tc.RepoPath)
	if dir == "" {
		return fmt.Errorf("push_to_pull_request_branch: WM_REPO_ROOT / repo path not set")
	}
	pol := p
	if pol == nil {
		pol = newPolicy(task)
	}
	title, labelNames, headRef, err := prHeadTitleLabels(ctx, dir, tc.Repo, n)
	if err != nil {
		return err
	}
	if err := pol.PushPRTitleMatchesPolicy(title); err != nil {
		return err
	}
	if err := pol.PushPRHasRequiredLabels(labelNames); err != nil {
		return err
	}
	cur, err := gitbranch.CurrentBranch(dir)
	if err != nil {
		return fmt.Errorf("push_to_pull_request_branch: git current branch: %w", err)
	}
	if cur != headRef {
		return fmt.Errorf("push_to_pull_request_branch: current branch %q must match PR head %q", cur, headRef)
	}
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "push", "-u", "origin", "HEAD")
	cmd.Env = os.Environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("push_to_pull_request_branch: git push: %w: %s", err, string(out))
	}
	return nil
}

func prHeadTitleLabels(ctx context.Context, dir, repo string, prNumber int) (title string, labelNames []string, headRef string, err error) {
	args := []string{"pr", "view", fmt.Sprintf("%d", prNumber), "--json", "headRefName,title,labels"}
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
		return "", nil, "", fmt.Errorf("gh pr view: %w: %s", err, strings.TrimSpace(string(out)))
	}
	var v prViewForPush
	if err := json.Unmarshal(out, &v); err != nil {
		return "", nil, "", fmt.Errorf("parse gh pr view json: %w", err)
	}
	if strings.TrimSpace(v.HeadRefName) == "" {
		return "", nil, "", fmt.Errorf("gh pr view: empty headRefName")
	}
	for _, l := range v.Labels {
		if l.Name != "" {
			labelNames = append(labelNames, l.Name)
		}
	}
	return v.Title, labelNames, v.HeadRefName, nil
}
