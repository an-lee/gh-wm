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

func fakeGhAPI(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("unix fake gh only")
	}
	dir := t.TempDir()
	gh := filepath.Join(dir, "gh")
	script := `#!/bin/sh
if [ "$1" = "api" ] && echo "$*" | grep -q -- '-X POST' && echo "$*" | grep -q '/labels'; then
  exit 0
fi
exit 1
`
	if err := os.WriteFile(gh, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestRunLabelOutput_AddsLabels(t *testing.T) {
	fakeGhAPI(t)
	task := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"add-labels": map[string]any{"labels": []any{"bug", "triage"}},
	}}}
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 9}
	if err := runLabelOutputLegacy(context.Background(), task, tc); err != nil {
		t.Fatal(err)
	}
}
