package output

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

func fakeGhForComments(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("unix fake gh only")
	}
	dir := t.TempDir()
	gh := filepath.Join(dir, "gh")
	script := `#!/bin/sh
if [ "$1" = "issue" ] && [ "$2" = "comment" ]; then
  exit 0
fi
if [ "$1" = "pr" ] && [ "$2" = "comment" ]; then
  exit 0
fi
exit 1
`
	if err := os.WriteFile(gh, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestRunCommentOutput_Issue(t *testing.T) {
	fakeGhForComments(t)
	task := &config.Task{Name: "t"}
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", IssueNumber: 3}
	res := &types.AgentResult{Summary: "done"}
	if err := runCommentOutputLegacy(context.Background(), nil, task, tc, res); err != nil {
		t.Fatal(err)
	}
}

func TestRunCommentOutput_PR(t *testing.T) {
	fakeGhForComments(t)
	task := &config.Task{Name: "t"}
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", PRNumber: 2, IssueNumber: 0}
	res := &types.AgentResult{Stdout: "out only"}
	if err := runCommentOutputLegacy(context.Background(), nil, task, tc, res); err != nil {
		t.Fatal(err)
	}
}

func TestRunCommentOutput_EmptyOutputFallback(t *testing.T) {
	fakeGhForComments(t)
	task := &config.Task{Name: "named"}
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", IssueNumber: 1}
	res := &types.AgentResult{}
	if err := runCommentOutputLegacy(context.Background(), nil, task, tc, res); err != nil {
		t.Fatal(err)
	}
}
