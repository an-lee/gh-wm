package engines

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
)

// Engine builds the agent subprocess for one run (claude, codex, or WM_AGENT_CMD custom).
type Engine interface {
	// ID returns a short name for logging (claude, codex, custom).
	ID() string
	// BuildCommand returns the command, stdin reader, and artifact format basename key.
	// appendSystemPrompt is passed to the built-in claude CLI as --append-system-prompt when non-empty; other engines ignore it.
	BuildCommand(ctx context.Context, glob *config.GlobalConfig, prompt string, forceStreamJSON bool, appendSystemPrompt string) (*exec.Cmd, io.Reader, string, error)
}

type claudeEngine struct{}

func (claudeEngine) ID() string { return "claude" }

func (claudeEngine) BuildCommand(ctx context.Context, glob *config.GlobalConfig, prompt string, forceStreamJSON bool, appendSystemPrompt string) (*exec.Cmd, io.Reader, string, error) {
	artifactFormat := AgentOutputFormatForRun("", "claude", glob)
	if forceStreamJSON {
		artifactFormat = config.ClaudeOutputFormatStreamJSON
	}
	args := AgentCLIArgs(glob, artifactFormat)
	if s := strings.TrimSpace(appendSystemPrompt); s != "" {
		args = append(args, "--append-system-prompt", s)
	}
	cmd := exec.CommandContext(ctx, "claude", args...)
	return cmd, strings.NewReader(prompt), artifactFormat, nil
}

type codexEngine struct{}

func (codexEngine) ID() string { return "codex" }

func (codexEngine) BuildCommand(ctx context.Context, glob *config.GlobalConfig, prompt string, _ bool, _ string) (*exec.Cmd, io.Reader, string, error) {
	artifactFormat := AgentOutputFormatForRun("", "codex", glob)
	if alt := strings.TrimSpace(os.Getenv("WM_ENGINE_CODEX_CMD")); alt != "" {
		name, args, stdin, err := ParseWMAgentCmd(alt, prompt)
		if err != nil {
			return nil, nil, "", err
		}
		cmd := exec.CommandContext(ctx, name, args...)
		return cmd, stdin, artifactFormat, nil
	}
	cmd := exec.CommandContext(ctx, "codex", AgentCLIArgs(glob, config.ClaudeOutputFormatText)...)
	return cmd, strings.NewReader(prompt), artifactFormat, nil
}

// customEngine honors WM_AGENT_CMD (or inline codex override is handled in codexEngine).
type customEngine struct {
	cmdLine string
}

func (c customEngine) ID() string { return "custom" }

func (c customEngine) BuildCommand(ctx context.Context, glob *config.GlobalConfig, prompt string, _ bool, _ string) (*exec.Cmd, io.Reader, string, error) {
	name, args, stdin, err := ParseWMAgentCmd(c.cmdLine, prompt)
	if err != nil {
		return nil, nil, "", err
	}
	artifactFormat := AgentOutputFormatForRun(c.cmdLine, "", glob)
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd, stdin, artifactFormat, nil
}

// ResolveEngine selects the engine implementation. engineName is from task/global; wmAgentCmd is WM_AGENT_CMD.
func ResolveEngine(engineName, wmAgentCmd string) (Engine, error) {
	if strings.TrimSpace(wmAgentCmd) != "" {
		return customEngine{cmdLine: wmAgentCmd}, nil
	}
	switch strings.ToLower(strings.TrimSpace(engineName)) {
	case "codex":
		return codexEngine{}, nil
	case "copilot":
		return nil, fmt.Errorf(`engine "copilot" is no longer supported; use WM_AGENT_CMD or set engine to "claude" or "codex"`)
	default:
		return claudeEngine{}, nil
	}
}

// BuildAgentCommand delegates to [ResolveEngine]. appendSystemPrompt is passed to the built-in claude CLI only; see [output.SafeOutputsSystemPromptAppend].
func BuildAgentCommand(ctx context.Context, glob *config.GlobalConfig, engineName, wmAgentCmd, prompt string, forceStreamJSON bool, appendSystemPrompt string) (*exec.Cmd, io.Reader, string, error) {
	eng, err := ResolveEngine(engineName, wmAgentCmd)
	if err != nil {
		return nil, nil, "", err
	}
	return eng.BuildCommand(ctx, glob, prompt, forceStreamJSON, appendSystemPrompt)
}
