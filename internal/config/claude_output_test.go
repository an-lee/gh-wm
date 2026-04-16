package config

import (
	"os"
	"testing"
)

func TestEffectiveClaudeOutputFormat(t *testing.T) {
	t.Cleanup(func() { _ = os.Unsetenv("WM_CLAUDE_OUTPUT_FORMAT") })

	g := &GlobalConfig{ClaudeOutputFormat: "json"}
	if got := EffectiveClaudeOutputFormat(g); got != ClaudeOutputFormatJSON {
		t.Fatalf("config: got %q", got)
	}

	t.Setenv("WM_CLAUDE_OUTPUT_FORMAT", "stream-json")
	if got := EffectiveClaudeOutputFormat(g); got != ClaudeOutputFormatStreamJSON {
		t.Fatalf("env overrides config: got %q", got)
	}

	_ = os.Unsetenv("WM_CLAUDE_OUTPUT_FORMAT")
	if got := EffectiveClaudeOutputFormat(nil); got != ClaudeOutputFormatText {
		t.Fatalf("nil glob default: got %q", got)
	}
}
