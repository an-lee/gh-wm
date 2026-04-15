// Package output runs post-agent steps inferred from safe-outputs: keys (hints, not enforced).
package output

import (
	"context"
	"fmt"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

// RunSuccessOutputs runs enabled outputs after a successful agent run (order: PR then comment).
func RunSuccessOutputs(ctx context.Context, glob *config.GlobalConfig, task *config.Task, tc *types.TaskContext, res *types.AgentResult) error {
	if glob == nil || task == nil || tc == nil || res == nil {
		return nil
	}
	if task.HasSafeOutputKey("create-pull-request") {
		if err := runPROutput(ctx, glob, task, tc, res); err != nil {
			return fmt.Errorf("create-pull-request: %w", err)
		}
	}
	if task.HasSafeOutputKey("add-labels") {
		if err := runLabelOutput(ctx, glob, task, tc, res); err != nil {
			return fmt.Errorf("add-labels: %w", err)
		}
	}
	if task.HasSafeOutputKey("add-comment") {
		if err := runCommentOutput(ctx, glob, task, tc, res); err != nil {
			return fmt.Errorf("add-comment: %w", err)
		}
	}
	return nil
}
