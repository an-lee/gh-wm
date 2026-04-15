package config

import (
	"encoding/json"
	"strings"
)

// GlobalConfig is .wm/config.yml
type GlobalConfig struct {
	Version  int    `yaml:"version"`
	Engine   string `yaml:"engine"`
	Model    string `yaml:"model"`
	MaxTurns int    `yaml:"max_turns"`
	Context  struct {
		Files []string `yaml:"files"`
	} `yaml:"context"`
	PR struct {
		Draft     bool     `yaml:"draft"`
		Reviewers []string `yaml:"reviewers"`
	} `yaml:"pr"`
}

// WMExtension is wm: in task frontmatter
type WMExtension struct {
	StateLabels map[string]string `yaml:"state_labels"`
}

// Task holds one .wm/tasks/*.md file
type Task struct {
	Name        string         // filename without .md
	Path        string         // absolute path
	Frontmatter map[string]any // raw YAML
	Body        string         // markdown prompt
}

// OnMap returns the on: block
func (t *Task) OnMap() map[string]any {
	if t == nil || t.Frontmatter == nil {
		return nil
	}
	on, _ := t.Frontmatter["on"].(map[string]any)
	return on
}

// Engine returns engine: from frontmatter or empty
func (t *Task) Engine() string {
	if t == nil {
		return ""
	}
	if e, ok := t.Frontmatter["engine"].(string); ok {
		return e
	}
	return ""
}

// ScheduleString extracts schedule from on: block for union in wm-agent.yml
func (t *Task) ScheduleString() string {
	on := t.OnMap()
	if on == nil {
		return ""
	}
	s, ok := on["schedule"]
	if !ok {
		return ""
	}
	switch v := s.(type) {
	case string:
		return v
	default:
		return ""
	}
}

// SafeOutputsMap returns the safe-outputs: block from frontmatter (keys are hints for enabled outputs).
func (t *Task) SafeOutputsMap() map[string]any {
	if t == nil || t.Frontmatter == nil {
		return nil
	}
	so, _ := t.Frontmatter["safe-outputs"].(map[string]any)
	return so
}

// HasSafeOutputKey reports whether safe-outputs contains a top-level key (e.g. "add-comment", "create-pull-request").
func (t *Task) HasSafeOutputKey(key string) bool {
	m := t.SafeOutputsMap()
	if m == nil {
		return false
	}
	_, ok := m[key]
	return ok
}

// TimeoutMinutes returns timeout-minutes from frontmatter, or defaultM if unset/invalid.
func (t *Task) TimeoutMinutes(defaultM int) int {
	if t == nil || t.Frontmatter == nil {
		return defaultM
	}
	switch v := t.Frontmatter["timeout-minutes"].(type) {
	case int:
		return clampTimeout(v, defaultM)
	case int64:
		return clampTimeout(int(v), defaultM)
	case float64:
		return clampTimeout(int(v), defaultM)
	default:
		return defaultM
	}
}

const maxTimeoutMinutes = 480 // 8h cap

func clampTimeout(m, defaultM int) int {
	if m <= 0 {
		return defaultM
	}
	if m > maxTimeoutMinutes {
		return maxTimeoutMinutes
	}
	return m
}

// WM returns parsed wm: extension from frontmatter.
func (t *Task) WM() WMExtension {
	var w WMExtension
	if t == nil || t.Frontmatter == nil {
		return w
	}
	raw, ok := t.Frontmatter["wm"].(map[string]any)
	if !ok {
		return w
	}
	if sl, ok := raw["state_labels"].(map[string]any); ok {
		w.StateLabels = make(map[string]string)
		for k, v := range sl {
			if s, ok := v.(string); ok {
				w.StateLabels[k] = s
			}
		}
	}
	return w
}

// ToolsYAML returns raw tools: value as YAML-friendly string for env (Phase 3 / WM_TASK_TOOLS).
func (t *Task) ToolsYAML() string {
	if t == nil || t.Frontmatter == nil {
		return ""
	}
	if tools, ok := t.Frontmatter["tools"]; ok {
		return toolsString(tools)
	}
	return ""
}

func toolsString(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return ""
		}
		return string(b)
	}
}
