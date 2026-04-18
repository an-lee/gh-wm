// Package output runs post-agent safe-output steps from agent-written output.json.
package output

// OutputKind is the JSON `type` field value (underscore form, gh-aw style).
type OutputKind string

const (
	KindCreatePullRequest OutputKind = "create_pull_request"
	KindAddComment        OutputKind = "add_comment"
	KindAddLabels         OutputKind = "add_labels"
	KindRemoveLabels      OutputKind = "remove_labels"
	KindCreateIssue       OutputKind = "create_issue"
	KindUpdatePullRequest OutputKind = "update_pull_request"
	KindUpdateIssue       OutputKind = "update_issue"
	KindCloseIssue        OutputKind = "close_issue"
	KindClosePullRequest  OutputKind = "close_pull_request"
	KindAddReviewer       OutputKind = "add_reviewer"
	KindNoop              OutputKind = "noop"
	KindMissingTool       OutputKind = "missing_tool"
	KindMissingData       OutputKind = "missing_data"
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
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Labels    []string `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
}

// ItemUpdateIssue fields for update_issue.
type ItemUpdateIssue struct {
	Title  string `json:"title,omitempty"`
	Body   string `json:"body,omitempty"`
	Target int    `json:"target"`
}

// ItemUpdatePullRequest fields for update_pull_request.
type ItemUpdatePullRequest struct {
	Title  string `json:"title,omitempty"`
	Body   string `json:"body,omitempty"`
	Target int    `json:"target"`
}

// ItemCloseIssue fields for close_issue.
type ItemCloseIssue struct {
	Comment     string `json:"comment,omitempty"`
	StateReason string `json:"state_reason,omitempty"`
	Target      int    `json:"target"`
}

// ItemClosePullRequest fields for close_pull_request.
type ItemClosePullRequest struct {
	Comment string `json:"comment,omitempty"`
	Target  int    `json:"target"`
}

// ItemAddReviewer fields for add_reviewer.
type ItemAddReviewer struct {
	Reviewers []string `json:"reviewers"`
	Target    int      `json:"target"`
}

// ItemNoop records completion without GitHub writes.
type ItemNoop struct {
	Message string `json:"message"`
}

// ItemMissingTool reports unavailable functionality (log-only; no GitHub API).
type ItemMissingTool struct {
	Tool   string `json:"tool"`
	Reason string `json:"reason"`
}

// ItemMissingData reports unavailable information (log-only).
type ItemMissingData struct {
	What   string `json:"what"`
	Reason string `json:"reason"`
}
