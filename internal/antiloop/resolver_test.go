package antiloop

import (
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

// --- CollectStateLabelValues ---

func TestCollectStateLabelValues_NilTasks(t *testing.T) {
	t.Parallel()
	got := CollectStateLabelValues(nil)
	if len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}

func TestCollectStateLabelValues_NilInSlice(t *testing.T) {
	t.Parallel()
	got := CollectStateLabelValues([]*config.Task{nil})
	if len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}

func TestCollectStateLabelValues_NoWM(t *testing.T) {
	t.Parallel()
	task := &config.Task{Name: "test"}
	got := CollectStateLabelValues([]*config.Task{task})
	if len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}

func TestCollectStateLabelValues_WithStateLabels(t *testing.T) {
	t.Parallel()
	task := &config.Task{
		Name: "test",
		Frontmatter: map[string]any{
			"wm": map[string]any{
				"state_labels": map[string]any{
					"working": "agent:working",
					"done":    "agent:review",
					"failed":  "agent:failed",
				},
			},
		},
	}
	got := CollectStateLabelValues([]*config.Task{task})
	want := map[string]struct{}{
		"agent:working": {},
		"agent:review":  {},
		"agent:failed":  {},
	}
	for k, v := range want {
		if _, ok := got[k]; !ok {
			t.Errorf("missing label %q", k)
		}
		_, ok := got[k]
		if ok != (v == struct{}{}) {
			t.Errorf("label %q: got ok=%v, want ok=%v", k, ok, v == struct{}{})
		}
	}
	if len(got) != len(want) {
		t.Errorf("got %d labels, want %d", len(got), len(want))
	}
}

func TestCollectStateLabelValues_TrimsWhitespace(t *testing.T) {
	t.Parallel()
	task := &config.Task{
		Name: "test",
		Frontmatter: map[string]any{
			"wm": map[string]any{
				"state_labels": map[string]any{
					"working": "  agent:working  ",
				},
			},
		},
	}
	got := CollectStateLabelValues([]*config.Task{task})
	if _, ok := got["agent:working"]; !ok {
		t.Errorf("got %v, want agent:working present", got)
	}
	if _, ok := got["  agent:working  "]; ok {
		t.Errorf("got %v, should not contain untrimmed key", got)
	}
}

func TestCollectStateLabelValues_EmptyValueSkipped(t *testing.T) {
	t.Parallel()
	task := &config.Task{
		Name: "test",
		Frontmatter: map[string]any{
			"wm": map[string]any{
				"state_labels": map[string]any{
					"working": "",
					"done":    "  ",
				},
			},
		},
	}
	got := CollectStateLabelValues([]*config.Task{task})
	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

// --- ShouldSkipIssuesLabeledStateLabel ---

func TestShouldSkipIssuesLabeledStateLabel_NilEvent(t *testing.T) {
	t.Parallel()
	got := ShouldSkipIssuesLabeledStateLabel(nil, map[string]struct{}{"foo": {}})
	if got {
		t.Fatal("nil event should return false")
	}
}

func TestShouldSkipIssuesLabeledStateLabel_EmptyStateLabels(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"action": "labeled", "label": map[string]any{"name": "foo"}}}
	got := ShouldSkipIssuesLabeledStateLabel(ev, map[string]struct{}{})
	if got {
		t.Fatal("empty state labels should return false")
	}
}

func TestShouldSkipIssuesLabeledStateLabel_NonIssuesEvent(t *testing.T) {
	t.Parallel()
	stateLabels := map[string]struct{}{"agent:working": {}}
	ev := &types.GitHubEvent{Name: "pull_request", Payload: map[string]any{"action": "labeled"}}
	got := ShouldSkipIssuesLabeledStateLabel(ev, stateLabels)
	if got {
		t.Fatal("non-issues event should return false")
	}
}

func TestShouldSkipIssuesLabeledStateLabel_NonLabeledAction(t *testing.T) {
	t.Parallel()
	stateLabels := map[string]struct{}{"agent:working": {}}
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"action": "opened"}}
	got := ShouldSkipIssuesLabeledStateLabel(ev, stateLabels)
	if got {
		t.Fatal("non-labeled action should return false")
	}
}

func TestShouldSkipIssuesLabeledStateLabel_LabelNotInStateLabels(t *testing.T) {
	t.Parallel()
	stateLabels := map[string]struct{}{"agent:working": {}}
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"action": "labeled", "label": map[string]any{"name": "bug"}}}
	got := ShouldSkipIssuesLabeledStateLabel(ev, stateLabels)
	if got {
		t.Fatal("label not in state labels should return false")
	}
}

func TestShouldSkipIssuesLabeledStateLabel_LabelInStateLabels(t *testing.T) {
	t.Parallel()
	stateLabels := map[string]struct{}{"agent:working": {}}
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"action": "labeled", "label": map[string]any{"name": "agent:working"}}}
	got := ShouldSkipIssuesLabeledStateLabel(ev, stateLabels)
	if !got {
		t.Fatal("label in state labels should return true")
	}
}

func TestShouldSkipIssuesLabeledStateLabel_CaseSensitive(t *testing.T) {
	t.Parallel()
	stateLabels := map[string]struct{}{"Agent:Working": {}}
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"action": "labeled", "label": map[string]any{"name": "agent:working"}}}
	got := ShouldSkipIssuesLabeledStateLabel(ev, stateLabels)
	if got {
		t.Fatal("state label matching should be case-sensitive")
	}
}

func TestShouldSkipIssuesLabeledStateLabel_MissingLabelInPayload(t *testing.T) {
	t.Parallel()
	stateLabels := map[string]struct{}{"agent:working": {}}
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"action": "labeled"}}
	got := ShouldSkipIssuesLabeledStateLabel(ev, stateLabels)
	if got {
		t.Fatal("missing label in payload should return false")
	}
}

func TestShouldSkipIssuesLabeledStateLabel_NoSenderKey(t *testing.T) {
	t.Parallel()
	stateLabels := map[string]struct{}{"agent:working": {}}
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"action": "labeled", "label": map[string]any{"name": "agent:working"}}}
	got := ShouldSkipIssuesLabeledStateLabel(ev, stateLabels)
	if !got {
		t.Fatal("should return true when label is in state labels")
	}
}

// --- ShouldSkipAutomatedSender ---

func TestShouldSkipAutomatedSender_NilEvent(t *testing.T) {
	t.Parallel()
	got := ShouldSkipAutomatedSender(nil)
	if got {
		t.Fatal("nil event should return false")
	}
}

func TestShouldSkipAutomatedSender_ScheduleEvent(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "schedule", Payload: map[string]any{"sender": map[string]any{"type": "Bot"}}}
	got := ShouldSkipAutomatedSender(ev)
	if got {
		t.Fatal("schedule event should never be skipped")
	}
}

func TestShouldSkipAutomatedSender_WorkflowDispatchEvent(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "workflow_dispatch", Payload: map[string]any{"sender": map[string]any{"type": "Bot"}}}
	got := ShouldSkipAutomatedSender(ev)
	if got {
		t.Fatal("workflow_dispatch event should never be skipped")
	}
}

func TestShouldSkipAutomatedSender_EmptyName(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "", Payload: map[string]any{"sender": map[string]any{"type": "Bot"}}}
	got := ShouldSkipAutomatedSender(ev)
	if got {
		t.Fatal("empty name should return false")
	}
}

func TestShouldSkipAutomatedSender_BotType(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"sender": map[string]any{"type": "Bot"}}}
	got := ShouldSkipAutomatedSender(ev)
	if !got {
		t.Fatal("Bot type should be skipped")
	}
}

func TestShouldSkipAutomatedSender_BotTypeCaseInsensitive(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"sender": map[string]any{"type": "bot"}}}
	got := ShouldSkipAutomatedSender(ev)
	if !got {
		t.Fatal("bot type (lowercase) should be skipped")
	}
}

func TestShouldSkipAutomatedSender_BotLoginSuffix(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"sender": map[string]any{"login": "github-actions[bot]"}}}
	got := ShouldSkipAutomatedSender(ev)
	if !got {
		t.Fatal("login ending with [bot] should be skipped")
	}
}

func TestShouldSkipAutomatedSender_BotLoginSuffixCaseInsensitive(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"sender": map[string]any{"login": "CI-BOT[BOT]"}}}
	got := ShouldSkipAutomatedSender(ev)
	if !got {
		t.Fatal("login ending with [BOT] (uppercase) should be skipped")
	}
}

func TestShouldSkipAutomatedSender_HumanUser(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"sender": map[string]any{"type": "User", "login": "an-lee"}}}
	got := ShouldSkipAutomatedSender(ev)
	if got {
		t.Fatal("human user should not be skipped")
	}
}

func TestShouldSkipAutomatedSender_NilPayload(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "issues", Payload: nil}
	got := ShouldSkipAutomatedSender(ev)
	if got {
		t.Fatal("nil payload should return false")
	}
}

func TestShouldSkipAutomatedSender_NoSender(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{}}
	got := ShouldSkipAutomatedSender(ev)
	if got {
		t.Fatal("no sender should return false")
	}
}

func TestShouldSkipAutomatedSender_SenderNotMap(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"sender": "not-a-map"}}
	got := ShouldSkipAutomatedSender(ev)
	if got {
		t.Fatal("sender not a map should return false")
	}
}

func TestShouldSkipAutomatedSender_WhitespaceName(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "  schedule  ", Payload: nil}
	got := ShouldSkipAutomatedSender(ev)
	if got {
		t.Fatal("whitespace schedule name should return false")
	}
}

// --- labelNameFromIssuesPayload ---

func TestLabelNameFromIssuesPayload_NilPayload(t *testing.T) {
	t.Parallel()
	got := labelNameFromIssuesPayload(nil)
	if got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestLabelNameFromIssuesPayload_NoLabel(t *testing.T) {
	t.Parallel()
	got := labelNameFromIssuesPayload(map[string]any{})
	if got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestLabelNameFromIssuesPayload_LabelNotMap(t *testing.T) {
	t.Parallel()
	got := labelNameFromIssuesPayload(map[string]any{"label": "not-a-map"})
	if got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestLabelNameFromIssuesPayload_NoNameField(t *testing.T) {
	t.Parallel()
	got := labelNameFromIssuesPayload(map[string]any{"label": map[string]any{"color": "red"}})
	if got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestLabelNameFromIssuesPayload_NameWrongType(t *testing.T) {
	t.Parallel()
	got := labelNameFromIssuesPayload(map[string]any{"label": map[string]any{"name": 42}})
	if got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestLabelNameFromIssuesPayload_Valid(t *testing.T) {
	t.Parallel()
	got := labelNameFromIssuesPayload(map[string]any{"label": map[string]any{"name": "  agent:working  "}})
	if got != "agent:working" {
		t.Fatalf("got %q, want %q", got, "agent:working")
	}
}

func TestLabelNameFromIssuesPayload_NameNotTrimmed(t *testing.T) {
	t.Parallel()
	got := labelNameFromIssuesPayload(map[string]any{"label": map[string]any{"name": "  hello  "}})
	if got == "  hello  " {
		t.Fatal("should trim whitespace")
	}
	if got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}
