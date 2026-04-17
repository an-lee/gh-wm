package gh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// DefaultRepoLabelColor is a neutral GitHub label color (6 hex chars, no leading #).
const DefaultRepoLabelColor = "ededed"

// EnsureRepoLabel creates the label in the repository if it does not already exist.
func EnsureRepoLabel(ctx context.Context, repo, label string) error {
	if strings.TrimSpace(label) == "" {
		return nil
	}
	c, err := REST()
	if err != nil {
		return err
	}
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	enc := url.PathEscape(label)
	path := fmt.Sprintf("repos/%s/%s/labels/%s", owner, name, enc)
	resp, err := c.RequestWithContext(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("get label %q: HTTP %d: %s", label, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	createPath := fmt.Sprintf("repos/%s/%s/labels", owner, name)
	payload := map[string]string{"name": label, "color": DefaultRepoLabelColor}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp2, err := c.RequestWithContext(ctx, "POST", createPath, bytes.NewReader(b))
	if err != nil {
		return err
	}
	body2, _ := io.ReadAll(resp2.Body)
	_ = resp2.Body.Close()
	if resp2.StatusCode >= 200 && resp2.StatusCode < 300 {
		return nil
	}
	// Race: another process created the label between GET and POST.
	if resp2.StatusCode == http.StatusUnprocessableEntity && strings.Contains(strings.ToLower(string(body2)), "already exists") {
		return nil
	}
	return fmt.Errorf("create label %q: HTTP %d: %s", label, resp2.StatusCode, strings.TrimSpace(string(body2)))
}
