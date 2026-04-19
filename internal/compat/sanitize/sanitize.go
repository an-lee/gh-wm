// Package sanitize derives gh-aw–style sanitized title/body/text from a GitHub webhook payload (github.event).
package sanitize

import (
	"fmt"
	"strings"
)

// FromPayload returns sanitized text, title, and body matching gh-aw semantics:
//   - issues / pull_request: text ≈ title + body; title and body from the issue or PR
//   - issue_comment / review comment: text is comment body; title from parent issue/PR when present
func FromPayload(payload map[string]any) (text, title, body string) {
	if payload == nil {
		return "", "", ""
	}
	// Comment events include both issue/PR and comment — prefer comment body for text (gh-aw semantics).
	if comment, ok := asMap(payload["comment"]); ok {
		body = stringField(comment["body"])
		if issue, ok := asMap(payload["issue"]); ok {
			title = stringField(issue["title"])
		} else if pr, ok := asMap(payload["pull_request"]); ok {
			title = stringField(pr["title"])
		}
		text = body
		return trimAll(text, title, body)
	}
	// issues event
	if issue, ok := asMap(payload["issue"]); ok {
		title = stringField(issue["title"])
		body = stringField(issue["body"])
		text = joinTitleBody(title, body)
		return trimAll(text, title, body)
	}
	// pull_request event
	if pr, ok := asMap(payload["pull_request"]); ok {
		title = stringField(pr["title"])
		body = stringField(pr["body"])
		text = joinTitleBody(title, body)
		return trimAll(text, title, body)
	}
	return "", "", ""
}

func joinTitleBody(title, body string) string {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if title == "" {
		return body
	}
	if body == "" {
		return title
	}
	return title + "\n\n" + body
}

func stringField(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case float64:
		return strings.TrimSpace(fmt.Sprint(x))
	case int:
		return fmt.Sprintf("%d", x)
	case int64:
		return fmt.Sprintf("%d", x)
	default:
		return strings.TrimSpace(fmt.Sprint(x))
	}
}

func asMap(v any) (map[string]any, bool) {
	if v == nil {
		return nil, false
	}
	m, ok := v.(map[string]any)
	return m, ok
}

func trimAll(text, title, body string) (string, string, string) {
	return strings.TrimSpace(text), strings.TrimSpace(title), strings.TrimSpace(body)
}
