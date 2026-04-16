package config

import (
	"os"
	"strings"
)

// Claude print-mode output formats (see Claude Code CLI --output-format).
const (
	ClaudeOutputFormatText       = "text"
	ClaudeOutputFormatJSON       = "json"
	ClaudeOutputFormatStreamJSON = "stream-json"
)

// EffectiveClaudeOutputFormat returns the claude -p output format: WM_CLAUDE_OUTPUT_FORMAT
// overrides .wm/config.yml claude_output_format; empty means text.
func EffectiveClaudeOutputFormat(glob *GlobalConfig) string {
	if v := strings.TrimSpace(os.Getenv("WM_CLAUDE_OUTPUT_FORMAT")); v != "" {
		return normalizeClaudeOutputFormat(v)
	}
	if glob != nil {
		if v := strings.TrimSpace(glob.ClaudeOutputFormat); v != "" {
			return normalizeClaudeOutputFormat(v)
		}
	}
	return ClaudeOutputFormatText
}

func normalizeClaudeOutputFormat(s string) string {
	switch strings.ToLower(s) {
	case ClaudeOutputFormatJSON:
		return ClaudeOutputFormatJSON
	case ClaudeOutputFormatStreamJSON:
		return ClaudeOutputFormatStreamJSON
	case ClaudeOutputFormatText, "":
		return ClaudeOutputFormatText
	default:
		return ClaudeOutputFormatText
	}
}
