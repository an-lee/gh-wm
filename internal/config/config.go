package config

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// loadCache caches Load results per repo root for the lifetime of a process.
// Safe for gh-wm's single-shot-per-event model where config files don't change
// between resolve→run calls within the same process.
var loadCache sync.Map

type cachedLoad struct {
	cfg   *GlobalConfig
	tasks []*Task
	err   error
}

// ResetLoadCache clears the config cache. Intended for use in tests to prevent
// cross-test pollution; not needed in normal operation.
func ResetLoadCache() { loadCache = sync.Map{} }

const (
	DirName  = ".wm"
	TasksDir = "tasks"
)

// Load loads global config from repoRoot/.wm/config.yml and tasks from .wm/tasks/*.md.
// Results are cached per repoRoot for the lifetime of the process to avoid redundant
// file I/O and YAML parsing across multiple calls (e.g. resolve→run in one CLI invocation).
func Load(repoRoot string) (*GlobalConfig, []*Task, error) {
	if cached, ok := loadCache.Load(repoRoot); ok {
		c := cached.(*cachedLoad)
		return c.cfg, c.tasks, c.err
	}
	cfgPath := filepath.Join(repoRoot, DirName, "config.yml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		loadCache.Store(repoRoot, &cachedLoad{err: err})
		return nil, nil, err
	}
	var g GlobalConfig
	if err := yaml.Unmarshal(data, &g); err != nil {
		loadCache.Store(repoRoot, &cachedLoad{err: err})
		return nil, nil, err
	}
	tasksDir := filepath.Join(repoRoot, DirName, TasksDir)
	tasks, err := LoadTasksDir(tasksDir)
	result := &cachedLoad{cfg: &g, tasks: tasks, err: err}
	loadCache.Store(repoRoot, result)
	return result.cfg, result.tasks, result.err
}

// LoadGlobalOnly reads only repoRoot/.wm/config.yml (no tasks). If the file is missing, returns (nil, nil).
func LoadGlobalOnly(repoRoot string) (*GlobalConfig, error) {
	cfgPath := filepath.Join(repoRoot, DirName, "config.yml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return ParseGlobal(data)
}

// WorkflowRunsOnLabels returns GitHub Actions runner labels for generated wm-agent.yml.
// If g is nil or workflow.runs_on is empty, returns ["ubuntu-latest"].
func WorkflowRunsOnLabels(g *GlobalConfig) []string {
	if g == nil || len(g.Workflow.RunsOn) == 0 {
		return []string{"ubuntu-latest"}
	}
	return g.Workflow.RunsOn
}

// WorkflowInstallClaudeCode reports whether generated wm-agent.yml should pass install_claude_code: true
// to the agent-run reusable workflow. When unset in config, defaults to true.
func WorkflowInstallClaudeCode(g *GlobalConfig) bool {
	if g == nil || g.Workflow.InstallClaudeCode == nil {
		return true
	}
	return *g.Workflow.InstallClaudeCode
}

// DefaultGlobal returns minimal defaults when config.yml missing pieces
// ParseGlobal unmarshals config.yml bytes
func ParseGlobal(data []byte) (*GlobalConfig, error) {
	var g GlobalConfig
	if err := yaml.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

// DefaultGlobal fills defaults
func DefaultGlobal(g *GlobalConfig) *GlobalConfig {
	if g == nil {
		g = &GlobalConfig{Version: 1}
	}
	if g.Engine == "" {
		g.Engine = "claude"
	}
	if g.MaxTurns == 0 {
		g.MaxTurns = 100
	}
	return g
}
