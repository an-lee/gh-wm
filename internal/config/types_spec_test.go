package config

import "testing"

func TestGlobalConfig_TypedSpec(t *testing.T) {
	t.Parallel()
	g := &GlobalConfig{
		Version:  1,
		Engine:   "claude",
		Model:    "x",
		MaxTurns: 10,
	}
	g.Workflow.RunsOn = []string{"ubuntu-latest"}
	g.Context.Files = []string{"README.md"}
	g.PR.Draft = true
	s := g.TypedSpec()
	if s == nil || s.Engine != "claude" || len(s.WorkflowRunsOn) != 1 || s.WorkflowRunsOn[0] != "ubuntu-latest" {
		t.Fatalf("TypedSpec: %+v", s)
	}
}
