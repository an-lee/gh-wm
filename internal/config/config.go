package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DirName  = ".wm"
	TasksDir = "tasks"
)

// Load loads global config from repoRoot/.wm/config.yml and tasks from .wm/tasks/*.md
func Load(repoRoot string) (*GlobalConfig, []*Task, error) {
	cfgPath := filepath.Join(repoRoot, DirName, "config.yml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, nil, err
	}
	var g GlobalConfig
	if err := yaml.Unmarshal(data, &g); err != nil {
		return nil, nil, err
	}
	tasksDir := filepath.Join(repoRoot, DirName, TasksDir)
	tasks, err := LoadTasksDir(tasksDir)
	if err != nil {
		return nil, nil, err
	}
	return &g, tasks, nil
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
