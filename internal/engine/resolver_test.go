package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gh-wm/gh-wm/internal/types"
)

func writeMinimalRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	wm := filepath.Join(root, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte(`version: 1
engine: claude
max_turns: 10
`), 0o644); err != nil {
		t.Fatal(err)
	}
	task := `---
on:
  issues:
    types: [opened]
---

prompt
`
	if err := os.WriteFile(filepath.Join(wm, "tasks", "a.md"), []byte(task), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestResolveMatchingTasks(t *testing.T) {
	t.Parallel()
	root := writeMinimalRepo(t)
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"action": "opened"}}
	names, err := ResolveMatchingTasks(root, ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "a" {
		t.Fatalf("got %v", names)
	}
}

func TestResolveMatchingTasks_ScheduleCronEnv(t *testing.T) {
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
  schedule: daily
---

x
`
	if err := os.WriteFile(filepath.Join(wm, "tasks", "sched.md"), []byte(task), 0o644); err != nil {
		t.Fatal(err)
	}
	ev := &types.GitHubEvent{Name: "schedule", Payload: map[string]any{}}
	t.Setenv("WM_SCHEDULE_CRON", "0 0 * * *")
	names, err := ResolveMatchingTasks(root, ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "sched" {
		t.Fatalf("got %v", names)
	}
}

func TestParseEventFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "ev.json")
	if err := os.WriteFile(p, []byte(`{"action":"opened"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_EVENT_NAME", "")
	ev, err := ParseEventFile("issues", p)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Name != "issues" || ev.Payload["action"] != "opened" {
		t.Fatalf("%+v", ev)
	}
	if _, err := ParseEventFile("", "/nonexistent"); err == nil {
		t.Fatal("expected read error")
	}
	bad := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(bad, []byte(`{`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseEventFile("x", bad); err == nil {
		t.Fatal("bad json")
	}
}

func TestParseEventFile_DefaultNameFromEnv(t *testing.T) {
	p := filepath.Join(t.TempDir(), "ev.json")
	if err := os.WriteFile(p, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_EVENT_NAME", "workflow_dispatch")
	ev, err := ParseEventFile("", p)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Name != "workflow_dispatch" {
		t.Fatal(ev.Name)
	}
}
