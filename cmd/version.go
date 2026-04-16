package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the release version (set via -ldflags at build time).
var Version = "dev"

// Commit is an optional VCS short SHA (set via -ldflags at build time).
var Commit = ""

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print gh-wm version",
	Args:  cobra.NoArgs,
	RunE:  runVersion,
}

func runVersion(cmd *cobra.Command, _ []string) error {
	fmt.Fprintf(cmd.OutOrStdout(), "gh-wm %s\n", Version)
	if Commit != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "commit: %s\n", Commit)
	}
	return nil
}
