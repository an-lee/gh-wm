package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func chdirTemp(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}

func TestInitCommand(t *testing.T) {
	root := t.TempDir()
	chdirTemp(t, root)
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".wm", "config.yml")); err != nil {
		t.Fatal(err)
	}
}

func TestUpgradeCommand(t *testing.T) {
	root := t.TempDir()
	wm := filepath.Join(root, ".wm", "tasks")
	if err := os.MkdirAll(wm, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "s.md"), []byte(`---
on:
  schedule: hourly
---

x
`), 0o644); err != nil {
		t.Fatal(err)
	}
	chdirTemp(t, root)
	t.Setenv("GH_WM_REPO", "test/hello")
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"upgrade"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "workflows", "wm-agent.yml")); err != nil {
		t.Fatal(err)
	}
}

func TestAddCommand_LocalFile(t *testing.T) {
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "task.md")
	content := `---
on:
  issues: {}
---

body
`
	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".wm"), 0o755); err != nil {
		t.Fatal(err)
	}
	chdirTemp(t, root)
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"add", src})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, ".wm", "tasks", "task.md")
	if _, err := os.Stat(dst); err != nil {
		t.Fatal(err)
	}
}
