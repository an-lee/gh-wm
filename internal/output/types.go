// Package output runs post-agent safe-output steps from WM_SAFE_OUTPUT_FILE (output.jsonl NDJSON).
package output

// OutputKind is the JSON `type` field value (underscore form, gh-aw style).
type OutputKind string

const (
	KindCreatePullRequest               OutputKind = "create_pull_request"
	KindAddComment                      OutputKind = "add_comment"
	KindAddLabels                       OutputKind = "add_labels"
	KindRemoveLabels                    OutputKind = "remove_labels"
	KindCreateIssue                     OutputKind = "create_issue"
	KindUpdatePullRequest               OutputKind = "update_pull_request"
	KindUpdateIssue                     OutputKind = "update_issue"
	KindCloseIssue                      OutputKind = "close_issue"
	KindClosePullRequest                OutputKind = "close_pull_request"
	KindAddReviewer                     OutputKind = "add_reviewer"
	KindCreatePullRequestReviewComment  OutputKind = "create_pull_request_review_comment"
	KindSubmitPullRequestReview         OutputKind = "submit_pull_request_review"
	KindReplyToPullRequestReviewComment OutputKind = "reply_to_pull_request_review_comment"
	KindResolvePullRequestReviewThread  OutputKind = "resolve_pull_request_review_thread"
	KindPushToPullRequestBranch         OutputKind = "push_to_pull_request_branch"
	KindNoop                            OutputKind = "noop"
	KindMissingTool                     OutputKind = "missing_tool"
	KindMissingData                     OutputKind = "missing_data"
)

// AgentOutputFile is the in-memory aggregate of parsed NDJSON items for one run.
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
	Title     string `json:"title,omitempty"`
	Body      string `json:"body,omitempty"`
	Target    int    `json:"target"`
	Operation string `json:"operation,omitempty"` // replace (default), append, prepend, replace-island
}

// ItemUpdatePullRequest fields for update_pull_request.
type ItemUpdatePullRequest struct {
	Title     string `json:"title,omitempty"`
	Body      string `json:"body,omitempty"`
	Target    int    `json:"target"`
	Operation string `json:"operation,omitempty"`
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

// ItemCreatePullRequestReviewComment fields for create_pull_request_review_comment (inline review comment).
type ItemCreatePullRequestReviewComment struct {
	Body      string `json:"body"`
	CommitID  string `json:"commit_id,omitempty"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Side      string `json:"side,omitempty"` // LEFT or RIGHT
	StartLine int    `json:"start_line,omitempty"`
	Target    int    `json:"target"`
}

// ItemSubmitPullRequestReview fields for submit_pull_request_review.
type ItemSubmitPullRequestReview struct {
	Event    string `json:"event"`
	Body     string `json:"body,omitempty"`
	CommitID string `json:"commit_id,omitempty"`
	Target   int    `json:"target"`
}

// ItemReplyToPullRequestReviewComment fields for reply_to_pull_request_review_comment.
type ItemReplyToPullRequestReviewComment struct {
	Body      string `json:"body"`
	CommentID int    `json:"comment_id"`
	Target    int    `json:"target"`
}

// ItemResolvePullRequestReviewThread fields for resolve_pull_request_review_thread (GraphQL thread id).
type ItemResolvePullRequestReviewThread struct {
	ThreadID string `json:"thread_id"`
	Target   int    `json:"target"`
}

// ItemPushToPullRequestBranch fields for push_to_pull_request_branch (git push to PR head).
type ItemPushToPullRequestBranch struct {
	Target int `json:"target"` // PR number; 0 = WM_PR_NUMBER / event
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
