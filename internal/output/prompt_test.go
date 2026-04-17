package output

import (
	"strings"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
)

func TestSafeOutputsSystemPromptAppend(t *testing.T) {
	t.Parallel()
	if s := SafeOutputsSystemPromptAppend(nil); s != "" {
		t.Fatalf("nil task: want empty, got %q", s)
	}
	task := &config.Task{Name: "x"}
	if s := SafeOutputsSystemPromptAppend(task); s != "" {
		t.Fatalf("no safe-outputs: want empty, got %q", s)
	}
	task.Frontmatter = map[string]any{
		"safe-outputs": map[string]any{"noop": map[string]any{}},
	}
	s := SafeOutputsSystemPromptAppend(task)
	if s == "" {
		t.Fatal("expected non-empty append text")
	}
	for _, sub := range []string{"gh-wm emit", "WM_SAFE_OUTPUT_FILE", "read-only CI"} {
		if !strings.Contains(s, sub) {
			t.Fatalf("append text missing %q:\n%s", sub, s)
		}
	}
}
