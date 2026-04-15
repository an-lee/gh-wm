package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs <issue-number>",
	Short: "List wm-agent workflow runs (best-effort filter by issue # in run title)",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogs,
}

type workflowRunRow struct {
	DisplayTitle string `json:"displayTitle"`
	URL          string `json:"url"`
	CreatedAt    string `json:"createdAt"`
}

func runLogs(_ *cobra.Command, args []string) error {
	n, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	out, err := exec.Command("gh", "run", "list", "--workflow", "wm-agent.yml", "--limit", "40", "--json", "displayTitle,url,createdAt").Output()
	if err != nil {
		return err
	}
	var rows []workflowRunRow
	if err := json.Unmarshal(out, &rows); err != nil {
		return err
	}
	needle := fmt.Sprintf("#%d", n)
	var matched []workflowRunRow
	for _, r := range rows {
		if strings.Contains(r.DisplayTitle, needle) {
			matched = append(matched, r)
		}
	}
	if len(matched) == 0 {
		fmt.Fprintf(os.Stderr, "No runs whose title contains %q (showing last %d runs):\n", needle, len(rows))
		for _, r := range rows {
			fmt.Printf("%s  %s\n  %s\n", r.CreatedAt, r.DisplayTitle, r.URL)
		}
		return nil
	}
	for _, r := range matched {
		fmt.Printf("%s  %s\n  %s\n", r.CreatedAt, r.DisplayTitle, r.URL)
	}
	return nil
}
