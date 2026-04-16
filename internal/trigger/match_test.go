package trigger

import (
	"path/filepath"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/gen"
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

func TestMatchOnOR_Issues_LabelsFilter(t *testing.T) {
	t.Parallel()
	on := map[string]any{
		"issues": map[string]any{
			"types":  []any{"labeled"},
			"labels": []any{"implement"},
		},
	}
	ev := &types.GitHubEvent{
		Name: "issues",
		Payload: map[string]any{
			"action": "labeled",
			"label":  map[string]any{"name": "implement"},
		},
	}
	if !MatchOnOR(ev, on) {
		t.Fatal("expected match when label matches")
	}
	ev.Payload["label"] = map[string]any{"name": "bug"}
	if MatchOnOR(ev, on) {
		t.Fatal("expected no match when label does not match filter")
	}
	ev.Payload["action"] = "opened"
	ev.Payload["label"] = map[string]any{"name": "implement"}
	if MatchOnOR(ev, on) {
		t.Fatal("labels filter requires action labeled")
	}
}

func TestMatchOnOR_IssueComment_WMAgentMarker(t *testing.T) {
	t.Parallel()
	ev := &types.GitHubEvent{
		Name: "issue_comment",
		Payload: map[string]any{
			"action": "created",
			"comment": map[string]any{
				"body": "hello\n\n" + WMAgentCommentMarkerPrefix + "x -->",
			},
		},
	}
	on := map[string]any{"issue_comment": map[string]any{"types": []any{"created"}}}
	if MatchOnOR(ev, on) {
		t.Fatal("wm-authored comment should not match")
	}
}

func TestIssueCommentBodyFromWMAgent(t *testing.T) {
	t.Parallel()
	if !IssueCommentBodyFromWMAgent(map[string]any{
		"comment": map[string]any{"body": "<!-- wm-agent:task -->"},
	}) {
		t.Fatal("expected true")
	}
	if IssueCommentBodyFromWMAgent(map[string]any{
		"comment": map[string]any{"body": "human text"},
	}) {
		t.Fatal("expected false")
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
	// wm-authored comment should not match slash_command either
	evWM := &types.GitHubEvent{
		Name: "issue_comment",
		Payload: map[string]any{
			"comment": map[string]any{"body": "/deploy prod\n\n" + WMAgentCommentMarkerPrefix + "t -->"},
		},
	}
	if MatchOnOR(evWM, on) {
		t.Fatal("wm marker should block slash_command match")
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

func testTaskWithSchedule(path, schedule string) *config.Task {
	return &config.Task{
		Name: filepath.Base(path),
		Path: path,
		Frontmatter: map[string]any{
			"on": map[string]any{"schedule": schedule},
		},
	}
}

func BenchmarkMatchOnOR_Issues(b *testing.B) {
	ev := &types.GitHubEvent{
		Name:    "issues",
		Payload: map[string]any{"action": "opened"},
	}
	on := map[string]any{"issues": map[string]any{"types": []any{"opened", "closed", "labeled"}}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MatchOnOR(ev, on)
	}
}

func BenchmarkMatchOnOR_SlashCommand(b *testing.B) {
	ev := &types.GitHubEvent{
		Name: "issue_comment",
		Payload: map[string]any{
			"comment": map[string]any{"body": "/deploy prod"},
		},
	}
	on := map[string]any{"slash_command": map[string]any{"name": "deploy"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MatchOnOR(ev, on)
	}
}

func TestScheduleCronMatches(t *testing.T) {
	t.Parallel()
	p := "/repo/.wm/tasks/t.md"
	if !ScheduleCronMatches(testTaskWithSchedule(p, ""), "0 0 * * *") {
		t.Fatal("empty task schedule matches any")
	}
	if ScheduleCronMatches(nil, "0 0 * * *") {
		t.Fatal("nil task should not match")
	}
	dailyCron := gen.FuzzyNormalizeSchedule("daily", p)
	if !ScheduleCronMatches(testTaskWithSchedule(p, "daily"), dailyCron) {
		t.Fatal("daily fuzzy cron")
	}
	weeklyCron := gen.FuzzyNormalizeSchedule("weekly", p)
	if !ScheduleCronMatches(testTaskWithSchedule(p, "weekly"), weeklyCron) {
		t.Fatal("weekly fuzzy cron")
	}
	hourlyCron := gen.FuzzyNormalizeSchedule("hourly", p)
	if !ScheduleCronMatches(testTaskWithSchedule(p, "hourly"), hourlyCron) {
		t.Fatal("hourly fuzzy cron")
	}
	if !ScheduleCronMatches(testTaskWithSchedule(p, "0  0   * * *"), "0 0 * * *") {
		t.Fatal("whitespace normalization for raw cron")
	}
	if ScheduleCronMatches(testTaskWithSchedule(p, "0 1 * * *"), "0 0 * * *") {
		t.Fatal("should not match different crons")
	}
}
