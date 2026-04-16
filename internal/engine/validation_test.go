package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/an-lee/gh-wm/internal/types"
)

func TestValidateAgentOutputErr(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	if err := validateAgentOutputErr(ctx, nil); err == nil {
		t.Fatal("expected error for nil result")
	}
	if err := validateAgentOutputErr(ctx, &types.AgentResult{Success: false, ExitCode: 1}); err == nil {
		t.Fatal("expected error when agent failed")
	}
	if err := validateAgentOutputErr(ctx, &types.AgentResult{Success: true, ExitCode: 0, Summary: "ok"}); err != nil {
		t.Fatal(err)
	}
	long := strings.Repeat("x", maxAgentOutputBytes+1)
	if err := validateAgentOutputErr(ctx, &types.AgentResult{Success: true, ExitCode: 0, Summary: long}); err == nil {
		t.Fatal("expected error for oversized output")
	}
}

func TestValidateAgentOutputErr_OversizedLogFile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dir := t.TempDir()
	p := filepath.Join(dir, "agent.log")
	if err := os.WriteFile(p, []byte(strings.Repeat("x", maxAgentOutputBytes+1)), 0o644); err != nil {
		t.Fatal(err)
	}
	err := validateAgentOutputErr(ctx, &types.AgentResult{
		Success:         true,
		ExitCode:        0,
		AgentStdoutPath: p,
	})
	if err == nil {
		t.Fatal("expected error for oversized log file")
	}
}
