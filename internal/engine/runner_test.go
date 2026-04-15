package engine

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/an-lee/gh-wm/internal/types"
)

func TestRunTask_Minimal(t *testing.T) {
	t.Setenv("WM_AGENT_CMD", "true")
	t.Cleanup(func() { _ = os.Unsetenv("WM_AGENT_CMD") })
	root := writeMinimalRepo(t)
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"action": "opened"}}
	res, err := RunTask(context.Background(), root, "a", ev)
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || !res.Success {
		t.Fatalf("%+v", res)
	}
}

func TestRunTask_NotFound(t *testing.T) {
	root := writeMinimalRepo(t)
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{}}
	if _, err := RunTask(context.Background(), root, "missing", ev); err == nil {
		t.Fatal("expected error")
	}
}

func TestRunTask_LoadError(t *testing.T) {
	if _, err := RunTask(context.Background(), "/nonexistent-root-12345", "x", &types.GitHubEvent{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestPostCheckpoint_NoEnv(t *testing.T) {
	t.Setenv("WM_CHECKPOINT", "")
	t.Cleanup(func() { _ = os.Unsetenv("WM_CHECKPOINT") })
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 1}
	postCheckpoint(tc, &types.AgentResult{Summary: "s"})
}

func TestPostCheckpoint_NoRepo(t *testing.T) {
	t.Setenv("WM_CHECKPOINT", "1")
	t.Cleanup(func() { _ = os.Unsetenv("WM_CHECKPOINT") })
	postCheckpoint(&types.TaskContext{IssueNumber: 1}, &types.AgentResult{Summary: "s"})
}

func TestLoadCheckpointHint_NoEnv(t *testing.T) {
	t.Setenv("WM_CHECKPOINT", "")
	t.Cleanup(func() { _ = os.Unsetenv("WM_CHECKPOINT") })
	tc := &types.TaskContext{}
	loadCheckpointHint(tc)
}

func TestLoadCheckpointHint_NoIssue(t *testing.T) {
	t.Setenv("WM_CHECKPOINT", "1")
	t.Cleanup(func() { _ = os.Unsetenv("WM_CHECKPOINT") })
	loadCheckpointHint(&types.TaskContext{Repo: "o/r"})
}

func TestPostCheckpoint_TruncatesSummary(t *testing.T) {
	t.Setenv("WM_CHECKPOINT", "1")
	t.Cleanup(func() { _ = os.Unsetenv("WM_CHECKPOINT") })
	long := strings.Repeat("x", 2500)
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 1}
	postCheckpoint(tc, &types.AgentResult{Summary: long})
}

func TestPostCheckpoint_UsesStdoutWhenSummaryEmpty(t *testing.T) {
	t.Setenv("WM_CHECKPOINT", "1")
	t.Cleanup(func() { _ = os.Unsetenv("WM_CHECKPOINT") })
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 1}
	postCheckpoint(tc, &types.AgentResult{Stdout: "from stdout"})
}
