package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/an-lee/gh-wm/internal/compat/awexpr"
	"github.com/an-lee/gh-wm/internal/config"
	"github.com/spf13/cobra"
)

var validateRepoRoot string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate task markdown bodies for gh-aw–compatible ${{ }} expressions",
	Long:  "Scans .wm/tasks/*.md bodies (same rules as gh wm upgrade). Uses compat.gh_aw_expressions from .wm/config.yml (error | warn | off).",
	RunE:  runValidate,
}

func init() {
	validateCmd.Flags().StringVar(&validateRepoRoot, "repo-root", ".", "repository root")
	rootCmd.AddCommand(validateCmd)
}

func runValidate(_ *cobra.Command, _ []string) error {
	cwd := validateRepoRoot
	if abs, err := filepath.Abs(cwd); err == nil {
		cwd = abs
	}
	tasksDir := filepath.Join(cwd, ".wm", "tasks")
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
	fmt.Fprintln(os.Stderr, "wm validate: ok")
	return nil
}
