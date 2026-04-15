package templates

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed data/config.yml data/CLAUDE.md data/tasks/*.md
var embedded embed.FS

// WriteConfig writes default config.yml
func WriteConfig(wmDir string) error {
	b, err := embedded.ReadFile("data/config.yml")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(wmDir, "config.yml"), b, 0o644)
}

// WriteCLAUDE writes CLAUDE.md at repo root if missing
func WriteCLAUDE(repoRoot string) error {
	p := filepath.Join(repoRoot, "CLAUDE.md")
	if _, err := os.Stat(p); err == nil {
		return nil
	}
	b, err := embedded.ReadFile("data/CLAUDE.md")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}

// WriteStarterTasks extracts embedded data/tasks/*.md
func WriteStarterTasks(tasksDir string) error {
	return fs.WalkDir(embedded, "data/tasks", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, err := embedded.ReadFile(path)
		if err != nil {
			return err
		}
		dst := filepath.Join(tasksDir, filepath.Base(path))
		return os.WriteFile(dst, b, 0o644)
	})
}
