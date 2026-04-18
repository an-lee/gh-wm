package output

import (
	"fmt"
	"strings"
)

const (
	islandStartMarker = "<!-- gh-wm:island -->"
	islandEndMarker   = "<!-- /gh-wm:island -->"
)

// normalizeUpdateOperation returns canonical operation: replace, append, prepend, replace_island.
// Empty or "replace" means full body replacement (default). Accepts gh-aw hyphen or underscore forms.
func normalizeUpdateOperation(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "-", "_")
	switch s {
	case "", "replace":
		return "replace"
	case "append":
		return "append"
	case "prepend":
		return "prepend"
	case "replace_island":
		return "replace_island"
	default:
		return s
	}
}

func isValidUpdateOperation(s string) bool {
	if strings.TrimSpace(s) == "" {
		return true
	}
	switch normalizeUpdateOperation(s) {
	case "replace", "append", "prepend", "replace_island":
		return true
	default:
		return false
	}
}

func replaceGhWMIsland(current, replacement string) (string, error) {
	start := strings.Index(current, islandStartMarker)
	end := strings.Index(current, islandEndMarker)
	if start < 0 || end < 0 || end <= start {
		return "", fmt.Errorf("replace_island: markers %q … %q not found or invalid order", islandStartMarker, islandEndMarker)
	}
	prefix := current[:start+len(islandStartMarker)]
	suffix := current[end:]
	return prefix + "\n" + replacement + "\n" + suffix, nil
}
