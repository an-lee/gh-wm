package ghclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/an-lee/gh-wm/internal/gh"
)

// EnsureRepoLabel creates the label in the repository if it does not already exist.
func EnsureRepoLabel(ctx context.Context, repo, label string) error {
	if strings.TrimSpace(label) == "" {
		return nil
	}
	if useREST() {
		return gh.EnsureRepoLabel(ctx, repo, label)
	}
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	enc := url.PathEscape(label)
	getPath := fmt.Sprintf("/repos/%s/%s/labels/%s", owner, name, enc)
	cmd := exec.CommandContext(ctx, "gh", "api", getPath)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	if !labelGetNotFound(string(out)) {
		return fmt.Errorf("gh api get label %q: %w: %s", label, err, strings.TrimSpace(string(out)))
	}
	create := exec.CommandContext(ctx, "gh", "label", "create", label, "--repo", repo, "--color", gh.DefaultRepoLabelColor)
	createOut, err := create.CombinedOutput()
	if err == nil {
		return nil
	}
	s := string(createOut)
	if strings.Contains(strings.ToLower(s), "already exists") {
		return nil
	}
	return fmt.Errorf("gh label create %q: %w: %s", label, err, strings.TrimSpace(s))
}

func labelGetNotFound(apiOut string) bool {
	s := strings.TrimSpace(apiOut)
	var msg struct {
		Message string `json:"message"`
	}
	if json.Unmarshal([]byte(s), &msg) == nil && strings.EqualFold(strings.TrimSpace(msg.Message), "Not Found") {
		return true
	}
	return strings.Contains(strings.ToLower(s), "not found") && strings.Contains(s, "404")
}
