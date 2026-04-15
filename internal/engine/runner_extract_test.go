package engine

import (
	"testing"

	"github.com/an-lee/gh-wm/internal/types"
)

func TestExtractNumbers(t *testing.T) {
	t.Parallel()
	tc := &types.TaskContext{}
	extractNumbers(nil, tc)
	if tc.IssueNumber != 0 {
		t.Fatal("nil payload")
	}
	extractNumbers(map[string]any{
		"issue":        map[string]any{"number": float64(3)},
		"pull_request": map[string]any{"number": float64(5)},
	}, tc)
	if tc.IssueNumber != 3 || tc.PRNumber != 5 {
		t.Fatalf("%+v", tc)
	}
	tc2 := &types.TaskContext{}
	extractNumbers(map[string]any{
		"pull_request": map[string]any{"number": float64(7)},
	}, tc2)
	if tc2.IssueNumber != 7 || tc2.PRNumber != 7 {
		t.Fatalf("issue should mirror PR when issue missing: %+v", tc2)
	}
}

func TestNum(t *testing.T) {
	t.Parallel()
	if num(float64(4)) != 4 || num(4) != 4 || num(int64(4)) != 4 || num("x") != 0 {
		t.Fatal("num")
	}
}
