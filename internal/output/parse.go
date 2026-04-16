package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ParseAgentOutputFile reads and parses output.json. Returns nil, nil if file is missing or empty.
func ParseAgentOutputFile(path string) (*AgentOutputFile, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read output file: %w", err)
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return nil, nil
	}
	var root AgentOutputFile
	if err := json.Unmarshal(b, &root); err != nil {
		return nil, fmt.Errorf("parse output.json: %w", err)
	}
	return &root, nil
}

// ItemType returns the normalized type string from an item map (underscore form).
func ItemType(m map[string]any) string {
	if m == nil {
		return ""
	}
	raw, ok := m["type"]
	if !ok {
		return ""
	}
	switch v := raw.(type) {
	case string:
		s := strings.TrimSpace(strings.ToLower(v))
		return strings.ReplaceAll(s, "-", "_")
	default:
		return ""
	}
}

// ParseOutputKind maps a type string to OutputKind.
func ParseOutputKind(s string) OutputKind {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "-", "_")
	switch s {
	case "create_pull_request":
		return KindCreatePullRequest
	case "add_comment":
		return KindAddComment
	case "add_labels":
		return KindAddLabels
	case "remove_labels":
		return KindRemoveLabels
	case "create_issue":
		return KindCreateIssue
	case "noop":
		return KindNoop
	default:
		return ""
	}
}

func stringField(m map[string]any, key string) string {
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

func stringSliceField(m map[string]any, key string) []string {
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

func intField(m map[string]any, key string) int {
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

func boolPtrField(m map[string]any, key string) *bool {
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

func mapToCreatePR(m map[string]any) ItemCreatePullRequest {
	return ItemCreatePullRequest{
		Title:  stringField(m, "title"),
		Body:   stringField(m, "body"),
		Draft:  boolPtrField(m, "draft"),
		Labels: stringSliceField(m, "labels"),
	}
}

func mapToAddComment(m map[string]any) ItemAddComment {
	return ItemAddComment{
		Body:   stringField(m, "body"),
		Target: intField(m, "target"),
	}
}

func mapToLabels(m map[string]any) ItemLabels {
	return ItemLabels{
		Labels: stringSliceField(m, "labels"),
		Target: intField(m, "target"),
	}
}

func mapToCreateIssue(m map[string]any) ItemCreateIssue {
	return ItemCreateIssue{
		Title:     stringField(m, "title"),
		Body:      stringField(m, "body"),
		Labels:    stringSliceField(m, "labels"),
		Assignees: stringSliceField(m, "assignees"),
	}
}

func mapToNoop(m map[string]any) ItemNoop {
	return ItemNoop{Message: stringField(m, "message")}
}
