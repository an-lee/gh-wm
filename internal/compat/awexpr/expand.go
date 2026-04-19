package awexpr

import (
	"fmt"
	"sort"
)

// Expand replaces all `${{ }}` spans in body using ctx. Spans outside code fences are expanded;
// order is right-to-left so offsets stay valid.
func Expand(body string, ctx *Context) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("nil expansion context")
	}
	spans := ScanBody(body)
	if len(spans) == 0 {
		return body, nil
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].Start > spans[j].Start })
	out := body
	for _, sp := range spans {
		val, err := ctx.Evaluate(sp.Expr)
		if err != nil {
			return "", fmt.Errorf("expand %q: %w", sp.Expr, err)
		}
		out = out[:sp.Start] + val + out[sp.End:]
	}
	return out, nil
}
