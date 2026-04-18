package ghclient

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/an-lee/gh-wm/internal/gh"
	"github.com/an-lee/gh-wm/internal/retry"
)

const ghLabelListMax = 9999

// EnsureRepoLabel creates the label in the repository if it does not already exist.
func EnsureRepoLabel(ctx context.Context, repo, label string) error {
	return EnsureRepoLabels(ctx, repo, []string{label})
}

// EnsureRepoLabels ensures each non-empty label exists on the repo (batched list + create missing).
func EnsureRepoLabels(ctx context.Context, repo string, labels []string) error {
	want := dedupeNonEmptyLabels(labels)
	if len(want) == 0 {
		return nil
	}
	if useREST() {
		return gh.EnsureRepoLabels(ctx, repo, want)
	}
	existing, err := listRepoLabelsSubprocess(ctx, repo)
	if err != nil {
		return err
	}
	for _, l := range want {
		if _, ok := existing[l]; ok {
			continue
		}
		if err := createRepoLabelSubprocess(ctx, repo, l); err != nil {
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

func listRepoLabelsSubprocess(ctx context.Context, repo string) (map[string]struct{}, error) {
	var names map[string]struct{}
	err := retry.WithAttempts(ctx, 3, func() error {
		cmd := exec.CommandContext(ctx, "gh", "label", "list", "--repo", repo, "--json", "name", "--limit", fmt.Sprintf("%d", ghLabelListMax))
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("gh label list: %w: %s", err, strings.TrimSpace(string(out)))
		}
		var rows []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(out, &rows); err != nil {
			return fmt.Errorf("parse gh label list json: %w", err)
		}
		names = make(map[string]struct{}, len(rows))
		for _, r := range rows {
			if r.Name != "" {
				names[r.Name] = struct{}{}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return names, nil
}

func createRepoLabelSubprocess(ctx context.Context, repo, label string) error {
	return retry.WithAttempts(ctx, 3, func() error {
		cmd := exec.CommandContext(ctx, "gh", "label", "create", label, "--repo", repo, "--color", gh.DefaultRepoLabelColor)
		out, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		s := string(out)
		if strings.Contains(strings.ToLower(s), "already exists") {
			return nil
		}
		return fmt.Errorf("gh label create %q: %w: %s", label, err, strings.TrimSpace(s))
	})
}
