// Package spec provides typed views and validation for .wm task and global configuration.
package spec

import (
	"fmt"
	"strings"
)

// TaskSpec is a typed view of task YAML frontmatter (subset of gh-aw fields gh-wm interprets).
type TaskSpec struct {
	Description    string
	Source         string
	Engine         string
	TimeoutMinutes int
	// OnReaction is from on.reaction (GitHub reaction content string).
	OnReaction string
	// RawSafeOutputKeys lists top-level keys under safe-outputs: (dash form).
	RawSafeOutputKeys []string
}

var knownEngines = map[string]struct{}{
	"claude": {},
	"codex":  {},
}

// ParseTaskFrontmatter builds a TaskSpec from raw frontmatter and returns validation warnings (non-fatal).
func ParseTaskFrontmatter(fm map[string]any) (*TaskSpec, []string, error) {
	if fm == nil {
		return &TaskSpec{}, nil, nil
	}
	var s TaskSpec
	var warnings []string

	if v, ok := fm["description"].(string); ok {
		s.Description = strings.TrimSpace(v)
	}
	if v, ok := fm["source"].(string); ok {
		s.Source = strings.TrimSpace(v)
	}
	if v, ok := fm["engine"].(string); ok {
		s.Engine = strings.TrimSpace(v)
		if s.Engine != "" {
			el := strings.ToLower(s.Engine)
			if el == "copilot" {
				warnings = append(warnings, `engine "copilot" is no longer supported; use WM_AGENT_CMD or set engine to claude or codex`)
			} else if _, ok := knownEngines[el]; !ok {
				warnings = append(warnings, fmt.Sprintf("engine %q is not a built-in name (claude, codex); use WM_AGENT_CMD for custom CLIs", s.Engine))
			}
		}
	}
	switch v := fm["timeout-minutes"].(type) {
	case int:
		s.TimeoutMinutes = v
	case int64:
		s.TimeoutMinutes = int(v)
	case float64:
		s.TimeoutMinutes = int(v)
	}
	if on, ok := fm["on"].(map[string]any); ok && on != nil {
		if r, ok := on["reaction"].(string); ok {
			s.OnReaction = strings.TrimSpace(r)
			if s.OnReaction != "" && !validGitHubReaction(s.OnReaction) {
				warnings = append(warnings, fmt.Sprintf("on.reaction %q is not a known GitHub reaction API value", s.OnReaction))
			}
		}
	}
	if so, ok := fm["safe-outputs"].(map[string]any); ok && so != nil {
		for k := range so {
			s.RawSafeOutputKeys = append(s.RawSafeOutputKeys, k)
			if strings.Contains(k, "_") {
				warnings = append(warnings, fmt.Sprintf("safe-outputs key %q uses underscores; gh-aw style prefers dash form (e.g. add-labels). v2 may canonicalize keys.", k))
			}
		}
	}
	return &s, warnings, nil
}
