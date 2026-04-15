package trigger

import (
	"strings"

	"github.com/gh-wm/gh-wm/internal/types"
)

// MatchOnOR returns true if any sub-trigger in on: matches (gh-aw / Actions semantics).
func MatchOnOR(event *types.GitHubEvent, on map[string]any) bool {
	if on == nil || event == nil {
		return false
	}
	matched := false
	if event.Name == "schedule" {
		if _, ok := on["schedule"]; ok {
			matched = matchScheduleBlock(event, on["schedule"])
		}
		if matched {
			return true
		}
	}
	if issues, ok := on["issues"].(map[string]any); ok && matchIssues(event, issues) {
		return true
	}
	if ic, ok := on["issue_comment"].(map[string]any); ok && matchIssueComment(event, ic) {
		return true
	}
	if pr, ok := on["pull_request"].(map[string]any); ok && matchPullRequest(event, pr) {
		return true
	}
	if sc, ok := on["slash_command"].(map[string]any); ok && matchSlashCommand(event, sc) {
		return true
	}
	if _, ok := on["workflow_dispatch"]; ok && event.Name == "workflow_dispatch" {
		return true
	}
	if _, ok := on["schedule"]; ok && event.Name == "schedule" {
		return matchScheduleBlock(event, on["schedule"])
	}
	return false
}

func matchIssues(event *types.GitHubEvent, issues map[string]any) bool {
	if event.Name != "issues" {
		return false
	}
	action, _ := event.Payload["action"].(string)
	typesVal, _ := issues["types"].([]any)
	if len(typesVal) == 0 {
		return true
	}
	for _, t := range typesVal {
		if s, ok := t.(string); ok && s == action {
			return true
		}
	}
	return false
}

func matchIssueComment(event *types.GitHubEvent, ic map[string]any) bool {
	if event.Name != "issue_comment" {
		return false
	}
	action, _ := event.Payload["action"].(string)
	typesVal, _ := ic["types"].([]any)
	if len(typesVal) > 0 {
		ok := false
		for _, t := range typesVal {
			if s, ok2 := t.(string); ok2 && s == action {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

func matchPullRequest(event *types.GitHubEvent, pr map[string]any) bool {
	if event.Name != "pull_request" && event.Name != "pull_request_target" {
		return false
	}
	action, _ := event.Payload["action"].(string)
	typesVal, _ := pr["types"].([]any)
	if len(typesVal) == 0 {
		return true
	}
	for _, t := range typesVal {
		if s, ok := t.(string); ok && s == action {
			return true
		}
	}
	return false
}

func matchSlashCommand(event *types.GitHubEvent, sc map[string]any) bool {
	if event.Name != "issue_comment" {
		return false
	}
	name, _ := sc["name"].(string)
	if name == "" {
		return false
	}
	comment, _ := event.Payload["comment"].(map[string]any)
	body, _ := comment["body"].(string)
	prefix := "/" + strings.TrimPrefix(name, "/")
	body = strings.TrimSpace(body)
	return strings.HasPrefix(body, prefix) || strings.HasPrefix(body, prefix+" ")
}

func matchScheduleBlock(event *types.GitHubEvent, sched any) bool {
	// GitHub schedule events don't include which cron in payload easily in all cases.
	// We pass GITHUB_EVENT_SCHEDULE or compare via env in runner; for resolve, accept all schedule tasks when event is schedule.
	_ = sched
	return event.Name == "schedule"
}

// ScheduleCronMatches checks task schedule string vs cron from workflow (for filtering in run).
func ScheduleCronMatches(taskSchedule string, workflowCron string) bool {
	if taskSchedule == "" {
		return true
	}
	norm := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.Join(strings.Fields(s), " ")
		return s
	}
	switch strings.ToLower(strings.TrimSpace(taskSchedule)) {
	case "daily":
		taskSchedule = "0 0 * * *"
	case "weekly":
		taskSchedule = "0 0 * * 0"
	case "hourly":
		taskSchedule = "0 * * * *"
	}
	return norm(taskSchedule) == norm(workflowCron)
}
