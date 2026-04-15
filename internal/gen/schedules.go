package gen

import (
	"github.com/an-lee/gh-wm/internal/config"
)

// CollectSchedulesFromTasksDir reads all tasks and unions on.schedule strings.
func CollectSchedulesFromTasksDir(tasksDir string) ([]string, error) {
	tasks, err := config.LoadTasksDir(tasksDir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, t := range tasks {
		s := t.ScheduleString()
		if s != "" {
			out = append(out, normalizeSchedule(s))
		}
	}
	return dedupe(out), nil
}

func normalizeSchedule(s string) string {
	switch {
	case s == "daily":
		return "0 0 * * *"
	case s == "weekly":
		return "0 0 * * 0"
	case s == "hourly":
		return "0 * * * *"
	default:
		return s
	}
}
