package ghclient

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// prepends a fake gh that succeeds for common api calls used in tests.
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
if [ "$1" != "api" ]; then
  exit 1
fi
# GET comments list
if echo "$2" | grep -q '/issues/[0-9]*/comments$' && ! echo "$*" | grep -q -- '-X'; then
  echo '[{"body":"<!-- wm-checkpoint: {\"summary\":\"checkpoint summary\"} -->"}]'
  exit 0
fi
# POST reaction on issue or issue comment
if echo "$*" | grep -q -- '-X POST' && echo "$*" | grep -q '/reactions'; then
  cat >/dev/null
  exit 0
fi
# POST comment with stdin
if echo "$*" | grep -q -- '-X POST' && echo "$*" | grep -q '/comments'; then
  cat >/dev/null
  exit 0
fi
# POST label
if echo "$*" | grep -q -- '-X POST' && echo "$*" | grep -q '/labels'; then
  exit 0
fi
# DELETE label
if echo "$*" | grep -q -- '-X DELETE' && echo "$*" | grep -q '/labels/'; then
  exit 0
fi
exit 1
`
	if err := os.WriteFile(gh, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestListIssueCommentBodies_WithFakeGh(t *testing.T) {
	WithFakeGH(t)
	bodies, err := ListIssueCommentBodies("o/r", 42)
	if err != nil {
		t.Fatal(err)
	}
	if len(bodies) != 1 || bodies[0] == "" {
		t.Fatalf("%v", bodies)
	}
}

func TestPostIssueComment_WithFakeGh(t *testing.T) {
	WithFakeGH(t)
	if err := PostIssueComment("o/r", 1, "hello"); err != nil {
		t.Fatal(err)
	}
}

func TestAddIssueLabel_WithFakeGh(t *testing.T) {
	WithFakeGH(t)
	if err := AddIssueLabel("o/r", 1, "lb"); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveIssueLabel_WithFakeGh(t *testing.T) {
	WithFakeGH(t)
	if err := RemoveIssueLabel("o/r", 1, "lb"); err != nil {
		t.Fatal(err)
	}
}

func TestAddIssueReaction_WithFakeGh(t *testing.T) {
	WithFakeGH(t)
	if err := AddIssueReaction("o/r", 1, "eyes"); err != nil {
		t.Fatal(err)
	}
}

func TestAddIssueCommentReaction_WithFakeGh(t *testing.T) {
	WithFakeGH(t)
	if err := AddIssueCommentReaction("o/r", 42, "eyes"); err != nil {
		t.Fatal(err)
	}
}
