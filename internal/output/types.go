// Package output runs post-agent safe-output steps from agent-written output.json.
package output

// OutputKind is the JSON `type` field value (underscore form, gh-aw style).
type OutputKind string

const (
	KindCreatePullRequest              OutputKind = "create_pull_request"
	KindAddComment                     OutputKind = "add_comment"
	KindAddLabels                      OutputKind = "add_labels"
	KindRemoveLabels                   OutputKind = "remove_labels"
	KindCreateIssue                    OutputKind = "create_issue"
	KindCreatePullRequestReviewComment OutputKind = "create_pull_request_review_comment"
	KindSubmitPullRequestReview        OutputKind = "submit_pull_request_review"
	KindNoop                           OutputKind = "noop"
	KindMissingTool                    OutputKind = "missing_tool"
	KindMissingData                    OutputKind = "missing_data"
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

// ItemCreatePullRequestReviewComment fields for inline PR review comments.
type ItemCreatePullRequestReviewComment struct {
	Body     string `json:"body"`
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Side     string `json:"side,omitempty"`
	CommitID string `json:"commit_id,omitempty"`
	Target   int    `json:"target,omitempty"`
}

// ItemSubmitPullRequestReview fields for submitting a PR review.
type ItemSubmitPullRequestReview struct {
	Event    string `json:"event"`
	Body     string `json:"body,omitempty"`
	CommitID string `json:"commit_id,omitempty"`
	Target   int    `json:"target,omitempty"`
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
