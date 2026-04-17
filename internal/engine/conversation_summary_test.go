package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

func TestAppendClaudeGitHubStepSummary_WritesFile(t *testing.T) {
	dir := t.TempDir()
	summaryFile := filepath.Join(t.TempDir(), "summary.md")
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)
	t.Setenv("WM_AGENT_CMD", "")
	if err := os.WriteFile(filepath.Join(dir, conversationJSONLFileName), []byte(`{"type":"result","subtype":"success","total_cost_usd":0.1}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result := &types.RunResult{Success: true, Duration: time.Second}
	a := &concludeArgs{
		task: &config.Task{Name: "hello"},
		glob: &config.GlobalConfig{},
		rd:   &RunDir{Path: dir},
	}
	appendClaudeGitHubStepSummary(result, a)
	b, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "hello") || !strings.Contains(string(b), "0.100000") {
		t.Fatalf("summary file: %s", b)
	}
}

// Regression: concludeArgs must pass glob from RunTask so the Models row is not always "—".
func TestAppendClaudeGitHubStepSummary_IncludesGlobModelWhenNoModelUsage(t *testing.T) {
	dir := t.TempDir()
	summaryFile := filepath.Join(t.TempDir(), "step-summary.md")
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)
	t.Setenv("WM_AGENT_CMD", "")
	if err := os.WriteFile(filepath.Join(dir, conversationJSONLFileName), []byte(`{"type":"result","subtype":"success","total_cost_usd":0.01}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result := &types.RunResult{Success: true, Duration: time.Second}
	a := &concludeArgs{
		task: &config.Task{Name: "t"},
		glob: &config.GlobalConfig{Model: "haiku-from-config"},
		rd:   &RunDir{Path: dir},
	}
	appendClaudeGitHubStepSummary(result, a)
	b, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "haiku-from-config") {
		t.Fatalf("expected glob.Model in step summary; got:\n%s", b)
	}
}

func TestParseClaudeConversationJSONL_ResultUsageModelUsage(t *testing.T) {
	t.Parallel()
	lines := []string{
		`{"type":"system","subtype":"init"}`,
		`{"type":"assistant","message":{"content":[]}}`,
		`{"type":"result","subtype":"success","total_cost_usd":0.0079825,"num_turns":3,"duration_ms":1998,"session_id":"abc-123","usage":{"input_tokens":3,"output_tokens":50,"cache_read_input_tokens":15635},"modelUsage":{"claude-opus-4-6":{"inputTokens":3,"outputTokens":50,"costUSD":0.0079825}}}`,
	}
	b := strings.Join(lines, "\n")
	stats := parseClaudeConversationJSONL([]byte(b))
	if stats.EventCount != 3 {
		t.Fatalf("EventCount: got %d want 3", stats.EventCount)
	}
	if stats.EventsByType["system"] != 1 || stats.EventsByType["assistant"] != 1 || stats.EventsByType["result"] != 1 {
		t.Fatalf("EventsByType: %+v", stats.EventsByType)
	}
	if stats.LastResult == nil {
		t.Fatal("expected LastResult")
	}
	lr := stats.LastResult
	if lr.Subtype != "success" || !lr.HasCost || lr.TotalCostUSD < 0.0079 {
		t.Fatalf("result fields: %+v", lr)
	}
	if !lr.HasNumTurns || lr.NumTurns != 3 {
		t.Fatalf("num_turns: %+v", lr)
	}
	if !lr.HasDurationMs || lr.DurationMs != 1998 {
		t.Fatalf("duration_ms: %+v", lr)
	}
	if lr.SessionID != "abc-123" {
		t.Fatalf("session_id: %q", lr.SessionID)
	}
	if !lr.HasUsage || lr.InputTokens != 3 || lr.OutputTokens != 50 || lr.CacheReadInputTokens != 15635 {
		t.Fatalf("usage: %+v", lr)
	}
	if len(lr.ModelIDs) != 1 || lr.ModelIDs[0] != "claude-opus-4-6" {
		t.Fatalf("modelUsage keys: %+v", lr.ModelIDs)
	}
}

func TestParseClaudeConversationJSONL_InvalidLineSkips(t *testing.T) {
	t.Parallel()
	b := "not json\n" + `{"type":"result","subtype":"success","total_cost_usd":0.01}` + "\n"
	stats := parseClaudeConversationJSONL([]byte(b))
	if stats.ParseErrors != 1 {
		t.Fatalf("ParseErrors: got %d want 1", stats.ParseErrors)
	}
	if stats.EventCount != 1 {
		t.Fatalf("EventCount: got %d", stats.EventCount)
	}
	if stats.LastResult == nil || !stats.LastResult.HasCost {
		t.Fatal("expected last result")
	}
}

func TestParseClaudeConversationJSON_SingleEnvelope(t *testing.T) {
	t.Parallel()
	b := `{"type":"result","subtype":"success","num_turns":1,"total_cost_usd":0.001,"usage":{"input_tokens":10,"output_tokens":20}}`
	stats, err := parseClaudeConversationJSON([]byte(b))
	if err != nil {
		t.Fatal(err)
	}
	if stats.EventCount != 1 {
		t.Fatalf("EventCount: %d", stats.EventCount)
	}
	if stats.LastResult == nil || stats.LastResult.InputTokens != 10 || stats.LastResult.OutputTokens != 20 {
		t.Fatalf("LastResult: %+v", stats.LastResult)
	}
}

func TestParseClaudeConversationArtifacts_PrefersJSONL(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, conversationJSONLFileName), []byte(`{"type":"result","subtype":"success","total_cost_usd":0.05}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, conversationJSONFileName), []byte(`{"type":"result","total_cost_usd":0.99}`), 0o644); err != nil {
		t.Fatal(err)
	}
	stats, err := parseClaudeConversationArtifacts(dir)
	if err != nil {
		t.Fatal(err)
	}
	if stats == nil || stats.LastResult == nil || stats.LastResult.TotalCostUSD != 0.05 {
		t.Fatalf("expected jsonl parsed: %+v", stats.LastResult)
	}
}

func TestMarkdownClaudeStepSummary(t *testing.T) {
	t.Parallel()
	stats := parseClaudeConversationJSONL([]byte(`{"type":"result","subtype":"success","total_cost_usd":0.042,"num_turns":12,"usage":{"input_tokens":100,"output_tokens":200}}`))
	res := &types.RunResult{Success: true, Duration: 5 * time.Second}
	md := markdownClaudeStepSummary("my-task", res, &config.GlobalConfig{Model: "sonnet"}, stats)
	if !strings.Contains(md, "my-task") || !strings.Contains(md, "Pipeline success") {
		t.Fatalf("markdown: %s", md)
	}
	if !strings.Contains(md, "Wall-clock") || !strings.Contains(md, "5s") {
		t.Fatalf("expected duration: %s", md)
	}
	if !strings.Contains(md, "0.042000") || !strings.Contains(md, "12") {
		t.Fatalf("expected cost/turns: %s", md)
	}
}

func TestFormatModelsLine_FromModelUsage(t *testing.T) {
	t.Parallel()
	stats := &ClaudeConversationStats{
		LastResult: &claudeResultSnapshot{ModelIDs: []string{"m-a", "m-b"}},
	}
	s := formatModelsLine(stats, "ignored")
	if s != "m-a, m-b" {
		t.Fatalf("got %q", s)
	}
}

func TestFormatModelsLine_ConfigFallback(t *testing.T) {
	t.Parallel()
	stats := &ClaudeConversationStats{}
	s := formatModelsLine(stats, "haiku")
	if !strings.Contains(s, "haiku") || !strings.Contains(s, "config") {
		t.Fatalf("got %q", s)
	}
}
