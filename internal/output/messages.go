package output

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/config/scalar"
	"github.com/an-lee/gh-wm/internal/ghclient"
	"github.com/an-lee/gh-wm/internal/types"
)

// safeOutputMessagesMap returns the optional messages: block under safe-outputs:.
func safeOutputMessagesMap(task *config.Task) map[string]any {
	if task == nil {
		return nil
	}
	so := task.SafeOutputsMap()
	if so == nil {
		return nil
	}
	raw, ok := so["messages"]
	if !ok {
		return nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	return m
}

// ExpandMessageTemplate substitutes {workflow_name}, {run_url}, {event_type}, and {status}.
func ExpandMessageTemplate(task *config.Task, tc *types.TaskContext, template string, status string) string {
	s := template
	wf := strings.TrimSpace(os.Getenv("GITHUB_WORKFLOW"))
	if wf == "" && task != nil {
		wf = task.Name
	}
	s = strings.ReplaceAll(s, "{workflow_name}", wf)
	s = strings.ReplaceAll(s, "{run_url}", runURL())
	ev := ""
	if tc != nil && tc.Event != nil {
		ev = strings.TrimSpace(tc.Event.Name)
	}
	s = strings.ReplaceAll(s, "{event_type}", ev)
	s = strings.ReplaceAll(s, "{status}", status)
	return s
}

func runURL() string {
	server := strings.TrimSuffix(strings.TrimSpace(os.Getenv("GITHUB_SERVER_URL")), "/")
	if server == "" {
		server = "https://github.com"
	}
	repo := strings.TrimSpace(os.Getenv("GITHUB_REPOSITORY"))
	runID := strings.TrimSpace(os.Getenv("GITHUB_RUN_ID"))
	if repo == "" || runID == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/actions/runs/%s", server, repo, runID)
}

// AppendMessagesFooter appends safe-outputs.messages.footer when set (before wm marker footers).
func AppendMessagesFooter(task *config.Task, tc *types.TaskContext, body string) string {
	m := safeOutputMessagesMap(task)
	if m == nil {
		return body
	}
	footer := strings.TrimSpace(scalar.StringFromMap(m, "footer"))
	if footer == "" {
		return body
	}
	footer = ExpandMessageTemplate(task, tc, footer, "")
	if strings.TrimSpace(body) == "" {
		return footer
	}
	return body + "\n\n" + footer
}

// PostConfiguredStatusComment posts run-started, run-success, or run-failure if configured.
// Best-effort: logs and returns nil on API failure (e.g. read-only token in CI agent job).
func PostConfiguredStatusComment(task *config.Task, tc *types.TaskContext, which string, statusDetail string) error {
	if task == nil || tc == nil {
		return nil
	}
	m := safeOutputMessagesMap(task)
	if m == nil {
		return nil
	}
	var key string
	switch which {
	case "run-started":
		key = "run-started"
	case "run-success":
		key = "run-success"
	case "run-failure":
		key = "run-failure"
	default:
		return nil
	}
	tpl := strings.TrimSpace(scalar.StringFromMap(m, key))
	if tpl == "" {
		return nil
	}
	n := commentTargetNumber(tc)
	if n <= 0 || strings.TrimSpace(tc.Repo) == "" {
		return nil
	}
	body := ExpandMessageTemplate(task, tc, tpl, statusDetail)
	body = body + WMAgentCommentMarkerFooter(tc.TaskName)
	if err := ghclient.PostIssueComment(tc.Repo, n, body); err != nil {
		slog.Debug("wm: safe-outputs messages: post status comment", "which", which, "err", err)
		return nil
	}
	return nil
}
