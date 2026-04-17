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

func installFakeGHForLabels(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("unix shell fake gh only")
	}
	dir := t.TempDir()
	gh := filepath.Join(dir, "gh")
	script := `#!/bin/sh
set -e
if [ "$1" = "api" ]; then
  if echo "$*" | grep -q 'POST' && echo "$*" | grep -q '/issues/'; then
    exit 0
  fi
  if echo "$*" | grep -q '/repos/' && echo "$*" | grep -q '/labels/'; then
    echo '{"name":"x"}'
    exit 0
  fi
fi
exit 1
`
	if err := os.WriteFile(gh, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GH_WM_REST", "")
}

func TestRunAddLabelsFromItem_AddsLabels(t *testing.T) {
	installFakeGHForLabels(t)
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{},
		},
	}}
	p := newPolicy(task)
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", IssueNumber: 3}
	item := ItemLabels{Labels: []string{"a", "b"}}
	if err := runAddLabelsFromItem(context.Background(), tc, p, item); err != nil {
		t.Fatal(err)
	}
}
