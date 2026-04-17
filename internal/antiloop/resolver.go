package antiloop

import (
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

// CollectStateLabelValues returns all non-empty wm.state_labels values from tasks (for loop-guard skips).
func CollectStateLabelValues(tasks []*config.Task) map[string]struct{} {
	out := make(map[string]struct{})
	for _, t := range tasks {
		if t == nil {
			continue
		}
		wm := t.WM()
		for _, v := range wm.StateLabels {
			v = strings.TrimSpace(v)
			if v != "" {
				out[v] = struct{}{}
			}
		}
	}
	return out
}

// ShouldSkipIssuesLabeledStateLabel returns true when the event is issues+labeled with a label
// that is any configured state label (avoids state-machine churn re-resolving tasks).
func ShouldSkipIssuesLabeledStateLabel(ev *types.GitHubEvent, stateLabels map[string]struct{}) bool {
	if len(stateLabels) == 0 || ev == nil || ev.Name != "issues" {
		return false
	}
	action, _ := ev.Payload["action"].(string)
	if action != "labeled" {
		return false
	}
	name := labelNameFromIssuesPayload(ev.Payload)
	if name == "" {
		return false
	}
	_, ok := stateLabels[name]
	return ok
}

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

func labelNameFromIssuesPayload(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	lab, ok := payload["label"].(map[string]any)
	if !ok {
		return ""
	}
	s, _ := lab["name"].(string)
	return strings.TrimSpace(s)
}
