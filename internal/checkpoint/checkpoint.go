// Package checkpoint persists agent state in HTML comments on issues (optional).
package checkpoint

import (
	"encoding/json"
	"fmt"
	"strings"
)

const marker = "<!-- wm-checkpoint:"

// Checkpoint is stored in issue comments.
type Checkpoint struct {
	Branch       string   `json:"branch"`
	SHA          string   `json:"sha"`
	Step         string   `json:"step"`
	FilesChanged []string `json:"files_changed"`
	Summary      string   `json:"summary"`
	Timestamp    string   `json:"timestamp"`
}

// Encode serializes checkpoint as HTML comment.
func Encode(c Checkpoint) string {
	b, _ := json.Marshal(c)
	return fmt.Sprintf("%s %s -->", marker, string(b))
}

// ParseLatest extracts the last checkpoint from comment bodies.
func ParseLatest(comments []string) (*Checkpoint, error) {
	var last *Checkpoint
	for _, body := range comments {
		idx := strings.Index(body, marker)
		if idx < 0 {
			continue
		}
		rest := body[idx+len(marker):]
		end := strings.Index(rest, "-->")
		if end < 0 {
			continue
		}
		var c Checkpoint
		if err := json.Unmarshal([]byte(strings.TrimSpace(rest[:end])), &c); err != nil {
			continue
		}
		last = &c
	}
	return last, nil
}
