package output

import (
	"log"
	"strings"
)

// runNoop logs a completion message (no GitHub API).
func runNoop(item ItemNoop) {
	msg := strings.TrimSpace(item.Message)
	if msg == "" {
		msg = "(noop: no message)"
	}
	log.Printf("wm: safe-output noop: %s", msg)
}
