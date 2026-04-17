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
