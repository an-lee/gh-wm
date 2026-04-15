package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SplitFrontmatter extracts YAML between first pair of --- lines and body.
func SplitFrontmatter(content string) (yamlRaw string, body string, err error) {
	content = strings.TrimPrefix(content, "\ufeff")
	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return "", content, fmt.Errorf("missing opening ---")
	}
	var ylines []string
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			body = strings.Join(lines[i+1:], "\n")
			return strings.TrimSpace(strings.Join(ylines, "\n")), strings.TrimSpace(body), nil
		}
		ylines = append(ylines, lines[i])
	}
	return "", "", fmt.Errorf("unclosed frontmatter")
}

// LoadTaskFile reads and parses a single task markdown file.
func LoadTaskFile(path string) (*Task, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	yamlRaw, body, err := SplitFrontmatter(string(b))
	if err != nil {
		return nil, err
	}
	var fm map[string]any
	if err := yaml.Unmarshal([]byte(yamlRaw), &fm); err != nil {
		return nil, fmt.Errorf("yaml frontmatter: %w", err)
	}
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return &Task{
		Name:        name,
		Path:        path,
		Frontmatter: fm,
		Body:        body,
	}, nil
}

// LoadTasksDir loads all .md files from dir (non-recursive).
func LoadTasksDir(dir string) ([]*Task, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []*Task
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if !strings.HasSuffix(n, ".md") || strings.HasSuffix(n, ".md.disabled") {
			continue
		}
		t, err := LoadTaskFile(filepath.Join(dir, n))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", n, err)
		}
		out = append(out, t)
	}
	return out, nil
}
