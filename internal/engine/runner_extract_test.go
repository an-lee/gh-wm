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
	tc3 := &types.TaskContext{}
	extractNumbers(map[string]any{
		"issue":   map[string]any{"number": float64(1)},
		"comment": map[string]any{"id": float64(999888777)},
	}, tc3)
	if tc3.CommentID != 999888777 {
		t.Fatalf("comment id: %+v", tc3)
	}
}

func TestNum(t *testing.T) {
	t.Parallel()
	if num(float64(4)) != 4 || num(4) != 4 || num(int64(4)) != 4 || num("x") != 0 {
		t.Fatal("num")
	}
}

func TestInt64FromID(t *testing.T) {
	t.Parallel()
	if int64FromID(float64(42)) != 42 || int64FromID(int(7)) != 7 || int64FromID(int64(9)) != 9 || int64FromID("x") != 0 {
		t.Fatal("int64FromID")
	}
}
