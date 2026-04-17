// Package engines builds agent subprocess commands for [engine.runAgent].
package engines

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
)

const WMAgentPromptPlaceholder = "{prompt}"

// LogStreamMode is true when built-in claude should use stream-json for live stderr.
type LogStreamMode bool

// IsBuiltinClaude reports whether we invoke the stock claude binary (not WM_AGENT_CMD, codex).
func IsBuiltinClaude(wmAgentCmd, engineName string) bool {
	if strings.TrimSpace(wmAgentCmd) != "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(engineName)) {
	case "codex":
		return false
	default:
		return true
	}
}

// AgentOutputFormatForRun selects the claude -p --output-format and run-dir filename base.
func AgentOutputFormatForRun(wmAgentCmd, engineName string, glob *config.GlobalConfig) string {
	if strings.TrimSpace(wmAgentCmd) != "" {
		return config.ClaudeOutputFormatText
	}
	switch strings.ToLower(strings.TrimSpace(engineName)) {
	case "codex":
		return config.ClaudeOutputFormatText
	default:
		return config.EffectiveClaudeOutputFormat(glob)
	}
}

// AgentCLIArgs returns argv for claude/codex non-interactive runs (prompt via stdin).
func AgentCLIArgs(glob *config.GlobalConfig, claudeOutputFormat string) []string {
	args := []string{"-p", "--dangerously-skip-permissions"}
	if glob == nil {
		if claudeOutputFormat == config.ClaudeOutputFormatJSON || claudeOutputFormat == config.ClaudeOutputFormatStreamJSON {
			if claudeOutputFormat == config.ClaudeOutputFormatStreamJSON {
				args = append(args, "--verbose")
			}
			args = append(args, "--output-format", claudeOutputFormat)
		}
		return args
	}
	if glob.Model != "" {
		args = append(args, "--model", glob.Model)
	}
	if glob.MaxTurns > 0 {
		args = append(args, "--max-turns", strconv.Itoa(glob.MaxTurns))
	}
	if claudeOutputFormat == config.ClaudeOutputFormatJSON || claudeOutputFormat == config.ClaudeOutputFormatStreamJSON {
		if claudeOutputFormat == config.ClaudeOutputFormatStreamJSON {
			args = append(args, "--verbose")
		}
		args = append(args, "--output-format", claudeOutputFormat)
	}
	return args
}

// ParseWMAgentCmd builds exec name and args from WM_AGENT_CMD / WM_ENGINE_CODEX_CMD.
func ParseWMAgentCmd(cmdLine, prompt string) (name string, args []string, stdin io.Reader, err error) {
	if strings.Contains(cmdLine, WMAgentPromptPlaceholder) {
		parts := strings.SplitN(cmdLine, WMAgentPromptPlaceholder, 2)
		pre := strings.Fields(strings.TrimSpace(parts[0]))
		suf := strings.Fields(strings.TrimSpace(parts[1]))
		if len(pre) == 0 {
			return "", nil, nil, fmt.Errorf("WM_AGENT_CMD: missing executable before {prompt}")
		}
		name = pre[0]
		args = append(append(append([]string{}, pre[1:]...), prompt), suf...)
		return name, args, nil, nil
	}
	fields := strings.Fields(cmdLine)
	if len(fields) == 0 {
		return "", nil, nil, fmt.Errorf("WM_AGENT_CMD: empty")
	}
	return fields[0], append(fields[1:], prompt), nil, nil
}
