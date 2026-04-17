package output

import (
	"fmt"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/config/scalar"
)

// AvailableOutputsSection builds markdown appended to the agent prompt describing WM_OUTPUT_FILE.
func AvailableOutputsSection(glob *config.GlobalConfig, task *config.Task) string {
	if task == nil {
		return ""
	}
	so := task.SafeOutputsMap()
	if so == nil || len(so) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n\n---\n## Safe outputs (structured)\n\n")
	b.WriteString("After completing your work, write **structured safe-output requests** to the file path in **`WM_OUTPUT_FILE`** (JSON). ")
	b.WriteString("If you do not need any GitHub follow-up actions, either omit the file or include a `noop` item.\n\n")
	b.WriteString("```json\n{\n  \"items\": [\n    { \"type\": \"noop\", \"message\": \"…\" }\n  ]\n}\n```\n\n")
	b.WriteString("Use these **`type`** values (underscores):\n\n")

	order := []struct {
		fmKey string
		kind  OutputKind
		desc  string
	}{
		{fmCreatePullRequest, KindCreatePullRequest, "`title`, `body`, optional `draft`, `labels`"},
		{fmAddComment, KindAddComment, "`body`, optional `target` (issue/PR number; default: triggering thread)"},
		{fmAddLabels, KindAddLabels, "`labels` (array), optional `target` (issue/PR number)"},
		{fmRemoveLabels, KindRemoveLabels, "`labels`, optional `target`"},
		{fmCreateIssue, KindCreateIssue, "`title`, `body`, optional `labels`, `assignees`"},
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
		line := fmt.Sprintf("- **`%s`** — max **%d** per run; fields: %s", row.kind, maxN, row.desc)
		if p := scalar.StringFromMap(block, "title-prefix"); p != "" {
			line += fmt.Sprintf("; titles must start with `%s` where configured", p)
		}
		if al := scalar.StringSliceFromMap(block, "allowed"); len(al) > 0 {
			line += fmt.Sprintf("; allowed labels: %v", al)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString(fmt.Sprintf("- **`%s`** — always available; `message` only (no GitHub writes)\n", KindNoop))

	if glob != nil && glob.PR.Draft {
		b.WriteString("\nDefault PR draft mode from config: **true** (override per item with `draft`).\n")
	}

	return b.String()
}
