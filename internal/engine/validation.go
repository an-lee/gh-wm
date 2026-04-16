package engine

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/an-lee/gh-wm/internal/types"
)

// MaxAgentOutputBytes is the upper bound for combined agent stdout/summary size after a successful run.
const maxAgentOutputBytes = 12 << 20 // 12 MiB

// validateAgentOutputErr returns nil only if the agent exited successfully and output size is within bounds.
func validateAgentOutputErr(ctx context.Context, res *types.AgentResult) error {
	if res == nil {
		return fmt.Errorf("agent result is nil")
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || res.TimedOut {
		return fmt.Errorf("agent timed out (context deadline exceeded)")
	}
	if !res.Success {
		return fmt.Errorf("agent process failed (exit code %d)", res.ExitCode)
	}
	if res.AgentStdoutPath != "" {
		st, err := os.Stat(res.AgentStdoutPath)
		if err != nil {
			return fmt.Errorf("agent output stat: %w", err)
		}
		if st.Size() > maxAgentOutputBytes {
			return fmt.Errorf("agent output exceeds maximum size (%d bytes)", maxAgentOutputBytes)
		}
		return nil
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
