package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveCommand_JSON(t *testing.T) {
	root := t.TempDir()
	wm := filepath.Join(root, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte("version: 1\nengine: claude\nmax_turns: 10\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	task := `---
on:
  issues:
    types: [opened]
---

x
`
	if err := os.WriteFile(filepath.Join(wm, "tasks", "t.md"), []byte(task), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(root, "event.json")
	if err := os.WriteFile(payload, []byte(`{"action":"opened"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"resolve", "--repo-root", root, "--event-name", "issues", "--payload", payload, "--json"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"t"`)) {
		t.Fatalf("output: %s", buf.String())
	}
}

func TestResolveCommand_Plain(t *testing.T) {
	root := t.TempDir()
	wm := filepath.Join(root, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte("version: 1\nengine: claude\nmax_turns: 10\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "tasks", "only.md"), []byte(`---
on:
  issues: {}
---

x
`), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(root, "ev.json")
	if err := os.WriteFile(payload, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"resolve", "--repo-root", root, "--event-name", "issues", "--payload", payload, "--json=false"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if buf.Len() < 2 {
		t.Fatal("expected task name line")
	}
}

func TestResolveCommand_ForceTask(t *testing.T) {
	root := t.TempDir()
	wm := filepath.Join(root, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte("version: 1\nengine: claude\nmax_turns: 10\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "tasks", "pinned.md"), []byte(`---
on:
  issues: {}
---

x
`), 0o644); err != nil {
		t.Fatal(err)
	}
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"resolve", "--repo-root", root, "--force-task", "pinned", "--json"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"pinned"`)) {
		t.Fatalf("output: %s", buf.String())
	}
}
