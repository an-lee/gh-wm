package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitFrontmatter(t *testing.T) {
	t.Parallel()
	const doc = `---
foo: bar
---

# Hello
`
	y, body, err := SplitFrontmatter(doc)
	if err != nil {
		t.Fatal(err)
	}
	if y != "foo: bar" {
		t.Fatalf("yaml: %q", y)
	}
	if body != "# Hello" {
		t.Fatalf("body: %q", body)
	}
}

func TestSplitFrontmatter_BOM(t *testing.T) {
	t.Parallel()
	doc := "\ufeff---\nx: 1\n---\nbody"
	y, body, err := SplitFrontmatter(doc)
	if err != nil {
		t.Fatal(err)
	}
	if y != "x: 1" || body != "body" {
		t.Fatalf("y=%q body=%q", y, body)
	}
}

func TestSplitFrontmatter_Errors(t *testing.T) {
	t.Parallel()
	if _, _, err := SplitFrontmatter("no frontmatter"); err == nil {
		t.Fatal("expected error")
	}
	if _, _, err := SplitFrontmatter("---\nunclosed"); err == nil {
		t.Fatal("expected unclosed error")
	}
}

func TestLoadTaskFile_LoadTasksDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "t.md")
	if err := os.WriteFile(path, []byte(`---
on:
  issues: {}
---

hello **world**
`), 0o644); err != nil {
		t.Fatal(err)
	}
	task, err := LoadTaskFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if task.Name != "t" || !strings.Contains(task.Body, "hello") {
		t.Fatalf("%+v", task)
	}
	// disabled skip
	if err := os.WriteFile(filepath.Join(dir, "x.md.disabled"), []byte(`---
on: {}
---
`), 0o644); err != nil {
		t.Fatal(err)
	}
	tasks, err := LoadTasksDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks", len(tasks))
	}
}
