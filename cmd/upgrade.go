package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/gen"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Self-upgrade gh-wm extension and regenerate wm-agent.yml",
	RunE:  runUpgrade,
}

func runUpgrade(_ *cobra.Command, _ []string) error {
	fmt.Fprintln(os.Stderr, "Upgrading gh-wm extension...")
	out, err := exec.Command("gh", "extension", "upgrade", "an-lee/gh-wm").CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			fmt.Fprintf(os.Stderr, "gh extension upgrade skipped or failed: %v\n%s\n", err, msg)
		} else {
			fmt.Fprintf(os.Stderr, "gh extension upgrade skipped or failed: %v\n", err)
		}
		fmt.Fprintln(os.Stderr, "Continuing: regenerating .github/workflows/wm-agent.yml (independent of extension upgrade).")
	}

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
	glob, err := config.LoadGlobalOnly(cwd)
	if err != nil {
		return err
	}
	runsOn := config.WorkflowRunsOnLabels(glob)
	var preSteps []config.StepDef
	if glob != nil {
		preSteps = glob.Workflow.PreSteps
	}
	if err := gen.WriteWMAgent(ghDir, repo, schedules, runsOn, preSteps); err != nil {
		return err
	}
	if err := ensureWmGitignoreRuns(filepath.Join(cwd, ".wm")); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Updated .github/workflows/wm-agent.yml")
	return nil
}
