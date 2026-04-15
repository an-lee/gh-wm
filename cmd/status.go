package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show open issues with agent labels (via gh issue list)",
	RunE:  runStatus,
}

func runStatus(_ *cobra.Command, _ []string) error {
	cmd := exec.Command("gh", "issue", "list", "--label", "agent,agent:working,agent:review", "--json", "number,title,labels", "--limit", "50")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh issue list: %w", err)
	}
	return nil
}
