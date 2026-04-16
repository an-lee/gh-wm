// Package gitbranch creates feature branches before wm run when PR output is enabled.
package gitbranch

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// DefaultBaseBranch returns the remote default branch name (main, master, etc.).
func DefaultBaseBranch(dir string) string {
	cmd := exec.Command("git", "-C", dir, "symbolic-ref", "refs/remotes/origin/HEAD")
	out, err := cmd.Output()
	if err == nil {
		s := strings.TrimSpace(string(out))
		if i := strings.LastIndex(s, "/"); i >= 0 {
			return strings.TrimSpace(s[i+1:])
		}
	}
	for _, b := range []string{"main", "master"} {
		c := exec.Command("git", "-C", dir, "rev-parse", "--verify", "origin/"+b)
		if c.Run() == nil {
			return b
		}
	}
	return "main"
}

// CurrentBranch returns the current branch name or "HEAD" when detached.
func CurrentBranch(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// PrepareFeatureForPR creates and checks out wm/<slug>-<timestamp> when the repo is on
// the default branch or detached HEAD. When already on a non-default branch, it does nothing
// (created=false, newBranch=current).
func PrepareFeatureForPR(repoRoot, taskName string) (previousBranch string, newBranch string, created bool, err error) {
	abs, err := filepath.Abs(repoRoot)
	if err != nil {
		return "", "", false, err
	}
	dir := filepath.Clean(abs)
	cur, err := CurrentBranch(dir)
	if err != nil {
		return "", "", false, fmt.Errorf("git current branch: %w", err)
	}
	base := DefaultBaseBranch(dir)
	if cur != base && cur != "HEAD" {
		return cur, cur, false, nil
	}
	slug := slugTaskName(taskName)
	newB := fmt.Sprintf("wm/%s-%s", slug, time.Now().UTC().Format("20060102-150405"))
	out, err := exec.Command("git", "-C", dir, "checkout", "-b", newB).CombinedOutput()
	if err != nil {
		return "", "", false, fmt.Errorf("git checkout -b %s: %w: %s", newB, err, strings.TrimSpace(string(out)))
	}
	return cur, newB, true, nil
}

// Checkout switches to branch; no-op for empty or "HEAD" (cannot restore detached state).
func Checkout(repoRoot, branch string) error {
	if strings.TrimSpace(branch) == "" || branch == "HEAD" {
		return nil
	}
	abs, err := filepath.Abs(repoRoot)
	if err != nil {
		return err
	}
	dir := filepath.Clean(abs)
	out, err := exec.Command("git", "-C", dir, "checkout", branch).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout %s: %w: %s", branch, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func slugTaskName(name string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(name) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
		case r == '-' || r == '_' || unicode.IsSpace(r):
			if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
				b.WriteRune('-')
			}
		}
	}
	s := strings.Trim(b.String(), "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	if s == "" {
		s = "task"
	}
	if len(s) > 48 {
		s = s[:48]
		s = strings.TrimRight(s, "-")
	}
	return s
}
