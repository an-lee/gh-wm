package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAddCommand_HTTP(t *testing.T) {
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

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"add", srv.URL + "/task.md"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".wm", "tasks", "task.md")); err != nil {
		t.Fatal(err)
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
