package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var statusAll bool

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show open issues with agent labels (via gh issue list or gh search)",
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().BoolVar(&statusAll, "all", false, "search issues across visible repos (gh search issues)")
}

func runStatus(_ *cobra.Command, _ []string) error {
	if statusAll {
		cmd := exec.Command("gh", "search", "issues", "label:agent OR label:agent:working OR label:agent:review", "--limit", "30", "--json", "repository,number,title,state")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("gh search issues: %w", err)
		}
		return nil
	}
	cmd := exec.Command("gh", "issue", "list", "--label", "agent,agent:working,agent:review", "--json", "number,title,labels", "--limit", "50")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh issue list: %w", err)
	}
	return nil
}
