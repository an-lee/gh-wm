package config

import (
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
