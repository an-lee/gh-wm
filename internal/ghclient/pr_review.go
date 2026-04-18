package ghclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/an-lee/gh-wm/internal/gh"
)

// PullRequestHeadSHA returns the head commit SHA for a pull request.
func PullRequestHeadSHA(repo string, pr int) (string, error) {
	if useREST() {
		return gh.PullRequestHeadSHA(context.Background(), repo, pr)
	}
	owner, name, err := splitRepo(repo)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, name, pr)
	cmd := exec.Command("gh", "api", path, "--jq", ".head.sha")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("gh api pulls: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// SubmitPullRequestReview submits a pull request review.
func SubmitPullRequestReview(repo string, pr int, commitID, event, body string) error {
	if useREST() {
		return gh.SubmitPullRequestReview(context.Background(), repo, pr, commitID, event, body)
	}
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	apiPath := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", owner, name, pr)
	payload := map[string]any{"event": event}
	if strings.TrimSpace(commitID) != "" {
		payload["commit_id"] = commitID
	}
	if strings.TrimSpace(body) != "" {
		payload["body"] = body
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	cmd := exec.Command("gh", "api", "-X", "POST", apiPath, "--input", "-")
	cmd.Stdin = bytes.NewReader(b)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh api submit pr review: %w: %s", err, stderr.String())
	}
	return nil
}
