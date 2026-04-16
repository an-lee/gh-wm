package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/an-lee/gh-wm/internal/gen"
	"github.com/an-lee/gh-wm/internal/types"
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

func TestResolveMatchingTasks_SkipAutomatedSender(t *testing.T) {
	t.Parallel()
	root := writeMinimalRepo(t)
	ev := &types.GitHubEvent{
		Name: "issues",
		Payload: map[string]any{
			"action": "opened",
			"sender": map[string]any{"type": "Bot", "login": "app[bot]"},
		},
	}
	names, err := ResolveMatchingTasks(root, ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Fatalf("expected skip bot sender, got %v", names)
	}
}

func TestResolveMatchingTasks_ScheduleIgnoresSender(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	wm := filepath.Join(root, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte("version: 1\nengine: claude\nmax_turns: 10\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	schedPath := filepath.Join(wm, "tasks", "sched.md")
	task := `---
on:
  schedule: daily
---
x
`
	if err := os.WriteFile(schedPath, []byte(task), 0o644); err != nil {
		t.Fatal(err)
	}
	ev := &types.GitHubEvent{
		Name: "schedule",
		Payload: map[string]any{
			"sender": map[string]any{"type": "Bot", "login": "dependabot[bot]"},
		},
	}
	names, err := ResolveMatchingTasks(root, ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "sched" {
		t.Fatalf("schedule resolve should ignore bot sender, got %v", names)
	}
}

func TestResolveMatchingTasks_SkipStateLabeledEvent(t *testing.T) {
	t.Parallel()
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
    types: [labeled]
---

x
`
	if err := os.WriteFile(filepath.Join(wm, "tasks", "impl.md"), []byte(task), 0o644); err != nil {
		t.Fatal(err)
	}
	task2 := `---
on:
  issues:
    types: [opened]
wm:
  state_labels:
    working: "agent:working"
---

y
`
	if err := os.WriteFile(filepath.Join(wm, "tasks", "other.md"), []byte(task2), 0o644); err != nil {
		t.Fatal(err)
	}
	ev := &types.GitHubEvent{
		Name: "issues",
		Payload: map[string]any{
			"action": "labeled",
			"label":  map[string]any{"name": "agent:working"},
		},
	}
	names, err := ResolveMatchingTasks(root, ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Fatalf("state label event should match no tasks, got %v", names)
	}
}

func TestResolveForcedTask(t *testing.T) {
	t.Parallel()
	root := writeMinimalRepo(t)
	names, err := ResolveForcedTask(root, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "a" {
		t.Fatalf("got %v", names)
	}
	if _, err := ResolveForcedTask(root, "missing"); err == nil {
		t.Fatal("expected error")
	}
	if _, err := ResolveForcedTask(root, ""); err == nil {
		t.Fatal("expected error for empty")
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
	schedPath := filepath.Join(wm, "tasks", "sched.md")
	if err := os.WriteFile(schedPath, []byte(task), 0o644); err != nil {
		t.Fatal(err)
	}
	ev := &types.GitHubEvent{Name: "schedule", Payload: map[string]any{}}
	t.Setenv("WM_SCHEDULE_CRON", gen.FuzzyNormalizeSchedule("daily", schedPath))
	names, err := ResolveMatchingTasks(root, ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "sched" {
		t.Fatalf("got %v", names)
	}
}

func BenchmarkResolveMatchingTasks(b *testing.B) {
	root := b.TempDir()
	wm := filepath.Join(root, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte("version: 1\nengine: claude\nmax_turns: 10\n"), 0o644); err != nil {
		b.Fatal(err)
	}
	task := `---
on:
  issues:
    types: [opened]
---

prompt
`
	if err := os.WriteFile(filepath.Join(wm, "tasks", "a.md"), []byte(task), 0o644); err != nil {
		b.Fatal(err)
	}
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"action": "opened"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ResolveMatchingTasks(root, ev); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkResolveMatchingTasks_CacheHit measures the cached (warm) path: a second call
// to ResolveMatchingTasks with the same repo root hits the config cache, skipping
// all file I/O and YAML parsing. First call within each iteration establishes the cache.
func BenchmarkResolveMatchingTasks_CacheHit(b *testing.B) {
	root := b.TempDir()
	wm := filepath.Join(root, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte("version: 1\nengine: claude\nmax_turns: 10\n"), 0o644); err != nil {
		b.Fatal(err)
	}
	task := `---
on:
  issues:
    types: [opened]
---

prompt
`
	if err := os.WriteFile(filepath.Join(wm, "tasks", "a.md"), []byte(task), 0o644); err != nil {
		b.Fatal(err)
	}
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"action": "opened"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// cold call — primes the cache
		if _, err := ResolveMatchingTasks(root, ev); err != nil {
			b.Fatal(err)
		}
		// hot call — served from cache
		if _, err := ResolveMatchingTasks(root, ev); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(0.5, "calls/op")
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

func TestParseEvent_EmptyPath(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "")
	ev, err := ParseEvent("issues", "")
	if err != nil {
		t.Fatal(err)
	}
	if ev.Name != "issues" {
		t.Fatalf("name: %q", ev.Name)
	}
	if len(ev.Payload) != 0 {
		t.Fatalf("payload: %+v", ev.Payload)
	}
}

func TestParseEvent_EmptyPath_DefaultNameFromEnv(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "workflow_dispatch")
	ev, err := ParseEvent("", "")
	if err != nil {
		t.Fatal(err)
	}
	if ev.Name != "workflow_dispatch" {
		t.Fatal(ev.Name)
	}
}
