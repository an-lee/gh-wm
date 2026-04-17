package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/an-lee/gh-wm/internal/config/scalar"
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

func mapToCreatePR(m map[string]any) ItemCreatePullRequest {
	return ItemCreatePullRequest{
		Title:  scalar.StringField(m, "title"),
		Body:   scalar.StringField(m, "body"),
		Draft:  scalar.BoolPtrField(m, "draft"),
		Labels: scalar.StringSliceField(m, "labels"),
	}
}

func mapToAddComment(m map[string]any) ItemAddComment {
	return ItemAddComment{
		Body:   scalar.StringField(m, "body"),
		Target: scalar.IntField(m, "target"),
	}
}

func mapToLabels(m map[string]any) ItemLabels {
	return ItemLabels{
		Labels: scalar.StringSliceField(m, "labels"),
		Target: scalar.IntField(m, "target"),
	}
}

func mapToCreateIssue(m map[string]any) ItemCreateIssue {
	return ItemCreateIssue{
		Title:     scalar.StringField(m, "title"),
		Body:      scalar.StringField(m, "body"),
		Labels:    scalar.StringSliceField(m, "labels"),
		Assignees: scalar.StringSliceField(m, "assignees"),
	}
}

func mapToNoop(m map[string]any) ItemNoop {
	return ItemNoop{Message: scalar.StringField(m, "message")}
}
