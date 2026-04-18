package ghclient

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEnsureRepoLabel_Empty(t *testing.T) {
	t.Parallel()
	if err := EnsureRepoLabel(context.Background(), "o/r", "  "); err != nil {
		t.Fatal(err)
	}
	if err := EnsureRepoLabels(context.Background(), "o/r", nil); err != nil {
		t.Fatal(err)
	}
}

// TestEnsureRepoLabel_ListOK uses fake gh: label list returns the label (no create).
func TestEnsureRepoLabel_ListOK(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake gh script is unix-only")
	}
	dir := t.TempDir()
	ghBin := filepath.Join(dir, "gh")
	script := `#!/bin/sh
if [ "$1" = "label" ] && [ "$2" = "list" ]; then
  echo '[{"name":"bug"}]'
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

	if err := EnsureRepoLabel(context.Background(), "owner/repo", "bug"); err != nil {
		t.Fatal(err)
	}
}

// TestEnsureRepoLabel_CreateWhenMissing runs gh label list (empty) then gh label create.
func TestEnsureRepoLabel_CreateWhenMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake gh script is unix-only")
	}
	dir := t.TempDir()
	ghBin := filepath.Join(dir, "gh")
	script := `#!/bin/sh
if [ "$1" = "label" ] && [ "$2" = "list" ]; then
  echo '[]'
  exit 0
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

func TestEnsureRepoLabels_Batched(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake gh script is unix-only")
	}
	dir := t.TempDir()
	ghBin := filepath.Join(dir, "gh")
	script := `#!/bin/sh
if [ "$1" = "label" ] && [ "$2" = "list" ]; then
  echo '[{"name":"a"}]'
  exit 0
fi
if [ "$1" = "label" ] && [ "$2" = "create" ]; then
  if [ "$3" = "b" ]; then
    exit 0
  fi
fi
echo "unexpected $*" >&2
exit 1
`
	if err := os.WriteFile(ghBin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GH_WM_REST", "")

	if err := EnsureRepoLabels(context.Background(), "owner/repo", []string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
}
