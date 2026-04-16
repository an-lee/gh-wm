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

var updateCmd = &cobra.Command{
	Use:   "update [task-name...]",
	Short: "Re-fetch task files from their source: URL or owner/repo/path",
	Long: `Updates .wm/tasks/*.md files that have a source: field in frontmatter (set when adding via gh wm add).

source: may be an https URL or an owner/repo/path shorthand (same as gh aw), e.g. owner/repo/workflows/task.md.

With no arguments, updates every task with a non-empty source. Pass task names (with or without .md) to update specific tasks.`,
	RunE: runUpdate,
}

func runUpdate(_ *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	tasksDir := filepath.Join(cwd, ".wm", "tasks")
	if _, err := os.Stat(tasksDir); os.IsNotExist(err) {
		return fmt.Errorf(".wm/tasks not found; run gh wm init first")
	}
	tasks, err := config.LoadTasksDir(tasksDir)
	if err != nil {
		return err
	}

	var toUpdate []*config.Task
	if len(args) == 0 {
		for _, t := range tasks {
			if t.Source() != "" {
				toUpdate = append(toUpdate, t)
			}
		}
		if len(toUpdate) == 0 {
			fmt.Fprintln(os.Stderr, "No tasks with source: in frontmatter. Add tasks via gh wm add <url | owner/repo/task | path> to record a source.")
			return nil
		}
	} else {
		seen := make(map[string]bool)
		for _, a := range args {
			n := strings.TrimSuffix(strings.TrimSpace(a), ".md")
			var found *config.Task
			for _, t := range tasks {
				if t.Name == n {
					found = t
					break
				}
			}
			if found == nil {
				return fmt.Errorf("unknown task %q", n)
			}
			if found.Source() == "" {
				return fmt.Errorf("task %q has no source: field; cannot update", n)
			}
			if !seen[n] {
				seen[n] = true
				toUpdate = append(toUpdate, found)
			}
		}
	}

	client := &http.Client{Timeout: 60 * time.Second}
	updated := 0
	for _, t := range toUpdate {
		src := t.Source()
		fetchURL, err := resolveSourceToURL(src)
		if err != nil {
			return fmt.Errorf("%s: %w", t.Name, err)
		}
		resp, err := client.Get(fetchURL)
		if err != nil {
			return fmt.Errorf("%s: fetch %q: %w", t.Name, fetchURL, err)
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return fmt.Errorf("%s: read body: %w", t.Name, err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s: HTTP %s from %s (source %q)", t.Name, resp.Status, fetchURL, src)
		}
		yamlRaw, _, err := config.SplitFrontmatter(string(bodyBytes))
		if err != nil {
			return fmt.Errorf("%s: invalid markdown (frontmatter): %w", t.Name, err)
		}
		var fm map[string]any
		if err := yaml.Unmarshal([]byte(yamlRaw), &fm); err != nil {
			return fmt.Errorf("%s: task frontmatter YAML: %w", t.Name, err)
		}
		data := bodyBytes
		if fm == nil {
			data = injectSource(bodyBytes, src)
		} else if _, ok := fm["source"]; !ok {
			data = injectSource(bodyBytes, src)
		}
		if err := os.WriteFile(t.Path, data, 0o644); err != nil {
			return fmt.Errorf("%s: write: %w", t.Name, err)
		}
		fmt.Fprintf(os.Stderr, "Updated %s\n", t.Path)
		updated++
	}
	if updated > 0 {
		fmt.Fprintln(os.Stderr, "Run: gh wm upgrade")
	}
	return nil
}
