package engine

import (
	"strings"
	"testing"

	"github.com/an-lee/gh-wm/internal/types"
)

func TestValidateAgentOutputErr(t *testing.T) {
	t.Parallel()
	if err := validateAgentOutputErr(nil); err == nil {
		t.Fatal("expected error for nil result")
	}
	if err := validateAgentOutputErr(&types.AgentResult{Success: false, ExitCode: 1}); err == nil {
		t.Fatal("expected error when agent failed")
	}
	if err := validateAgentOutputErr(&types.AgentResult{Success: true, ExitCode: 0, Summary: "ok"}); err != nil {
		t.Fatal(err)
	}
	long := strings.Repeat("x", maxAgentOutputBytes+1)
	if err := validateAgentOutputErr(&types.AgentResult{Success: true, ExitCode: 0, Summary: long}); err == nil {
		t.Fatal("expected error for oversized output")
	}
}
