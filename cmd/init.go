package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	preSteps := glob.Workflow.PreSteps
	if err := gen.WriteWMAgent(ghDir, repo, schedules, runsOn, preSteps, config.WorkflowInstallClaudeCode(glob)); err != nil {
		return err
	}
	if err := ensureWmGitignoreRuns(wm); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Initialized .wm/ and .github/workflows/wm-agent.yml")
	return nil
}

const wmGitignoreRunsLine = "runs/"

// ensureWmGitignoreRuns writes .wm/.gitignore so per-run artifact dirs under .wm/runs/ are not tracked.
func ensureWmGitignoreRuns(wmDir string) error {
	if err := os.MkdirAll(wmDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(wmDir, ".gitignore")
	line := wmGitignoreRunsLine + "\n"
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(path, []byte(line), 0o644)
		}
		return err
	}
	s := string(b)
	if strings.Contains(s, "runs/") {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if len(s) > 0 && !strings.HasSuffix(s, "\n") {
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}
	_, err = f.WriteString(line)
	return err
}
