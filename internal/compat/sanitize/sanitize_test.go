package sanitize

import "testing"

func TestFromPayload_Issue(t *testing.T) {
	text, title, body := FromPayload(map[string]any{
		"issue": map[string]any{"title": "T", "body": "B"},
	})
	if title != "T" || body != "B" || text != "T\n\nB" {
		t.Fatalf("got text=%q title=%q body=%q", text, title, body)
	}
}

func TestFromPayload_Comment(t *testing.T) {
	text, title, body := FromPayload(map[string]any{
		"comment": map[string]any{"body": "cbody"},
		"issue":   map[string]any{"title": "It"},
	})
	if body != "cbody" || title != "It" || text != "cbody" {
		t.Fatalf("got text=%q title=%q body=%q", text, title, body)
	}
}
