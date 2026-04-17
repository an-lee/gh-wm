package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPersistedAgentResult_MissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, err := loadPersistedAgentResult(dir)
	if err == nil {
		t.Fatal("expected error for missing result.json")
	}
}

func TestLoadPersistedAgentResult_NoAgentResult(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, resultFileName)
	if err := os.WriteFile(path, []byte(`{"success":true,"phase":"validation"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := loadPersistedAgentResult(dir)
	if err == nil {
		t.Fatal("expected error when agent_result is absent")
	}
}

func TestLoadPersistedAgentResult_AgentNotSuccessful(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, resultFileName)
	raw := `{
  "success": true,
  "agent_result": {
    "success": false,
    "exit_code": 1
  }
}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := loadPersistedAgentResult(dir)
	if err == nil {
		t.Fatal("expected error when agent_result.success is false")
	}
}

func TestLoadPersistedAgentResult_TopLevelSuccessFalse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, resultFileName)
	raw := `{
  "success": false,
  "agent_result": {
    "success": true,
    "exit_code": 0
  }
}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := loadPersistedAgentResult(dir)
	if err == nil {
		t.Fatal("expected error when top-level success is false")
	}
}
