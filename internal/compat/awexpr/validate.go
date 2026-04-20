package awexpr

import (
	"fmt"
	"strings"
	"unicode"
)

// GhAWExpressionsMode controls validation strictness for task body expressions.
type GhAWExpressionsMode string

const (
	ModeError GhAWExpressionsMode = "error"
	ModeWarn  GhAWExpressionsMode = "warn"
	ModeOff   GhAWExpressionsMode = "off"
)

// ParseGhAWExpressionsMode returns mode or default "error".
func ParseGhAWExpressionsMode(s string) GhAWExpressionsMode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "error":
		return ModeError
	case "warn":
		return ModeWarn
	case "off":
		return ModeOff
	default:
		return ModeError
	}
}

// ValidationWarning is a non-fatal deprecation or style hint.
type ValidationWarning struct {
	Expr    string
	Message string
}

// ValidateBody checks all `${{ }}` spans and unsupported `{{#` in the markdown body.
func ValidateBody(body string) ([]ValidationWarning, error) {
	if ContainsUnsupportedTemplate(body) {
		return nil, fmt.Errorf("unsupported gh-aw template directive {{#...}} in task body; use ${{ }} expressions only or remove Handlebars-style blocks")
	}
	spans := ScanBody(body)
	var warns []ValidationWarning
	for _, sp := range spans {
		w, err := validateExpr(sp.Expr)
		if err != nil {
			return nil, err
		}
		warns = append(warns, w...)
	}
	return warns, nil
}

func validateExpr(expr string) ([]ValidationWarning, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("empty ${{ }} expression")
	}
	// OR chains: validate each arm
	if strings.Contains(expr, "||") {
		parts := splitTopLevelOR(expr)
		var all []ValidationWarning
		for _, p := range parts {
			w, err := validateAtom(strings.TrimSpace(p))
			if err != nil {
				return nil, err
			}
			all = append(all, w...)
		}
		return all, nil
	}
	return validateAtom(expr)
}

func splitTopLevelOR(s string) []string {
	// Split on "||" — good enough for gh-aw patterns (no nested || in typical prompts).
	parts := strings.Split(s, "||")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

func validateAtom(expr string) ([]ValidationWarning, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("empty expression in ${{ }}")
	}
	lower := strings.ToLower(expr)

	// Deny functions and secret/env paths (gh-aw markdown rules).
	if strings.Contains(lower, "secrets.") {
		return nil, fmt.Errorf("expression %q: secrets.* is not allowed in task body (gh-aw policy)", expr)
	}
	if strings.Contains(lower, "env.") {
		return nil, fmt.Errorf("expression %q: env.* is not allowed in task body (gh-aw policy)", expr)
	}
	if strings.Contains(lower, "vars.") {
		return nil, fmt.Errorf("expression %q: vars.* is not allowed in task body (gh-aw policy)", expr)
	}
	if containsCall(lower, "tojson") || containsCall(lower, "fromjson") {
		return nil, fmt.Errorf("expression %q: toJSON/fromJSON are not allowed in task body (gh-aw policy)", expr)
	}

	// Deny matrix, strategy — not available in agent prompt context
	if hasPrefixDot(lower, "matrix.") {
		return nil, fmt.Errorf("expression %q: matrix.* is not available in gh-wm task expansion", expr)
	}

	var warns []ValidationWarning

	switch {
	case strings.HasPrefix(lower, "wm."):
		if !isAllowedWM(expr) {
			return nil, fmt.Errorf("expression %q: unsupported wm.* path", expr)
		}
	case strings.HasPrefix(lower, "steps.sanitized.outputs."):
		if !isAllowedSanitized(expr) {
			return nil, fmt.Errorf("expression %q: use steps.sanitized.outputs.text, .title, or .body", expr)
		}
		warns = append(warns, ValidationWarning{
			Expr:    expr,
			Message: "prefer wm.sanitized.* (canonical in gh-wm); steps.sanitized.outputs.* is a gh-aw compatibility alias",
		})
	case strings.HasPrefix(lower, "needs.resolve.outputs."):
		if !isAllowedNeedsResolve(expr) {
			return nil, fmt.Errorf("expression %q: only needs.resolve.outputs.tasks and needs.resolve.outputs.has_tasks are supported", expr)
		}
	case strings.HasPrefix(lower, "github."):
		if !isAllowedGitHub(expr) {
			return nil, fmt.Errorf("expression %q: unsupported github.* path for gh-wm expansion", expr)
		}
	default:
		return nil, fmt.Errorf("expression %q: unsupported root (allowed: wm.*, github.*, steps.sanitized.outputs.*, needs.resolve.outputs.*)", expr)
	}
	return warns, nil
}

func containsCall(lower, name string) bool {
	// name( with optional space
	i := strings.Index(lower, name+"(")
	if i >= 0 {
		return true
	}
	return strings.Contains(lower, name+" (")
}

func hasPrefixDot(lower, prefix string) bool {
	return lower == prefix[:len(prefix)-1] || strings.HasPrefix(lower, prefix)
}

func isAllowedWM(expr string) bool {
	e := strings.TrimSpace(expr)
	switch e {
	case "wm.task_name", "wm.sanitized.text", "wm.sanitized.title", "wm.sanitized.body":
		return true
	default:
		return false
	}
}

func isAllowedSanitized(expr string) bool {
	e := strings.TrimSpace(strings.ToLower(expr))
	switch e {
	case "steps.sanitized.outputs.text", "steps.sanitized.outputs.title", "steps.sanitized.outputs.body":
		return true
	default:
		return false
	}
}

func isAllowedNeedsResolve(expr string) bool {
	e := strings.TrimSpace(strings.ToLower(expr))
	switch e {
	case "needs.resolve.outputs.tasks", "needs.resolve.outputs.has_tasks":
		return true
	default:
		return false
	}
}

// isAllowedGitHub permits gh-aw markdown subset: github.event.* and select top-level github fields.
func isAllowedGitHub(expr string) bool {
	e := strings.TrimSpace(expr)
	lower := strings.ToLower(e)
	if !strings.HasPrefix(lower, "github.") {
		return false
	}
	if strings.HasPrefix(lower, "github.event.") {
		path := e[len("github.event."):]
		return path != "" && isSafePath(path)
	}
	if lower == "github.event" {
		return false
	}
	// Direct github.* (no event subtree)
	switch lower {
	case "github.repository", "github.actor", "github.repository_owner", "github.run_id", "github.run_number",
		"github.run_attempt", "github.workflow", "github.server_url", "github.ref", "github.ref_name",
		"github.sha", "github.workspace", "github.event_name", "github.job", "github.api_url",
		"github.base_ref", "github.head_ref":
		return true
	default:
		return false
	}
}

// isSafePath allows github.event.<segments> property paths (letters, digits, _, -).
func isSafePath(path string) bool {
	if path == "" {
		return false
	}
	for _, p := range strings.Split(path, ".") {
		if p == "" {
			return false
		}
		for _, r := range p {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
				continue
			}
			return false
		}
	}
	return true
}
