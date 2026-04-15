// Package types defines shared types for triggers, outputs, and engine.
package types

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
}
