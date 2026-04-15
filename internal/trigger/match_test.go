package trigger

import (
	"testing"

	"github.com/an-lee/gh-wm/internal/types"
)

func TestMatchOnOR_Nil(t *testing.T) {
	t.Parallel()
	if MatchOnOR(nil, map[string]any{"issues": map[string]any{}}) {
		t.Fatal("expected false for nil event")
	}
	if MatchOnOR(&types.GitHubEvent{Name: "issues", Payload: map[string]any{}}, nil) {
		t.Fatal("expected false for nil on")
	}
}

func TestMatchOnOR_Issues(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{
		Name:    "issues",
		Payload: map[string]any{"action": "opened"},
	}
	on := map[string]any{"issues": map[string]any{"types": []any{"opened"}}}
	if !MatchOnOR(ev, on) {
		t.Fatal("expected match")
	}
	ev.Payload["action"] = "closed"
	if MatchOnOR(ev, on) {
		t.Fatal("expected no match")
	}
	// empty types = match any action
	ev.Payload["action"] = "labeled"
	on2 := map[string]any{"issues": map[string]any{}}
	if !MatchOnOR(ev, on2) {
		t.Fatal("empty types should match")
	}
}

func TestMatchOnOR_IssueComment(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{
		Name:    "issue_comment",
		Payload: map[string]any{"action": "created"},
	}
	on := map[string]any{"issue_comment": map[string]any{"types": []any{"created"}}}
	if !MatchOnOR(ev, on) {
		t.Fatal("expected match")
	}
	ev.Payload["action"] = "deleted"
	if MatchOnOR(ev, on) {
		t.Fatal("expected no match when types filter excludes")
	}
	// no types = match
	ev.Payload["action"] = "created"
	on2 := map[string]any{"issue_comment": map[string]any{}}
	if !MatchOnOR(ev, on2) {
		t.Fatal("expected match without types")
	}
}

func TestMatchOnOR_PullRequest(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"pull_request", "pull_request_target"} {
		ev := &types.GitHubEvent{
			Name:    name,
			Payload: map[string]any{"action": "opened"},
		}
		on := map[string]any{"pull_request": map[string]any{"types": []any{"opened", "synchronize"}}}
		if !MatchOnOR(ev, on) {
			t.Fatalf("%s: expected match", name)
		}
	}
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{}}
	on := map[string]any{"pull_request": map[string]any{}}
	if MatchOnOR(ev, on) {
		t.Fatal("wrong event name")
	}
}

func TestMatchOnOR_SlashCommand(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{
		Name: "issue_comment",
		Payload: map[string]any{
			"comment": map[string]any{"body": "/deploy prod"},
		},
	}
	on := map[string]any{"slash_command": map[string]any{"name": "deploy"}}
	if !MatchOnOR(ev, on) {
		t.Fatal("expected match")
	}
	ev.Payload["comment"] = map[string]any{"body": "/deploy  extra"}
	if !MatchOnOR(ev, on) {
		t.Fatal("expected match with space after command")
	}
	ev.Payload["comment"] = map[string]any{"body": "nope"}
	if MatchOnOR(ev, on) {
		t.Fatal("expected no match")
	}
	// wrong event type
	ev2 := &types.GitHubEvent{Name: "issues", Payload: ev.Payload}
	if MatchOnOR(ev2, on) {
		t.Fatal("slash_command only on issue_comment")
	}
	// empty name
	onBad := map[string]any{"slash_command": map[string]any{"name": ""}}
	if MatchOnOR(ev, onBad) {
		t.Fatal("empty name should not match")
	}
}

func TestMatchOnOR_WorkflowDispatch(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "workflow_dispatch", Payload: map[string]any{}}
	on := map[string]any{"workflow_dispatch": map[string]any{}}
	if !MatchOnOR(ev, on) {
		t.Fatal("expected match")
	}
}

func TestMatchOnOR_Schedule(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{Name: "schedule", Payload: map[string]any{}}
	on := map[string]any{"schedule": "0 0 * * *"}
	if !MatchOnOR(ev, on) {
		t.Fatal("expected match")
	}
	// non-schedule event with schedule key should still evaluate schedule branch at end
	ev2 := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"action": "opened"}}
	if MatchOnOR(ev2, on) {
		t.Fatal("schedule should not match issues event")
	}
}

func TestScheduleCronMatches(t *testing.T) {
	t.Parallel()
	if !ScheduleCronMatches("", "0 0 * * *") {
		t.Fatal("empty task schedule matches any")
	}
	if !ScheduleCronMatches("daily", "0 0 * * *") {
		t.Fatal("daily alias")
	}
	if !ScheduleCronMatches("weekly", "0 0 * * 0") {
		t.Fatal("weekly alias")
	}
	if !ScheduleCronMatches("hourly", "0 * * * *") {
		t.Fatal("hourly alias")
	}
	if !ScheduleCronMatches("0  0   * * *", "0 0 * * *") {
		t.Fatal("whitespace normalization")
	}
	if ScheduleCronMatches("0 1 * * *", "0 0 * * *") {
		t.Fatal("should not match different crons")
	}
}
