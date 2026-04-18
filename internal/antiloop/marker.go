package antiloop

import (
	"strings"

	"github.com/an-lee/gh-wm/internal/types"
)

// ShouldSkipAutomatedSender skips resolve for bot-originated events (defense when workflows run with PATs).
// schedule and workflow_dispatch are never skipped.
func ShouldSkipAutomatedSender(ev *types.GitHubEvent) bool {
	if ev == nil {
		return false
	}
	switch strings.TrimSpace(ev.Name) {
	case "schedule", "workflow_dispatch":
		return false
	case "":
		return false
	}
	p := ev.Payload
	if p == nil {
		return false
	}
	sender, ok := p["sender"].(map[string]any)
	if !ok {
		return false
	}
	if t, _ := sender["type"].(string); strings.EqualFold(strings.TrimSpace(t), "Bot") {
		return true
	}
	if login, _ := sender["login"].(string); strings.HasSuffix(strings.ToLower(strings.TrimSpace(login)), "[bot]") {
		return true
	}
	return false
}

// WMAgentCommentMarkerPrefix is embedded in WM Agent-authored issue/PR comments so resolve can skip re-entrancy loops.
const WMAgentCommentMarkerPrefix = "<!-- wm-agent:"

// WMAgentCommentMarkerFooter appends a hidden HTML marker so resolve can ignore wm-authored comments (loop guard).
func WMAgentCommentMarkerFooter(taskName string) string {
	t := strings.TrimSpace(taskName)
	if t == "" {
		t = "unknown"
	}
	return "\n\n" + WMAgentCommentMarkerPrefix + t + " -->"
}
