package engine

import (
	"testing"

	"github.com/gh-wm/gh-wm/internal/config"
	"github.com/gh-wm/gh-wm/internal/types"
)

func TestIssueOrPRNumber(t *testing.T) {
	t.Parallel()
	if issueOrPRNumber(nil) != 0 {
		t.Fatal("nil")
	}
	if issueOrPRNumber(&types.TaskContext{IssueNumber: 2}) != 2 {
		t.Fatal("issue")
	}
	if issueOrPRNumber(&types.TaskContext{PRNumber: 3}) != 3 {
		t.Fatal("pr")
	}
}

func TestApplyState_NoOp(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{}
	wm := config.WMExtension{}
	ApplyStateWorking(tc, wm)
	ApplyStateDone(tc, wm)
	ApplyStateFailed(tc, wm)
	tc2 := &types.TaskContext{Repo: "o/r", IssueNumber: 1}
	ApplyStateWorking(tc2, config.WMExtension{})
}
