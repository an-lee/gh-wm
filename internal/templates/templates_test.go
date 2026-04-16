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
