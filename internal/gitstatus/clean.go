// Package gitstatus checks repository working tree state for wm run guards.
package gitstatus

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// EnsureClean returns an error if repoRoot is not a git work tree or
// git status --porcelain reports any change (modified, staged, or untracked).
func EnsureClean(repoRoot string) error {
	abs, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("git: resolve repo root: %w", err)
	}
	abs = filepath.Clean(abs)

	out, err := exec.Command("git", "-C", abs, "rev-parse", "--is-inside-work-tree").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git: not a repository at %q: %w\n%s", abs, err, strings.TrimSpace(string(out)))
	}
	if strings.TrimSpace(string(out)) != "true" {
		return fmt.Errorf("git: not a work tree at %q", abs)
	}

	st, err := exec.Command("git", "-C", abs, "status", "--porcelain").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if strings.TrimSpace(string(st)) != "" {
		return fmt.Errorf("working tree is not clean at %q (commit or stash changes before running)", abs)
	}
	return nil
}
