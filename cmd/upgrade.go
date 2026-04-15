package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/an-lee/gh-wm/internal/gen"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Regenerate wm-agent.yml from .wm/tasks and optional GH_WM_REPO",
	RunE:  runUpgrade,
}

func runUpgrade(_ *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	tasksDir := filepath.Join(cwd, ".wm", "tasks")
	schedules, err := gen.CollectSchedulesFromTasksDir(tasksDir)
	if err != nil {
		return err
	}
	repo := os.Getenv("GH_WM_REPO")
	if repo == "" {
		repo = "an-lee/gh-wm"
	}
	ghDir := filepath.Join(cwd, ".github", "workflows")
	if err := os.MkdirAll(ghDir, 0o755); err != nil {
		return err
	}
	if err := gen.WriteWMAgent(ghDir, repo, schedules); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Updated .github/workflows/wm-agent.yml")
	return nil
}
