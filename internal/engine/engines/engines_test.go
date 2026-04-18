package engines

import (
	"strings"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
)

// --- IsBuiltinClaude ---

func TestIsBuiltinClaude_EmptyWMAgentCmd(t *testing.T) {
	t.Parallel()
	tests := []struct {
		engineName string
		want       bool
	}{
		{"claude", true},
		{"CLAUDE", true},
		{"", true},
		{"codex", false},
		{"CODEX", false},
	}
	for _, tt := range tests {
		got := IsBuiltinClaude("", tt.engineName)
		if got != tt.want {
			t.Errorf("IsBuiltinClaude(%q, %q) = %v, want %v", "", tt.engineName, got, tt.want)
		}
	}
}

func TestIsBuiltinClaude_WithWMAgentCmd(t *testing.T) {
	t.Parallel()
	tests := []struct {
		engineName string
		want       bool
	}{
		{"claude", false},
		{"", false},
		{"codex", false},
	}
	for _, tt := range tests {
		got := IsBuiltinClaude("claude -p", tt.engineName)
		if got != tt.want {
			t.Errorf("IsBuiltinClaude(%q, %q) = %v, want %v", "claude -p", tt.engineName, got, tt.want)
		}
	}
}

func TestIsBuiltinClaude_WhitespaceTrimmed(t *testing.T) {
	t.Parallel()
	// WM_AGENT_CMD with only whitespace should be treated as empty
	got := IsBuiltinClaude("   ", "claude")
	if !got {
		t.Error("whitespace-only WM_AGENT_CMD should be treated as empty")
	}
}

// --- AgentOutputFormatForRun ---

func TestAgentOutputFormatForRun_WithWMAgentCmd(t *testing.T) {
	t.Parallel()
	got := AgentOutputFormatForRun("claude -p", "", nil)
	if got != config.ClaudeOutputFormatText {
		t.Errorf("got %q, want %q", got, config.ClaudeOutputFormatText)
	}
}

func TestAgentOutputFormatForRun_Codex(t *testing.T) {
	t.Parallel()
	got := AgentOutputFormatForRun("", "codex", nil)
	if got != config.ClaudeOutputFormatText {
		t.Errorf("got %q, want %q", got, config.ClaudeOutputFormatText)
	}
}

func TestAgentOutputFormatForRun_CodexCaseInsensitive(t *testing.T) {
	t.Parallel()
	got := AgentOutputFormatForRun("", "CODEX", nil)
	if got != config.ClaudeOutputFormatText {
		t.Errorf("got %q, want %q", got, config.ClaudeOutputFormatText)
	}
}

func TestAgentOutputFormatForRun_BuiltinClaude(t *testing.T) {
	t.Parallel()
	glob := &config.GlobalConfig{ClaudeOutputFormat: config.ClaudeOutputFormatJSON}
	got := AgentOutputFormatForRun("", "claude", glob)
	if got != config.ClaudeOutputFormatJSON {
		t.Errorf("got %q, want %q", got, config.ClaudeOutputFormatJSON)
	}
}

func TestAgentOutputFormatForRun_BuiltinClaudeDefault(t *testing.T) {
	t.Parallel()
	got := AgentOutputFormatForRun("", "claude", nil)
	if got != config.ClaudeOutputFormatText {
		t.Errorf("got %q, want %q", got, config.ClaudeOutputFormatText)
	}
}

// --- AgentCLIArgs ---

func TestAgentCLIArgs_Basic(t *testing.T) {
	t.Parallel()
	glob := &config.GlobalConfig{}
	args := AgentCLIArgs(glob, config.ClaudeOutputFormatText)
	if len(args) < 2 || args[0] != "-p" || args[1] != "--dangerously-skip-permissions" {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestAgentCLIArgs_WithModel(t *testing.T) {
	t.Parallel()
	glob := &config.GlobalConfig{Model: "opus-4"}
	args := AgentCLIArgs(glob, config.ClaudeOutputFormatText)
	found := false
	for _, a := range args {
		if a == "--model" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected --model in args %v", args)
	}
}

func TestAgentCLIArgs_WithMaxTurns(t *testing.T) {
	t.Parallel()
	glob := &config.GlobalConfig{MaxTurns: 10}
	args := AgentCLIArgs(glob, config.ClaudeOutputFormatText)
	found := false
	for _, a := range args {
		if a == "--max-turns" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected --max-turns in args %v", args)
	}
}

func TestAgentCLIArgs_NilGlob(t *testing.T) {
	t.Parallel()
	args := AgentCLIArgs(nil, config.ClaudeOutputFormatText)
	if len(args) < 2 {
		t.Errorf("expected at least -p and --dangerously-skip-permissions, got %v", args)
	}
}

func TestAgentCLIArgs_JSONFormat(t *testing.T) {
	t.Parallel()
	glob := &config.GlobalConfig{}
	args := AgentCLIArgs(glob, config.ClaudeOutputFormatJSON)
	found := false
	for _, a := range args {
		if a == "--output-format" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected --output-format in JSON args %v", args)
	}
}

func TestAgentCLIArgs_StreamJSONFormat(t *testing.T) {
	t.Parallel()
	glob := &config.GlobalConfig{}
	args := AgentCLIArgs(glob, config.ClaudeOutputFormatStreamJSON)
	hasOutputFormat := false
	hasVerbose := false
	for i, a := range args {
		if a == "--output-format" && i+1 < len(args) && args[i+1] == config.ClaudeOutputFormatStreamJSON {
			hasOutputFormat = true
		}
		if a == "--verbose" {
			hasVerbose = true
		}
	}
	if !hasOutputFormat {
		t.Errorf("expected --output-format stream-json in args %v", args)
	}
	if !hasVerbose {
		t.Errorf("expected --verbose for stream-json in args %v", args)
	}
}

// --- ParseWMAgentCmd ---

func TestParseWMAgentCmd_NoPlaceholder(t *testing.T) {
	t.Parallel()
	name, args, stdin, err := ParseWMAgentCmd("claude -p --dangerously-skip-permissions", "my-prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "claude" {
		t.Errorf("name = %q, want %q", name, "claude")
	}
	if len(args) < 2 {
		t.Errorf("expected at least 2 args, got %v", args)
	}
	if stdin != nil {
		t.Error("stdin should be nil without placeholder")
	}
}

func TestParseWMAgentCmd_WithPlaceholder(t *testing.T) {
	t.Parallel()
	name, args, _, err := ParseWMAgentCmd("claude -p --dangerously-skip-permissions {prompt}", "my-prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "claude" {
		t.Errorf("name = %q, want %q", name, "claude")
	}
	// Should contain: -p, --dangerously-skip-permissions, prompt, (any trailing)
	promptSeen := false
	for _, a := range args {
		if a == "my-prompt" {
			promptSeen = true
			break
		}
	}
	if !promptSeen {
		t.Errorf("expected prompt in args %v", args)
	}
}

func TestParseWMAgentCmd_PlaceholderInMiddle(t *testing.T) {
	t.Parallel()
	name, args, _, err := ParseWMAgentCmd("claude {prompt} --flag", "my-prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "claude" {
		t.Errorf("name = %q, want %q", name, "claude")
	}
	// args should be: ["--flag", "my-prompt"]
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %v", args)
	}
	if args[0] != "my-prompt" || args[1] != "--flag" {
		t.Errorf("args = %v, want [my-prompt, --flag]", args)
	}
}

func TestParseWMAgentCmd_EmptyCmd(t *testing.T) {
	t.Parallel()
	_, _, _, err := ParseWMAgentCmd("", "my-prompt")
	if err == nil {
		t.Fatal("expected error for empty cmd")
	}
}

func TestParseWMAgentCmd_PlaceholderAtStart(t *testing.T) {
	t.Parallel()
	// {prompt} at start means no executable before it → error
	_, _, _, err := ParseWMAgentCmd("{prompt}", "my-prompt")
	if err == nil {
		t.Fatal("expected error when {prompt} is at start (no executable)")
	}
}

func TestParseWMAgentCmd_WhitespaceFields(t *testing.T) {
	t.Parallel()
	name, args, _, err := ParseWMAgentCmd("  claude  -p  {prompt}  --flag  ", "my-prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "claude" {
		t.Errorf("name = %q, want %q", name, "claude")
	}
	if args[0] != "-p" || args[len(args)-1] != "--flag" {
		t.Errorf("args = %v, unexpected order", args)
	}
}

// --- ResolveEngine ---

func TestResolveEngine_EmptyWMAgentCmd(t *testing.T) {
	t.Parallel()
	tests := []struct {
		engineName string
		wantID     string
	}{
		{"claude", "claude"},
		{"", "claude"},
		{"CODEX", "codex"},
		{"codex", "codex"},
	}
	for _, tt := range tests {
		eng, err := ResolveEngine(tt.engineName, "")
		if err != nil {
			t.Errorf("ResolveEngine(%q, %q) unexpected error: %v", tt.engineName, "", err)
			continue
		}
		if eng.ID() != tt.wantID {
			t.Errorf("ResolveEngine(%q, %q).ID() = %q, want %q", tt.engineName, "", eng.ID(), tt.wantID)
		}
	}
}

func TestResolveEngine_WithWMAgentCmd(t *testing.T) {
	t.Parallel()
	eng, err := ResolveEngine("claude", "claude -p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eng.ID() != "custom" {
		t.Errorf("ID = %q, want %q", eng.ID(), "custom")
	}
}

func TestResolveEngine_CopilotDeprecated(t *testing.T) {
	t.Parallel()
	_, err := ResolveEngine("copilot", "")
	if err == nil {
		t.Fatal("expected error for copilot")
	}
	if !strings.Contains(err.Error(), "no longer supported") {
		t.Errorf("error should mention removal: %v", err)
	}
}

func TestResolveEngine_UnknownEngineDefaultsToClaude(t *testing.T) {
	t.Parallel()
	eng, err := ResolveEngine("unknown-engine", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eng.ID() != "claude" {
		t.Errorf("ID = %q, want %q", eng.ID(), "claude")
	}
}
