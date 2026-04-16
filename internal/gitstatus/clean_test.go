package gitstatus

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEnsureClean_emptyRepo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t")
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644)
	runGit(t, dir, "add", "a.txt")
	runGit(t, dir, "commit", "-m", "init")

	if err := EnsureClean(dir); err != nil {
		t.Fatalf("EnsureClean: %v", err)
	}
}

func TestEnsureClean_dirtyUntracked(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t")
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644)
	runGit(t, dir, "add", "a.txt")
	runGit(t, dir, "commit", "-m", "init")
	_ = os.WriteFile(filepath.Join(dir, "b.txt"), []byte("y"), 0o644)

	if err := EnsureClean(dir); err == nil {
		t.Fatal("expected error for untracked file")
	}
}

func TestEnsureClean_notGitRepo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := EnsureClean(dir); err == nil {
		t.Fatal("expected error for non-repo")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
