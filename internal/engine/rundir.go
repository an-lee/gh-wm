package engine

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/types"
)

const (
	runDirEnv        = "WM_RUN_DIR"
	promptFileName   = "prompt.md"
	agentLogFileName = "agent-stdout.log"
	// conversationJSONFileName holds claude -p --output-format json (built-in claude only).
	conversationJSONFileName = "conversation.json"
	// conversationJSONLFileName holds claude -p --output-format stream-json (built-in claude only).
	conversationJSONLFileName = "conversation.jsonl"
	metaFileName              = "meta.json"
	resultFileName            = "result.json"
	// outputJSONFileName is the agent-written structured safe-outputs request (WM_OUTPUT_FILE).
	outputJSONFileName = "output.json"

	defaultPruneAge = 7 * 24 * time.Hour
)

// RunDir is a per-run artifact directory under .wm/runs/<id>/ or WM_RUN_DIR/<id>/.
type RunDir struct {
	Path string
}

type runMeta struct {
	RunID     string `json:"run_id"`
	TaskName  string `json:"task_name"`
	EventName string `json:"event_name"`
	StartedAt string `json:"started_at"` // RFC3339 UTC
	Phase     string `json:"phase"`
	Success   bool   `json:"success"`
}

// NewRunDir creates a unique run directory, writes initial meta.json, and prunes old runs.
func NewRunDir(repoRoot, taskName, eventName string) (*RunDir, error) {
	_ = PruneRunDirs(repoRoot, defaultPruneAge)
	if alt := strings.TrimSpace(os.Getenv(runDirEnv)); alt != "" {
		_ = pruneRunDirBase(filepath.Clean(alt), defaultPruneAge)
	}

	runID := makeRunID(taskName)
	var base string
	if alt := strings.TrimSpace(os.Getenv(runDirEnv)); alt != "" {
		base = filepath.Join(filepath.Clean(alt), runID)
	} else {
		abs, err := filepath.Abs(repoRoot)
		if err != nil {
			return nil, fmt.Errorf("run dir: %w", err)
		}
		base = filepath.Join(abs, ".wm", "runs", runID)
	}
	if err := os.MkdirAll(base, 0o755); err != nil {
		return nil, fmt.Errorf("run dir mkdir: %w", err)
	}

	rd := &RunDir{Path: base}
	meta := runMeta{
		RunID:     runID,
		TaskName:  taskName,
		EventName: eventName,
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		Phase:     string(types.PhaseActivation),
		Success:   false,
	}
	if err := rd.writeMeta(&meta); err != nil {
		return nil, err
	}
	return rd, nil
}

func makeRunID(taskName string) string {
	slug := slugTaskNameForRun(taskName)
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Extremely unlikely; fall back to timestamp-only id uniqueness via nanos in path is not ideal — use pid+time
		slug = fmt.Sprintf("%s-%d", slug, time.Now().UnixNano())
		return slug
	}
	return fmt.Sprintf("%s-%s-%s", slug, time.Now().UTC().Format("20060102T150405"), hex.EncodeToString(b))
}

func slugTaskNameForRun(name string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(name) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
		case r == '-' || r == '_' || unicode.IsSpace(r):
			if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
				b.WriteRune('-')
			}
		}
	}
	s := strings.Trim(b.String(), "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	if s == "" {
		s = "task"
	}
	if len(s) > 48 {
		s = s[:48]
		s = strings.TrimRight(s, "-")
	}
	return s
}

func (r *RunDir) writeMeta(m *runMeta) error {
	if r == nil || m == nil {
		return nil
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("run meta encode: %w", err)
	}
	path := filepath.Join(r.Path, metaFileName)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write meta.json: %w", err)
	}
	return nil
}

// UpdateMeta writes phase and optional success flag to meta.json.
func (r *RunDir) UpdateMeta(phase types.Phase, success bool) error {
	if r == nil {
		return nil
	}
	path := filepath.Join(r.Path, metaFileName)
	prev, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read meta.json: %w", err)
	}
	var m runMeta
	if err := json.Unmarshal(prev, &m); err != nil {
		return fmt.Errorf("parse meta.json: %w", err)
	}
	m.Phase = string(phase)
	m.Success = success
	return r.writeMeta(&m)
}

// WritePrompt writes the assembled agent prompt.
func (r *RunDir) WritePrompt(prompt string) error {
	if r == nil {
		return nil
	}
	path := filepath.Join(r.Path, promptFileName)
	return os.WriteFile(path, []byte(prompt), 0o644)
}

func agentArtifactFilename(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case config.ClaudeOutputFormatJSON:
		return conversationJSONFileName
	case config.ClaudeOutputFormatStreamJSON:
		return conversationJSONLFileName
	default:
		return agentLogFileName
	}
}

// AgentOutputPath returns the path to the combined agent stdout/stderr capture file for this run.
func (r *RunDir) AgentOutputPath(format string) string {
	if r == nil {
		return ""
	}
	return filepath.Join(r.Path, agentArtifactFilename(format))
}

// OutputJSONPath returns the path for structured safe-output JSON (see WM_OUTPUT_FILE).
func (r *RunDir) OutputJSONPath() string {
	if r == nil {
		return ""
	}
	return filepath.Join(r.Path, outputJSONFileName)
}

// OpenAgentOutput creates or truncates the agent output file for writing (see agentArtifactFilename).
func (r *RunDir) OpenAgentOutput(format string) (*os.File, error) {
	if r == nil {
		return nil, fmt.Errorf("run dir is nil")
	}
	path := filepath.Join(r.Path, agentArtifactFilename(format))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open agent log: %w", err)
	}
	return f, nil
}

// WriteResult writes a JSON snapshot of the run outcome (errors as strings).
func (r *RunDir) WriteResult(res *types.RunResult) error {
	if r == nil || res == nil {
		return nil
	}
	var errs []string
	for _, e := range res.Errors {
		if e != nil {
			errs = append(errs, e.Error())
		}
	}
	out := struct {
		Phase       string   `json:"phase"`
		Success     bool     `json:"success"`
		Duration    string   `json:"duration"`
		DurationMs  int64    `json:"duration_ms"`
		Errors      []string `json:"errors,omitempty"`
		RunDir      string   `json:"run_dir"`
		AgentResult *struct {
			Success         bool   `json:"success"`
			ExitCode        int    `json:"exit_code"`
			Stdout          string `json:"stdout,omitempty"`
			Stderr          string `json:"stderr,omitempty"`
			Summary         string `json:"summary,omitempty"`
			TimedOut        bool   `json:"timed_out,omitempty"`
			AgentStdoutPath string `json:"agent_stdout_path,omitempty"`
			OutputFilePath  string `json:"output_file_path,omitempty"`
		} `json:"agent_result,omitempty"`
	}{
		Phase:      string(res.Phase),
		Success:    res.Success,
		Duration:   res.Duration.String(),
		DurationMs: res.Duration.Milliseconds(),
		Errors:     errs,
		RunDir:     r.Path,
	}
	if res.AgentResult != nil {
		ar := res.AgentResult
		out.AgentResult = &struct {
			Success         bool   `json:"success"`
			ExitCode        int    `json:"exit_code"`
			Stdout          string `json:"stdout,omitempty"`
			Stderr          string `json:"stderr,omitempty"`
			Summary         string `json:"summary,omitempty"`
			TimedOut        bool   `json:"timed_out,omitempty"`
			AgentStdoutPath string `json:"agent_stdout_path,omitempty"`
			OutputFilePath  string `json:"output_file_path,omitempty"`
		}{
			Success:         ar.Success,
			ExitCode:        ar.ExitCode,
			Stdout:          ar.Stdout,
			Stderr:          ar.Stderr,
			Summary:         ar.Summary,
			TimedOut:        ar.TimedOut,
			AgentStdoutPath: ar.AgentStdoutPath,
			OutputFilePath:  ar.OutputFilePath,
		}
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("encode result.json: %w", err)
	}
	path := filepath.Join(r.Path, resultFileName)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write result.json: %w", err)
	}
	return nil
}

// PruneRunDirs removes run directories under repoRoot/.wm/runs older than maxAge.
func PruneRunDirs(repoRoot string, maxAge time.Duration) error {
	abs, err := filepath.Abs(repoRoot)
	if err != nil {
		return err
	}
	base := filepath.Join(abs, ".wm", "runs")
	return pruneRunDirBase(base, maxAge)
}

func pruneRunDirBase(base string, maxAge time.Duration) error {
	if maxAge <= 0 {
		return nil
	}
	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	cutoff := time.Now().Add(-maxAge)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p := filepath.Join(base, e.Name())
		st, err := os.Stat(p)
		if err != nil {
			continue
		}
		if st.ModTime().Before(cutoff) {
			_ = os.RemoveAll(p)
		}
	}
	return nil
}

const agentTailMaxBytes = 64 << 10 // 64 KiB retained in memory for checkpoint / comments

// tailBuffer keeps only the last max bytes written (for AgentResult.Stdout/Summary).
type tailBuffer struct {
	max int
	buf []byte
}

func (t *tailBuffer) Write(p []byte) (int, error) {
	if t.max <= 0 {
		t.max = agentTailMaxBytes
	}
	t.buf = append(t.buf, p...)
	if len(t.buf) > t.max {
		t.buf = t.buf[len(t.buf)-t.max:]
	}
	return len(p), nil
}

func (t *tailBuffer) String() string {
	if t == nil {
		return ""
	}
	return string(t.buf)
}
