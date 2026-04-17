package ghclient

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/an-lee/gh-wm/internal/gh"
)

// useREST returns true when GH_WM_REST=1 (go-gh REST); default is gh CLI subprocess for tests and minimal shells.
func useREST() bool {
	return os.Getenv("GH_WM_REST") == "1"
}

// AddIssueLabel adds a label to an issue (owner/repo from "owner/repo").
func AddIssueLabel(repo string, issue int, label string) error {
	if useREST() {
		return gh.AddIssueLabel(context.Background(), repo, issue, label)
	}
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo: %s", repo)
	}
	owner, name := parts[0], parts[1]
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/labels", owner, name, issue)
	cmd := exec.Command("gh", "api", "-X", "POST", path, "-f", "labels[]="+label)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh api: %w: %s", err, stderr.String())
	}
	return nil
}

// CurrentRepo returns owner/nameWithOwner from gh repo view.
func CurrentRepo() (string, error) {
	out, err := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner").Output()
	if err != nil {
		return "", fmt.Errorf("gh repo view: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
