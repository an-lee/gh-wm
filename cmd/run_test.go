package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunCommand(t *testing.T) {
	t.Setenv("WM_AGENT_CMD", "true")
	t.Cleanup(func() { _ = os.Unsetenv("WM_AGENT_CMD") })
	root := t.TempDir()
	wm := filepath.Join(root, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte("version: 1\nengine: claude\nmax_turns: 10\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "tasks", "mytask.md"), []byte(`---
on:
  issues: {}
---

run me
`), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(root, "ev.json")
	if err := os.WriteFile(payload, []byte(`{"action":"opened"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "t@t")
	runGit(t, root, "config", "user.name", "t")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "init")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"run", "--repo-root", root, "--task", "mytask", "--event-name", "issues", "--payload", payload})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRunCommand_Remote(t *testing.T) {
	prependFakeGh(t, `
if [ "$1" = "repo" ] && [ "$2" = "view" ]; then
  echo 'myorg/myrepo'
  exit 0
fi
if [ "$1" = "workflow" ] && [ "$2" = "run" ]; then
  if [ "$3" != "wm-agent.yml" ]; then exit 1; fi
  if [ "$4" != "-R" ] || [ "$5" != "myorg/myrepo" ]; then exit 1; fi
  shift 5
  seen_task=0
  seen_issue=0
  seen_ref=0
  while [ $# -gt 0 ]; do
    case "$1" in
      -f)
        case "$2" in
          task_name=mytask) seen_task=1 ;;
          issue_number=99) seen_issue=1 ;;
          *) ;;
        esac
        shift 2
        ;;
      --ref)
        if [ "$2" != "topic" ]; then exit 1; fi
        seen_ref=1
        shift 2
        ;;
      *) exit 1 ;;
    esac
  done
  if [ "$seen_task" != 1 ] || [ "$seen_issue" != 1 ] || [ "$seen_ref" != 1 ]; then exit 1; fi
  exit 0
fi
exit 1
`)
	root := t.TempDir()
	wm := filepath.Join(root, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte("version: 1\nengine: claude\nmax_turns: 10\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "tasks", "mytask.md"), []byte(`---
on:
  issues: {}
---

run me
`), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "t@t")
	runGit(t, root, "config", "user.name", "t")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "init")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"run", "--repo-root", root, "--task", "mytask", "--remote", "--issue", "99", "--ref", "topic"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	_ = buf
}

func TestRunCommand_Remote_ExplicitRepo(t *testing.T) {
	prependFakeGh(t, `
if [ "$1" = "workflow" ] && [ "$2" = "run" ]; then
  if [ "$4" != "-R" ] || [ "$5" != "other/explicit" ]; then exit 1; fi
  if ! echo "$*" | grep -q -- '-f task_name=t'; then exit 1; fi
  exit 0
fi
exit 1
`)
	root := t.TempDir()
	wm := filepath.Join(root, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte("version: 1\nengine: claude\nmax_turns: 10\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "tasks", "t.md"), []byte(`---
on:
  issues: {}
---
`), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "t@t")
	runGit(t, root, "config", "user.name", "t")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "init")

	rootCmd.SetOut(new(bytes.Buffer))
	rootCmd.SetErr(new(bytes.Buffer))
	rootCmd.SetArgs([]string{"run", "--repo-root", root, "--task", "t", "--remote", "--repo", "other/explicit"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRunCommand_TaskMissing(t *testing.T) {
	root := t.TempDir()
	wm := filepath.Join(root, ".wm")
	if err := os.MkdirAll(filepath.Join(wm, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wm, "config.yml"), []byte("version: 1\nengine: claude\nmax_turns: 10\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(root, "ev.json")
	if err := os.WriteFile(payload, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	rootCmd.SetArgs([]string{"run", "--repo-root", root, "--task", "nope", "--payload", payload})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
