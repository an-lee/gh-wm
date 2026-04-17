package antiloop

import "testing"

func TestWMAgentCommentMarkerFooter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		task string
		want string
	}{
		{"empty", "", "\n\n<!-- wm-agent:unknown -->"},
		{"whitespace", "   ", "\n\n<!-- wm-agent:unknown -->"},
		{"normal", "implement", "\n\n<!-- wm-agent:implement -->"},
		{"with-spaces", "  implement  ", "\n\n<!-- wm-agent:implement -->"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := WMAgentCommentMarkerFooter(tt.task)
			if got != tt.want {
				t.Errorf("WMAgentCommentMarkerFooter(%q) = %q, want %q", tt.task, got, tt.want)
			}
		})
	}
}

func TestWMAgentCommentMarkerPrefix(t *testing.T) {
	t.Parallel()
	if WMAgentCommentMarkerPrefix != "<!-- wm-agent:" {
		t.Errorf("WMAgentCommentMarkerPrefix = %q, want %q", WMAgentCommentMarkerPrefix, "<!-- wm-agent:")
	}
}
