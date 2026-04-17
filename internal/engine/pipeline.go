package engine

import (
	"log/slog"
	"os"

	"github.com/an-lee/gh-wm/internal/types"
)

func init() {
	if os.Getenv("WM_LOG_FORMAT") == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	}
}

// logPhase records a pipeline phase for structured logs (alongside progress lines).
func logPhase(task string, phase types.Phase) {
	slog.Info("wm pipeline", "task", task, "phase", string(phase))
}
