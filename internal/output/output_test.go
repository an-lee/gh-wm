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

func TestRunSuccessOutputs_NDJSONNoop(t *testing.T) {
	t.Parallel()
	nd := filepath.Join(t.TempDir(), "out.jsonl")
	if err := os.WriteFile(nd, []byte("{\"type\":\"noop\",\"message\":\"from ndjson\"}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	g := &config.GlobalConfig{}
	task := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"noop": map[string]any{},
	}}}
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", IssueNumber: 1}
	res := &types.AgentResult{SafeOutputFilePath: nd}
	if err := RunSuccessOutputs(context.Background(), g, task, tc, res); err != nil {
		t.Fatal(err)
	}
}

func TestRunSuccessOutputs_NestedNoopEnvelope(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "out.jsonl")
	if err := os.WriteFile(p, []byte(`{"noop":{"message":"nested noop"}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	g := &config.GlobalConfig{}
	task := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"noop": map[string]any{},
	}}}
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", IssueNumber: 1}
	res := &types.AgentResult{SafeOutputFilePath: p}
	if err := RunSuccessOutputs(context.Background(), g, task, tc, res); err != nil {
		t.Fatalf("nested noop: %v", err)
	}
}

func TestRunSuccessOutputs_AgentDrivenNoop(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "out.jsonl")
	if err := os.WriteFile(p, []byte(`{"type":"noop","message":"nothing"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	g := &config.GlobalConfig{}
	task := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"add-comment": map[string]any{},
	}}}
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", IssueNumber: 1}
	res := &types.AgentResult{SafeOutputFilePath: p}
	if err := RunSuccessOutputs(context.Background(), g, task, tc, res); err != nil {
		t.Fatal(err)
	}
}

func TestRunSuccessOutputs_ImplicitNoopWhenEmpty(t *testing.T) {
	t.Parallel()
	g := &config.GlobalConfig{}
	task := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"add-comment": map[string]any{},
	}}}
	tc := &types.TaskContext{RepoPath: t.TempDir()}
	res := &types.AgentResult{SafeOutputFilePath: filepath.Join(t.TempDir(), "missing.jsonl")}
	if err := RunSuccessOutputs(context.Background(), g, task, tc, res); err != nil {
		t.Fatalf("expected success (implicit noop), got: %v", err)
	}
}

func TestRunSuccessOutputs_NoSafeOutputsSkips(t *testing.T) {
	t.Parallel()
	g := &config.GlobalConfig{}
	task := &config.Task{}
	tc := &types.TaskContext{}
	res := &types.AgentResult{}
	if err := RunSuccessOutputs(context.Background(), g, task, tc, res); err != nil {
		t.Fatal(err)
	}
}

func TestRunSuccessOutputs_MissingExecutorFails(t *testing.T) {
	original, hadOriginal := kindRegistry[KindAddComment]
	delete(kindRegistry, KindAddComment)
	t.Cleanup(func() {
		if hadOriginal {
			kindRegistry[KindAddComment] = original
			return
		}
		delete(kindRegistry, KindAddComment)
	})

	p := filepath.Join(t.TempDir(), "out.jsonl")
	if err := os.WriteFile(p, []byte(`{"type":"add_comment","body":"hello"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	g := &config.GlobalConfig{}
	task := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"add-comment": map[string]any{},
	}}}
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", IssueNumber: 1}
	res := &types.AgentResult{SafeOutputFilePath: p}

	err := RunSuccessOutputs(context.Background(), g, task, tc, res)
	if err == nil {
		t.Fatal("expected error when executor is missing")
	}
	if !strings.Contains(err.Error(), "no executor registered") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSuccessOutputs_NilShortCircuit(t *testing.T) {
	t.Parallel()
	if err := RunSuccessOutputs(context.Background(), nil, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	g := &config.GlobalConfig{}
	task := &config.Task{}
	tc := &types.TaskContext{}
	res := &types.AgentResult{}
	if err := RunSuccessOutputs(context.Background(), g, task, tc, res); err != nil {
		t.Fatal(err)
	}
}

func TestRunCommentFromItem_NoNumber(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{RepoPath: t.TempDir()}
	if err := runCommentFromItem(context.Background(), nil, tc, ItemAddComment{Body: "x"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestPostFallbackComment_NoTarget(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r"}
	if err := postFallbackComment(tc, "hello"); err == nil {
		t.Fatal("expected error when no issue/PR number")
	}
}

func TestRunSuccessOutputs_ImplicitNoopWithLastResponseTextNoIssue(t *testing.T) {
	t.Parallel()
	g := &config.GlobalConfig{}
	task := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"add-comment": map[string]any{},
	}}}
	tc := &types.TaskContext{RepoPath: t.TempDir()}
	res := &types.AgentResult{
		SafeOutputFilePath: filepath.Join(t.TempDir(), "missing.jsonl"),
		LastResponseText:   "agent said something",
	}
	if err := RunSuccessOutputs(context.Background(), g, task, tc, res); err != nil {
		t.Fatalf("expected success without GitHub target: %v", err)
	}
}
