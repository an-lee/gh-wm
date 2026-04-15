package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := WriteConfig(dir); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "config.yml"))
	if err != nil || len(b) < 10 {
		t.Fatal("config missing")
	}
}

func TestWriteCLAUDE_SkipIfExists(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	existing := filepath.Join(root, "CLAUDE.md")
	if err := os.WriteFile(existing, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteCLAUDE(root); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(existing)
	if string(b) != "keep" {
		t.Fatal("should not overwrite")
	}
}

func TestWriteCLAUDE_Create(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := WriteCLAUDE(root); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil || len(b) < 20 {
		t.Fatal(err)
	}
}

func TestWriteStarterTasks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := WriteStarterTasks(dir); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		t.Fatal("no tasks written")
	}
}
