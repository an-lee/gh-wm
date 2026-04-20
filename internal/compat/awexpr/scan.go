package awexpr

import (
	"strings"
	"unicode"
)

// Span is a `${{ ... }}` region in the task body (byte offsets into the original string).
type Span struct {
	Start, End int // [Start, End) half-open, End exclusive
	Expr       string
}

// ScanBody finds `${{ ... }}` expressions outside Markdown fenced code blocks (```).
func ScanBody(body string) []Span {
	iv := fenceIntervals(body)
	var out []Span
	for i := 0; i < len(body); {
		j := strings.Index(body[i:], "${{")
		if j < 0 {
			break
		}
		start := i + j
		if overlapsIntervals(start, iv) {
			i = start + 3
			continue
		}
		// Find closing `}}`
		k := start + 3
		closeIdx := -1
		for k < len(body)-1 {
			if body[k] == '}' && body[k+1] == '}' {
				closeIdx = k
				break
			}
			k++
		}
		if closeIdx < 0 {
			break
		}
		expr := strings.TrimSpace(body[start+3 : closeIdx])
		end := closeIdx + 2
		out = append(out, Span{Start: start, End: end, Expr: expr})
		i = end
	}
	return out
}

func overlapsIntervals(pos int, iv [][2]int) bool {
	for _, p := range iv {
		if pos >= p[0] && pos < p[1] {
			return true
		}
	}
	return false
}

// fenceIntervals returns [start, end) byte ranges of ``` fenced blocks (GFM-style, line-based).
func fenceIntervals(body string) [][2]int {
	lines := strings.Split(body, "\n")
	var out [][2]int
	inFence := false
	var fenceStart int
	offset := 0
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			if !inFence {
				inFence = true
				fenceStart = offset
			} else {
				out = append(out, [2]int{fenceStart, offset + len(line)})
				inFence = false
			}
		}
		offset += len(line) + 1
	}
	if inFence {
		out = append(out, [2]int{fenceStart, len(body)})
	}
	return out
}

// ContainsUnsupportedTemplate returns true if the body contains gh-aw `{{#...}}` directives outside fences.
func ContainsUnsupportedTemplate(body string) bool {
	iv := fenceIntervals(body)
	for i := 0; i < len(body); {
		j := strings.Index(body[i:], "{{#")
		if j < 0 {
			return false
		}
		pos := i + j
		if overlapsIntervals(pos, iv) {
			i = pos + 3
			continue
		}
		return true
	}
	return false
}

// TrimExpression trims spaces and normalizes inner whitespace for comparison.
func TrimExpression(s string) string {
	return strings.TrimFunc(s, unicode.IsSpace)
}
