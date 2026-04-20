package awexpr

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/an-lee/gh-wm/internal/compat/sanitize"
	"github.com/an-lee/gh-wm/internal/types"
)

// Context holds values for expanding ${{ }} in task bodies.
type Context struct {
	Event map[string]any // github.event payload

	SanitizedText  string
	SanitizedTitle string
	SanitizedBody  string

	TaskName string

	ResolvedTasksJSON string
	HasTasks          string

	githubFlat map[string]string
}

// BuildContext builds an expander context from task context and environment.
func BuildContext(tc *types.TaskContext) *Context {
	var payload map[string]any
	if tc != nil && tc.Event != nil && tc.Event.Payload != nil {
		payload = tc.Event.Payload
	}
	st, sti, sb := sanitize.FromPayload(payload)
	c := &Context{
		Event:             payload,
		SanitizedText:     st,
		SanitizedTitle:    sti,
		SanitizedBody:     sb,
		TaskName:          taskNameOf(tc),
		ResolvedTasksJSON: strings.TrimSpace(os.Getenv("WM_RESOLVED_TASKS_JSON")),
		HasTasks:          strings.TrimSpace(os.Getenv("WM_HAS_TASKS")),
		githubFlat:        githubEnvFlat(),
	}
	return c
}

func taskNameOf(tc *types.TaskContext) string {
	if tc == nil {
		return strings.TrimSpace(os.Getenv("WM_TASK"))
	}
	if tc.TaskName != "" {
		return tc.TaskName
	}
	return strings.TrimSpace(os.Getenv("WM_TASK"))
}

func githubEnvFlat() map[string]string {
	keys := []string{
		"GITHUB_ACTOR", "GITHUB_REPOSITORY", "GITHUB_REPOSITORY_OWNER", "GITHUB_RUN_ID", "GITHUB_RUN_NUMBER",
		"GITHUB_RUN_ATTEMPT", "GITHUB_WORKFLOW", "GITHUB_SERVER_URL", "GITHUB_REF", "GITHUB_REF_NAME",
		"GITHUB_SHA", "GITHUB_WORKSPACE", "GITHUB_EVENT_NAME", "GITHUB_JOB", "GITHUB_API_URL",
		"GITHUB_BASE_REF", "GITHUB_HEAD_REF",
	}
	m := make(map[string]string, len(keys))
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			m[k] = v
		}
	}
	return m
}

// Evaluate expands a single expression (may contain top-level ||).
func (c *Context) Evaluate(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "", fmt.Errorf("empty expression")
	}
	if strings.Contains(expr, "||") {
		parts := splitTopLevelOR(expr)
		for _, p := range parts {
			p = strings.TrimSpace(p)
			v, err := c.evaluateAtom(p)
			if err != nil {
				return "", err
			}
			if truthy(v) {
				return v, nil
			}
		}
		return "", nil
	}
	return c.evaluateAtom(expr)
}

func truthy(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || s == "false" || s == "0" {
		return false
	}
	return true
}

func (c *Context) evaluateAtom(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	lower := strings.ToLower(expr)

	switch {
	case lower == "wm.task_name":
		return c.TaskName, nil
	case lower == "wm.sanitized.text":
		return c.SanitizedText, nil
	case lower == "wm.sanitized.title":
		return c.SanitizedTitle, nil
	case lower == "wm.sanitized.body":
		return c.SanitizedBody, nil
	case lower == "steps.sanitized.outputs.text":
		return c.SanitizedText, nil
	case lower == "steps.sanitized.outputs.title":
		return c.SanitizedTitle, nil
	case lower == "steps.sanitized.outputs.body":
		return c.SanitizedBody, nil
	case lower == "needs.resolve.outputs.tasks":
		return c.ResolvedTasksJSON, nil
	case lower == "needs.resolve.outputs.has_tasks":
		return c.HasTasks, nil
	case strings.HasPrefix(lower, "github.event."):
		path := expr[len("github.event."):]
		v, err := walkEvent(c.Event, path)
		if err != nil {
			return "", err
		}
		return formatAny(v), nil
	default:
		return c.githubScalar(lower)
	}
}

func (c *Context) githubScalar(lower string) (string, error) {
	switch lower {
	case "github.actor":
		return c.githubFlat["GITHUB_ACTOR"], nil
	case "github.repository":
		return c.githubFlat["GITHUB_REPOSITORY"], nil
	case "github.repository_owner":
		return c.githubFlat["GITHUB_REPOSITORY_OWNER"], nil
	case "github.run_id":
		return c.githubFlat["GITHUB_RUN_ID"], nil
	case "github.run_number":
		return c.githubFlat["GITHUB_RUN_NUMBER"], nil
	case "github.run_attempt":
		return c.githubFlat["GITHUB_RUN_ATTEMPT"], nil
	case "github.workflow":
		return c.githubFlat["GITHUB_WORKFLOW"], nil
	case "github.server_url":
		return c.githubFlat["GITHUB_SERVER_URL"], nil
	case "github.ref":
		return c.githubFlat["GITHUB_REF"], nil
	case "github.ref_name":
		return c.githubFlat["GITHUB_REF_NAME"], nil
	case "github.sha":
		return c.githubFlat["GITHUB_SHA"], nil
	case "github.workspace":
		return c.githubFlat["GITHUB_WORKSPACE"], nil
	case "github.event_name":
		return c.githubFlat["GITHUB_EVENT_NAME"], nil
	case "github.job":
		return c.githubFlat["GITHUB_JOB"], nil
	case "github.api_url":
		return c.githubFlat["GITHUB_API_URL"], nil
	case "github.base_ref":
		return c.githubFlat["GITHUB_BASE_REF"], nil
	case "github.head_ref":
		return c.githubFlat["GITHUB_HEAD_REF"], nil
	default:
		return "", fmt.Errorf("unsupported expression %q", lower)
	}
}

func walkEvent(root map[string]any, path string) (any, error) {
	if path == "" {
		return nil, fmt.Errorf("empty github.event path")
	}
	if root == nil {
		return nil, nil
	}
	parts := strings.Split(path, ".")
	var cur any = root
	for _, p := range parts {
		if cur == nil {
			return nil, nil
		}
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("github.event path %q: not an object at segment %q", path, p)
		}
		cur, ok = m[p]
		if !ok {
			return nil, nil
		}
	}
	return cur, nil
}

func formatAny(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	default:
		return fmt.Sprint(x)
	}
}
