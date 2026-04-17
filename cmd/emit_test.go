package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
)

func TestEmitNoop_AppendsNDJSONLine(t *testing.T) {
	config.ResetLoadCache()
	t.Cleanup(func() { config.ResetLoadCache() })

	root := t.TempDir()
	wm := filepath.Join(root, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte("version: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	task := `---
safe-outputs:
  noop:
---
body
`
	if err := os.WriteFile(filepath.Join(wm, "tasks", "a.md"), []byte(task), 0o644); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(t.TempDir(), "out.jsonl")

	t.Setenv("WM_REPO_ROOT", root)
	t.Setenv("WM_TASK", "a")
	t.Setenv("WM_SAFE_OUTPUT_FILE", outPath)
	t.Setenv("GITHUB_REPOSITORY", "o/r")
	t.Cleanup(func() {
		_ = os.Unsetenv("WM_REPO_ROOT")
		_ = os.Unsetenv("WM_TASK")
		_ = os.Unsetenv("WM_SAFE_OUTPUT_FILE")
		_ = os.Unsetenv("GITHUB_REPOSITORY")
	})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"emit", "noop", "--message", "from test"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 || b[len(b)-1] != '\n' {
		t.Fatalf("expected newline-terminated jsonl, got %q", string(b))
	}
}
