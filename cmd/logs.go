package cmd

import (
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs <issue-number>",
	Short: "Open latest workflow run for this repo (placeholder)",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogs,
}

func runLogs(_ *cobra.Command, args []string) error {
	_, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	cmd := exec.Command("gh", "run", "list", "--workflow", "wm-agent.yml", "--limit", "5")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
