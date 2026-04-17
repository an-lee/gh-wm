package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/output"
	"github.com/an-lee/gh-wm/internal/types"
)

// ProcessRunOutputs applies safe-outputs and the conclusion phase for a run directory produced by
// RunTask with RunOptions.AgentOnly (CI token sandbox: follow-up job with write permissions).
func ProcessRunOutputs(ctx context.Context, repoRoot, runDirPath string, event *types.GitHubEvent, opts *RunOptions) (*types.RunResult, error) {
	start := time.Now()
	runDirPath = filepath.Clean(runDirPath)
	st, err := os.Stat(runDirPath)
	if err != nil {
		return nil, fmt.Errorf("process-outputs: run directory: %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("process-outputs: not a directory: %s", runDirPath)
	}

	metaPath := filepath.Join(runDirPath, metaFileName)
	mb, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("process-outputs: read meta.json: %w", err)
	}
	var meta runMeta
	if err := json.Unmarshal(mb, &meta); err != nil {
		return nil, fmt.Errorf("process-outputs: parse meta.json: %w", err)
	}
	taskName := strings.TrimSpace(meta.TaskName)
	if taskName == "" {
		return nil, fmt.Errorf("process-outputs: empty task_name in meta.json")
	}

	glob, tasks, err := config.Load(repoRoot)
	if err != nil {
		return nil, err
	}
	glob = config.DefaultGlobal(glob)
	var task *config.Task
	for _, t := range tasks {
		if t.Name == taskName {
			task = t
			break
		}
	}
	if task == nil {
		return nil, fmt.Errorf("process-outputs: task not found: %s", taskName)
	}
	if err := validateEventContext(event); err != nil {
		return nil, err
	}
	if err := validateTaskConfig(task, glob); err != nil {
		return nil, err
	}

	tc := &types.TaskContext{
		TaskName: taskName,
		Repo:     os.Getenv("GITHUB_REPOSITORY"),
		RepoPath: repoRoot,
		Event:    event,
	}
	if event != nil {
		extractNumbers(event.Payload, tc)
	}

	wm := task.WM()
	am, err := ReadActivationMeta(runDirPath)
	if err != nil {
		return nil, err
	}
	branchCreated := am.BranchCreated
	prevBranch := am.PrevBranch

	rd := &RunDir{Path: runDirPath}
	ar, err := loadPersistedAgentResult(runDirPath)
	if err != nil {
		return nil, err
	}

	result := &types.RunResult{
		Phase:       types.PhaseOutputs,
		RunDir:      runDirPath,
		AgentResult: ar,
	}

	min := task.TimeoutMinutes(45)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(min)*time.Minute)
	defer cancel()

	m := task.SafeOutputsMap()
	if m != nil && len(m) > 0 {
		progressf(opts, "safe-outputs: applying allowed GitHub actions from output.json")
	}
	if outErr := output.RunSuccessOutputs(ctx, glob, task, tc, ar); outErr != nil {
		addRunErr(result, outErr)
		result.Duration = time.Since(start)
		concludeRun(result, &concludeArgs{
			runSucceeded:  false,
			tc:            tc,
			glob:          glob,
			task:          task,
			wm:            wm,
			repoRoot:      repoRoot,
			branchCreated: branchCreated,
			prevBranch:    prevBranch,
			rd:            rd,
		})
		return result, outErr
	}

	result.Success = true
	result.Duration = time.Since(start)
	concludeRun(result, &concludeArgs{
		runSucceeded:  true,
		tc:            tc,
		glob:          glob,
		task:          task,
		wm:            wm,
		repoRoot:      repoRoot,
		branchCreated: branchCreated,
		prevBranch:    prevBranch,
		rd:            rd,
	})
	return result, nil
}

func loadPersistedAgentResult(runDirPath string) (*types.AgentResult, error) {
	path := filepath.Join(runDirPath, resultFileName)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("process-outputs: missing %s (expected after gh wm run --agent-only)", resultFileName)
		}
		return nil, fmt.Errorf("process-outputs: read result.json: %w", err)
	}
	var file struct {
		Success     *bool `json:"success"`
		AgentResult *struct {
			Success            bool   `json:"success"`
			ExitCode           int    `json:"exit_code"`
			Stdout             string `json:"stdout,omitempty"`
			Stderr             string `json:"stderr,omitempty"`
			Summary            string `json:"summary,omitempty"`
			TimedOut           bool   `json:"timed_out,omitempty"`
			AgentStdoutPath    string `json:"agent_stdout_path,omitempty"`
			OutputFilePath     string `json:"output_file_path,omitempty"`
			SafeOutputFilePath string `json:"safe_output_file_path,omitempty"`
		} `json:"agent_result"`
	}
	if err := json.Unmarshal(b, &file); err != nil {
		return nil, fmt.Errorf("process-outputs: parse result.json: %w", err)
	}
	if file.AgentResult == nil {
		return nil, fmt.Errorf("process-outputs: result.json has no agent_result snapshot")
	}
	if file.Success != nil && !*file.Success {
		return nil, fmt.Errorf("process-outputs: refusing to apply outputs: result.json success is false")
	}
	if !file.AgentResult.Success {
		return nil, fmt.Errorf("process-outputs: refusing to apply outputs: agent_result.success is false")
	}
	ar := &types.AgentResult{
		Success:            file.AgentResult.Success,
		ExitCode:           file.AgentResult.ExitCode,
		Stdout:             file.AgentResult.Stdout,
		Stderr:             file.AgentResult.Stderr,
		Summary:            file.AgentResult.Summary,
		TimedOut:           file.AgentResult.TimedOut,
		AgentStdoutPath:    file.AgentResult.AgentStdoutPath,
		OutputFilePath:     file.AgentResult.OutputFilePath,
		SafeOutputFilePath: file.AgentResult.SafeOutputFilePath,
	}
	rd := &RunDir{Path: runDirPath}
	if ar.OutputFilePath == "" {
		ar.OutputFilePath = rd.OutputJSONPath()
	}
	if ar.SafeOutputFilePath == "" {
		ar.SafeOutputFilePath = rd.SafeOutputJSONLPath()
	}
	return ar, nil
}
