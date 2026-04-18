// Package scalar provides shared coercion helpers for YAML/JSON map[string]any values.
package scalar

import "strings"

// StringFromMap returns a trimmed string field from m[key], or "" if missing or not a string.
func StringFromMap(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

// StringSliceFromMap returns a trimmed non-empty string slice from m[key] when it is []any of strings.
func StringSliceFromMap(m map[string]any, key string) []string {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, x := range arr {
		s, ok := x.(string)
		if ok && strings.TrimSpace(s) != "" {
			out = append(out, strings.TrimSpace(s))
		}
	}
	return out
}

// IntFromMap returns an int from m[key] for float64/int/int64 JSON numbers.
func IntFromMap(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	default:
		return 0
	}
}

// BoolPtrFromMap returns *bool when m[key] is a bool.
func BoolPtrFromMap(m map[string]any, key string) *bool {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	switch x := v.(type) {
	case bool:
		b := x
		return &b
	default:
		return nil
	}
}

// StringField is like StringFromMap but for item maps keyed by JSON field names (parse.go).
func StringField(m map[string]any, key string) string {
	return StringFromMap(m, key)
}

// StringSliceField extracts []string from m[key] when encoded as []any (parse.go item maps).
func StringSliceField(m map[string]any, key string) []string {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, x := range arr {
		s, ok := x.(string)
		if ok && strings.TrimSpace(s) != "" {
			out = append(out, strings.TrimSpace(s))
		}
	}
	return out
}

// IntField extracts int from item map (parse.go).
func IntField(m map[string]any, key string) int {
	return IntFromMap(m, key)
}

// IntFieldFirst walks keys in order and returns the first strictly positive int from m.
// Use with "target" first so explicit target wins over gh-aw aliases (issue_number, etc.).
func IntFieldFirst(m map[string]any, keys ...string) int {
	for _, key := range keys {
		n := IntField(m, key)
		if n > 0 {
			return n
		}
	}
	return 0
}

// BoolPtrField extracts *bool from item map (parse.go).
func BoolPtrField(m map[string]any, key string) *bool {
	return BoolPtrFromMap(m, key)
}
