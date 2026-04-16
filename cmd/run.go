package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/engine"
	"github.com/an-lee/gh-wm/internal/ghclient"
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
	runRemote     bool
	runGhRepo     string
	runWorkflow   string
	runRef        string
	runIssue      int
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
	runCmd.Flags().BoolVar(&runRemote, "remote", false, "dispatch wm-agent.yml on GitHub via gh workflow run (requires gh CLI; run gh wm upgrade so workflow has task_name input)")
	runCmd.Flags().StringVar(&runGhRepo, "repo", "", "GitHub repository OWNER/NAME for --remote (default: gh repo view)")
	runCmd.Flags().StringVar(&runWorkflow, "workflow", "wm-agent.yml", "workflow file for --remote")
	runCmd.Flags().StringVar(&runRef, "ref", "", "git ref for --remote (optional; default branch if unset)")
	runCmd.Flags().IntVar(&runIssue, "issue", 0, "optional issue/PR number for --remote (-f issue_number)")
	_ = runCmd.MarkFlagRequired("task")
}

func runRun(_ *cobra.Command, _ []string) error {
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

	if runRemote {
		return runRemoteDispatch(glob, task)
	}

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

func runRemoteDispatch(glob *config.GlobalConfig, task *config.Task) error {
	repo := strings.TrimSpace(runGhRepo)
	if repo == "" {
		var err error
		repo, err = ghclient.CurrentRepo()
		if err != nil {
			return fmt.Errorf("remote run: set --repo or run from a gh checkout: %w", err)
		}
	}
	repoDisplay := runRepoRoot
	if abs, err := filepath.Abs(runRepoRoot); err == nil {
		repoDisplay = abs
	}
	engineName := strings.TrimSpace(task.Engine())
	if engineName == "" {
		engineName = glob.Engine
	}
	fmt.Fprintf(os.Stderr, "wm run: remote workflow_dispatch workflow=%s github_repo=%s task=%q engine=%s repo_root=%s\n",
		strings.TrimSpace(runWorkflow), repo, runTask, engineName, repoDisplay)

	wf := strings.TrimSpace(runWorkflow)
	if wf == "" {
		wf = "wm-agent.yml"
	}
	args := []string{"workflow", "run", wf, "-R", repo, "-f", "task_name=" + runTask}
	if runIssue > 0 {
		args = append(args, "-f", fmt.Sprintf("issue_number=%d", runIssue))
	}
	if r := strings.TrimSpace(runRef); r != "" {
		args = append(args, "--ref", r)
	}
	cmd := exec.Command("gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh %v: %w", args, err)
	}
	return nil
}
