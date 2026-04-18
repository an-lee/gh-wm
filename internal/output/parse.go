package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
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

// ParseAgentOutputJSONLFile reads one JSON object per line (NDJSON). Malformed lines are skipped.
// Returns nil, nil if the path is empty, the file is missing, or there are no valid lines.
func ParseAgentOutputJSONLFile(path string) ([]map[string]any, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read output jsonl: %w", err)
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return nil, nil
	}
	var out []map[string]any
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			slog.Info("wm: safe-output jsonl: skip malformed line", "path", path, "line", lineNum, "err", err)
			continue
		}
		out = append(out, m)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan output jsonl: %w", err)
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
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
	case "create_pull_request_review_comment":
		return KindCreatePullRequestReviewComment
	case "submit_pull_request_review":
		return KindSubmitPullRequestReview
	case "noop":
		return KindNoop
	case "missing_tool":
		return KindMissingTool
	case "missing_data":
		return KindMissingData
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

func mapToMissingTool(m map[string]any) ItemMissingTool {
	return ItemMissingTool{
		Tool:   scalar.StringField(m, "tool"),
		Reason: scalar.StringField(m, "reason"),
	}
}

func mapToMissingData(m map[string]any) ItemMissingData {
	return ItemMissingData{
		What:   scalar.StringField(m, "what"),
		Reason: scalar.StringField(m, "reason"),
	}
}

func mapToCreatePullRequestReviewComment(m map[string]any) ItemCreatePullRequestReviewComment {
	return ItemCreatePullRequestReviewComment{
		Body:     scalar.StringField(m, "body"),
		Path:     scalar.StringField(m, "path"),
		Line:     scalar.IntField(m, "line"),
		Side:     scalar.StringField(m, "side"),
		CommitID: scalar.StringField(m, "commit_id"),
		Target:   scalar.IntField(m, "target"),
	}
}

func mapToSubmitPullRequestReview(m map[string]any) ItemSubmitPullRequestReview {
	return ItemSubmitPullRequestReview{
		Event:    scalar.StringField(m, "event"),
		Body:     scalar.StringField(m, "body"),
		CommitID: scalar.StringField(m, "commit_id"),
		Target:   scalar.IntField(m, "target"),
	}
}
