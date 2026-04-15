package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func prependFakeGh(t *testing.T, script string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("unix shell fake gh only")
	}
	dir := t.TempDir()
	gh := filepath.Join(dir, "gh")
	full := "#!/bin/sh\nset -e\n" + script
	if err := os.WriteFile(gh, []byte(full), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestLogsCommand(t *testing.T) {
	prependFakeGh(t, `
if [ "$1" = "run" ] && [ "$2" = "list" ]; then
  echo '[{"displayTitle":"WM #42 fix","url":"https://example","createdAt":"2020-01-01"}]'
  exit 0
fi
exit 1
`)
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"logs", "42"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestLogsCommand_NoMatch(t *testing.T) {
	prependFakeGh(t, `
if [ "$1" = "run" ] && [ "$2" = "list" ]; then
  echo '[{"displayTitle":"other","url":"https://x","createdAt":"2020"}]'
  exit 0
fi
exit 1
`)
	rootCmd.SetArgs([]string{"logs", "99"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestStatusCommand(t *testing.T) {
	prependFakeGh(t, `
if [ "$1" = "issue" ] && [ "$2" = "list" ]; then
  exit 0
fi
exit 1
`)
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestStatusCommand_All(t *testing.T) {
	prependFakeGh(t, `
if [ "$1" = "search" ] && [ "$2" = "issues" ]; then
  exit 0
fi
exit 1
`)
	rootCmd.SetArgs([]string{"status", "--all"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestAssignCommand(t *testing.T) {
	prependFakeGh(t, `
if [ "$1" = "repo" ] && [ "$2" = "view" ]; then
  echo 'myorg/myrepo'
  exit 0
fi
if [ "$1" = "api" ] && echo "$*" | grep -q -- '-X POST' && echo "$*" | grep -q '/labels'; then
  exit 0
fi
exit 1
`)
	rootCmd.SetArgs([]string{"assign", "10", "--label", "custom"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
