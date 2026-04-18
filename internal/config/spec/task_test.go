package spec

import (
	"strings"
	"testing"
)

func TestParseTaskFrontmatter_Basic(t *testing.T) {
	t.Parallel()
	fm := map[string]any{
		"description":     "hi",
		"engine":          "claude",
		"timeout-minutes": float64(30),
		"safe-outputs":    map[string]any{"noop": map[string]any{}},
		"on":              map[string]any{"reaction": "eyes"},
	}
	s, warns, err := ParseTaskFrontmatter(fm)
	if err != nil {
		t.Fatal(err)
	}
	if s.Description != "hi" || s.Engine != "claude" || s.TimeoutMinutes != 30 {
		t.Fatalf("spec: %+v", s)
	}
	if len(s.RawSafeOutputKeys) != 1 || s.RawSafeOutputKeys[0] != "noop" {
		t.Fatalf("keys: %v", s.RawSafeOutputKeys)
	}
	if s.OnReaction != "eyes" {
		t.Fatalf("on: %+v", s)
	}
	if len(warns) != 0 {
		t.Fatalf("unexpected warnings: %v", warns)
	}
}

func TestParseTaskFrontmatter_UnknownEngineWarning(t *testing.T) {
	t.Parallel()
	s, warns, err := ParseTaskFrontmatter(map[string]any{"engine": "custom-cli"})
	if err != nil {
		t.Fatal(err)
	}
	if s.Engine != "custom-cli" || len(warns) != 1 {
		t.Fatalf("engine=%q warns=%v", s.Engine, warns)
	}
}

func TestParseTaskFrontmatter_CopilotDeprecationWarning(t *testing.T) {
	t.Parallel()
	_, warns, err := ParseTaskFrontmatter(map[string]any{"engine": "copilot"})
	if err != nil {
		t.Fatal(err)
	}
	if len(warns) != 1 || !strings.Contains(warns[0], "deprecated") {
		t.Fatalf("warns=%v", warns)
	}
}

func TestParseTaskFrontmatter_SafeOutputsUnderscoreWarning(t *testing.T) {
	t.Parallel()
	_, warns, err := ParseTaskFrontmatter(map[string]any{
		"safe-outputs": map[string]any{"create_issue": map[string]any{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(warns) != 1 || !strings.Contains(warns[0], "underscore") {
		t.Fatalf("warns=%v", warns)
	}
}
