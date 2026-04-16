package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/engine"
	"github.com/spf13/cobra"
)

var (
	runRepoRoot string
	runTask     string
	runEvent    string
	runPayload  string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a single task for the current event",
	RunE:  runRun,
}

func init() {
	runCmd.Flags().StringVar(&runRepoRoot, "repo-root", ".", "repository root")
	runCmd.Flags().StringVar(&runTask, "task", "", "task name (filename without .md)")
	runCmd.Flags().StringVar(&runEvent, "event-name", "", "event name (default: GITHUB_EVENT_NAME)")
	runCmd.Flags().StringVar(&runPayload, "payload", "", "event JSON path (default: GITHUB_EVENT_PATH; if unset, `{}`)")
	_ = runCmd.MarkFlagRequired("task")
}

func runRun(_ *cobra.Command, _ []string) error {
	evName := runEvent
	if evName == "" {
		evName = os.Getenv("GITHUB_EVENT_NAME")
	}
	path := runPayload
	if path == "" {
		path = os.Getenv("GITHUB_EVENT_PATH")
	}
	ev, err := engine.ParseEvent(evName, path)
	if err != nil {
		return err
	}
	_, tasks, err := config.Load(runRepoRoot)
	if err != nil {
		return err
	}
	var task *config.Task
	for _, t := range tasks {
		if t.Name == runTask {
			task = t
			break
		}
	}
	if task == nil {
		return fmt.Errorf("task not found: %s", runTask)
	}
	min := task.TimeoutMinutes(45)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(min)*time.Minute)
	defer cancel()
	res, err := engine.RunTask(ctx, runRepoRoot, runTask, ev)
	if res != nil {
		fmt.Fprintln(os.Stderr, res.Stdout)
		if res.Stderr != "" {
			fmt.Fprintln(os.Stderr, res.Stderr)
		}
	}
	return err
}
