package output

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/an-lee/gh-wm/internal/types"
)

func installFakeGHForComment(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("unix shell fake gh only")
	}
	dir := t.TempDir()
	gh := filepath.Join(dir, "gh")
	script := `#!/bin/sh
set -e
if [ "$1" = "api" ] && echo "$*" | grep -q -- '-X POST' && echo "$*" | grep -q '/comments'; then
  cat >/dev/null
  exit 0
fi
exit 1
`
	if err := os.WriteFile(gh, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GH_WM_REST", "") // ensure exec path
}

func TestRunCommentFromItem_Issue(t *testing.T) {
	installFakeGHForComment(t)
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", IssueNumber: 1}
	if err := runCommentFromItem(context.Background(), tc, ItemAddComment{Body: "hello"}); err != nil {
		t.Fatal(err)
	}
}

func TestRunCommentFromItem_PR(t *testing.T) {
	installFakeGHForComment(t)
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", PRNumber: 2}
	if err := runCommentFromItem(context.Background(), tc, ItemAddComment{Body: "hello"}); err != nil {
		t.Fatal(err)
	}
}

func TestRunCommentFromItem_EmptyBody(t *testing.T) {
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", IssueNumber: 1}
	if err := runCommentFromItem(context.Background(), tc, ItemAddComment{Body: ""}); err == nil {
		t.Fatal("expected error")
	}
}
