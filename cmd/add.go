package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gh-wm/gh-wm/internal/config"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <url-or-path>",
	Short: "Download or copy a gh-aw-style task .md into .wm/tasks/",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func runAdd(_ *cobra.Command, args []string) error {
	src := strings.TrimSpace(args[0])
	var data []byte
	var err error
	switch {
	case strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://"):
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Get(src)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP %s", resp.Status)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	default:
		data, err = os.ReadFile(src)
		if err != nil {
			return err
		}
	}
	if _, _, err := config.SplitFrontmatter(string(data)); err != nil {
		return fmt.Errorf("task file must start with YAML frontmatter (---): %w", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	tasksDir := filepath.Join(cwd, ".wm", "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		return err
	}
	base := filepath.Base(src)
	if idx := strings.Index(base, "?"); idx >= 0 {
		base = base[:idx]
	}
	if !strings.HasSuffix(strings.ToLower(base), ".md") {
		base = base + ".md"
	}
	dest := filepath.Join(tasksDir, base)
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Wrote %s\nRun: gh wm upgrade\n", dest)
	return nil
}
