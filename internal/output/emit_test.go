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
