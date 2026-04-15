package engine

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/an-lee/gh-wm/internal/types"
)

func installFakeGH(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("unix fake gh only")
	}
	dir := t.TempDir()
	gh := filepath.Join(dir, "gh")
	script := `#!/bin/sh
set -e
if [ "$1" != "api" ]; then
  exit 1
fi
if echo "$2" | grep -q '/issues/[0-9]*/comments$' && ! echo "$*" | grep -q -- '-X'; then
  echo '[{"body":"<!-- wm-checkpoint: {\"summary\":\"from-comment\"} -->"}]'
  exit 0
fi
if echo "$*" | grep -q -- '-X POST' && echo "$*" | grep -q '/comments'; then
  cat >/dev/null
  exit 0
fi
exit 1
`
	if err := os.WriteFile(gh, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestLoadCheckpointHint_FromComments(t *testing.T) {
	installFakeGH(t)
	t.Setenv("WM_CHECKPOINT", "1")
	t.Cleanup(func() { _ = os.Unsetenv("WM_CHECKPOINT") })
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 7}
	loadCheckpointHint(tc)
	if tc.CheckpointHint != "from-comment" {
		t.Fatalf("got %q", tc.CheckpointHint)
	}
}
