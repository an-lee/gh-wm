package output

import (
	"context"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

// kindExecutor runs one parsed safe-output item (registry, one registration per [OutputKind]).
type kindExecutor func(ctx context.Context, glob *config.GlobalConfig, task *config.Task, tc *types.TaskContext, p *Policy, raw map[string]any) error

var kindRegistry = make(map[OutputKind]kindExecutor)

func registerKind(kind OutputKind, fn kindExecutor) {
	kindRegistry[kind] = fn
}

func executorFor(kind OutputKind) (kindExecutor, bool) {
	fn, ok := kindRegistry[kind]
	return fn, ok
}
