package engine

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gh-wm/gh-wm/internal/config"
	"github.com/gh-wm/gh-wm/internal/trigger"
	"github.com/gh-wm/gh-wm/internal/types"
)

// ResolveMatchingTasks returns task names that match the event.
func ResolveMatchingTasks(repoRoot string, event *types.GitHubEvent) ([]string, error) {
	_, tasks, err := config.Load(repoRoot)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, t := range tasks {
		on := t.OnMap()
		if trigger.MatchOnOR(event, on) {
			// For schedule: optionally filter by cron match via env
			if event.Name == "schedule" {
				wc := os.Getenv("WM_SCHEDULE_CRON")
				if wc != "" {
					ts := t.ScheduleString()
					if ts != "" && !trigger.ScheduleCronMatches(ts, wc) {
						continue
					}
				}
			}
			names = append(names, t.Name)
		}
	}
	return names, nil
}

// ParseEventFile reads event JSON from path (github.event payload).
func ParseEventFile(eventName, path string) (*types.GitHubEvent, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(b, &payload); err != nil {
		return nil, fmt.Errorf("parse event json: %w", err)
	}
	name := eventName
	if name == "" {
		name = os.Getenv("GITHUB_EVENT_NAME")
	}
	if name == "" {
		name = "unknown"
	}
	return &types.GitHubEvent{Name: name, Payload: payload}, nil
}
