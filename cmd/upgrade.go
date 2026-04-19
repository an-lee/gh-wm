package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Self-upgrade the gh-wm gh extension",
	RunE:  runUpgrade,
}

func runUpgrade(_ *cobra.Command, _ []string) error {
	fmt.Fprintln(os.Stderr, "Upgrading gh-wm extension...")
	out, err := exec.Command("gh", "extension", "upgrade", "an-lee/gh-wm").CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			fmt.Fprintf(os.Stderr, "gh extension upgrade failed: %v\n%s\n", err, msg)
		} else {
			fmt.Fprintf(os.Stderr, "gh extension upgrade failed: %v\n", err)
		}
		return err
	}
	fmt.Fprintln(os.Stderr, "gh extension upgrade completed.")
	return nil
}
