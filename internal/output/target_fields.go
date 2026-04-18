package output

import "github.com/an-lee/gh-wm/internal/config/scalar"

// intTargetComment resolves issue/PR number for add_comment, add_labels, remove_labels (gh-aw compatible aliases).
func intTargetComment(m map[string]any) int {
	return scalar.IntFieldFirst(m, "target", "issue_number", "pull_request_number", "item_number")
}

// intTargetIssue resolves issue number for update_issue, close_issue (gh-aw compatible aliases).
func intTargetIssue(m map[string]any) int {
	return scalar.IntFieldFirst(m, "target", "issue_number", "item_number")
}

// intTargetPR resolves PR number for PR-scoped outputs (gh-aw compatible aliases).
func intTargetPR(m map[string]any) int {
	return scalar.IntFieldFirst(m, "target", "pull_request_number", "issue_number", "item_number")
}
