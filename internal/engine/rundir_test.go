package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/an-lee/gh-wm/internal/types"
)

func TestNewRunDir_CreatesLayout(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	rd, err := NewRunDir(root, "My-Task", "issues")
	if err != nil {
		t.Fatal(err)
	}
	if rd == nil || rd.Path == "" {
		t.Fatal("expected run dir path")
	}
	meta := filepath.Join(rd.Path, metaFileName)
	if _, err := os.Stat(meta); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(meta)
	if err != nil {
		t.Fatal(err)
	}
	if !containsAll(string(b), `"task_name":`, `"My-Task"`, `"event_name":`, `"issues"`, `"phase":`, `"activation"`) {
		t.Fatalf("meta.json: %s", b)
	}
}

func TestRunDir_WritePromptAndResult(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	rd, err := NewRunDir(root, "t", "workflow_dispatch")
	if err != nil {
		t.Fatal(err)
	}
	if err := rd.WritePrompt("hello prompt"); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(rd.Path, promptFileName)
	b, err := os.ReadFile(p)
	if err != nil || string(b) != "hello prompt" {
		t.Fatalf("prompt: %q err=%v", b, err)
	}

	res := &types.RunResult{
		Phase:   types.PhaseValidation,
		Success: false,
		Errors:  nil,
	}
	if err := rd.WriteResult(res); err != nil {
		t.Fatal(err)
	}
	rb, err := os.ReadFile(filepath.Join(rd.Path, resultFileName))
	if err != nil {
		t.Fatal(err)
	}
	if !containsAll(string(rb), `"phase": "validation"`, `"success": false`) {
		t.Fatalf("result.json: %s", rb)
	}
}

func TestPruneRunDirs_RemovesOld(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	runs := filepath.Join(root, ".wm", "runs")
	if err := os.MkdirAll(runs, 0o755); err != nil {
		t.Fatal(err)
	}
	oldDir := filepath.Join(runs, "old-run")
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	if err := os.Chtimes(oldDir, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	if err := PruneRunDirs(root, 7*24*time.Hour); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Fatalf("expected old dir removed: %v", err)
	}
}

func TestTailBuffer(t *testing.T) {
	t.Parallel()
	var tb tailBuffer
	tb.max = 10
	if n, _ := tb.Write([]byte("0123456789")); n != 10 {
		t.Fatalf("n=%d", n)
	}
	if tb.String() != "0123456789" {
		t.Fatalf("got %q", tb.String())
	}
	if _, err := tb.Write([]byte("AB")); err != nil {
		t.Fatal(err)
	}
	if n := tb.String(); len(n) != 10 || n != "23456789AB" {
		t.Fatalf("got %q len=%d", n, len(n))
	}
}

func containsAll(s string, subs ...string) bool {
	for _, x := range subs {
		if !strings.Contains(s, x) {
			return false
		}
	}
	return true
}
