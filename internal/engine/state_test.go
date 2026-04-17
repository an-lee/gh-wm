package engine

import (
	"errors"
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
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

func TestIsLabelRemoveNotFound(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"no_match", errors.New("some other error"), false},
		{"lowercase_404", errors.New("GET https://api.github.com: 404 not found"), true},
		{"uppercase_404", errors.New("HTTP 404 Not Found"), true},
		{"lowercase_not_found", errors.New("label not found"), true},
		{"mixed_case", errors.New("Not Found label"), true},
		{"not_found_in_long_msg", errors.New("HTTP 422: could not remove label (not found)"), true},
		{"rate_limit_error_false", errors.New("API rate limit exceeded (429)"), false},
		{"contains_not_found_as_word", errors.New("label was not found in the list"), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isLabelRemoveNotFound(tc.err); got != tc.want {
				t.Fatalf("isLabelRemoveNotFound(%v): got %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
