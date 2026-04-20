package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/an-lee/gh-wm/internal/compat/awexpr"
	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/gen"
	"github.com/spf13/cobra"
)

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Regenerate .github/workflows/wm-agent.yml from tasks and config",
	RunE:  runCompile,
}

func runCompile(_ *cobra.Command, _ []string) error {
	if err := regenerateWMAgentFromCwd(); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Updated .github/workflows/wm-agent.yml")
	return nil
}

// regenerateWMAgentFromCwd validates tasks and writes wm-agent.yml plus .wm/.gitignore for runs/.
func regenerateWMAgentFromCwd() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	tasksDir := filepath.Join(cwd, ".wm", "tasks")
	triggers, err := gen.CollectTriggersFromTasksDir(tasksDir)
	if err != nil {
		return err
	}
	tasks, err := config.LoadTasksDir(tasksDir)
	if err != nil {
		return err
	}
	glob, err := config.LoadGlobalOnly(cwd)
	if err != nil {
		return err
	}
	mode := awexpr.ParseGhAWExpressionsMode(config.CompatGhAWExpressions(glob))
	warnf := func(path, msg string) {
		fmt.Fprintf(os.Stderr, "wm: %s: %s\n", path, msg)
	}
	if err := awexpr.ValidateTasks(tasks, mode, warnf); err != nil {
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
	runsOn := config.WorkflowRunsOnLabels(glob)
	var preSteps []config.StepDef
	if glob != nil {
		preSteps = glob.Workflow.PreSteps
	}
	if err := gen.WriteWMAgent(ghDir, repo, triggers, runsOn, preSteps, config.WorkflowInstallClaudeCode(glob), config.WorkflowGhWMExtensionVersion(glob)); err != nil {
		return err
	}
	return ensureWmGitignoreRuns(filepath.Join(cwd, ".wm"))
}
