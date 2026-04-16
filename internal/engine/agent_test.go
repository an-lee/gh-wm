package engine

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

func TestAgentCLIArgs(t *testing.T) {
	t.Parallel()
	args := agentCLIArgs(&config.GlobalConfig{Model: "m1", MaxTurns: 7})
	want := []string{"-p", "--dangerously-skip-permissions", "--model", "m1", "--max-turns", "7"}
	if len(args) != len(want) {
		t.Fatalf("got %v", args)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("i=%d got %v want %v", i, args, want)
		}
	}
	if args2 := agentCLIArgs(nil); len(args2) != 2 || args2[0] != "-p" || args2[1] != "--dangerously-skip-permissions" {
		t.Fatalf("nil glob: %v", args2)
	}
}

func TestParseWM_AGENT_CMD_Placeholder(t *testing.T) {
	t.Parallel()
	name, args, stdin, err := parseWM_AGENT_CMD("echo {prompt} tail", "BODY")
	if err != nil || stdin != nil {
		t.Fatalf("err=%v stdin=%v", err, stdin)
	}
	if name != "echo" || len(args) != 2 || args[0] != "BODY" || args[1] != "tail" {
		t.Fatalf("name=%q args=%v", name, args)
	}
}

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
	res, err := runAgent(ctx, glob, task, tc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.ExitCode != 0 {
		t.Fatalf("%+v", res)
	}
}

func TestRunAgent_LogWriter(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	t.Setenv("WM_AGENT_CMD", "echo")
	t.Cleanup(func() { _ = os.Unsetenv("WM_AGENT_CMD") })
	task := &config.Task{Name: "t", Body: "hello", Frontmatter: map[string]any{}}
	tc := &types.TaskContext{RepoPath: dir, Repo: "o/r", TaskName: "t"}
	var logBuf bytes.Buffer
	res, err := runAgent(ctx, &config.GlobalConfig{Engine: "claude"}, task, tc, &RunOptions{LogWriter: &logBuf}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(res.Stdout) != "hello" {
		t.Fatalf("stdout: %q", res.Stdout)
	}
	if logBuf.String() != res.Stdout {
		t.Fatalf("LogWriter tee: log=%q stdout=%q", logBuf.String(), res.Stdout)
	}
}

func TestRunAgent_CopilotError(t *testing.T) {
	task := &config.Task{Name: "t", Body: "hi", Frontmatter: map[string]any{"engine": "copilot"}}
	tc := &types.TaskContext{RepoPath: t.TempDir(), TaskName: "t"}
	res, err := runAgent(context.Background(), &config.GlobalConfig{Engine: "claude"}, task, tc, nil, nil)
	if err == nil || res == nil || res.ExitCode != -1 {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestRunAgent_CodexWithEnvAlt(t *testing.T) {
	t.Setenv("WM_ENGINE_CODEX_CMD", "true")
	t.Cleanup(func() { _ = os.Unsetenv("WM_ENGINE_CODEX_CMD") })
	task := &config.Task{Name: "t", Body: "x", Frontmatter: map[string]any{"engine": "codex"}}
	tc := &types.TaskContext{RepoPath: t.TempDir(), TaskName: "t"}
	res, err := runAgent(context.Background(), &config.GlobalConfig{Engine: "claude"}, task, tc, nil, nil)
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
	res, err := runAgent(context.Background(), &config.GlobalConfig{Engine: "claude"}, task, tc, nil, nil)
	if err == nil || res == nil || res.Success {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestRunAgent_CheckpointHintAppended(t *testing.T) {
	t.Setenv("WM_AGENT_CMD", "true")
	t.Cleanup(func() { _ = os.Unsetenv("WM_AGENT_CMD") })
	task := &config.Task{Name: "t", Body: "base", Frontmatter: map[string]any{}}
	tc := &types.TaskContext{RepoPath: t.TempDir(), TaskName: "t", CheckpointHint: "prev"}
	res, err := runAgent(context.Background(), &config.GlobalConfig{Engine: "claude"}, task, tc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatal(res)
	}
}
