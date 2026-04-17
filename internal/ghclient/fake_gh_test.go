package ghclient

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// WithFakeGH prepends a fake gh binary for commands that still shell out (e.g. CurrentRepo).
func WithFakeGH(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake gh script is unix-only")
	}
	dir := t.TempDir()
	gh := filepath.Join(dir, "gh")
	script := `#!/bin/sh
set -e
# gh repo view --json nameWithOwner [-q .nameWithOwner]
if [ "$1" = "repo" ] && [ "$2" = "view" ]; then
  echo 'test-owner/test-repo'
  exit 0
fi
exit 1
`
	if err := os.WriteFile(gh, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// API helpers use a PATH stub by default (`gh api`). Set GH_WM_REST=1 to use go-gh REST ([internal/gh](../gh)).
