package cmd

import (
	"fmt"
	"strconv"

	"github.com/an-lee/gh-wm/internal/ghclient"
	"github.com/spf13/cobra"
)

var assignLabel string

var assignCmd = &cobra.Command{
	Use:   "assign <issue-number>",
	Short: "Add trigger label to an issue (default: agent)",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssign,
}

func init() {
	assignCmd.Flags().StringVar(&assignLabel, "label", "agent", "label to add")
}

func runAssign(_ *cobra.Command, args []string) error {
	n, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("issue number: %w", err)
	}
	repo, err := ghclient.CurrentRepo()
	if err != nil {
		return err
	}
	return ghclient.AddIssueLabel(repo, n, assignLabel)
}
