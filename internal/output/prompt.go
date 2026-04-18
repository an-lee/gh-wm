package output

import (
	"fmt"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/config/scalar"
)

// emitSubcommand maps safe-output kind to `gh-wm emit` subcommand (dash form).
func emitSubcommand(kind OutputKind) string {
	switch kind {
	case KindCreatePullRequest:
		return "create-pull-request"
	case KindAddComment:
		return "add-comment"
	case KindAddLabels:
		return "add-labels"
	case KindRemoveLabels:
		return "remove-labels"
	case KindCreateIssue:
		return "create-issue"
	case KindCreatePullRequestReviewComment:
		return "create-pull-request-review-comment"
	case KindSubmitPullRequestReview:
		return "submit-pull-request-review"
	case KindNoop:
		return "noop"
	case KindMissingTool:
		return "missing-tool"
	case KindMissingData:
		return "missing-data"
	default:
		return ""
	}
}

// SafeOutputsSystemPromptAppend returns text for Claude Code's `--append-system-prompt` when the task declares
// `safe-outputs:`. Empty when there is no policy to enforce. Used only for the built-in `claude` engine (not codex/custom).
func SafeOutputsSystemPromptAppend(task *config.Task) string {
	if task == nil {
		return ""
	}
	so := task.SafeOutputsMap()
	if so == nil || len(so) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("gh-wm safe-outputs (system):\n\n")
	b.WriteString("In read-only CI, direct GitHub writes via `gh` (for example `gh issue comment`, `gh pr create`, `gh label`) typically fail with permission errors. ")
	b.WriteString("You MUST record each allowed follow-up by running `gh-wm emit <subcommand>` with the correct flags (or `gh wm emit` when gh-wm is installed as a gh extension). ")
	b.WriteString("Each call appends one validated JSON line to WM_SAFE_OUTPUT_FILE (output.jsonl). Do not use raw `gh` for mutations that belong in the safe-outputs pipeline.\n")
	return b.String()
}

// AvailableOutputsSection builds markdown appended to the agent prompt describing `gh-wm emit`.
func AvailableOutputsSection(glob *config.GlobalConfig, task *config.Task) string {
	if task == nil {
		return ""
	}
	so := task.SafeOutputsMap()
	if so == nil || len(so) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n\n---\n## Safe outputs\n\n")
	b.WriteString("Each call appends one validated JSON line to **`WM_SAFE_OUTPUT_FILE`** (`output.jsonl`). ")
	b.WriteString("The run sets **`WM_REPO_ROOT`**, **`WM_TASK`**, **`WM_SAFE_OUTPUT_FILE`**, and typically **`GITHUB_REPOSITORY`** plus **`WM_ISSUE_NUMBER`** / **`WM_PR_NUMBER`** when applicable.\n\n")
	b.WriteString("If you have nothing to post, run **`gh-wm emit noop --message \"…\"`** (optional; missing output is treated as an implicit noop with a warning).\n\n")
	b.WriteString("Legacy: writing a single JSON blob to **`WM_OUTPUT_FILE`** (`output.json` with `items`) is still supported and merged with the NDJSON log.\n\n")
	b.WriteString("**Available for this task:**\n\n")

	order := []struct {
		fmKey string
		kind  OutputKind
		flags string
	}{
		{fmCreatePullRequest, KindCreatePullRequest, "`--title`, `--body`, optional `--draft`, `--labels`"},
		{fmAddComment, KindAddComment, "`--body`, optional `--target` (issue/PR number; else event numbers from env)"},
		{fmAddLabels, KindAddLabels, "`--labels` (repeat or comma-separated), optional `--target`"},
		{fmRemoveLabels, KindRemoveLabels, "`--labels`, optional `--target`"},
		{fmCreateIssue, KindCreateIssue, "`--title`, optional `--body`, `--labels`, `--assignees`"},
		{fmCreatePullRequestReviewComment, KindCreatePullRequestReviewComment, "`--body`, `--path`, `--line`, optional `--side`, `--commit`, `--target`"},
		{fmSubmitPullRequestReview, KindSubmitPullRequestReview, "`--event` (APPROVE|REQUEST_CHANGES|COMMENT), optional `--body`, `--commit`, `--target`"},
	}

	for _, row := range order {
		if !task.HasSafeOutputKey(row.fmKey) {
			continue
		}
		block := map[string]any{}
		if raw, ok := so[row.fmKey]; ok {
			if m, ok := raw.(map[string]any); ok {
				block = m
			}
		}
		maxN := scalar.IntFromMap(block, "max")
		if maxN <= 0 {
			maxN = defaultMaxPerKind(row.kind)
		}
		sub := emitSubcommand(row.kind)
		line := fmt.Sprintf("- **`gh-wm emit %s`** — max **%d** per run; flags: %s", sub, maxN, row.flags)
		if p := scalar.StringFromMap(block, "title-prefix"); p != "" {
			line += fmt.Sprintf("; titles must start with `%s` where configured", p)
		}
		if al := scalar.StringSliceFromMap(block, "allowed"); len(al) > 0 {
			line += fmt.Sprintf("; allowed labels: %v", al)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString(fmt.Sprintf("- **`gh-wm emit %s`** — always available; `--message` (no GitHub writes)\n", emitSubcommand(KindNoop)))
	b.WriteString("- **`gh-wm emit missing-tool`** — always available; `--tool`, `--reason`\n")
	b.WriteString("- **`gh-wm emit missing-data`** — always available; `--what`, `--reason`\n")

	if glob != nil && glob.PR.Draft {
		b.WriteString("\nDefault PR draft mode from config: **true** (override with `gh-wm emit create-pull-request --draft`).\n")
	}

	return b.String()
}
