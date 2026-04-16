package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var addCmd = &cobra.Command{
	Use:   "add <owner/repo/task | url | path>",
	Short: "Download a task by GitHub shorthand, URL, or local path into .wm/tasks/",
	Long: `Adds a task Markdown file under .wm/tasks/.

  gh wm add owner/repo/task-name
    Fetches from the upstream repo on the default branch (main), trying
    workflows/<task>.md first (gh aw layout), then .wm/tasks/<task>.md (gh wm layout).
    Records source: as owner/repo/workflows/... or owner/repo/.wm/tasks/...

  gh wm add https://...
    Downloads the file and records source: as the URL.

  gh wm add ./path/to/task.md
    Copies a local file (no source: unless already in frontmatter).

After a successful add, gh wm upgrade runs automatically so wm-agent.yml matches the new task.`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func runAdd(_ *cobra.Command, args []string) error {
	src := strings.TrimSpace(args[0])
	var data []byte
	var err error
	var sourceToInject string // non-empty => inject if frontmatter lacks source:
	var destBase string       // when set, used as output filename (basename)

	client := &http.Client{Timeout: 60 * time.Second}

	switch {
	case isGitHubShorthand(src):
		owner, repo, task := parseGitHubShorthand(src)
		taskFile := normalizeTaskFileName(task)
		if taskFile == "" {
			return fmt.Errorf("invalid task name in shorthand %q", src)
		}
		var body []byte
		var source string
		body, _, source, err = fetchShorthand(client, owner, repo, taskFile)
		if err != nil {
			return err
		}
		data = body
		sourceToInject = source
		destBase = filepath.Base(taskFile)

	case strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://"):
		resp, err := client.Get(src)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP %s", resp.Status)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		sourceToInject = src

	default:
		data, err = os.ReadFile(src)
		if err != nil {
			return err
		}
	}
	yamlRaw, _, err := config.SplitFrontmatter(string(data))
	if err != nil {
		return fmt.Errorf("task file must start with YAML frontmatter (---): %w", err)
	}
	var fm map[string]any
	if err := yaml.Unmarshal([]byte(yamlRaw), &fm); err != nil {
		return fmt.Errorf("task frontmatter YAML: %w", err)
	}
	if sourceToInject != "" && fm != nil {
		if _, ok := fm["source"]; !ok {
			data = injectSource(data, sourceToInject)
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	tasksDir := filepath.Join(cwd, ".wm", "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		return err
	}
	base := destBase
	if base == "" {
		base = filepath.Base(src)
		if idx := strings.Index(base, "?"); idx >= 0 {
			base = base[:idx]
		}
		if !strings.HasSuffix(strings.ToLower(base), ".md") {
			base = base + ".md"
		}
	}
	dest := filepath.Join(tasksDir, base)
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Wrote %s\n", dest)
	return runUpgrade(nil, nil)
}

// injectSource inserts source: <ref> immediately after the opening --- line.
// ref is either an https URL or an owner/repo/path shorthand (gh aw style).
func injectSource(data []byte, ref string) []byte {
	s := string(data)
	nl := strings.Index(s, "\n")
	if nl < 0 {
		return data
	}
	line := fmt.Sprintf("source: %q\n", ref)
	return []byte(s[:nl+1] + line + s[nl+1:])
}
