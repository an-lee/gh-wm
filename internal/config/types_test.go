package config

import (
	"testing"
)

func TestTaskTimeoutMinutes(t *testing.T) {
	t.Parallel()
	task := &Task{Frontmatter: map[string]any{"timeout-minutes": 15}}
	if got := task.TimeoutMinutes(45); got != 15 {
		t.Fatalf("timeout: got %d want 15", got)
	}
	if got := (&Task{}).TimeoutMinutes(45); got != 45 {
		t.Fatalf("default: got %d want 45", got)
	}
	if got := (&Task{Frontmatter: map[string]any{"timeout-minutes": 0}}).TimeoutMinutes(45); got != 45 {
		t.Fatal("zero uses default")
	}
	if got := (&Task{Frontmatter: map[string]any{"timeout-minutes": 99999}}).TimeoutMinutes(45); got != maxTimeoutMinutes {
		t.Fatalf("capped: got %d", got)
	}
	for _, v := range []any{float64(20), int64(20)} {
		if got := (&Task{Frontmatter: map[string]any{"timeout-minutes": v}}).TimeoutMinutes(45); got != 20 {
			t.Fatalf("type %T: got %d", v, got)
		}
	}
}

func TestTaskSafeOutputs(t *testing.T) {
	t.Parallel()
	task := &Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-comment":         map[string]any{},
			"create-pull-request": map[string]any{"draft": true},
		},
	}}
	if !task.HasSafeOutputKey("add-comment") {
		t.Fatal("expected add-comment")
	}
	if !task.HasSafeOutputKey("create-pull-request") {
		t.Fatal("expected create-pull-request")
	}
}

func TestTaskOnMap(t *testing.T) {
	t.Parallel()
	if (&Task{}).OnMap() != nil {
		t.Fatal("nil frontmatter")
	}
	tk := &Task{Frontmatter: map[string]any{"on": map[string]any{"issues": map[string]any{}}}}
	if tk.OnMap() == nil {
		t.Fatal("expected on map")
	}
}

func TestTaskEngineScheduleString(t *testing.T) {
	t.Parallel()
	if (&Task{Frontmatter: map[string]any{"engine": "codex"}}).Engine() != "codex" {
		t.Fatal("engine")
	}
	if (&Task{Frontmatter: map[string]any{"on": map[string]any{"schedule": "daily"}}}).ScheduleString() != "daily" {
		t.Fatal("schedule string")
	}
	if (&Task{Frontmatter: map[string]any{"on": map[string]any{"schedule": 42}}}).ScheduleString() != "" {
		t.Fatal("non-string schedule")
	}
}

func TestTaskToolsYAML(t *testing.T) {
	t.Parallel()
	if got := (&Task{Frontmatter: map[string]any{"tools": "x"}}).ToolsYAML(); got != "x" {
		t.Fatal(got)
	}
	if got := (&Task{Frontmatter: map[string]any{"tools": []any{"a"}}}).ToolsYAML(); got == "" {
		t.Fatal("expected json marshaled tools")
	}
}
