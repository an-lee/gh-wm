package output

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

// RunSuccessOutputs runs agent-driven safe outputs from WM_SAFE_OUTPUT_FILE (output.jsonl NDJSON)
// and/or legacy WM_OUTPUT_FILE (output.json). If both are empty, logs a warning and succeeds (implicit noop).
func RunSuccessOutputs(ctx context.Context, glob *config.GlobalConfig, task *config.Task, tc *types.TaskContext, res *types.AgentResult) error {
	if glob == nil || task == nil || tc == nil || res == nil {
		return nil
	}
	if !taskDeclaresSafeOutputs(task) {
		return nil
	}
	ao, err := ParseAgentOutputFile(res.OutputFilePath)
	if err != nil {
		return err
	}
	ndItems, err := ParseAgentOutputJSONLFile(res.SafeOutputFilePath)
	if err != nil {
		return err
	}
	var legacy []map[string]any
	if ao != nil {
		legacy = ao.Items
	}
	merged := append(append([]map[string]any(nil), ndItems...), legacy...)
	if len(merged) == 0 {
		slog.Warn("wm: safe-outputs: no structured output (implicit noop); use `gh-wm emit noop` or other emit subcommands when safe-outputs is set",
			"safe_output_file", strings.TrimSpace(res.SafeOutputFilePath),
			"legacy_output_file", strings.TrimSpace(res.OutputFilePath))
		return nil
	}
	return runAgentDrivenOutputs(ctx, glob, task, tc, &AgentOutputFile{Items: merged})
}

func taskDeclaresSafeOutputs(task *config.Task) bool {
	m := task.SafeOutputsMap()
	return m != nil && len(m) > 0
}

func runAgentDrivenOutputs(ctx context.Context, glob *config.GlobalConfig, task *config.Task, tc *types.TaskContext, ao *AgentOutputFile) error {
	p := newPolicy(task)
	for _, raw := range ao.Items {
		if raw == nil {
			continue
		}
		kind := ParseOutputKind(ItemType(raw))
		if kind == "" {
			slog.Info("wm: safe-output: unknown item type, skipping", "type", ItemType(raw))
			continue
		}
		if kind == KindNoop {
			runNoop(mapToNoop(raw))
			continue
		}
		if kind == KindMissingTool {
			runMissingTool(mapToMissingTool(raw))
			continue
		}
		if kind == KindMissingData {
			runMissingData(mapToMissingData(raw))
			continue
		}
		if !p.Allowed(kind) {
			slog.Info("wm: safe-output: type not permitted by safe-outputs:, skipping", "kind", kind)
			continue
		}
		if err := p.CheckMax(kind); err != nil {
			return err
		}

		fn, ok := executorFor(kind)
		if !ok {
			return fmt.Errorf("safe-outputs: no executor registered for kind %q", kind)
		}
		execErr := fn(ctx, glob, task, tc, p, raw)
		if execErr != nil {
			return fmt.Errorf("%s: %w", kind, execErr)
		}
		p.RecordSuccess(kind)
	}
	return nil
}
