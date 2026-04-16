package output

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

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

func TestRunCommentFromItem_Issue(t *testing.T) {
	fakeGhForComments(t)
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", IssueNumber: 3}
	if err := runCommentFromItem(context.Background(), tc, ItemAddComment{Body: "done"}); err != nil {
		t.Fatal(err)
	}
}

func TestRunCommentFromItem_PR(t *testing.T) {
	fakeGhForComments(t)
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", PRNumber: 2, IssueNumber: 0}
	if err := runCommentFromItem(context.Background(), tc, ItemAddComment{Body: "out only"}); err != nil {
		t.Fatal(err)
	}
}

func TestRunCommentFromItem_EmptyBody(t *testing.T) {
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", IssueNumber: 1}
	if err := runCommentFromItem(context.Background(), tc, ItemAddComment{Body: ""}); err == nil {
		t.Fatal("expected error")
	}
}
