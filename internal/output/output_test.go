package output

import (
	"context"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

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
	if err := runCommentOutput(context.Background(), nil, task, tc, res); err == nil {
		t.Fatal("expected error")
	}
}

func TestRunLabelOutput_Validation(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"add-labels": "not-a-map",
	}}}
	tc := &types.TaskContext{Repo: "o/r", IssueNumber: 1}
	if err := runLabelOutput(context.Background(), nil, task, tc, nil); err == nil {
		t.Fatal("expected type error")
	}
	task2 := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"add-labels": map[string]any{"labels": []any{}},
	}}}
	if err := runLabelOutput(context.Background(), nil, task2, tc, nil); err != nil {
		t.Fatal(err)
	}
	task3 := &config.Task{Frontmatter: map[string]any{"safe-outputs": map[string]any{
		"add-labels": map[string]any{"labels": []any{"l"}},
	}}}
	if err := runLabelOutput(context.Background(), nil, task3, &types.TaskContext{IssueNumber: 1}, nil); err == nil {
		t.Fatal("expected missing repo")
	}
}
