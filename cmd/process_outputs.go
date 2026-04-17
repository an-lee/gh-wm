package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/an-lee/gh-wm/internal/engine"
	"github.com/an-lee/gh-wm/internal/types"
	"github.com/spf13/cobra"
)

var (
	processRepoRoot string
	processRunDir   string
	processTask     string
	processEvent    string
	processPayload  string
)

var processOutputsCmd = &cobra.Command{
	Use:   "process-outputs",
	Short: "Apply safe-outputs and conclusion for a run directory (after --agent-only)",
	Long: `Loads the run directory from a prior gh wm run --agent-only invocation and executes
the safe-outputs phase plus conclusion (checkpoint, state labels). Use in a follow-up CI job
with write permissions while the agent job used a read-only token.

Pass either --run-dir or --task (CI: --task "$TASK_NAME" resolves the newest .wm/runs/<id> for that task).`,
	RunE: runProcessOutputs,
}

func init() {
	processOutputsCmd.Flags().StringVar(&processRepoRoot, "repo-root", ".", "repository root")
	processOutputsCmd.Flags().StringVar(&processRunDir, "run-dir", "", "path to per-run directory (.wm/runs/<id>); mutually exclusive with --task")
	processOutputsCmd.Flags().StringVar(&processTask, "task", "", "task name: use newest run dir under .wm/runs for this task (mutually exclusive with --run-dir)")
	processOutputsCmd.Flags().StringVar(&processEvent, "event-name", "", "event name (default: GITHUB_EVENT_NAME)")
	processOutputsCmd.Flags().StringVar(&processPayload, "payload", "", "event JSON path (default: GITHUB_EVENT_PATH; if unset, `{}`)")
	rootCmd.AddCommand(processOutputsCmd)
}

func runProcessOutputs(_ *cobra.Command, _ []string) error {
	evName := processEvent
	if evName == "" {
		evName = os.Getenv("GITHUB_EVENT_NAME")
	}
	path := processPayload
	if path == "" {
		path = os.Getenv("GITHUB_EVENT_PATH")
	}
	ev, err := engine.ParseEvent(evName, path)
	if err != nil {
		return err
	}

	if processRunDir != "" && processTask != "" {
		return fmt.Errorf("wm process-outputs: specify either --run-dir or --task, not both")
	}
	if processRunDir == "" && processTask == "" {
		return fmt.Errorf("wm process-outputs: required flag: --run-dir or --task")
	}
	var runDir string
	if processTask != "" {
		runDir, err = engine.FindLatestRunDirForTask(processRepoRoot, processTask)
		if err != nil {
			return err
		}
	} else {
		runDir = filepath.Clean(processRunDir)
	}
	repoDisplay := processRepoRoot
	if abs, err := filepath.Abs(processRepoRoot); err == nil {
		repoDisplay = abs
	}
	fmt.Fprintf(os.Stderr, "wm process-outputs: run_dir=%s repo=%s\n\n", runDir, repoDisplay)

	start := time.Now()
	runResult, err := engine.ProcessRunOutputs(context.Background(), processRepoRoot, runDir, ev, &engine.RunOptions{
		LogWriter:      os.Stderr,
		ProgressWriter: os.Stderr,
	})
	dur := time.Since(start)

	exitCode := -1
	phase := types.PhaseOutputs
	success := false
	if runResult != nil {
		phase = runResult.Phase
		success = runResult.Success
		if runResult.AgentResult != nil {
			exitCode = runResult.AgentResult.ExitCode
		}
	}

	fmt.Fprintf(os.Stderr, "\n---\nwm process-outputs: run_dir=%s duration=%s exit_code=%d success=%v phase=%s\n",
		runDir, dur.Round(time.Millisecond), exitCode, success, phase)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
	return err
}
