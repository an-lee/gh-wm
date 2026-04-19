package output

import (
	"context"
	"strings"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

// ---------------------------------------------------------------------------
// postComment error paths
// ---------------------------------------------------------------------------

func TestPostComment_MissingRepo(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{}
	err := postComment(tc, 1, "hello")
	if err == nil {
		t.Fatal("expected error for missing repo")
	}
	if !strings.Contains(err.Error(), "GITHUB_REPOSITORY not set") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostComment_Truncation(t *testing.T) {
	t.Parallel()
	installFakeGHForComment(t)
	tc := &types.TaskContext{Repo: "o/r"}
	// Body exactly maxCommentBody bytes → no truncation.
	body := strings.Repeat("x", maxCommentBody)
	err := postComment(tc, 1, body)
	if err != nil {
		t.Fatalf("body == maxCommentBody should succeed: %v", err)
	}
	// Body one byte over the limit → truncation marker appended.
	longBody := strings.Repeat("x", maxCommentBody+1)
	err = postComment(tc, 1, longBody)
	if err != nil {
		t.Fatalf("truncated body should not error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// postFallbackComment paths
// ---------------------------------------------------------------------------

func TestPostFallbackComment_EmptyBody(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 1}
	// Empty/whitespace body returns nil without calling postComment.
	err := postFallbackComment(tc, "")
	if err != nil {
		t.Fatalf("empty response should return nil: %v", err)
	}
	err = postFallbackComment(tc, "   ")
	if err != nil {
		t.Fatalf("whitespace-only response should return nil: %v", err)
	}
}

func TestPostFallbackComment_MissingIssueAndPR(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{Repo: "o/r"}
	// No issue number and no PR number → error.
	err := postFallbackComment(tc, "some response")
	if err == nil {
		t.Fatal("expected error when no issue or PR number")
	}
	if !strings.Contains(err.Error(), "no issue or PR number") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostFallbackComment_UsesPRNumber(t *testing.T) {
	installFakeGHForComment(t)
	tc := &types.TaskContext{Repo: "o/r", PRNumber: 42}
	// Falls back to PRNumber when IssueNumber is 0.
	err := postFallbackComment(tc, "response body")
	if err != nil {
		t.Fatalf("postFallbackComment with PRNumber should succeed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// resolveCommentTarget
// ---------------------------------------------------------------------------

func TestResolveCommentTarget_Explicit(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{IssueNumber: 1, PRNumber: 2}
	got := resolveCommentTarget(tc, 99)
	if got != 99 {
		t.Fatalf("explicit target should take priority: got %d, want 99", got)
	}
}

func TestResolveCommentTarget_ZeroUsesIssue(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{IssueNumber: 5, PRNumber: 2}
	got := resolveCommentTarget(tc, 0)
	if got != 5 {
		t.Fatalf("zero target should use IssueNumber: got %d, want 5", got)
	}
}

func TestResolveCommentTarget_ZeroUsesPRWhenNoIssue(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{IssueNumber: 0, PRNumber: 7}
	got := resolveCommentTarget(tc, 0)
	if got != 7 {
		t.Fatalf("zero target with no IssueNumber should use PRNumber: got %d, want 7", got)
	}
}

func TestResolveCommentTarget_ZeroBoth(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{IssueNumber: 0, PRNumber: 0}
	got := resolveCommentTarget(tc, 0)
	if got != 0 {
		t.Fatalf("zero target with no numbers should return 0: got %d", got)
	}
}

// ---------------------------------------------------------------------------
// commentTargetNumber
// ---------------------------------------------------------------------------

func TestCommentTargetNumber_PRNumber(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{IssueNumber: 0, PRNumber: 10}
	got := commentTargetNumber(tc)
	if got != 10 {
		t.Fatalf("got %d, want 10", got)
	}
}

func TestCommentTargetNumber_IssueNumberFallback(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{IssueNumber: 7, PRNumber: 0}
	got := commentTargetNumber(tc)
	if got != 7 {
		t.Fatalf("got %d, want 7", got)
	}
}

func TestCommentTargetNumber_BothZero(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{IssueNumber: 0, PRNumber: 0}
	got := commentTargetNumber(tc)
	if got != 0 {
		t.Fatalf("got %d, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// runAddLabelsFromItem error paths
// ---------------------------------------------------------------------------

func TestRunAddLabelsFromItem_EmptyLabels(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{"add-labels": map[string]any{}},
	}}
	p := newPolicy(task)
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 3}
	err := runAddLabelsFromItem(context.Background(), tc, p, ItemLabels{Labels: []string{}})
	if err == nil {
		t.Fatal("expected error for empty labels")
	}
	if !strings.Contains(err.Error(), "empty labels") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunAddLabelsFromItem_MissingIssueNumber(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{"add-labels": map[string]any{}},
	}}
	p := newPolicy(task)
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 0, PRNumber: 0}
	err := runAddLabelsFromItem(context.Background(), tc, p, ItemLabels{Labels: []string{"bug"}})
	if err == nil {
		t.Fatal("expected error for missing issue/PR number")
	}
	if !strings.Contains(err.Error(), "no issue/PR number") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunAddLabelsFromItem_MissingRepo(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{"add-labels": map[string]any{}},
	}}
	p := newPolicy(task)
	tc := &types.TaskContext{Repo: "", IssueNumber: 3}
	err := runAddLabelsFromItem(context.Background(), tc, p, ItemLabels{Labels: []string{"bug"}})
	if err == nil {
		t.Fatal("expected error for missing repo")
	}
	if !strings.Contains(err.Error(), "no issue/PR number or repository") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunAddLabelsFromItem_DisallowedLabel(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{"add-labels": map[string]any{
			"allow": []any{"bug"},
		}},
	}}
	p := newPolicy(task)
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 3}
	err := runAddLabelsFromItem(context.Background(), tc, p, ItemLabels{Labels: []string{"security"}})
	if err == nil {
		t.Fatal("expected error for disallowed label")
	}
	if !strings.Contains(err.Error(), "not allowed by policy") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunAddLabelsFromItem_MixedAllowedAndEmpty(t *testing.T) {
	t.Parallel()
	installFakeGHForLabels(t)
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{"add-labels": map[string]any{}},
	}}
	p := newPolicy(task)
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 3}
	// Mix of empty strings (filtered out) and valid labels.
	err := runAddLabelsFromItem(context.Background(), tc, p, ItemLabels{Labels: []string{"", "  ", "bug"}})
	if err != nil {
		t.Fatalf("empty/whitespace labels should be filtered: %v", err)
	}
}

// ---------------------------------------------------------------------------
// resolveLabelTarget
// ---------------------------------------------------------------------------

func TestResolveLabelTarget_Explicit(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{IssueNumber: 1, PRNumber: 2}
	got := resolveLabelTarget(tc, 99)
	if got != 99 {
		t.Fatalf("explicit target should take priority: got %d, want 99", got)
	}
}

func TestResolveLabelTarget_ZeroUsesIssue(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{IssueNumber: 5, PRNumber: 2}
	got := resolveLabelTarget(tc, 0)
	if got != 5 {
		t.Fatalf("zero target should use IssueNumber: got %d, want 5", got)
	}
}

func TestResolveLabelTarget_ZeroUsesPRWhenNoIssue(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{IssueNumber: 0, PRNumber: 7}
	got := resolveLabelTarget(tc, 0)
	if got != 7 {
		t.Fatalf("zero target with no IssueNumber should use PRNumber: got %d, want 7", got)
	}
}

func TestResolveLabelTarget_ZeroBoth(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{IssueNumber: 0, PRNumber: 0}
	got := resolveLabelTarget(tc, 0)
	if got != 0 {
		t.Fatalf("zero target with no numbers should return 0: got %d", got)
	}
}

// ---------------------------------------------------------------------------
// WMAgentCommentMarkerFooter edge cases
// ---------------------------------------------------------------------------

func TestWMAgentCommentMarkerFooter_DefaultTaskName(t *testing.T) {
	t.Parallel()
	got := WMAgentCommentMarkerFooter("")
	if !strings.Contains(got, "unknown") {
		t.Fatalf("empty task name should default to unknown: %q", got)
	}
	if !strings.Contains(got, "wm-agent:") {
		t.Fatalf("should contain wm-agent marker: %q", got)
	}
}

func TestWMAgentCommentMarkerFooter_WhitespaceName(t *testing.T) {
	t.Parallel()
	got := WMAgentCommentMarkerFooter("   ")
	if !strings.Contains(got, "unknown") {
		t.Fatalf("whitespace task name should default to unknown: %q", got)
	}
}
