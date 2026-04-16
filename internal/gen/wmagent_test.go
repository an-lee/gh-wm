package gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
)

func TestWriteWMAgent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := WriteWMAgent(dir, "owner/name", []string{"0 1 * * *", "0 1 * * *", ""}, []string{"ubuntu-latest"}, nil, true); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "wm-agent.yml"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, p := range []string{"owner/name", "0 1 * * *", "agent-resolve.yml", "cron:", `runs_on: '["ubuntu-latest"]'`, "install_claude_code: true", "task_name:", "force_task:", "has_tasks == 'true'"} {
		if !strings.Contains(s, p) {
			t.Fatalf("missing %q in %s", p, s)
		}
	}
}

func TestWriteWMAgent_DefaultSchedule(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := WriteWMAgent(dir, "o/r", nil, nil, nil, true); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "wm-agent.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "0 22 * * 1-5") {
		t.Fatal("default schedule missing")
	}
	if !strings.Contains(string(b), `runs_on: '["ubuntu-latest"]'`) {
		t.Fatal("default runs_on missing")
	}
}

func TestWriteWMAgent_SelfHostedLabels(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	labels := []string{"self-hosted", "linux"}
	if err := WriteWMAgent(dir, "o/r", []string{"0 1 * * *"}, labels, nil, true); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "wm-agent.yml"))
	if err != nil {
		t.Fatal(err)
	}
	want := `runs_on: '["self-hosted","linux"]'`
	if !strings.Contains(string(b), want) {
		t.Fatalf("want %q in output", want)
	}
}

func TestWriteWMAgent_InstallClaudeCodeFalse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := WriteWMAgent(dir, "o/r", []string{"0 1 * * *"}, []string{"ubuntu-latest"}, nil, false); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "wm-agent.yml"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "install_claude_code: false") {
		t.Fatal("expected install_claude_code: false in reusable workflow call")
	}
}

func TestWriteWMAgent_PreStepsInline(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	preSteps := []config.StepDef{
		{Uses: "jdx/mise-action@v4", With: map[string]any{"cache": true}},
		{Name: "Bundle install", Run: "bundle install"},
	}
	if err := WriteWMAgent(dir, "o/r", []string{"0 1 * * *"}, []string{"ubuntu-latest"}, preSteps, true); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "wm-agent.yml"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if strings.Contains(s, "agent-run.yml") {
		t.Fatal("expected inline run job without reusable agent-run.yml")
	}
	for _, p := range []string{
		"jdx/mise-action@v4",
		"Bundle install",
		"bundle install",
		"has_tasks == 'true'",
		"runs-on: ${{ fromJson('[\"ubuntu-latest\"]') }}",
		"actions/setup-go@v5",
		"Install Claude Code",
		"claude.ai/install.sh",
	} {
		if !strings.Contains(s, p) {
			t.Fatalf("missing %q in generated yaml", p)
		}
	}
}

func TestWriteWMAgent_PreStepsInline_NoClaudeInstall(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	preSteps := []config.StepDef{
		{Uses: "jdx/mise-action@v4"},
	}
	if err := WriteWMAgent(dir, "o/r", []string{"0 1 * * *"}, []string{"ubuntu-latest"}, preSteps, false); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "wm-agent.yml"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if strings.Contains(s, "Install Claude Code") || strings.Contains(s, "claude.ai/install.sh") {
		t.Fatal("did not expect Claude install steps when installClaudeCode is false")
	}
}
