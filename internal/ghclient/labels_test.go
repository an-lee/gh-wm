package ghclient

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLabelGetNotFound(t *testing.T) {
	t.Parallel()
	tests := []struct {
		out  string
		want bool
	}{
		{`{"message":"Not Found"}`, true},
		{`{"message":"not found"}`, true},
		{`HTTP 404: Not Found`, true},
		{`error 404`, false},
		{`{"message":"Server Error"}`, false},
		{"", false},
	}
	for _, tc := range tests {
		if got := labelGetNotFound(tc.out); got != tc.want {
			t.Errorf("labelGetNotFound(%q) = %v, want %v", tc.out, got, tc.want)
		}
	}
}

func TestEnsureRepoLabel_Empty(t *testing.T) {
	t.Parallel()
	if err := EnsureRepoLabel(context.Background(), "o/r", "  "); err != nil {
		t.Fatal(err)
	}
}

// TestEnsureRepoLabel_GetOK uses a fake gh that succeeds on first api call (label exists).
func TestEnsureRepoLabel_GetOK(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake gh script is unix-only")
	}
	dir := t.TempDir()
	ghBin := filepath.Join(dir, "gh")
	script := `#!/bin/sh
if [ "$1" = "api" ]; then
  echo '{"name":"bug"}'
  exit 0
fi
echo "unexpected $*" >&2
exit 1
`
	if err := os.WriteFile(ghBin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GH_WM_REST", "") // subprocess path

	if err := EnsureRepoLabel(context.Background(), "owner/repo", "bug"); err != nil {
		t.Fatal(err)
	}
}

// TestEnsureRepoLabel_CreateWhenMissing runs gh api (404) then gh label create.
func TestEnsureRepoLabel_CreateWhenMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake gh script is unix-only")
	}
	dir := t.TempDir()
	ghBin := filepath.Join(dir, "gh")
	script := `#!/bin/sh
if [ "$1" = "api" ]; then
  echo '{"message":"Not Found"}' >&2
  exit 1
fi
if [ "$1" = "label" ] && [ "$2" = "create" ]; then
  exit 0
fi
echo "unexpected $*" >&2
exit 1
`
	if err := os.WriteFile(ghBin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GH_WM_REST", "")

	if err := EnsureRepoLabel(context.Background(), "owner/repo", "new-label"); err != nil {
		t.Fatal(err)
	}
}
