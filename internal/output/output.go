package output

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

// RunSuccessOutputs runs post-agent safe outputs: agent-driven output.json when present, else legacy behavior.
func RunSuccessOutputs(ctx context.Context, glob *config.GlobalConfig, task *config.Task, tc *types.TaskContext, res *types.AgentResult) error {
	if glob == nil || task == nil || tc == nil || res == nil {
		return nil
	}
	ao, err := ParseAgentOutputFile(res.OutputFilePath)
	if err != nil {
		return err
	}
	if ao != nil && len(ao.Items) > 0 {
		return runAgentDrivenOutputs(ctx, glob, task, tc, ao)
	}
	return runLegacyOutputs(ctx, glob, task, tc, res)
}

func runLegacyOutputs(ctx context.Context, glob *config.GlobalConfig, task *config.Task, tc *types.TaskContext, res *types.AgentResult) error {
	if task.HasSafeOutputKey(fmCreatePullRequest) {
		if err := runPROutputLegacy(ctx, glob, task, tc); err != nil {
			return fmt.Errorf("create-pull-request: %w", err)
		}
	}
	if task.HasSafeOutputKey(fmAddLabels) {
		if err := runLabelOutputLegacy(ctx, task, tc); err != nil {
			return fmt.Errorf("add-labels: %w", err)
		}
	}
	if task.HasSafeOutputKey(fmAddComment) {
		if err := runCommentOutputLegacy(ctx, glob, task, tc, res); err != nil {
			return fmt.Errorf("add-comment: %w", err)
		}
	}
	return nil
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
