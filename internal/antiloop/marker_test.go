package antiloop

import (
	"testing"

	"github.com/an-lee/gh-wm/internal/types"
)

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

func TestShouldSkipAutomatedSender_NilPayload(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "issues"}
	if ShouldSkipAutomatedSender(ev) {
		t.Fatal("nil payload should return false")
	}
}

func TestShouldSkipAutomatedSender_ScheduleAndWorkflowDispatchNeverSkip(t *testing.T) {
	t.Parallel()
	tests := []string{"schedule", "workflow_dispatch"}
	for _, name := range tests {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ev := &types.GitHubEvent{
				Name:    name,
				Payload: map[string]any{"sender": map[string]any{"type": "Bot", "login": "github-actions[bot]"}},
			}
			if ShouldSkipAutomatedSender(ev) {
				t.Fatalf("%s event should never be skipped", name)
			}
		})
	}
}

func TestShouldSkipAutomatedSender_BotTypeHeuristic(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{
		Name:    "issue_comment",
		Payload: map[string]any{"sender": map[string]any{"type": " bot "}},
	}
	if !ShouldSkipAutomatedSender(ev) {
		t.Fatal("sender type Bot should be skipped")
	}
}

func TestShouldSkipAutomatedSender_BotLoginSuffixHeuristic(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{
		Name:    "pull_request",
		Payload: map[string]any{"sender": map[string]any{"login": "  CI-AUTOMATION[BOT]  "}},
	}
	if !ShouldSkipAutomatedSender(ev) {
		t.Fatal("sender login ending with [bot] should be skipped")
	}
}

func TestShouldSkipAutomatedSender_HumanSender(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{
		Name:    "issues",
		Payload: map[string]any{"sender": map[string]any{"type": "User", "login": "an-lee"}},
	}
	if ShouldSkipAutomatedSender(ev) {
		t.Fatal("human sender should not be skipped")
	}
}
