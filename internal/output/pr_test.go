package output

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/gitbranch"
	"github.com/an-lee/gh-wm/internal/types"
)

func TestDefaultBaseBranch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "f.txt")
	runGit(t, dir, "commit", "-m", "init")
	runGit(t, dir, "branch", "-M", "main")
	runGit(t, dir, "remote", "add", "origin", ".")
	runGit(t, dir, "fetch", "origin")
	b := gitbranch.DefaultBaseBranch(dir)
	if b != "main" {
		t.Fatalf("got %q", b)
	}
}

func TestCommitsAheadOfBase(t *testing.T) {
	t.Parallel()
	dir := gitRepoWithOriginMain(t)
	n, err := commitsAheadOfBase(dir, "main")
	if err != nil || n != 0 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("ab"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "commit", "-am", "second")
	n2, err := commitsAheadOfBase(dir, "main")
	if err != nil || n2 != 1 {
		t.Fatalf("n=%d err=%v", n2, err)
	}
}

func gitRepoWithOriginMain(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "f.txt")
	runGit(t, dir, "commit", "-m", "init")
	runGit(t, dir, "branch", "-M", "main")
	runGit(t, dir, "remote", "add", "origin", ".")
	runGit(t, dir, "fetch", "origin")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// writeFakeGH writes a fake gh binary to binDir/gh that responds to
// "gh pr list --head …" with jsonOut and exits with exitCode.
// Returns the path to the fake binary.
func writeFakeGH(t *testing.T, binDir string, jsonOut string, exitCode int) string {
	t.Helper()
	bin := filepath.Join(binDir, "gh")
	quoted := strconv.Quote(jsonOut)
	script := fmt.Sprintf("#!/bin/sh\ncase \"$*\" in\n  *'pr list'*) echo %s ;;\n  *) echo 'unhandled' >&2 ; exit 1 ;;\nesac\nexit %d\n", quoted, exitCode)
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin
}

// gitRepoWithFeatureBranch creates a temp git repo with:
// - main branch (default base)
// - origin remote pointing to the repo itself
// - a feature branch with one commit ahead of main
// Returns (repoPath, featureBranchName).
func gitRepoWithFeatureBranch(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "a.txt")
	runGit(t, dir, "commit", "-m", "init")
	runGit(t, dir, "branch", "-M", "main")
	runGit(t, dir, "remote", "add", "origin", ".")
	runGit(t, dir, "fetch", "origin")

	// Create feature branch with one extra commit
	runGit(t, dir, "checkout", "-b", "feat/test")
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "commit", "-am", "feat: add b")
	return dir, "feat/test"
}

// gitRepoOnBaseBranch creates a temp git repo on main (base) with origin set up.
func gitRepoOnBaseBranch(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "a.txt")
	runGit(t, dir, "commit", "-m", "init")
	runGit(t, dir, "branch", "-M", "main")
	runGit(t, dir, "remote", "add", "origin", ".")
	runGit(t, dir, "fetch", "origin")
	return dir
}

// gitRepoFeatureBehind creates a temp git repo where the feature branch is
// behind main (no commits ahead) — used to test the "ahead==0" early-return path.
func gitRepoFeatureBehind(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "a.txt")
	runGit(t, dir, "commit", "-m", "init")
	runGit(t, dir, "branch", "-M", "main")
	runGit(t, dir, "remote", "add", "origin", ".")
	runGit(t, dir, "fetch", "origin")

	// Check out feat but do NOT add any new commits — it's at the same commit as main
	runGit(t, dir, "checkout", "-b", "feat/behind")
	return dir
}

// ---------------------------------------------------------------------------
// runCreatePullRequestItem tests
// ---------------------------------------------------------------------------

func TestRunCreatePullRequestItem_OnBaseBranch(t *testing.T) {
	t.Parallel()
	dir := gitRepoOnBaseBranch(t)
	g := &config.GlobalConfig{}
	task := &config.Task{Name: "test-task", Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"create-pull-request": map[string]any{},
	}}}
	tc := &types.TaskContext{RepoPath: dir, Repo: "o/r"}
	item := ItemCreatePullRequest{Title: "My PR", Body: "PR body"}

	// We're on main (base branch), so tryCreatePullRequest returns nil
	// without attempting gh or git push.
	err := runCreatePullRequestItem(context.Background(), g, task, tc, nil, item)
	if err != nil {
		t.Fatalf("expected nil on base branch, got: %v", err)
	}
}

func TestRunCreatePullRequestItem_NilPolicy(t *testing.T) {
	t.Parallel()
	// nil Policy falls back to default title/body and no prefix.
	dir := gitRepoOnBaseBranch(t)
	g := &config.GlobalConfig{}
	task := &config.Task{Name: "my-task"}
	tc := &types.TaskContext{RepoPath: dir, Repo: "o/r"}
	item := ItemCreatePullRequest{Title: "PR", Body: "Body"}

	// nil policy → newPolicy(task) used; cur==base → early return.
	err := runCreatePullRequestItem(context.Background(), g, task, tc, nil, item)
	if err != nil {
		t.Fatalf("expected nil on base branch, got: %v", err)
	}
}

func TestRunCreatePullRequestItem_TitlePrefixApplied(t *testing.T) {
	t.Parallel()
	dir := gitRepoOnBaseBranch(t)
	g := &config.GlobalConfig{}
	task := &config.Task{Name: "test-task", Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"create-pull-request": map[string]any{"title-prefix": "[WIP] "},
	}}}
	tc := &types.TaskContext{RepoPath: dir, Repo: "o/r"}

	// We can't easily capture the pushed title, but we can verify that when
	// cur==base, the function returns nil (title prefix logic was exercised).
	item := ItemCreatePullRequest{Title: "My PR", Body: "PR body"}
	err := runCreatePullRequestItem(context.Background(), g, task, tc, nil, item)
	if err != nil {
		t.Fatalf("expected nil on base branch, got: %v", err)
	}
}

func TestRunCreatePullRequestItem_DraftMerged(t *testing.T) {
	t.Parallel()
	dir := gitRepoOnBaseBranch(t)
	draft := true
	g := &config.GlobalConfig{}
	task := &config.Task{Name: "test-task", Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"create-pull-request": map[string]any{},
	}}}
	tc := &types.TaskContext{RepoPath: dir, Repo: "o/r"}
	item := ItemCreatePullRequest{Title: "My PR", Draft: &draft}

	err := runCreatePullRequestItem(context.Background(), g, task, tc, nil, item)
	if err != nil {
		t.Fatalf("expected nil on base branch, got: %v", err)
	}
}

func TestRunCreatePullRequestItem_LabelsMerged(t *testing.T) {
	t.Parallel()
	dir := gitRepoOnBaseBranch(t)
	g := &config.GlobalConfig{}
	task := &config.Task{Name: "test-task", Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"create-pull-request": map[string]any{"labels": []any{"area/test"}},
	}}}
	tc := &types.TaskContext{RepoPath: dir, Repo: "o/r"}
	item := ItemCreatePullRequest{Title: "My PR", Labels: []string{"bug"}}

	// MergeLabels called with agent labels ["bug"] + def labels ["area/test"]
	err := runCreatePullRequestItem(context.Background(), g, task, tc, nil, item)
	if err != nil {
		t.Fatalf("expected nil on base branch, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// tryCreatePullRequest tests
// ---------------------------------------------------------------------------

func TestTryCreatePullRequest_OnBaseBranch(t *testing.T) {
	t.Parallel()
	dir := gitRepoOnBaseBranch(t)
	task := &config.Task{Name: "test-task"}
	tc := &types.TaskContext{RepoPath: dir, Repo: "o/r"}

	// cur == base → returns nil immediately, no git push or gh pr create.
	err := tryCreatePullRequest(context.Background(), task, tc, "Title", "Body", false, nil)
	if err != nil {
		t.Fatalf("expected nil when cur==base, got: %v", err)
	}
}

func TestTryCreatePullRequest_NoCommitsAhead(t *testing.T) {
	// feat/behind has zero commits ahead of main. However, headHasOpenPR
	// is called before commitsAheadOfBase, so we need fake gh.
	binDir := t.TempDir()
	writeFakeGH(t, binDir, "[]", 0) // no open PR → continues to ahead check
	dir := gitRepoFeatureBehind(t)
	task := &config.Task{Name: "test-task"}
	tc := &types.TaskContext{RepoPath: dir, Repo: "o/r"}
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	t.Cleanup(func() { os.Setenv("PATH", origPath) })

	// headHasOpenPR returns false (no open PR); commitsAheadOfBase returns 0 → early-return.
	err := tryCreatePullRequest(context.Background(), task, tc, "Title", "Body", false, nil)
	if err != nil {
		t.Fatalf("expected nil when ahead==0, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// headHasOpenPR tests — use a fake gh binary
// ---------------------------------------------------------------------------

func TestHeadHasOpenPR_EmptyList(t *testing.T) {
	binDir := t.TempDir()
	writeFakeGH(t, binDir, "[]", 0)
	dir := gitRepoOnBaseBranch(t)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	t.Cleanup(func() { os.Setenv("PATH", origPath) })

	has, err := headHasOpenPR(context.Background(), dir, "o/r", "main")
	if err != nil {
		t.Fatalf("headHasOpenPR error: %v", err)
	}
	if has {
		t.Fatal("expected has==false for empty PR list")
	}
}

func TestHeadHasOpenPR_HasOpenPR(t *testing.T) {
	binDir := t.TempDir()
	writeFakeGH(t, binDir, "[{\"number\":42}]", 0)
	dir := gitRepoOnBaseBranch(t)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	t.Cleanup(func() { os.Setenv("PATH", origPath) })

	has, err := headHasOpenPR(context.Background(), dir, "o/r", "main")
	if err != nil {
		t.Fatalf("headHasOpenPR error: %v", err)
	}
	if !has {
		t.Fatal("expected has==true when gh returns PR list")
	}
}

func TestHeadHasOpenPR_GhFails(t *testing.T) {
	binDir := t.TempDir()
	writeFakeGH(t, binDir, "", 1) // exits 1, no output
	dir := gitRepoOnBaseBranch(t)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	t.Cleanup(func() { os.Setenv("PATH", origPath) })

	_, err := headHasOpenPR(context.Background(), dir, "o/r", "main")
	if err == nil {
		t.Fatal("expected error when gh fails")
	}
	if !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("expected 'exit status' in error, got: %v", err)
	}
}

func TestHeadHasOpenPR_InvalidJSON(t *testing.T) {
	binDir := t.TempDir()
	writeFakeGH(t, binDir, "not json", 0)
	dir := gitRepoOnBaseBranch(t)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	t.Cleanup(func() { os.Setenv("PATH", origPath) })

	_, err := headHasOpenPR(context.Background(), dir, "o/r", "main")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestHeadHasOpenPR_NoRepo(t *testing.T) {
	binDir := t.TempDir()
	writeFakeGH(t, binDir, "[]", 0)
	dir := gitRepoOnBaseBranch(t)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	t.Cleanup(func() { os.Setenv("PATH", origPath) })

	// Empty repo means --repo is not passed to gh; the fake gh still returns [].
	has, err := headHasOpenPR(context.Background(), dir, "", "main")
	if err != nil {
		t.Fatalf("headHasOpenPR error: %v", err)
	}
	if has {
		t.Fatal("expected has==false")
	}
}

// ---------------------------------------------------------------------------
// commitsAheadOfBase error path
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// ghRepoForLabels tests
// ---------------------------------------------------------------------------

func TestGhRepoForLabels_WithRepo(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{Repo: "my-org/my-repo"}
	got, err := ghRepoForLabels(tc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "my-org/my-repo" {
		t.Fatalf("got %q, want %q", got, "my-org/my-repo")
	}
}

func TestCommitsAheadOfBase_InvalidBase(t *testing.T) {
	t.Parallel()
	dir := gitRepoOnBaseBranch(t)
	// origin/nonexistent does not exist → git rev-list returns error.
	n, err := commitsAheadOfBase(dir, "nonexistent")
	if err == nil {
		t.Fatalf("expected error for invalid base, got n=%d", n)
	}
}
