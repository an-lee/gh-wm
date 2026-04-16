package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/gen"
	"github.com/an-lee/gh-wm/internal/templates"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create .wm/ layout, starter tasks, and wm-agent.yml",
	RunE:  runInit,
}

func runInit(_ *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	wm := filepath.Join(cwd, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		return err
	}
	if err := templates.WriteConfig(wm); err != nil {
		return err
	}
	if err := templates.WriteStarterTasks(filepath.Join(wm, "tasks")); err != nil {
		return err
	}
	if err := templates.WriteCLAUDE(cwd); err != nil {
		return err
	}
	ghDir := filepath.Join(cwd, ".github", "workflows")
	if err := os.MkdirAll(ghDir, 0o755); err != nil {
		return err
	}
	schedules, err := gen.CollectSchedulesFromTasksDir(filepath.Join(wm, "tasks"))
	if err != nil {
		return err
	}
	repo := os.Getenv("GH_WM_REPO")
	if repo == "" {
		repo = "an-lee/gh-wm"
	}
	glob, _, err := config.Load(cwd)
	if err != nil {
		return err
	}
	runsOn := config.WorkflowRunsOnLabels(glob)
	if err := gen.WriteWMAgent(ghDir, repo, schedules, runsOn); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Initialized .wm/ and .github/workflows/wm-agent.yml")
	return nil
}
