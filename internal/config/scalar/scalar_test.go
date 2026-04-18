package scalar

import "testing"

// --- StringFromMap ---

func TestStringFromMap_NilMap(t *testing.T) {
	t.Parallel()
	if got := StringFromMap(nil, "key"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestStringFromMap_MissingKey(t *testing.T) {
	t.Parallel()
	if got := StringFromMap(map[string]any{}, "key"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestStringFromMap_WrongType(t *testing.T) {
	t.Parallel()
	if got := StringFromMap(map[string]any{"key": 42}, "key"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestStringFromMap_Valid(t *testing.T) {
	t.Parallel()
	if got := StringFromMap(map[string]any{"key": "  hello  "}, "key"); got != "hello" {
		t.Fatalf("got %q", got)
	}
}

func TestStringFromMap_TrimsWhitespace(t *testing.T) {
	t.Parallel()
	if got := StringFromMap(map[string]any{"key": "  hello  "}, "key"); got == "  hello  " {
		t.Fatal("should trim whitespace")
	}
}

// --- StringSliceFromMap ---

func TestStringSliceFromMap_NilMap(t *testing.T) {
	t.Parallel()
	if got := StringSliceFromMap(nil, "key"); got != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestStringSliceFromMap_MissingKey(t *testing.T) {
	t.Parallel()
	if got := StringSliceFromMap(map[string]any{}, "key"); got != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestStringSliceFromMap_WrongType(t *testing.T) {
	t.Parallel()
	if got := StringSliceFromMap(map[string]any{"key": "not-an-array"}, "key"); got != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestStringSliceFromMap_Valid(t *testing.T) {
	t.Parallel()
	got := StringSliceFromMap(map[string]any{"key": []any{"a", "b", "  c  "}}, "key")
	if len(got) != 3 {
		t.Fatalf("got %d items, want 3", len(got))
	}
	if got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("got %#v", got)
	}
}

func TestStringSliceFromMap_IgnoresEmptyStrings(t *testing.T) {
	t.Parallel()
	got := StringSliceFromMap(map[string]any{"key": []any{"a", "", "  "}}, "key")
	if len(got) != 1 {
		t.Fatalf("got %#v", got)
	}
}

func TestStringSliceFromMap_IgnoresNonStrings(t *testing.T) {
	t.Parallel()
	got := StringSliceFromMap(map[string]any{"key": []any{"a", 42, true}}, "key")
	if len(got) != 1 || got[0] != "a" {
		t.Fatalf("got %#v", got)
	}
}

func TestStringSliceFromMap_TrimsStrings(t *testing.T) {
	t.Parallel()
	got := StringSliceFromMap(map[string]any{"key": []any{"  a  "}}, "key")
	if len(got) != 1 || got[0] != "a" {
		t.Fatalf("got %#v", got)
	}
}

// --- IntFromMap ---

func TestIntFromMap_NilMap(t *testing.T) {
	t.Parallel()
	if got := IntFromMap(nil, "key"); got != 0 {
		t.Fatalf("got %d", got)
	}
}

func TestIntFromMap_MissingKey(t *testing.T) {
	t.Parallel()
	if got := IntFromMap(map[string]any{}, "key"); got != 0 {
		t.Fatalf("got %d", got)
	}
}

func TestIntFromMap_WrongType(t *testing.T) {
	t.Parallel()
	if got := IntFromMap(map[string]any{"key": "not-an-int"}, "key"); got != 0 {
		t.Fatalf("got %d", got)
	}
}

func TestIntFromMap_Float64(t *testing.T) {
	t.Parallel()
	if got := IntFromMap(map[string]any{"key": float64(42)}, "key"); got != 42 {
		t.Fatalf("got %d", got)
	}
}

func TestIntFromMap_Int(t *testing.T) {
	t.Parallel()
	if got := IntFromMap(map[string]any{"key": 42}, "key"); got != 42 {
		t.Fatalf("got %d", got)
	}
}

func TestIntFromMap_Int64(t *testing.T) {
	t.Parallel()
	if got := IntFromMap(map[string]any{"key": int64(42)}, "key"); got != 42 {
		t.Fatalf("got %d", got)
	}
}

func TestIntFromMap_NegativeFloat(t *testing.T) {
	t.Parallel()
	if got := IntFromMap(map[string]any{"key": float64(-10)}, "key"); got != -10 {
		t.Fatalf("got %d", got)
	}
}

func TestIntFieldFirst_PrefersFirstPositive(t *testing.T) {
	t.Parallel()
	m := map[string]any{"target": 0, "issue_number": float64(42)}
	if got := IntFieldFirst(m, "target", "issue_number"); got != 42 {
		t.Fatalf("got %d, want 42", got)
	}
}

func TestIntFieldFirst_TargetWins(t *testing.T) {
	t.Parallel()
	m := map[string]any{"target": 3, "issue_number": 99}
	if got := IntFieldFirst(m, "target", "issue_number"); got != 3 {
		t.Fatalf("got %d, want 3", got)
	}
}

func TestIntFieldFirst_SkipsNonPositive(t *testing.T) {
	t.Parallel()
	m := map[string]any{"target": -1, "pull_request_number": 7}
	if got := IntFieldFirst(m, "target", "pull_request_number"); got != 7 {
		t.Fatalf("got %d, want 7", got)
	}
}

func TestIntFieldFirst_None(t *testing.T) {
	t.Parallel()
	if got := IntFieldFirst(map[string]any{}, "target", "issue_number"); got != 0 {
		t.Fatalf("got %d", got)
	}
}

// --- BoolPtrFromMap ---

func TestBoolPtrFromMap_NilMap(t *testing.T) {
	t.Parallel()
	if got := BoolPtrFromMap(nil, "key"); got != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestBoolPtrFromMap_MissingKey(t *testing.T) {
	t.Parallel()
	if got := BoolPtrFromMap(map[string]any{}, "key"); got != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestBoolPtrFromMap_WrongType(t *testing.T) {
	t.Parallel()
	if got := BoolPtrFromMap(map[string]any{"key": "not-a-bool"}, "key"); got != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestBoolPtrFromMap_True(t *testing.T) {
	t.Parallel()
	got := BoolPtrFromMap(map[string]any{"key": true}, "key")
	if got == nil || !*got {
		t.Fatalf("got %#v", got)
	}
}

func TestBoolPtrFromMap_False(t *testing.T) {
	t.Parallel()
	got := BoolPtrFromMap(map[string]any{"key": false}, "key")
	if got == nil || *got {
		t.Fatalf("got %#v", got)
	}
}
