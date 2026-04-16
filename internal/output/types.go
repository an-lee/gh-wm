// Package output runs post-agent safe-output steps (agent-driven or legacy).
package output

// OutputKind is the JSON `type` field value (underscore form, gh-aw style).
type OutputKind string

const (
	KindCreatePullRequest OutputKind = "create_pull_request"
	KindAddComment        OutputKind = "add_comment"
	KindAddLabels         OutputKind = "add_labels"
	KindRemoveLabels      OutputKind = "remove_labels"
	KindCreateIssue       OutputKind = "create_issue"
	KindNoop              OutputKind = "noop"
)

// AgentOutputFile is the root JSON shape written to WM_OUTPUT_FILE (output.json).
type AgentOutputFile struct {
	Items []map[string]any `json:"items"`
}

// ItemCreatePullRequest fields for create_pull_request.
type ItemCreatePullRequest struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Draft  *bool    `json:"draft,omitempty"`
	Labels []string `json:"labels,omitempty"`
}

// ItemAddComment fields for add_comment.
type ItemAddComment struct {
	Body   string `json:"body"`
	Target int    `json:"target"` // issue or PR number; 0 = use event context
}

// ItemAddLabels / ItemRemoveLabels
type ItemLabels struct {
	Labels []string `json:"labels"`
	Target int      `json:"target"`
}

// ItemCreateIssue fields for create_issue.
type ItemCreateIssue struct {
	Title      string   `json:"title"`
	Body       string   `json:"body"`
	Labels     []string `json:"labels,omitempty"`
	Assignees  []string `json:"assignees,omitempty"`
}

// ItemNoop records completion without GitHub writes.
type ItemNoop struct {
	Message string `json:"message"`
}
