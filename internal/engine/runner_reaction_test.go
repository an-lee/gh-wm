package engine

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

// withFakeGHReactions extends the standard withFakeGH to also handle reaction API calls.
func withFakeGHReactions(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("unix shell fake gh only")
	}
	dir := t.TempDir()
	gh := filepath.Join(dir, "gh")
	script := `#!/bin/sh
set -e
# gh repo view
if [ "$1" = "repo" ] && [ "$2" = "view" ]; then
  echo 'test-owner/test-repo'
  exit 0
fi
if [ "$1" != "api" ]; then
  exit 1
fi
# POST reaction (issue or comment)
if echo "$*" | grep -q -- '-X POST' && echo "$*" | grep -q '/reactions'; then
  cat >/dev/null
  exit 0
fi
# GET comments list (checkpoint)
if echo "$2" | grep -q '/issues/[0-9]*/comments$' && ! echo "$*" | grep -q -- '-X'; then
  echo '[]'
  exit 0
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
// applyOnReactionBestEffort tests
// ---------------------------------------------------------------------------

func TestApplyOnReactionBestEffort_NilTask(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{Repo: "o/r"}
	result := &types.RunResult{}
	applyOnReactionBestEffort(nil, nil, tc, result)
	if len(result.Errors) != 0 {
		t.Fatalf("nil task should not add errors, got: %v", result.Errors)
	}
}

func TestApplyOnReactionBestEffort_NilTC(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{"on": map[string]any{"reaction": "+1"}}}
	result := &types.RunResult{}
	applyOnReactionBestEffort(nil, task, nil, result)
	if len(result.Errors) != 0 {
		t.Fatalf("nil tc should not add errors, got: %v", result.Errors)
	}
}

func TestApplyOnReactionBestEffort_NilResult(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{"on": map[string]any{"reaction": "+1"}}}
	tc := &types.TaskContext{Repo: "o/r"}
	// Should not panic.
	applyOnReactionBestEffort(nil, task, tc, nil)
}

func TestApplyOnReactionBestEffort_EmptyContent(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{"on": map[string]any{"reaction": ""}}}
	tc := &types.TaskContext{}
	result := &types.RunResult{}
	applyOnReactionBestEffort(nil, task, tc, result)
	if len(result.Errors) != 0 {
		t.Fatalf("empty content should not add errors, got: %v", result.Errors)
	}
}

func TestApplyOnReactionBestEffort_WhitespaceContent(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{"on": map[string]any{"reaction": "   "}}}
	tc := &types.TaskContext{Repo: "o/r"}
	result := &types.RunResult{}
	applyOnReactionBestEffort(nil, task, tc, result)
	if len(result.Errors) != 0 {
		t.Fatalf("whitespace-only content should not add errors, got: %v", result.Errors)
	}
}

func TestApplyOnReactionBestEffort_EmptyRepo(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{"on": map[string]any{"reaction": "+1"}}}
	tc := &types.TaskContext{Repo: "  "}
	result := &types.RunResult{}
	applyOnReactionBestEffort(nil, task, tc, result)
	if len(result.Errors) != 0 {
		t.Fatalf("empty repo should not add errors, got: %v", result.Errors)
	}
}

func TestApplyOnReactionBestEffort_NilEvent(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{"on": map[string]any{"reaction": "+1"}}}
	tc := &types.TaskContext{Repo: "o/r"}
	result := &types.RunResult{}
	applyOnReactionBestEffort(nil, task, tc, result)
	if len(result.Errors) != 0 {
		t.Fatalf("nil event should not add errors, got: %v", result.Errors)
	}
}

func TestApplyOnReactionBestEffort_EmptyEventName(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{"on": map[string]any{"reaction": "+1"}}}
	tc := &types.TaskContext{Repo: "o/r"}
	result := &types.RunResult{}
	applyOnReactionBestEffort(nil, task, tc, result)
	if len(result.Errors) != 0 {
		t.Fatalf("empty event name should not add errors, got: %v", result.Errors)
	}
}

func TestApplyOnReactionBestEffort_NoCommentIDOrIssueNumber(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{"on": map[string]any{"reaction": "+1"}}}
	tc := &types.TaskContext{Repo: "o/r"}
	result := &types.RunResult{}
	applyOnReactionBestEffort(nil, task, tc, result)
	// No comment ID and no issue/PR number → returns without calling ghclient.
	if len(result.Errors) != 0 {
		t.Fatalf("no target should not add errors, got: %v", result.Errors)
	}
}

func TestApplyOnReactionBestEffort_IssueCommentWithCommentID(t *testing.T) {
	withFakeGHReactions(t)
	task := &config.Task{Frontmatter: map[string]any{"on": map[string]any{"reaction": "+1"}}}
	tc := &types.TaskContext{Repo: "o/r", CommentID: 12345, Event: &types.GitHubEvent{Name: "issue_comment"}}
	result := &types.RunResult{}
	applyOnReactionBestEffort(nil, task, tc, result)
	if len(result.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
}

func TestApplyOnReactionBestEffort_IssuesWithIssueNumber(t *testing.T) {
	withFakeGHReactions(t)
	task := &config.Task{Frontmatter: map[string]any{"on": map[string]any{"reaction": "🎉"}}}
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 42, Event: &types.GitHubEvent{Name: "issues"}}
	result := &types.RunResult{}
	applyOnReactionBestEffort(nil, task, tc, result)
	if len(result.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
}

func TestApplyOnReactionBestEffort_PullRequestWithPRNumber(t *testing.T) {
	withFakeGHReactions(t)
	task := &config.Task{Frontmatter: map[string]any{"on": map[string]any{"reaction": "+1"}}}
	tc := &types.TaskContext{Repo: "o/r", PRNumber: 99, Event: &types.GitHubEvent{Name: "pull_request"}}
	result := &types.RunResult{}
	applyOnReactionBestEffort(nil, task, tc, result)
	if len(result.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
}

func TestApplyOnReactionBestEffort_GhClientError(t *testing.T) {
	// Without fake gh for reactions, ghclient calls fail and error is recorded.
	task := &config.Task{Frontmatter: map[string]any{"on": map[string]any{"reaction": "+1"}}}
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 1, Event: &types.GitHubEvent{Name: "issues"}}
	result := &types.RunResult{}
	applyOnReactionBestEffort(nil, task, tc, result)
	if len(result.Errors) == 0 {
		t.Fatal("expected error when ghclient fails")
	}
	if !strings.Contains(result.Errors[0].Error(), "on.reaction") {
		t.Fatalf("error should contain 'on.reaction': %v", result.Errors[0])
	}
}

// ---------------------------------------------------------------------------
// issueOrPRNumber tests
// ---------------------------------------------------------------------------

func TestIssueOrPRNumber_Nil(t *testing.T) {
	t.Parallel()
	if got := issueOrPRNumber(nil); got != 0 {
		t.Fatalf("nil tc: got %d, want 0", got)
	}
}

func TestIssueOrPRNumber_IssueNumber(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{IssueNumber: 5}
	if got := issueOrPRNumber(tc); got != 5 {
		t.Fatalf("got %d, want 5", got)
	}
}

func TestIssueOrPRNumber_PRNumber(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{PRNumber: 7}
	if got := issueOrPRNumber(tc); got != 7 {
		t.Fatalf("got %d, want 7", got)
	}
}

func TestIssueOrPRNumber_IssueNumberPreferred(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{IssueNumber: 5, PRNumber: 7}
	if got := issueOrPRNumber(tc); got != 5 {
		t.Fatalf("issue number should be preferred: got %d, want 5", got)
	}
}
