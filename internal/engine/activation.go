package engine

import (
	"fmt"
	"os"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

// validateEventContext checks that the GitHub event is usable for a task run.
func validateEventContext(ev *types.GitHubEvent) error {
	if ev == nil {
		return fmt.Errorf("event is nil")
	}
	if ev.Payload == nil {
		return fmt.Errorf("event payload is nil")
	}
	name := strings.TrimSpace(ev.Name)
	if name == "" {
		return fmt.Errorf("event name is empty")
	}
	// "unknown" is used when GITHUB_EVENT_NAME is unset (see parseEventJSON).
	if strings.EqualFold(name, "unknown") {
		return nil
	}
	return nil
}

// validateTaskConfig checks engine and basic task constraints before running the agent.
func validateTaskConfig(task *config.Task, glob *config.GlobalConfig) error {
	if task == nil {
		return fmt.Errorf("task is nil")
	}
	if glob == nil {
		return fmt.Errorf("config is nil")
	}
	if wmAgent := strings.TrimSpace(os.Getenv("WM_AGENT_CMD")); wmAgent != "" {
		return nil
	}
	eng := strings.ToLower(strings.TrimSpace(task.Engine()))
	if eng == "" {
		eng = strings.ToLower(strings.TrimSpace(glob.Engine))
	}
	switch eng {
	case "", "claude", "codex":
		return nil
	case "copilot":
		return fmt.Errorf("engine copilot requires WM_AGENT_CMD")
	default:
		return fmt.Errorf("unknown engine %q (use claude, codex, copilot, or set WM_AGENT_CMD)", eng)
	}
}
