package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/an-lee/gh-wm/internal/engine"
	"github.com/spf13/cobra"
)

var (
	resolveRepoRoot  string
	resolveEvent     string
	resolvePayload   string
	resolveJSON      bool
	resolveForceTask string
)

var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "List task names matching the GitHub event",
	RunE:  runResolve,
}

func init() {
	resolveCmd.Flags().StringVar(&resolveRepoRoot, "repo-root", ".", "repository root (contains .wm/)")
	resolveCmd.Flags().StringVar(&resolveEvent, "event-name", "", "GitHub event name (default: GITHUB_EVENT_NAME)")
	resolveCmd.Flags().StringVar(&resolvePayload, "payload", "", "path to event JSON (default: GITHUB_EVENT_PATH; if unset, `{}`)")
	resolveCmd.Flags().BoolVar(&resolveJSON, "json", true, "print JSON array to stdout")
	resolveCmd.Flags().StringVar(&resolveForceTask, "force-task", "", "pin a single task by name (skips event matching; for manual/CI use)")
}

func runResolve(cmd *cobra.Command, _ []string) error {
	var names []string
	var err error
	if strings.TrimSpace(resolveForceTask) != "" {
		names, err = engine.ResolveForcedTask(resolveRepoRoot, resolveForceTask)
	} else {
		evName := resolveEvent
		if evName == "" {
			evName = os.Getenv("GITHUB_EVENT_NAME")
		}
		path := resolvePayload
		if path == "" {
			path = os.Getenv("GITHUB_EVENT_PATH")
		}
		ev, err := engine.ParseEvent(evName, path)
		if err != nil {
			return err
		}
		names, err = engine.ResolveMatchingTasks(resolveRepoRoot, ev)
	}
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if !resolveJSON {
		for _, n := range names {
			fmt.Fprintln(out, n)
		}
		return nil
	}
	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	return enc.Encode(names)
}
