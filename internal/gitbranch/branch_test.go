package gitbranch

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlugTaskName(t *testing.T) {
	t.Parallel()
	if g, w := slugTaskName("Daily Doc Updater!"), "daily-doc-updater"; g != w {
		t.Fatalf("got %q want %q", g, w)
	}
	if g := slugTaskName("   "); g != "task" {
		t.Fatalf("got %q", g)
	}
}

func TestPrepareFeatureForPR_OnMainCreatesBranch(t *testing.T) {
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

	prev, newB, created, err := PrepareFeatureForPR(dir, "my-task")
	if err != nil {
		t.Fatal(err)
	}
	if !created || prev != "main" {
		t.Fatalf("prev=%q created=%v", prev, created)
	}
	if !strings.HasPrefix(newB, "wm/my-task-") {
		t.Fatalf("new branch: %q", newB)
	}
	cur, err := CurrentBranch(dir)
	if err != nil || cur != newB {
		t.Fatalf("cur=%q err=%v", cur, err)
	}
}

func TestPrepareFeatureForPR_AlreadyOnFeatureSkips(t *testing.T) {
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
	runGit(t, dir, "checkout", "-b", "feature")

	prev, newB, created, err := PrepareFeatureForPR(dir, "x")
	if err != nil {
		t.Fatal(err)
	}
	if created || prev != "feature" || newB != "feature" {
		t.Fatalf("prev=%q new=%q created=%v", prev, newB, created)
	}
}

func TestCheckout(t *testing.T) {
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
	runGit(t, dir, "checkout", "-b", "feature")

	if err := Checkout(dir, "main"); err != nil {
		t.Fatal(err)
	}
	cur, err := CurrentBranch(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cur != "main" {
		t.Fatalf("got %q, want main", cur)
	}
}

func TestCheckout_EmptyBranchNoOp(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t")
	if err := Checkout(dir, ""); err != nil {
		t.Fatal("empty branch should be a no-op")
	}
	if err := Checkout(dir, "   "); err != nil {
		t.Fatal("whitespace-only branch should be a no-op")
	}
}

func TestCheckout_HEADNoOp(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t")
	if err := Checkout(dir, "HEAD"); err != nil {
		t.Fatal("HEAD should be a no-op")
	}
}

func TestCheckout_NonexistentBranch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@t")
	runGit(t, dir, "config", "user.name", "t")
	if err := Checkout(dir, "nonexistent-branch"); err == nil {
		t.Fatal("expected error for nonexistent branch")
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
