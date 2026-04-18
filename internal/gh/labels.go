package gh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/an-lee/gh-wm/internal/retry"
)

// DefaultRepoLabelColor is a neutral GitHub label color (6 hex chars, no leading #).
const DefaultRepoLabelColor = "ededed"

// EnsureRepoLabel creates the label in the repository if it does not already exist.
func EnsureRepoLabel(ctx context.Context, repo, label string) error {
	return EnsureRepoLabels(ctx, repo, []string{label})
}

// EnsureRepoLabels ensures each non-empty label exists (lists all labels via REST with pagination, then POSTs missing).
func EnsureRepoLabels(ctx context.Context, repo string, labels []string) error {
	want := dedupeNonEmptyLabels(labels)
	if len(want) == 0 {
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
	existing, err := listAllRepoLabelsREST(ctx, c, owner, name)
	if err != nil {
		return err
	}
	for _, l := range want {
		if _, ok := existing[l]; ok {
			continue
		}
		if err := createRepoLabelREST(ctx, c, owner, name, l); err != nil {
			return fmt.Errorf("ensure label %q: %w", l, err)
		}
		existing[l] = struct{}{}
	}
	return nil
}

func dedupeNonEmptyLabels(in []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, l := range in {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if _, ok := seen[l]; ok {
			continue
		}
		seen[l] = struct{}{}
		out = append(out, l)
	}
	return out
}

func listAllRepoLabelsREST(ctx context.Context, c restGetter, owner, name string) (map[string]struct{}, error) {
	existing := make(map[string]struct{})
	for page := 1; ; page++ {
		path := fmt.Sprintf("repos/%s/%s/labels?per_page=100&page=%d", owner, name, page)
		var rows []struct {
			Name string `json:"name"`
		}
		err := retry.WithAttempts(ctx, 3, func() error {
			resp, err := c.RequestWithContext(ctx, "GET", path, nil)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, rerr := io.ReadAll(resp.Body)
			if rerr != nil {
				return rerr
			}
			if resp.StatusCode == http.StatusOK {
				if err := json.Unmarshal(body, &rows); err != nil {
					return err
				}
				return nil
			}
			if retry.IsTransientHTTPStatus(resp.StatusCode) {
				return fmt.Errorf("list labels: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
			}
			return fmt.Errorf("list labels: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		})
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			if r.Name != "" {
				existing[r.Name] = struct{}{}
			}
		}
		if len(rows) < 100 {
			break
		}
	}
	return existing, nil
}

type restGetter interface {
	RequestWithContext(ctx context.Context, method string, path string, body io.Reader) (*http.Response, error)
}

func createRepoLabelREST(ctx context.Context, c restGetter, owner, name, label string) error {
	createPath := fmt.Sprintf("repos/%s/%s/labels", owner, name)
	payload := map[string]string{"name": label, "color": DefaultRepoLabelColor}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return retry.WithAttempts(ctx, 3, func() error {
		resp, err := c.RequestWithContext(ctx, "POST", createPath, bytes.NewReader(b))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, rerr := io.ReadAll(resp.Body)
		if rerr != nil {
			return rerr
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		if resp.StatusCode == http.StatusUnprocessableEntity && strings.Contains(strings.ToLower(string(body)), "already exists") {
			return nil
		}
		if retry.IsTransientHTTPStatus(resp.StatusCode) {
			return fmt.Errorf("create label %q: HTTP %d: %s", label, resp.StatusCode, strings.TrimSpace(string(body)))
		}
		return fmt.Errorf("create label %q: HTTP %d: %s", label, resp.StatusCode, strings.TrimSpace(string(body)))
	})
}
