package engine

import (
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

func TestValidateEventContext(t *testing.T) {
	t.Parallel()
	if err := validateEventContext(nil); err == nil {
		t.Fatal("expected error for nil event")
	}
	if err := validateEventContext(&types.GitHubEvent{Name: "issues", Payload: nil}); err == nil {
		t.Fatal("expected error for nil payload")
	}
	if err := validateEventContext(&types.GitHubEvent{Name: "", Payload: map[string]any{}}); err == nil {
		t.Fatal("expected error for empty name")
	}
	if err := validateEventContext(&types.GitHubEvent{Name: "unknown", Payload: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	if err := validateEventContext(&types.GitHubEvent{Name: "issues", Payload: map[string]any{"action": "opened"}}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateTaskConfig(t *testing.T) {
	t.Parallel()
	glob := &config.GlobalConfig{}
	glob.Engine = "claude"
	task := &config.Task{Frontmatter: map[string]any{}}
	if err := validateTaskConfig(task, glob); err != nil {
		t.Fatal(err)
	}
	task.Frontmatter["engine"] = "bogus"
	if err := validateTaskConfig(task, glob); err == nil {
		t.Fatal("expected error for unknown engine")
	}
}

func TestValidateTaskConfig_Reaction(t *testing.T) {
	glob := &config.GlobalConfig{Engine: "claude"}
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		task := &config.Task{Frontmatter: map[string]any{
			"on": map[string]any{"reaction": "eyes"},
		}}
		if err := validateTaskConfig(task, glob); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		task := &config.Task{Frontmatter: map[string]any{
			"on": map[string]any{"reaction": "not-a-reaction"},
		}}
		if err := validateTaskConfig(task, glob); err == nil {
			t.Fatal("expected error for invalid on.reaction")
		}
	})
	t.Run("invalid even with WM_AGENT_CMD", func(t *testing.T) {
		t.Setenv("WM_AGENT_CMD", "echo")
		task := &config.Task{Frontmatter: map[string]any{
			"on": map[string]any{"reaction": "not-a-reaction"},
		}}
		if err := validateTaskConfig(task, glob); err == nil {
			t.Fatal("expected error for invalid on.reaction when WM_AGENT_CMD is set")
		}
	})
}
