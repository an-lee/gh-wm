package output

import "github.com/an-lee/gh-wm/internal/types"

func resolveIssueTarget(tc *types.TaskContext, target int) int {
	if target > 0 {
		return target
	}
	if tc == nil {
		return 0
	}
	if tc.IssueNumber > 0 {
		return tc.IssueNumber
	}
	return tc.PRNumber
}

func resolvePRTarget(tc *types.TaskContext, target int) int {
	if target > 0 {
		return target
	}
	if tc == nil {
		return 0
	}
	if tc.PRNumber > 0 {
		return tc.PRNumber
	}
	return tc.IssueNumber
}
