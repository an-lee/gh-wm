package antiloop

import "strings"

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
