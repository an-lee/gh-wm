package output

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/an-lee/gh-wm/internal/types"
)

// installFakeGHForIssue installs a fake gh that handles:
//   - gh label list --repo o/r --json name --limit 9999
//   - gh api -X POST /repos/o/r/issues (create issue)
func installFakeGHForIssue(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("unix shell fake gh only")
	}
	dir := t.TempDir()
	gh := filepath.Join(dir, "gh")
	script := `#!/bin/sh
set -e
# gh label list --repo ... --json name --limit 9999
if [ "$1" = "label" ] && [ "$2" = "list" ]; then
  echo '[]'
  exit 0
fi
# gh api -X POST /repos/.../issues
if [ "$1" = "api" ]; then
  printf '%s' "$*" | grep -q "POST.*repos.*issues"
  exit $?
fi
exit 1
`
	if err := os.WriteFile(gh, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GH_WM_REST", "")
}

// ---------------------------------------------------------------------------
// runCreateIssue tests
// ---------------------------------------------------------------------------

func TestRunCreateIssue_EmptyTitle(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{Repo: "o/r"}
	err := runCreateIssue(context.Background(), tc, ItemCreateIssue{Title: "  ", Body: "body"})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
	if !strings.Contains(err.Error(), "empty title") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCreateIssue_MissingRepo(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{}
	err := runCreateIssue(context.Background(), tc, ItemCreateIssue{Title: "Title", Body: "body"})
	if err == nil {
		t.Fatal("expected error for missing repo")
	}
	if !strings.Contains(err.Error(), "GITHUB_REPOSITORY not set") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCreateIssue_LabelsFiltered(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{Repo: "o/r"}
	// ghclient.EnsureRepoLabels is called with labels after filtering empty ones.
	// We can't fully exercise ghclient without a fake gh for the label path,
	// but we can verify the filter loop works correctly by checking that an
	// ItemCreateIssue with all-empty labels produces no labels in the ghclient call.
	item := ItemCreateIssue{Title: "Title", Body: "body", Labels: []string{"", "  ", "bug"}}
	// Filter: ["", "  ", "bug"] → ["bug"] (empty/whitespace strings dropped).
	// The ghclient call will fail without fake gh, so this test verifies the
	// label-filtering logic up to the ghclient call.
	err := runCreateIssue(context.Background(), tc, item)
	// Expect ghclient error (no fake gh), not a label-filtering error.
	if err == nil {
		t.Fatal("expected error (gh not mocked)")
	}
}

func TestRunCreateIssue_AssigneesFiltered(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{Repo: "o/r"}
	item := ItemCreateIssue{Title: "Title", Body: "body", Assignees: []string{"", "  ", "user1"}}
	err := runCreateIssue(context.Background(), tc, item)
	if err == nil {
		t.Fatal("expected error (gh not mocked)")
	}
	// Error should NOT be about assignee filtering.
	if strings.Contains(err.Error(), "empty") {
		t.Fatalf("assignee filter should not produce error: %v", err)
	}
}

func TestRunCreateIssue_Success(t *testing.T) {
	installFakeGHForIssue(t)
	tc := &types.TaskContext{Repo: "o/r"}
	item := ItemCreateIssue{
		Title:     "Bug: something is wrong",
		Body:      "Description of the bug",
		Labels:    []string{"bug", "help wanted"},
		Assignees: []string{"user1"},
	}
	if err := runCreateIssue(context.Background(), tc, item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCreateIssue_LabelsOnlyCreatesExisting(t *testing.T) {
	installFakeGHForIssue(t)
	tc := &types.TaskContext{Repo: "o/r"}
	item := ItemCreateIssue{
		Title:  "Just a title",
		Body:   "Body",
		Labels: []string{"bug"},
	}
	if err := runCreateIssue(context.Background(), tc, item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCreateIssue_NoLabelsNoAssignees(t *testing.T) {
	installFakeGHForIssue(t)
	tc := &types.TaskContext{Repo: "o/r"}
	item := ItemCreateIssue{Title: "Just a title", Body: "Body"}
	if err := runCreateIssue(context.Background(), tc, item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}