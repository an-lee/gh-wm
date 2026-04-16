package engine

import (
	"fmt"

	"github.com/an-lee/gh-wm/internal/types"
)

// MaxAgentOutputBytes is the upper bound for combined agent stdout/summary size after a successful run.
const maxAgentOutputBytes = 12 << 20 // 12 MiB

// validateAgentOutputErr returns nil only if the agent exited successfully and output size is within bounds.
func validateAgentOutputErr(res *types.AgentResult) error {
	if res == nil {
		return fmt.Errorf("agent result is nil")
	}
	if !res.Success {
		return fmt.Errorf("agent process failed (exit code %d)", res.ExitCode)
	}
	combined := res.Summary
	if combined == "" {
		combined = res.Stdout
	}
	if len(combined) > maxAgentOutputBytes {
		return fmt.Errorf("agent output exceeds maximum size (%d bytes)", maxAgentOutputBytes)
	}
	return nil
}
