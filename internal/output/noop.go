package output

import (
	"log/slog"
	"strings"
)

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
