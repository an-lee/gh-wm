package output

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

func TestDetectDefaultBaseBranch(t *testing.T) {
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
	b := detectDefaultBaseBranch(dir)
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

func TestRunPROutput_NoCommitsAhead(t *testing.T) {
	t.Parallel()
	dir := gitRepoWithOriginMain(t)
	glob := &config.GlobalConfig{}
	task := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"create-pull-request": map[string]any{},
	}}}
	tc := &types.TaskContext{RepoPath: dir, Repo: "o/r"}
	if err := runPROutput(context.Background(), glob, task, tc, nil); err != nil {
		t.Fatal(err)
	}
}
