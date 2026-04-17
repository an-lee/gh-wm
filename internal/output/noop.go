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
