package output

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

func TestValidateAndAppend_MaxPerKind(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"add-comment": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{
		Repo:        "o/r",
		RepoPath:    dir,
		TaskName:    "t",
		IssueNumber: 42,
	}
	ctx := context.Background()
	item1 := map[string]any{"body": "one", "target": 0}
	if err := ValidateAndAppend(ctx, g, task, tc, KindAddComment, item1, path); err != nil {
		t.Fatal(err)
	}
	item2 := map[string]any{"body": "two", "target": 0}
	if err := ValidateAndAppend(ctx, g, task, tc, KindAddComment, item2, path); err == nil {
		t.Fatal("expected policy max error")
	}
}

func TestValidateAndAppend_TitlePrefixCreateIssue(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				// Note: scalar trims string fields; use a prefix without trailing space.
				"create-issue": map[string]any{"title-prefix": "[wm]"},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir}
	ctx := context.Background()
	item := map[string]any{"title": "hello", "body": "b"}
	if err := ValidateAndAppend(ctx, g, task, tc, KindCreateIssue, item, path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"title":"[wm]hello"`) {
		t.Fatalf("expected prefixed title in file, got %q", string(b))
	}
}

func TestValidateAndAppend_UpdateIssueTitlePrefix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"update-issue": map[string]any{"title-prefix": "[fix]"},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, IssueNumber: 7}
	ctx := context.Background()
	item := map[string]any{"title": "hello", "body": "", "target": 0}
	if err := ValidateAndAppend(ctx, g, task, tc, KindUpdateIssue, item, path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"title":"[fix]hello"`) {
		t.Fatalf("expected prefixed title in file, got %q", string(b))
	}
}

func TestValidateAndAppend_UpdateIssueInvalidOperation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"update-issue": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, IssueNumber: 3}
	item := map[string]any{"title": "", "body": "b", "operation": "bogus", "target": 0}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindUpdateIssue, item, path)
	if err == nil {
		t.Fatal("expected validation error for operation")
	}
}

func TestValidateAndAppend_CloseIssueInvalidStateReason(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"close-issue": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, IssueNumber: 3}
	item := map[string]any{"target": 0, "state_reason": "bogus"}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindCloseIssue, item, path)
	if err == nil {
		t.Fatal("expected validation error for state_reason")
	}
}

func TestValidateAndAppend_CreatePullRequestReviewCommentInvalidSide(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"create-pull-request-review-comment": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 9}
	item := map[string]any{
		"body": "hi", "commit_id": "abc1234", "path": "f.go", "line": 3, "side": "MIDDLE", "target": 0,
	}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindCreatePullRequestReviewComment, item, path)
	if err == nil {
		t.Fatal("expected validation error for side")
	}
}

func TestValidateAndAppend_CreatePullRequestReviewCommentOK(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"create-pull-request-review-comment": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 9}
	item := map[string]any{
		"body": "hi", "commit_id": "abc1234", "path": "f.go", "line": 3, "side": "right", "target": 0,
	}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindCreatePullRequestReviewComment, item, path); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAndAppend_PushToPullRequestBranchOK(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"push-to-pull-request-branch": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 5}
	item := map[string]any{"target": 0}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindPushToPullRequestBranch, item, path); err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// validIssueCloseReason edge cases
// ---------------------------------------------------------------------------

func TestValidIssueCloseReason_VariousFormats(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  bool
	}{
		{"completed", true},
		{"not_planned", true},
		{"duplicate", true},
		{"COMPLETED", true},
		{"Not_Planned", true},
		{"not-planned", true},
		{"COMPLETED", true},
		{"completed", true},
		{"not planned", true}, // space → underscore normalized
		{"bogus", false},
		{"", false},
		{"maybe", false},
	}
	for _, tc := range tests {
		got := validIssueCloseReason(tc.input)
		if got != tc.want {
			t.Errorf("validIssueCloseReason(%q): got %v, want %v", tc.input, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// validateEmitPayload error paths
// ---------------------------------------------------------------------------

func TestValidateEmitPayload_AddReviewer_EmptyReviewers(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"add-reviewer": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 7}
	item := map[string]any{"reviewers": []any{}, "target": 0}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindAddReviewer, item, path)
	if err == nil {
		t.Fatal("expected validation error for empty reviewers")
	}
}

func TestValidateEmitPayload_AddReviewer_NoRepo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"add-reviewer": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{RepoPath: dir, PRNumber: 7} // no Repo
	item := map[string]any{"reviewers": []any{"alice"}, "target": 0}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindAddReviewer, item, path)
	if err == nil {
		t.Fatal("expected validation error for no repo")
	}
}

func TestValidateEmitPayload_ReplyToPullRequestReviewComment_EmptyBody(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"reply-to-pull-request-review-comment": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 7}
	item := map[string]any{"comment_id": 42, "body": "  ", "target": 0}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindReplyToPullRequestReviewComment, item, path)
	if err == nil {
		t.Fatal("expected validation error for empty body")
	}
}

func TestValidateEmitPayload_ReplyToPullRequestReviewComment_InvalidCommentID(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"reply-to-pull-request-review-comment": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 7}
	item := map[string]any{"comment_id": 0, "body": "hello", "target": 0}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindReplyToPullRequestReviewComment, item, path)
	if err == nil {
		t.Fatal("expected validation error for comment_id==0")
	}
}

func TestValidateEmitPayload_ResolvePullRequestReviewThread_EmptyThreadID(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"resolve-pull-request-review-thread": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 7}
	item := map[string]any{"thread_id": "  ", "target": 0}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindResolvePullRequestReviewThread, item, path)
	if err == nil {
		t.Fatal("expected validation error for empty thread_id")
	}
}

func TestValidateEmitPayload_SubmitPullRequestReview_InvalidEvent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"submit-pull-request-review": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 7}
	item := map[string]any{"event": "MAYBE", "target": 0}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindSubmitPullRequestReview, item, path)
	if err == nil {
		t.Fatal("expected validation error for invalid event")
	}
}

func TestValidateEmitPayload_SubmitPullRequestReview_EmptyEvent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"submit-pull-request-review": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 7}
	item := map[string]any{"event": "  ", "target": 0}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindSubmitPullRequestReview, item, path)
	if err == nil {
		t.Fatal("expected validation error for empty event")
	}
}

func TestValidateEmitPayload_UpdateIssue_BothTitleAndBodyEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"update-issue": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, IssueNumber: 3}
	item := map[string]any{"title": "  ", "body": "  ", "target": 0}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindUpdateIssue, item, path)
	if err == nil {
		t.Fatal("expected validation error for both title and body empty")
	}
}

func TestValidateEmitPayload_UpdateIssue_NoRepo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"update-issue": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{RepoPath: dir, IssueNumber: 3} // no Repo
	item := map[string]any{"title": "hello", "body": "", "target": 0}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindUpdateIssue, item, path)
	if err == nil {
		t.Fatal("expected validation error for no repo")
	}
}

func TestValidateEmitPayload_UpdatePullRequest_BothTitleAndBodyEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"update-pull-request": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 5}
	item := map[string]any{"title": "  ", "body": "  ", "target": 0}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindUpdatePullRequest, item, path)
	if err == nil {
		t.Fatal("expected validation error for both title and body empty")
	}
}

func TestValidateEmitPayload_UpdatePullRequest_NoRepo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"update-pull-request": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{RepoPath: dir, PRNumber: 5} // no Repo
	item := map[string]any{"title": "hello", "body": "", "target": 0}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindUpdatePullRequest, item, path)
	if err == nil {
		t.Fatal("expected validation error for no repo")
	}
}

func TestValidateEmitPayload_CloseIssue_NoRepo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"close-issue": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{RepoPath: dir, IssueNumber: 3} // no Repo
	item := map[string]any{"target": 0, "state_reason": "completed"}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindCloseIssue, item, path)
	if err == nil {
		t.Fatal("expected validation error for no repo")
	}
}

func TestValidateEmitPayload_ClosePullRequest_NoRepo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"close-pull-request": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{RepoPath: dir, PRNumber: 5} // no Repo
	item := map[string]any{"target": 0}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindClosePullRequest, item, path)
	if err == nil {
		t.Fatal("expected validation error for no repo")
	}
}

func TestValidateEmitPayload_CloseIssue_NoIssueNumber(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"close-issue": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir} // no IssueNumber
	item := map[string]any{"target": 0, "state_reason": "completed"}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindCloseIssue, item, path)
	if err == nil {
		t.Fatal("expected validation error for no issue number")
	}
}

func TestValidateEmitPayload_ClosePullRequest_NoPRNumber(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"close-pull-request": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir} // no PRNumber
	item := map[string]any{"target": 0}
	err := ValidateAndAppend(context.Background(), g, task, tc, KindClosePullRequest, item, path)
	if err == nil {
		t.Fatal("expected validation error for no PR number")
	}
}

// ---------------------------------------------------------------------------
// applyPolicyMutations branches
// ---------------------------------------------------------------------------

func TestApplyPolicyMutations_CreatePullRequest_EmptyTitleGetsDefault(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"create-pull-request": map[string]any{},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir}
	item := map[string]any{"title": "  ", "body": "my body"}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindCreatePullRequest, item, path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"title":"[t] wm task"`) {
		t.Fatalf("expected default title in file, got %q", string(b))
	}
}

func TestApplyPolicyMutations_UpdateIssue_BodyOnlyNoPrefix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				// Note: no title-prefix set
				"update-issue": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, IssueNumber: 3}
	item := map[string]any{"body": "updated body", "target": 0}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindUpdateIssue, item, path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"title"`) {
		t.Fatalf("title should not be set when only body is provided, got %q", string(b))
	}
}

func TestApplyPolicyMutations_UpdatePullRequest_TitlePrefixApplied(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"update-pull-request": map[string]any{"title-prefix": "[WIP]"},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 5}
	item := map[string]any{"title": "fix bug", "body": "", "target": 0}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindUpdatePullRequest, item, path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"title":"[WIP]fix bug"`) {
		t.Fatalf("expected prefixed title in file, got %q", string(b))
	}
}

func TestApplyPolicyMutations_CreatePullRequest_EmptyTitleNoPrefix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"create-pull-request": map[string]any{"title-prefix": "[auto]"},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir}
	item := map[string]any{"title": "  ", "body": "body"}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindCreatePullRequest, item, path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"title":"[auto][t] wm task"`) {
		t.Fatalf("expected prefixed default title in file, got %q", string(b))
	}
}

// ---------------------------------------------------------------------------
// ValidateAndAppend edge cases
// ---------------------------------------------------------------------------

func TestValidateAndAppend_CloseIssue_ValidStateReasons(t *testing.T) {
	for _, sr := range []string{"completed", "not_planned", "duplicate", "COMPLETED", "Not-Planned"} {
		sr := sr
		t.Run(sr, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := filepath.Join(dir, "out.jsonl")
			g := &config.GlobalConfig{}
			task := &config.Task{
				Name: "t",
				Frontmatter: map[string]any{
					"safe-outputs": map[string]any{
						"close-issue": map[string]any{"max": 1},
					},
				},
			}
			tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, IssueNumber: 3}
			item := map[string]any{"target": 0, "state_reason": sr}
			if err := ValidateAndAppend(context.Background(), g, task, tc, KindCloseIssue, item, path); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAndAppend_CloseIssue_NoStateReason(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"close-issue": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, IssueNumber: 3}
	item := map[string]any{"target": 0}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindCloseIssue, item, path); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAndAppend_CloseIssue_TargetOverridesEventContext(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"close-issue": map[string]any{"max": 1},
			},
		},
	}
	// tc has IssueNumber 3, but item has target 7 — item wins
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, IssueNumber: 3}
	item := map[string]any{"target": 7}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindCloseIssue, item, path); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAndAppend_ClosePullRequest_TargetOverridesEventContext(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"close-pull-request": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 3}
	item := map[string]any{"target": 9}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindClosePullRequest, item, path); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAndAppend_UpdateIssue_TargetOverridesEventContext(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"update-issue": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, IssueNumber: 3}
	item := map[string]any{"title": "hello", "body": "", "target": 9}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindUpdateIssue, item, path); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAndAppend_UpdatePullRequest_TargetOverridesEventContext(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"update-pull-request": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 3}
	item := map[string]any{"title": "hello", "body": "", "target": 9}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindUpdatePullRequest, item, path); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAndAppend_AddReviewer_TargetOverridesEventContext(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"add-reviewer": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 3}
	item := map[string]any{"reviewers": []any{"alice"}, "target": 9}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindAddReviewer, item, path); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAndAppend_ReplyToPullRequestReviewComment_TargetOverridesEventContext(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"reply-to-pull-request-review-comment": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 3}
	item := map[string]any{"comment_id": 42, "body": "hi", "target": 9}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindReplyToPullRequestReviewComment, item, path); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAndAppend_ResolvePullRequestReviewThread_TargetOverridesEventContext(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"resolve-pull-request-review-thread": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 3}
	item := map[string]any{"thread_id": "gid://...", "target": 9}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindResolvePullRequestReviewThread, item, path); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAndAppend_SubmitPullRequestReview_ValidEvents(t *testing.T) {
	for _, ev := range []string{"APPROVE", "approve", "REQUEST_CHANGES", "COMMENT"} {
		ev := ev
		t.Run(ev, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := filepath.Join(dir, "out.jsonl")
			g := &config.GlobalConfig{}
			task := &config.Task{
				Name: "t",
				Frontmatter: map[string]any{
					"safe-outputs": map[string]any{
						"submit-pull-request-review": map[string]any{"max": 1},
					},
				},
			}
			tc := &types.TaskContext{Repo: "o/r", RepoPath: dir, PRNumber: 5}
			item := map[string]any{"event": ev, "target": 0}
			if err := ValidateAndAppend(context.Background(), g, task, tc, KindSubmitPullRequestReview, item, path); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// toStringSliceAny edge cases
// ---------------------------------------------------------------------------

func TestToStringSliceAny_Empty(t *testing.T) {
	t.Parallel()
	got := toStringSliceAny(nil)
	if got != nil {
		t.Fatalf("nil input: got %#v", got)
	}
	got = toStringSliceAny([]string{})
	if got != nil {
		t.Fatalf("empty slice: got %#v", got)
	}
}

func TestToStringSliceAny_Valid(t *testing.T) {
	t.Parallel()
	got := toStringSliceAny([]string{"a", "b"})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("got %#v", got)
	}
}

func TestValidateAndAppend_SkipsMalformedExistingLine(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	if err := os.WriteFile(path, []byte("{bad\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	g := &config.GlobalConfig{}
	task := &config.Task{
		Name: "t",
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"add-comment": map[string]any{"max": 1},
			},
		},
	}
	tc := &types.TaskContext{
		Repo:        "o/r",
		RepoPath:    dir,
		TaskName:    "t",
		IssueNumber: 42,
	}
	item := map[string]any{"body": "ok", "target": 0}
	if err := ValidateAndAppend(context.Background(), g, task, tc, KindAddComment, item, path); err != nil {
		t.Fatalf("expected append to succeed with malformed existing line: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"body":"ok"`) {
		t.Fatalf("expected appended line in file, got %q", string(b))
	}
}
