package output

import (
	"log/slog"
	"strings"
)

// normalizeNestedNoopItem accepts gh-aw-style `{"noop": {"message": "…"}}` without a top-level `type`
// and returns a map with `type` + `message` for the noop executor.
func normalizeNestedNoopItem(raw map[string]any) map[string]any {
	if raw == nil {
		return nil
	}
	if _, has := raw["type"]; has {
		return raw
	}
	noop, ok := raw["noop"].(map[string]any)
	if !ok {
		return raw
	}
	out := make(map[string]any, len(raw)+2)
	for k, v := range raw {
		out[k] = v
	}
	out["type"] = string(KindNoop)
	if msg, ok := noop["message"].(string); ok {
		out["message"] = msg
	}
	return out
}

// runNoop logs a completion message (no GitHub API).
func runNoop(item ItemNoop) {
	msg := strings.TrimSpace(item.Message)
	if msg == "" {
		msg = "(noop: no message)"
	}
	slog.Info("wm: safe-output noop", "message", msg)
}

func runMissingTool(item ItemMissingTool) {
	slog.Info("wm: safe-output missing_tool", "tool", strings.TrimSpace(item.Tool), "reason", strings.TrimSpace(item.Reason))
}

func runMissingData(item ItemMissingData) {
	slog.Info("wm: safe-output missing_data", "what", strings.TrimSpace(item.What), "reason", strings.TrimSpace(item.Reason))
}
