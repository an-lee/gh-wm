package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

// ClaudeConversationStats aggregates parsed Claude Code print-mode output (json / stream-json).
type ClaudeConversationStats struct {
	EventCount   int
	EventsByType map[string]int
	ParseErrors  int
	LastResult   *claudeResultSnapshot
}

type claudeResultSnapshot struct {
	Subtype              string
	TotalCostUSD         float64
	HasCost              bool
	NumTurns             int
	HasNumTurns          bool
	DurationMs           int
	HasDurationMs        bool
	SessionID            string
	InputTokens          int
	OutputTokens         int
	CacheReadInputTokens int
	HasUsage             bool
	ModelIDs             []string
}

func pickFloatM(m map[string]any, keys ...string) (float64, bool) {
	for _, k := range keys {
		switch x := m[k].(type) {
		case float64:
			return x, true
		case int:
			return float64(x), true
		case int64:
			return float64(x), true
		}
	}
	return 0, false
}

func pickIntM(m map[string]any, keys ...string) (int, bool) {
	for _, k := range keys {
		switch x := m[k].(type) {
		case float64:
			return int(x), true
		case int:
			return x, true
		case int64:
			return int(x), true
		}
	}
	return 0, false
}

func extractResultSnapshot(ev map[string]any) *claudeResultSnapshot {
	if ev == nil {
		return nil
	}
	r := &claudeResultSnapshot{}
	r.Subtype, _ = ev["subtype"].(string)
	if v, ok := pickFloatM(ev, "total_cost_usd", "cost_usd", "cost"); ok {
		r.TotalCostUSD = v
		r.HasCost = true
	}
	if n, ok := pickIntM(ev, "num_turns", "turns"); ok {
		r.NumTurns = n
		r.HasNumTurns = true
	}
	if n, ok := pickIntM(ev, "duration_ms"); ok {
		r.DurationMs = n
		r.HasDurationMs = true
	}
	r.SessionID, _ = ev["session_id"].(string)
	if u, ok := ev["usage"].(map[string]any); ok {
		if n, ok := pickIntM(u, "input_tokens"); ok {
			r.InputTokens = n
			r.HasUsage = true
		}
		if n, ok := pickIntM(u, "output_tokens"); ok {
			r.OutputTokens = n
			r.HasUsage = true
		}
		if n, ok := pickIntM(u, "cache_read_input_tokens"); ok {
			r.CacheReadInputTokens = n
			r.HasUsage = true
		}
	}
	if mu, ok := ev["modelUsage"].(map[string]any); ok {
		for k := range mu {
			r.ModelIDs = append(r.ModelIDs, k)
		}
		sort.Strings(r.ModelIDs)
	}
	return r
}

func parseClaudeConversationJSONL(b []byte) *ClaudeConversationStats {
	stats := &ClaudeConversationStats{EventsByType: make(map[string]int)}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var ev map[string]any
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			stats.ParseErrors++
			continue
		}
		stats.EventCount++
		t, _ := ev["type"].(string)
		if t != "" {
			stats.EventsByType[t]++
		}
		if t == "result" {
			stats.LastResult = extractResultSnapshot(ev)
		}
	}
	return stats
}

func parseClaudeConversationJSON(b []byte) (*ClaudeConversationStats, error) {
	var ev map[string]any
	if err := json.Unmarshal(b, &ev); err != nil {
		return nil, err
	}
	stats := &ClaudeConversationStats{EventsByType: make(map[string]int)}
	stats.EventCount = 1
	t, _ := ev["type"].(string)
	if t != "" {
		stats.EventsByType[t] = 1
	}
	if t == "result" {
		stats.LastResult = extractResultSnapshot(ev)
	}
	return stats, nil
}

// parseClaudeConversationArtifacts reads conversation.jsonl (preferred) or conversation.json from runDir.
// Returns (nil, nil) when no non-empty artifact exists.
func parseClaudeConversationArtifacts(runDir string) (*ClaudeConversationStats, error) {
	jsonl := filepath.Join(runDir, conversationJSONLFileName)
	if st, err := os.Stat(jsonl); err == nil && st.Size() > 0 {
		b, err := os.ReadFile(jsonl)
		if err != nil {
			return nil, err
		}
		return parseClaudeConversationJSONL(b), nil
	}
	js := filepath.Join(runDir, conversationJSONFileName)
	if st, err := os.Stat(js); err == nil && st.Size() > 0 {
		b, err := os.ReadFile(js)
		if err != nil {
			return nil, err
		}
		return parseClaudeConversationJSON(b)
	}
	return nil, nil
}

func formatModelsLine(stats *ClaudeConversationStats, modelFromConfig string) string {
	if stats != nil && stats.LastResult != nil && len(stats.LastResult.ModelIDs) > 0 {
		return strings.Join(stats.LastResult.ModelIDs, ", ")
	}
	if strings.TrimSpace(modelFromConfig) != "" {
		return strings.TrimSpace(modelFromConfig) + " (config)"
	}
	return "—"
}

func formatEventTypesLine(stats *ClaudeConversationStats) string {
	if stats == nil || len(stats.EventsByType) == 0 {
		return "—"
	}
	keys := make([]string, 0, len(stats.EventsByType))
	for k := range stats.EventsByType {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s: %d", k, stats.EventsByType[k]))
	}
	return strings.Join(parts, ", ")
}

// markdownClaudeStepSummary returns markdown for GITHUB_STEP_SUMMARY (no trailing newline required).
func markdownClaudeStepSummary(taskName string, result *types.RunResult, glob *config.GlobalConfig, stats *ClaudeConversationStats) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## gh-wm · Claude run · task `%s`\n\n", taskName)
	fmt.Fprintf(&b, "| Metric | Value |\n| --- | --- |\n")
	success := "—"
	if result != nil {
		if result.Success {
			success = "yes"
		} else {
			success = "no"
		}
	}
	fmt.Fprintf(&b, "| Pipeline success | %s |\n", success)
	if result != nil {
		fmt.Fprintf(&b, "| Wall-clock (gh-wm) | %s |\n", result.Duration.Round(time.Millisecond).String())
	}
	if stats != nil {
		fmt.Fprintf(&b, "| Stream events (lines) | %d |\n", stats.EventCount)
		if stats.ParseErrors > 0 {
			fmt.Fprintf(&b, "| JSONL parse errors | %d |\n", stats.ParseErrors)
		}
		fmt.Fprintf(&b, "| Event types | %s |\n", formatEventTypesLine(stats))
	}
	modelCfg := ""
	if glob != nil {
		modelCfg = glob.Model
	}
	fmt.Fprintf(&b, "| Model(s) | %s |\n", formatModelsLine(stats, modelCfg))
	if stats != nil && stats.LastResult != nil {
		lr := stats.LastResult
		if lr.Subtype != "" {
			fmt.Fprintf(&b, "| Result subtype | %s |\n", lr.Subtype)
		}
		if lr.HasCost {
			fmt.Fprintf(&b, "| Total cost (USD) | $%.6f |\n", lr.TotalCostUSD)
		}
		if lr.HasNumTurns {
			fmt.Fprintf(&b, "| Turns | %d |\n", lr.NumTurns)
		}
		if lr.HasDurationMs {
			fmt.Fprintf(&b, "| Agent duration_ms | %d |\n", lr.DurationMs)
		}
		if lr.SessionID != "" {
			fmt.Fprintf(&b, "| Session ID | %s |\n", lr.SessionID)
		}
		if lr.HasUsage {
			fmt.Fprintf(&b, "| Input tokens | %d |\n", lr.InputTokens)
			fmt.Fprintf(&b, "| Output tokens | %d |\n", lr.OutputTokens)
			if lr.CacheReadInputTokens > 0 {
				fmt.Fprintf(&b, "| Cache read input tokens | %d |\n", lr.CacheReadInputTokens)
			}
		}
	}
	return b.String()
}

// appendClaudeGitHubStepSummary appends Claude usage stats to GITHUB_STEP_SUMMARY when set (e.g. GitHub Actions).
// Best-effort: ignores I/O errors and missing artifacts.
func appendClaudeGitHubStepSummary(result *types.RunResult, a *concludeArgs) {
	if result == nil || a == nil || a.rd == nil || a.task == nil {
		return
	}
	path := strings.TrimSpace(os.Getenv("GITHUB_STEP_SUMMARY"))
	if path == "" {
		return
	}
	if strings.TrimSpace(os.Getenv("WM_AGENT_CMD")) != "" {
		return
	}
	engineName := strings.TrimSpace(a.task.Engine())
	if engineName == "" && a.glob != nil {
		engineName = a.glob.Engine
	}
	if !isBuiltinClaude("", engineName) {
		return
	}
	stats, err := parseClaudeConversationArtifacts(a.rd.Path)
	if err != nil || stats == nil {
		return
	}
	if stats.EventCount == 0 && stats.LastResult == nil && stats.ParseErrors == 0 {
		return
	}
	md := markdownClaudeStepSummary(a.task.Name, result, a.glob, stats)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(f, "\n%s\n", md)
	_ = f.Close()
}
