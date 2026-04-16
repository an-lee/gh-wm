package engine

import (
	"testing"

	"github.com/an-lee/gh-wm/internal/types"
)

func TestConcludeRun_NoPanicNilArgs(t *testing.T) {
	t.Parallel()
	concludeRun(nil, nil)
	concludeRun(&types.RunResult{}, nil)
	r := &types.RunResult{}
	concludeRun(r, &concludeArgs{})
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
}
