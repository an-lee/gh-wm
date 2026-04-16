package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveCommand_MissingPayload(t *testing.T) {
	t.Setenv("GITHUB_EVENT_PATH", "")
	t.Cleanup(func() { _ = os.Unsetenv("GITHUB_EVENT_PATH") })
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
  issues: {}
---

x
`
	if err := os.WriteFile(filepath.Join(wm, "tasks", "a.md"), []byte(task), 0o644); err != nil {
		t.Fatal(err)
	}
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"resolve", "--repo-root", root, "--event-name", "issues"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var names []string
	if err := json.Unmarshal(buf.Bytes(), &names); err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "a" {
		t.Fatalf("got %v", names)
	}
}
