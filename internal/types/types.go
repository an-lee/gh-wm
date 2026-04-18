// Package types defines shared types for triggers, outputs, and engine.
package types

import "time"

// Phase identifies which stage of RunTask produced the current outcome.
type Phase string

const (
	PhaseActivation Phase = "activation"
	PhaseAgent      Phase = "agent"
	PhaseValidation Phase = "validation"
	PhaseOutputs    Phase = "safe-outputs"
	PhaseConclusion Phase = "conclusion"
)

// RunResult is the structured outcome of RunTask (phases, errors, timing).
type RunResult struct {
	AgentResult *AgentResult
	Phase       Phase // phase where execution stopped or last completed phase on success
	Success     bool
	Errors      []error
	Duration    time.Duration
	// RunDir is the per-run artifact directory (.wm/runs/... or WM_RUN_DIR/...), if created.
	RunDir string
}

// GitHubEvent wraps event name and parsed payload (github.event JSON).
type GitHubEvent struct {
	Name    string
	Payload map[string]any
}

// TaskContext is built when a task matches an event.
type TaskContext struct {
	TaskName string
	Repo     string // owner/repo
	RepoPath string // local checkout path (Actions) or cwd
	Event    *GitHubEvent
	// IssueNumber, PRNumber, etc. extracted from payload
	IssueNumber  int
	PRNumber     int
	CommentID    int64
	LabelName    string
	ScheduleCron string // when event is schedule, which cron fired
	// CheckpointHint is injected into the agent prompt when WM_CHECKPOINT=1 and a prior checkpoint exists.
	CheckpointHint string
}

// AgentResult is produced after running the agent subprocess.
type AgentResult struct {
	Success  bool
	Stdout   string
	Stderr   string
	Summary  string
	ExitCode int
	// TimedOut is true when the run ended because the context deadline was exceeded.
	TimedOut bool
	// AgentStdoutPath is the on-disk combined agent log when RunDir is used (full transcript).
	AgentStdoutPath string
	// SafeOutputFilePath is the NDJSON log from `gh wm emit` (output.jsonl), read by RunSuccessOutputs.
	SafeOutputFilePath string
	// LastResponseText is the final assistant text from Claude print-mode conversation.json(l) (result.result), when available.
	LastResponseText string
}
