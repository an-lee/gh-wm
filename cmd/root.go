package cmd

import (
	"github.com/spf13/cobra"
)

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "gh-wm",
	Short: "GitHub Workflow Manager — run gh-aw-style task markdown in CI",
	Long:  `gh-wm loads .wm/tasks/*.md (gh-aw compatible), resolves events, and runs agents.`,
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(assignCmd)
	rootCmd.AddCommand(resolveCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(addCmd)
}
