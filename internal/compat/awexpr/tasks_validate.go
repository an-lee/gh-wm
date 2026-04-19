package awexpr

import (
	"fmt"

	"github.com/an-lee/gh-wm/internal/config"
)

// ValidateTasks runs expression validation on every task body. ModeOff skips.
// ModeError fails on the first invalid expression.
// ModeWarn logs invalid expressions and deprecation hints via warnf, then returns nil.
func ValidateTasks(tasks []*config.Task, mode GhAWExpressionsMode, warnf func(taskPath, message string)) error {
	if mode == ModeOff {
		return nil
	}
	for _, t := range tasks {
		if t == nil {
			continue
		}
		warns, err := ValidateBody(t.Body)
		if err != nil {
			if mode == ModeWarn && warnf != nil {
				warnf(t.Path, err.Error())
				continue
			}
			return fmt.Errorf("%s: %w", t.Path, err)
		}
		if warnf != nil {
			for _, w := range warns {
				if w.Message != "" {
					warnf(t.Path, w.Message+" (expression: "+w.Expr+")")
				}
			}
		}
	}
	return nil
}
