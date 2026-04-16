package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/engine"
	"github.com/an-lee/gh-wm/internal/gitbranch"
	"github.com/an-lee/gh-wm/internal/gitstatus"
	"github.com/an-lee/gh-wm/internal/types"
	"github.com/spf13/cobra"
)

var (
	runRepoRoot   string
	runTask       string
	runEvent      string
	runPayload    string
	runAllowDirty bool
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
	runCmd.Flags().BoolVar(&runAllowDirty, "allow-dirty", false, "skip git clean working tree check (git status --porcelain must be empty otherwise)")
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
	glob, tasks, err := config.Load(runRepoRoot)
	if err != nil {
		return err
	}
	glob = config.DefaultGlobal(glob)
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
	if !runAllowDirty {
		if err := gitstatus.EnsureClean(runRepoRoot); err != nil {
			return err
		}
	}
	min := task.TimeoutMinutes(45)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(min)*time.Minute)
	defer cancel()

	repoDisplay := runRepoRoot
	if abs, err := filepath.Abs(runRepoRoot); err == nil {
		repoDisplay = abs
	}
	engineName := strings.TrimSpace(task.Engine())
	if engineName == "" {
		engineName = glob.Engine
	}
	branch := "(unknown)"
	if b, err := gitbranch.CurrentBranch(repoDisplay); err == nil {
		branch = b
	}
	fmt.Fprintf(os.Stderr, "wm run: task=%q repo=%s branch=%s engine=%s\n", runTask, repoDisplay, branch, engineName)
	fmt.Fprintf(os.Stderr, "wm run: agent subprocess starting (streaming stderr)...\n\n")

	start := time.Now()
	runResult, err := engine.RunTask(ctx, runRepoRoot, runTask, ev, &engine.RunOptions{LogWriter: os.Stderr})
	dur := time.Since(start)

	exitCode := -1
	phase := types.PhaseActivation
	success := false
	if runResult != nil {
		phase = runResult.Phase
		success = runResult.Success
		if runResult.AgentResult != nil {
			exitCode = runResult.AgentResult.ExitCode
		}
	}

	if runResult != nil && runResult.RunDir != "" {
		fmt.Fprintf(os.Stderr, "wm run: artifacts=%s\n", runResult.RunDir)
	}
	fmt.Fprintf(os.Stderr, "\n---\nwm run: task=%q repo=%s duration=%s exit_code=%d success=%v phase=%s\n",
		runTask, repoDisplay, dur.Round(time.Millisecond), exitCode, success, phase)
	if err != nil {
		if runResult != nil && runResult.Phase == types.PhaseOutputs {
			fmt.Fprintf(os.Stderr, "failure phase: safe-outputs (post-agent)\n")
		} else {
			fmt.Fprintf(os.Stderr, "failure phase: %s\n", phase)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
	return err
}
