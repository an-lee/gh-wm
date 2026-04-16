package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// rawGitHubBaseURL is the prefix for raw.githubusercontent.com-style URLs.
// Tests may override it to point at httptest.Server.
var rawGitHubBaseURL = "https://raw.githubusercontent.com"

// isGitHubShorthand reports whether src is owner/repo/task (three path segments, not a URL or local path).
func isGitHubShorthand(src string) bool {
	src = strings.TrimSpace(src)
	if src == "" {
		return false
	}
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		return false
	}
	if strings.HasPrefix(src, "/") || strings.HasPrefix(src, ".") {
		return false
	}
	parts := strings.SplitN(src, "/", 3)
	return len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != ""
}

// parseGitHubShorthand returns owner, repo, and the third path segment (task name, optional .md).
func parseGitHubShorthand(src string) (owner, repo, task string) {
	parts := strings.SplitN(strings.TrimSpace(src), "/", 3)
	if len(parts) != 3 {
		return "", "", ""
	}
	return parts[0], parts[1], parts[2]
}

func normalizeTaskFileName(task string) string {
	t := strings.TrimSpace(task)
	if t == "" {
		return t
	}
	if !strings.HasSuffix(strings.ToLower(t), ".md") {
		return t + ".md"
	}
	return t
}

func rawGitHubFileURL(owner, repo, pathUnderMain string) string {
	base := strings.TrimSuffix(rawGitHubBaseURL, "/")
	return fmt.Sprintf("%s/%s/%s/main/%s", base, owner, repo, pathUnderMain)
}

// fetchShorthand tries workflows/<file> then .wm/tasks/<file> on the default branch, same as gh aw + gh wm layouts.
func fetchShorthand(client *http.Client, owner, repo, taskFile string) (data []byte, fetchedURL string, source string, err error) {
	candidates := []struct {
		pathInRepo string
		source     string
	}{
		{"workflows/" + taskFile, fmt.Sprintf("%s/%s/workflows/%s", owner, repo, taskFile)},
		{".wm/tasks/" + taskFile, fmt.Sprintf("%s/%s/.wm/tasks/%s", owner, repo, taskFile)},
	}
	for _, c := range candidates {
		u := rawGitHubFileURL(owner, repo, c.pathInRepo)
		resp, err := client.Get(u)
		if err != nil {
			return nil, "", "", err
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, "", "", err
		}
		if resp.StatusCode == http.StatusOK {
			return body, u, c.source, nil
		}
		if resp.StatusCode == http.StatusNotFound {
			continue
		}
		return nil, "", "", fmt.Errorf("HTTP %s from %s", resp.Status, u)
	}
	return nil, "", "", fmt.Errorf("task %q not found in %s/%s on main (tried workflows/ and .wm/tasks/)", taskFile, owner, repo)
}

// resolveSourceToURL turns a stored source: value into a fetch URL.
// Accepts https URLs unchanged, or owner/repo/path/to/file.md (path after repo root on main).
func resolveSourceToURL(src string) (string, error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return "", fmt.Errorf("empty source")
	}
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		return src, nil
	}
	parts := strings.SplitN(src, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", fmt.Errorf("invalid source %q: expected https URL or owner/repo/path", src)
	}
	return rawGitHubFileURL(parts[0], parts[1], parts[2]), nil
}
