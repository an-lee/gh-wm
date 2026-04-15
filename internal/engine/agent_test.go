package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

func TestRunAgent_WM_AGENT_CMD(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WM_AGENT_CMD", "true")
	t.Cleanup(func() { _ = os.Unsetenv("WM_AGENT_CMD") })
	glob := &config.GlobalConfig{Engine: "claude", Context: struct {
		Files []string `yaml:"files"`
	}{Files: []string{"note.txt"}}}
	task := &config.Task{Name: "t", Body: "  ", Frontmatter: map[string]any{}}
	tc := &types.TaskContext{RepoPath: dir, Repo: "o/r", TaskName: "t"}
	res, err := runAgent(ctx, glob, task, tc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.ExitCode != 0 {
		t.Fatalf("%+v", res)
	}
}

func TestRunAgent_CopilotError(t *testing.T) {
	task := &config.Task{Name: "t", Body: "hi", Frontmatter: map[string]any{"engine": "copilot"}}
	tc := &types.TaskContext{RepoPath: t.TempDir(), TaskName: "t"}
	res, err := runAgent(context.Background(), &config.GlobalConfig{Engine: "claude"}, task, tc)
	if err == nil || res == nil || res.ExitCode != -1 {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestRunAgent_CodexWithEnvAlt(t *testing.T) {
	t.Setenv("WM_ENGINE_CODEX_CMD", "true")
	t.Cleanup(func() { _ = os.Unsetenv("WM_ENGINE_CODEX_CMD") })
	task := &config.Task{Name: "t", Body: "x", Frontmatter: map[string]any{"engine": "codex"}}
	tc := &types.TaskContext{RepoPath: t.TempDir(), TaskName: "t"}
	res, err := runAgent(context.Background(), &config.GlobalConfig{Engine: "claude"}, task, tc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("%+v", res)
	}
}

func TestRunAgent_NonZeroExit(t *testing.T) {
	t.Setenv("WM_AGENT_CMD", "false")
	t.Cleanup(func() { _ = os.Unsetenv("WM_AGENT_CMD") })
	task := &config.Task{Name: "t", Body: "x", Frontmatter: map[string]any{}}
	tc := &types.TaskContext{RepoPath: t.TempDir(), TaskName: "t"}
	res, err := runAgent(context.Background(), &config.GlobalConfig{Engine: "claude"}, task, tc)
	if err == nil || res == nil || res.Success {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestRunAgent_CheckpointHintAppended(t *testing.T) {
	t.Setenv("WM_AGENT_CMD", "true")
	t.Cleanup(func() { _ = os.Unsetenv("WM_AGENT_CMD") })
	task := &config.Task{Name: "t", Body: "base", Frontmatter: map[string]any{}}
	tc := &types.TaskContext{RepoPath: t.TempDir(), TaskName: "t", CheckpointHint: "prev"}
	res, err := runAgent(context.Background(), &config.GlobalConfig{Engine: "claude"}, task, tc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatal(res)
	}
}
