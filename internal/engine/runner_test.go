package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/an-lee/gh-wm/internal/types"
)

// withFakeGH prepends a fake gh that instantly succeeds for common api calls used in tests.
func withFakeGH(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake gh script is unix-only")
	}
	dir := t.TempDir()
	gh := filepath.Join(dir, "gh")
	script := `#!/bin/sh
set -e
# gh repo view --json nameWithOwner
if [ "$1" = "repo" ] && [ "$2" = "view" ]; then
  echo 'test-owner/test-repo'
  exit 0
fi
if [ "$1" != "api" ]; then
  exit 1
fi
# GET comments list
if echo "$2" | grep -q '/issues/[0-9]*/comments$' && ! echo "$*" | grep -q -- '-X'; then
  echo '[{"body":"<!-- wm-checkpoint: {\"summary\":\"checkpoint summary\"} -->"}]'
  exit 0
fi
# POST comment with stdin
if echo "$*" | grep -q -- '-X POST' && echo "$*" | grep -q '/comments'; then
  cat >/dev/null
  exit 0
fi
# POST label
if echo "$*" | grep -q -- '-X POST' && echo "$*" | grep -q '/labels'; then
  exit 0
fi
# DELETE label
if echo "$*" | grep -q -- '-X DELETE' && echo "$*" | grep -q '/labels/'; then
  exit 0
fi
exit 1
`
	if err := os.WriteFile(gh, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestAddRunErr(t *testing.T) {
	t.Parallel()
	r := &types.RunResult{}
	addRunErr(r, nil)
	addRunErr(r, fmt.Errorf("e1"))
	if len(r.Errors) != 1 {
		t.Fatalf("got %d errors", len(r.Errors))
	}
	addRunErr(nil, fmt.Errorf("ignored"))
}

func TestRunTask_Minimal(t *testing.T) {
	t.Setenv("WM_AGENT_CMD", "true")
	t.Cleanup(func() { _ = os.Unsetenv("WM_AGENT_CMD") })
	root := writeMinimalRepo(t)
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{"action": "opened"}}
	out, err := RunTask(context.Background(), root, "a", ev, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out == nil || !out.Success || out.AgentResult == nil || !out.AgentResult.Success {
		t.Fatalf("%+v", out)
	}
}

func TestRunTask_NotFound(t *testing.T) {
	root := writeMinimalRepo(t)
	ev := &types.GitHubEvent{Name: "issues", Payload: map[string]any{}}
	if _, err := RunTask(context.Background(), root, "missing", ev, nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestRunTask_LoadError(t *testing.T) {
	if _, err := RunTask(context.Background(), "/nonexistent-root-12345", "x", &types.GitHubEvent{}, nil); err == nil {
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
	withFakeGH(t)
	t.Setenv("WM_CHECKPOINT", "1")
	t.Cleanup(func() { _ = os.Unsetenv("WM_CHECKPOINT") })
	long := strings.Repeat("x", 2500)
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 1}
	postCheckpoint(tc, &types.AgentResult{Summary: long})
}

func TestPostCheckpoint_UsesStdoutWhenSummaryEmpty(t *testing.T) {
	withFakeGH(t)
	t.Setenv("WM_CHECKPOINT", "1")
	t.Cleanup(func() { _ = os.Unsetenv("WM_CHECKPOINT") })
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 1}
	postCheckpoint(tc, &types.AgentResult{Stdout: "from stdout"})
}
