package output

import "testing"

func TestNormalizeNestedNoopItem(t *testing.T) {
	t.Parallel()
	raw := map[string]any{"noop": map[string]any{"message": "nested"}}
	got := normalizeNestedNoopItem(raw)
	if ItemType(got) != "noop" {
		t.Fatalf("type: got %q", ItemType(got))
	}
	if got["message"] != "nested" {
		t.Fatalf("message: got %v", got["message"])
	}
}

func TestNormalizeNestedNoopItem_WithTypeUnchanged(t *testing.T) {
	t.Parallel()
	raw := map[string]any{"type": "noop", "message": "flat"}
	got := normalizeNestedNoopItem(raw)
	if got["message"] != "flat" {
		t.Fatal("should not wrap")
	}
}
