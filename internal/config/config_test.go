package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGlobal(t *testing.T) {
	t.Parallel()
	g, err := ParseGlobal([]byte(`version: 1
engine: codex
max_turns: 50
`))
	if err != nil {
		t.Fatal(err)
	}
	if g.Engine != "codex" || g.MaxTurns != 50 {
		t.Fatalf("got %+v", g)
	}
	if _, err := ParseGlobal([]byte(`[`)); err == nil {
		t.Fatal("expected yaml error")
	}
}

func TestParseGlobal_WorkflowPreSteps(t *testing.T) {
	t.Parallel()
	g, err := ParseGlobal([]byte(`version: 1
workflow:
  runs_on: [ubuntu-latest]
  pre_steps:
    - uses: jdx/mise-action@v4
      with:
        cache: true
    - name: Deps
      run: bundle install
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Workflow.PreSteps) != 2 {
		t.Fatalf("pre_steps: got %d items", len(g.Workflow.PreSteps))
	}
	if g.Workflow.PreSteps[0].Uses != "jdx/mise-action@v4" {
		t.Fatalf("step 0: %+v", g.Workflow.PreSteps[0])
	}
	if v, ok := g.Workflow.PreSteps[0].With["cache"].(bool); !ok || !v {
		t.Fatalf("step 0 with.cache: %+v", g.Workflow.PreSteps[0].With)
	}
	if g.Workflow.PreSteps[1].Name != "Deps" || g.Workflow.PreSteps[1].Run != "bundle install" {
		t.Fatalf("step 1: %+v", g.Workflow.PreSteps[1])
	}
}

func TestParseGlobal_WorkflowRunsOn(t *testing.T) {
	t.Parallel()
	g, err := ParseGlobal([]byte(`version: 1
workflow:
  runs_on:
    - self-hosted
    - linux
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Workflow.RunsOn) != 2 || g.Workflow.RunsOn[0] != "self-hosted" {
		t.Fatalf("got %+v", g.Workflow.RunsOn)
	}
	if got := WorkflowRunsOnLabels(g); len(got) != 2 {
		t.Fatalf("WorkflowRunsOnLabels: %v", got)
	}
}

func TestWorkflowRunsOnLabels_Default(t *testing.T) {
	t.Parallel()
	if got := WorkflowRunsOnLabels(nil); len(got) != 1 || got[0] != "ubuntu-latest" {
		t.Fatalf("got %v", got)
	}
	g := &GlobalConfig{Version: 1}
	if got := WorkflowRunsOnLabels(g); len(got) != 1 || got[0] != "ubuntu-latest" {
		t.Fatalf("got %v", got)
	}
}

func TestLoadGlobalOnly(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if g, err := LoadGlobalOnly(root); err != nil || g != nil {
		t.Fatalf("missing config: g=%v err=%v", g, err)
	}
	wm := filepath.Join(root, DirName)
	if err := os.MkdirAll(wm, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(wm, "config.yml")
	if err := os.WriteFile(cfg, []byte(`version: 1
workflow:
  runs_on:
    - custom
`), 0o644); err != nil {
		t.Fatal(err)
	}
	g, err := LoadGlobalOnly(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Workflow.RunsOn) != 1 || g.Workflow.RunsOn[0] != "custom" {
		t.Fatalf("got %+v", g.Workflow.RunsOn)
	}
}

func TestDefaultGlobal(t *testing.T) {
	t.Parallel()
	g := DefaultGlobal(nil)
	if g.Engine != "claude" || g.MaxTurns != 100 {
		t.Fatalf("defaults: %+v", g)
	}
	g2 := DefaultGlobal(&GlobalConfig{Version: 2})
	if g2.Version != 2 || g2.Engine != "claude" {
		t.Fatalf("preserve version: %+v", g2)
	}
}

func TestLoad(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	wm := filepath.Join(root, DirName)
	if err := os.MkdirAll(wm, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(wm, "config.yml")
	if err := os.WriteFile(cfg, []byte(`version: 1
engine: claude
max_turns: 10
`), 0o644); err != nil {
		t.Fatal(err)
	}
	tasksDir := filepath.Join(wm, TasksDir)
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	taskPath := filepath.Join(tasksDir, "hello.md")
	content := `---
name: hello
on:
  issues:
    types: [opened]
---

body
`
	if err := os.WriteFile(taskPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	g, tasks, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if g.MaxTurns != 10 || len(tasks) != 1 {
		t.Fatalf("g=%+v tasks=%d", g, len(tasks))
	}
	if tasks[0].Name != "hello" || tasks[0].Body != "body" {
		t.Fatalf("task: %+v", tasks[0])
	}
}

func BenchmarkConfigLoad(b *testing.B) {
	root := b.TempDir()
	wm := filepath.Join(root, DirName)
	if err := os.MkdirAll(filepath.Join(wm, TasksDir), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte("version: 1\nengine: claude\nmax_turns: 10\n"), 0o644); err != nil {
		b.Fatal(err)
	}
	task := `---
name: hello
on:
  issues:
    types: [opened]
---

body
`
	if err := os.WriteFile(filepath.Join(wm, TasksDir, "hello.md"), []byte(task), 0o644); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := Load(root); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseGlobal(b *testing.B) {
	data := []byte(`version: 1
engine: claude
max_turns: 100
workflow:
  runs_on: [ubuntu-latest]
`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ParseGlobal(data); err != nil {
			b.Fatal(err)
		}
	}
}

func TestLoad_MissingConfig(t *testing.T) {
	t.Parallel()
	if _, _, err := Load(t.TempDir()); err == nil {
		t.Fatal("expected error")
	}
}
