package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateCommand_AllWithSource(t *testing.T) {
	content := `---
on:
  issues: {}
---

from url
`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(content))
	}))
	t.Cleanup(srv.Close)

	root := t.TempDir()
	wm := filepath.Join(root, ".wm", "tasks")
	if err := os.MkdirAll(wm, 0o755); err != nil {
		t.Fatal(err)
	}
	task := fmt.Sprintf(`---
source: %q
on:
  issues: {}
---

old body
`, srv.URL+"/task.md")
	if err := os.WriteFile(filepath.Join(wm, "task.md"), []byte(task), 0o644); err != nil {
		t.Fatal(err)
	}
	chdirTemp(t, root)
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"update"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(wm, "task.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "from url") {
		t.Fatalf("expected updated body: %s", b)
	}
}

func TestUpdateCommand_ShorthandSource(t *testing.T) {
	content := `---
on:
  issues: {}
---

from shorthand
`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/o/r/main/workflows/x.md") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(content))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	prev := rawGitHubBaseURL
	rawGitHubBaseURL = srv.URL
	t.Cleanup(func() { rawGitHubBaseURL = prev })

	root := t.TempDir()
	wm := filepath.Join(root, ".wm", "tasks")
	if err := os.MkdirAll(wm, 0o755); err != nil {
		t.Fatal(err)
	}
	task := `---
source: "o/r/workflows/x.md"
on:
  issues: {}
---
old
`
	if err := os.WriteFile(filepath.Join(wm, "x.md"), []byte(task), 0o644); err != nil {
		t.Fatal(err)
	}
	chdirTemp(t, root)
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"update"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(wm, "x.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "from shorthand") {
		t.Fatalf("expected updated body: %s", b)
	}
}

func TestUpdateCommand_SpecificTask(t *testing.T) {
	content := `---
on:
  issues: {}
---

specific
`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(content))
	}))
	t.Cleanup(srv.Close)

	root := t.TempDir()
	wm := filepath.Join(root, ".wm", "tasks")
	if err := os.MkdirAll(wm, 0o755); err != nil {
		t.Fatal(err)
	}
	a := fmt.Sprintf(`---
source: %q
on:
  issues: {}
---
a`, srv.URL+"/a.md")
	b := `---
on:
  issues: {}
---
b`
	if err := os.WriteFile(filepath.Join(wm, "only.md"), []byte(a), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "other.md"), []byte(b), 0o644); err != nil {
		t.Fatal(err)
	}
	chdirTemp(t, root)
	rootCmd.SetArgs([]string{"update", "only"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(filepath.Join(wm, "only.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "specific") {
		t.Fatalf("expected updated only.md: %s", out)
	}
	other, err := os.ReadFile(filepath.Join(wm, "other.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(other) != b {
		t.Fatalf("other.md should be unchanged")
	}
}
