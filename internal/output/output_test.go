package output

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

func TestRunSuccessOutputs_AgentDrivenNoop(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "output.json")
	if err := os.WriteFile(p, []byte(`{"items":[{"type":"noop","message":"nothing"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	g := &config.GlobalConfig{}
	task := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"add-comment": map[string]any{},
	}}}
	tc := &types.TaskContext{RepoPath: t.TempDir(), Repo: "o/r", IssueNumber: 1}
	res := &types.AgentResult{OutputFilePath: p}
	if err := RunSuccessOutputs(context.Background(), g, task, tc, res); err != nil {
		t.Fatal(err)
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

func TestRunCommentOutput_NoNumber(t *testing.T) {
	t.Parallel()
	task := &config.Task{Name: "x"}
	tc := &types.TaskContext{RepoPath: t.TempDir()}
	res := &types.AgentResult{Summary: "hi"}
	if err := runCommentOutputLegacy(context.Background(), nil, task, tc, res); err == nil {
		t.Fatal("expected error")
	}
}

func TestRunLabelOutput_Validation(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"add-labels": "not-a-map",
	}}}
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 1}
	if err := runLabelOutputLegacy(context.Background(), task, tc); err == nil {
		t.Fatal("expected type error")
	}
	task2 := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"add-labels": map[string]any{"labels": []any{}},
	}}}
	if err := runLabelOutputLegacy(context.Background(), task2, tc); err != nil {
		t.Fatal(err)
	}
	task3 := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"add-labels": map[string]any{"labels": []any{"l"}},
	}}}
	if err := runLabelOutputLegacy(context.Background(), task3, &types.TaskContext{IssueNumber: 1}); err == nil {
		t.Fatal("expected missing repo")
	}
}
