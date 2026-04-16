package output

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

// RunSuccessOutputs runs agent-driven safe outputs from WM_OUTPUT_FILE (output.json).
// If the task declares safe-outputs: with at least one key, a non-empty {"items":[...]} is required
// (use type noop when no GitHub follow-up is needed). If safe-outputs is absent or empty, this is a no-op.
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
	if ao == nil || len(ao.Items) == 0 {
		path := strings.TrimSpace(res.OutputFilePath)
		if path == "" {
			path = "$WM_OUTPUT_FILE (per-run output.json path)"
		}
		return fmt.Errorf("safe-outputs: missing or empty structured output; write JSON to %s with {\"items\":[{\"type\":\"noop\",\"message\":\"…\"}]} or other allowed types", path)
	}
	return runAgentDrivenOutputs(ctx, glob, task, tc, ao)
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
			log.Printf("wm: safe-output: unknown item type %q, skipping", ItemType(raw))
			continue
		}
		if kind == KindNoop {
			runNoop(mapToNoop(raw))
			continue
		}
		if !p.Allowed(kind) {
			log.Printf("wm: safe-output: type %q not permitted by safe-outputs:, skipping", kind)
			continue
		}
		if err := p.CheckMax(kind); err != nil {
			return err
		}

		var execErr error
		switch kind {
		case KindCreatePullRequest:
			item := mapToCreatePR(raw)
			execErr = runCreatePullRequestItem(ctx, glob, task, tc, p, item)
		case KindAddComment:
			item := mapToAddComment(raw)
			execErr = runCommentFromItem(ctx, tc, item)
		case KindAddLabels:
			item := mapToLabels(raw)
			execErr = runAddLabelsFromItem(ctx, tc, p, item)
		case KindRemoveLabels:
			item := mapToLabels(raw)
			if len(item.Labels) == 0 {
				execErr = fmt.Errorf("remove_labels: empty labels")
			} else {
				execErr = runRemoveLabelsFromItemWithPolicy(ctx, tc, p, item)
			}
		case KindCreateIssue:
			item := mapToCreateIssue(raw)
			item.Title = p.ApplyTitlePrefix(KindCreateIssue, strings.TrimSpace(item.Title))
			if strings.TrimSpace(item.Title) == "" {
				execErr = fmt.Errorf("create_issue: empty title")
			} else {
				item.Labels = p.MergeLabels(KindCreateIssue, item.Labels)
				execErr = runCreateIssue(ctx, tc, item)
			}
		default:
			continue
		}
		if execErr != nil {
			return fmt.Errorf("%s: %w", kind, execErr)
		}
		p.RecordSuccess(kind)
	}
	return nil
}
