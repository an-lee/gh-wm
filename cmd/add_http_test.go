package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func fakeGhForAdd(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "windows" {
		prependFakeGh(t, `
if [ "$1" = "extension" ] && [ "$2" = "upgrade" ] && [ "$3" = "an-lee/gh-wm" ]; then
  exit 0
fi
exit 1
`)
	}
}

func TestAddCommand_HTTP(t *testing.T) {
	fakeGhForAdd(t)
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
	if err := os.MkdirAll(filepath.Join(root, ".wm"), 0o755); err != nil {
		t.Fatal(err)
	}
	chdirTemp(t, root)
	t.Setenv("GH_WM_REPO", "test/hello")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"add", srv.URL + "/task.md"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, ".wm", "tasks", "task.md")
	if _, err := os.Stat(dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "workflows", "wm-agent.yml")); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "source:") || !strings.Contains(string(b), srv.URL) {
		t.Fatalf("expected source: with URL in frontmatter, got:\n%s", b)
	}
}

func TestAddCommand_GitHubShorthand(t *testing.T) {
	fakeGhForAdd(t)
	content := `---
on:
  issues: {}
---

from workflows
`
	var order []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, r.URL.Path)
		if strings.HasSuffix(r.URL.Path, "/workflows/daily-doc-updater.md") {
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
	if err := os.MkdirAll(filepath.Join(root, ".wm"), 0o755); err != nil {
		t.Fatal(err)
	}
	chdirTemp(t, root)
	t.Setenv("GH_WM_REPO", "test/hello")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"add", "githubnext/agentics/daily-doc-updater"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "workflows", "wm-agent.yml")); err != nil {
		t.Fatal(err)
	}
	if len(order) < 1 || !strings.HasSuffix(order[0], "/workflows/daily-doc-updater.md") {
		t.Fatalf("expected workflows path tried first, got paths %v", order)
	}
	dst := filepath.Join(root, ".wm", "tasks", "daily-doc-updater.md")
	b, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "source:") || !strings.Contains(s, "githubnext/agentics/workflows/daily-doc-updater.md") {
		t.Fatalf("expected gh aw-style source shorthand, got:\n%s", b)
	}
	if strings.Contains(s, "raw.githubusercontent.com") {
		t.Fatalf("source should not be raw URL, got:\n%s", b)
	}
}

func TestAddCommand_GitHubShorthand_WmFallback(t *testing.T) {
	fakeGhForAdd(t)
	content := `---
on:
  issues: {}
---

from wm tasks
`
	var order []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, r.URL.Path)
		if strings.HasSuffix(r.URL.Path, "/workflows/only-wm.md") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/.wm/tasks/only-wm.md") {
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
	if err := os.MkdirAll(filepath.Join(root, ".wm"), 0o755); err != nil {
		t.Fatal(err)
	}
	chdirTemp(t, root)
	t.Setenv("GH_WM_REPO", "test/hello")

	rootCmd.SetArgs([]string{"add", "o/r/only-wm"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "workflows", "wm-agent.yml")); err != nil {
		t.Fatal(err)
	}
	if len(order) < 2 {
		t.Fatalf("expected two probes, got %v", order)
	}
	if !strings.HasSuffix(order[0], "/workflows/only-wm.md") || !strings.HasSuffix(order[1], "/.wm/tasks/only-wm.md") {
		t.Fatalf("unexpected probe order: %v", order)
	}
	dst := filepath.Join(root, ".wm", "tasks", "only-wm.md")
	b, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "o/r/.wm/tasks/only-wm.md") {
		t.Fatalf("expected .wm/tasks source, got:\n%s", b)
	}
}

func TestAddCommand_HTTPNotOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".wm"), 0o755); err != nil {
		t.Fatal(err)
	}
	chdirTemp(t, root)
	rootCmd.SetArgs([]string{"add", srv.URL})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected HTTP error")
	}
}
