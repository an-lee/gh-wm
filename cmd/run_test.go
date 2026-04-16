package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunCommand(t *testing.T) {
	t.Setenv("WM_AGENT_CMD", "true")
	t.Cleanup(func() { _ = os.Unsetenv("WM_AGENT_CMD") })
	root := t.TempDir()
	wm := filepath.Join(root, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte("version: 1\nengine: claude\nmax_turns: 10\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "tasks", "mytask.md"), []byte(`---
on:
  issues: {}
---

run me
`), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(root, "ev.json")
	if err := os.WriteFile(payload, []byte(`{"action":"opened"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "t@t")
	runGit(t, root, "config", "user.name", "t")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "init")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"run", "--repo-root", root, "--task", "mytask", "--event-name", "issues", "--payload", payload})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRunCommand_TaskMissing(t *testing.T) {
	root := t.TempDir()
	wm := filepath.Join(root, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte("version: 1\nengine: claude\nmax_turns: 10\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(root, "ev.json")
	if err := os.WriteFile(payload, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	rootCmd.SetArgs([]string{"run", "--repo-root", root, "--task", "nope", "--payload", payload})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error")
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
