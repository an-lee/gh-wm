package config

import "testing"

func TestTaskTimeoutMinutes(t *testing.T) {
	t.Parallel()
	task := &Task{Frontmatter: map[string]any{"timeout-minutes": 15}}
	if got := task.TimeoutMinutes(45); got != 15 {
		t.Fatalf("timeout: got %d want 15", got)
	}
	if got := (&Task{}).TimeoutMinutes(45); got != 45 {
		t.Fatalf("default: got %d want 45", got)
	}
}

func TestTaskSafeOutputs(t *testing.T) {
	t.Parallel()
	task := &Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-comment":          map[string]any{},
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
