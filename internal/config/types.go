package config

import (
	"encoding/json"
	"strings"

	"github.com/an-lee/gh-wm/internal/config/spec"
)

// StepDef is one GitHub Actions job step (uses or run) for workflow.pre_steps.
type StepDef struct {
	Name string         `yaml:"name,omitempty"`
	Uses string         `yaml:"uses,omitempty"`
	Run  string         `yaml:"run,omitempty"`
	With map[string]any `yaml:"with,omitempty"`
	Env  map[string]any `yaml:"env,omitempty"`
	If   string         `yaml:"if,omitempty"`
}

// GlobalConfig is .wm/config.yml
type GlobalConfig struct {
	Version            int    `yaml:"version"`
	Engine             string `yaml:"engine"`
	Model              string `yaml:"model"`
	MaxTurns           int    `yaml:"max_turns"`
	ClaudeOutputFormat string `yaml:"claude_output_format,omitempty"`
	// Compat configures gh-aw–style ${{ }} validation and expansion in task bodies.
	Compat struct {
		// GhAWExpressions is error | warn | off (default error). When off, compile/init skip body scans.
		GhAWExpressions string `yaml:"gh_aw_expressions,omitempty"`
		// GhAWExpand enables runtime expansion of ${{ }} in task bodies at gh wm run (default true).
		GhAWExpand *bool `yaml:"gh_aw_expand,omitempty"`
	} `yaml:"compat,omitempty"`
	Workflow struct {
		RunsOn               []string  `yaml:"runs_on"`
		PreSteps             []StepDef `yaml:"pre_steps"`
		InstallClaudeCode    *bool     `yaml:"install_claude_code,omitempty"`
		GhWMExtensionVersion string    `yaml:"gh_wm_extension_version,omitempty"`
	} `yaml:"workflow"`
	Context struct {
		Files []string `yaml:"files"`
	} `yaml:"context"`
	PR struct {
		Draft     bool     `yaml:"draft"`
		Reviewers []string `yaml:"reviewers"`
	} `yaml:"pr"`
}

// TypedSpec returns a [spec.GlobalSpec] snapshot for tooling and JSON schema alignment.
func (g *GlobalConfig) TypedSpec() *spec.GlobalSpec {
	if g == nil {
		return nil
	}
	return &spec.GlobalSpec{
		Version:                      g.Version,
		Engine:                       g.Engine,
		Model:                        g.Model,
		MaxTurns:                     g.MaxTurns,
		ClaudeOutputFormat:           g.ClaudeOutputFormat,
		WorkflowRunsOn:               append([]string(nil), g.Workflow.RunsOn...),
		WorkflowGhWMExtensionVersion: WorkflowGhWMExtensionVersion(g),
		ContextFiles:                 append([]string(nil), g.Context.Files...),
		PRDraft:                      g.PR.Draft,
		PRReviewers:                  append([]string(nil), g.PR.Reviewers...),
	}
}

// Task holds one .wm/tasks/*.md file
type Task struct {
	Name        string         // filename without .md
	Path        string         // absolute path
	Frontmatter map[string]any // raw YAML
	Body        string         // markdown prompt
	// Spec is a typed view of Frontmatter from [spec.ParseTaskFrontmatter] (filled by LoadTaskFile).
	Spec *spec.TaskSpec `json:"-"`
}

var validGitHubReactionContents = map[string]struct{}{
	"+1":       {},
	"-1":       {},
	"laugh":    {},
	"confused": {},
	"heart":    {},
	"hooray":   {},
	"rocket":   {},
	"eyes":     {},
}

// ValidGitHubReaction reports whether s is a valid GitHub reactions API content string.
func ValidGitHubReaction(s string) bool {
	_, ok := validGitHubReactionContents[s]
	return ok
}

// OnMap returns the on: block
func (t *Task) OnMap() map[string]any {
	if t == nil || t.Frontmatter == nil {
		return nil
	}
	on, _ := t.Frontmatter["on"].(map[string]any)
	return on
}

// OnReactionContent returns on.reaction from frontmatter (trimmed), or empty if unset or not a string.
func (t *Task) OnReactionContent() string {
	on := t.OnMap()
	if on == nil {
		return ""
	}
	s, ok := on["reaction"].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

// Source returns source: from frontmatter (upstream URL for gh wm update), or empty.
func (t *Task) Source() string {
	if t == nil || t.Frontmatter == nil {
		return ""
	}
	s, _ := t.Frontmatter["source"].(string)
	return strings.TrimSpace(s)
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
