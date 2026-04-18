package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/an-lee/gh-wm/internal/antiloop"
	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/trigger"
	"github.com/an-lee/gh-wm/internal/types"
)

// ResolveMatchingTasks returns task names that match the event.
func ResolveMatchingTasks(repoRoot string, event *types.GitHubEvent) ([]string, error) {
	_, tasks, err := config.Load(repoRoot)
	if err != nil {
		return nil, err
	}
	if antiloop.ShouldSkipAutomatedSender(event) {
		return nil, nil
	}
	var names []string
	for _, t := range tasks {
		on := t.OnMap()
		if trigger.MatchOnOR(event, on) {
			// For schedule: optionally filter by cron match via env
			if event.Name == "schedule" {
				wc := os.Getenv("WM_SCHEDULE_CRON")
				if wc != "" {
					if !trigger.ScheduleCronMatches(t, wc) {
						continue
					}
				}
			}
			names = append(names, t.Name)
		}
	}
	return names, nil
}

// ResolveForcedTask returns [taskName] if a task with that name exists under .wm/tasks.
// It does not evaluate on: triggers (same semantics as local gh wm run).
func ResolveForcedTask(repoRoot, taskName string) ([]string, error) {
	taskName = strings.TrimSpace(taskName)
	if taskName == "" {
		return nil, fmt.Errorf("force-task: empty")
	}
	_, tasks, err := config.Load(repoRoot)
	if err != nil {
		return nil, err
	}
	for _, t := range tasks {
		if t.Name == taskName {
			return []string{t.Name}, nil
		}
	}
	return nil, fmt.Errorf("task not found: %s", taskName)
}

func parseEventJSON(eventName string, b []byte) (*types.GitHubEvent, error) {
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

// ParseEventFile reads event JSON from path (github.event payload).
func ParseEventFile(eventName, path string) (*types.GitHubEvent, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseEventJSON(eventName, b)
}

// ParseEvent loads a GitHub event from path, or uses an empty JSON object when path is empty
// (neither --payload nor GITHUB_EVENT_PATH), for local quick runs.
func ParseEvent(eventName, path string) (*types.GitHubEvent, error) {
	if strings.TrimSpace(path) == "" {
		return parseEventJSON(eventName, []byte("{}"))
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseEventJSON(eventName, b)
}
