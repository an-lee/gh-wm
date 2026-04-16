package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const logStreamAgentTextMax = 120

// logStreamWriter parses Claude Code --output-format stream-json (JSONL) and writes
// human-readable lines to the underlying writer. Raw JSONL is unchanged on disk (tee before this).
type logStreamWriter struct {
	w       io.Writer
	lineBuf bytes.Buffer
	textBuf strings.Builder
}

func newLogStreamWriter(w io.Writer) *logStreamWriter {
	if w == nil {
		w = io.Discard
	}
	return &logStreamWriter{w: w}
}

func (l *logStreamWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	l.lineBuf.Write(p)
	for {
		data := l.lineBuf.Bytes()
		i := bytes.IndexByte(data, '\n')
		if i < 0 {
			break
		}
		line := strings.TrimSpace(string(data[:i]))
		l.lineBuf.Next(i + 1)
		if line != "" {
			if err := l.writeFormattedLine(line); err != nil {
				return 0, err
			}
		}
	}
	return len(p), nil
}

func (l *logStreamWriter) writeFormattedLine(line string) error {
	var ev map[string]any
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		_, err := fmt.Fprintf(l.w, "%s\n", truncateRunes(line, 400))
		return err
	}
	s := formatStreamJSONEvent(ev, &l.textBuf)
	if s == "" {
		return nil
	}
	_, err := fmt.Fprintln(l.w, s)
	return err
}

func formatStreamJSONEvent(ev map[string]any, textBuf *strings.Builder) string {
	typ, _ := ev["type"].(string)
	switch typ {
	case "result":
		return formatResultEvent(ev)
	case "system":
		return formatSystemEvent(ev)
	case "assistant":
		return formatAssistantEvent(ev)
	case "user":
		return formatUserEvent(ev)
	case "stream_event":
		return formatStreamEvent(ev, textBuf)
	default:
		if typ != "" {
			return fmt.Sprintf("[event] type=%s", typ)
		}
		return ""
	}
}

func formatResultEvent(ev map[string]any) string {
	sub, _ := ev["subtype"].(string)
	if sub == "" {
		sub = "done"
	}
	parts := []string{fmt.Sprintf("[result] %s", sub)}
	if v, ok := pickFloat(ev, "total_cost_usd", "cost_usd", "cost"); ok {
		parts = append(parts, fmt.Sprintf("cost=$%.4f", v))
	}
	if n, ok := pickInt(ev, "num_turns", "turns"); ok {
		parts = append(parts, fmt.Sprintf("turns=%d", n))
	}
	return strings.Join(parts, " | ")
}

func formatSystemEvent(ev map[string]any) string {
	sub, _ := ev["subtype"].(string)
	if sub == "" {
		sub = "system"
	}
	switch sub {
	case "init":
		return "[session] started"
	default:
		return fmt.Sprintf("[system] %s", sub)
	}
}

func formatAssistantEvent(ev map[string]any) string {
	msg, ok := ev["message"].(map[string]any)
	if !ok {
		return ""
	}
	content, ok := msg["content"].([]any)
	if !ok {
		return ""
	}
	var lines []string
	for _, block := range content {
		m, ok := block.(map[string]any)
		if !ok {
			continue
		}
		bt, _ := m["type"].(string)
		switch bt {
		case "tool_use":
			name, _ := m["name"].(string)
			if name == "" {
				name = "tool"
			}
			hint := toolInputHint(m["input"])
			if hint != "" {
				lines = append(lines, fmt.Sprintf("[tool] %s → %s", name, hint))
			} else {
				lines = append(lines, fmt.Sprintf("[tool] %s", name))
			}
		case "text":
			txt, _ := m["text"].(string)
			txt = strings.TrimSpace(txt)
			if txt != "" {
				lines = append(lines, fmt.Sprintf("[agent] %s", truncateRunes(txt, logStreamAgentTextMax)))
			}
		}
	}
	return strings.Join(lines, "\n")
}

func formatUserEvent(ev map[string]any) string {
	msg, ok := ev["message"].(map[string]any)
	if !ok {
		return ""
	}
	content, ok := msg["content"].([]any)
	if !ok {
		return ""
	}
	var lines []string
	for _, block := range content {
		m, ok := block.(map[string]any)
		if !ok {
			continue
		}
		bt, _ := m["type"].(string)
		if bt == "tool_result" {
			tid, _ := m["tool_use_id"].(string)
			if tid != "" {
				lines = append(lines, fmt.Sprintf("[tool_result] id=%s", truncateRunes(tid, 40)))
			} else {
				lines = append(lines, "[tool_result]")
			}
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

func formatStreamEvent(ev map[string]any, textBuf *strings.Builder) string {
	inner, ok := ev["event"].(map[string]any)
	if !ok {
		return ""
	}
	et, _ := inner["type"].(string)
	switch et {
	case "content_block_start":
		var parts []string
		if textBuf != nil && textBuf.Len() > 0 {
			s := strings.TrimSpace(textBuf.String())
			textBuf.Reset()
			if s != "" {
				parts = append(parts, fmt.Sprintf("[agent] %s", truncateRunes(s, logStreamAgentTextMax)))
			}
		}
		cb, ok := inner["content_block"].(map[string]any)
		if !ok {
			return strings.Join(parts, "\n")
		}
		cbt, _ := cb["type"].(string)
		if cbt == "tool_use" {
			name, _ := cb["name"].(string)
			if name == "" {
				name = "tool"
			}
			hint := toolInputHint(cb["input"])
			var toolLine string
			if hint != "" {
				toolLine = fmt.Sprintf("[tool] %s → %s", name, hint)
			} else {
				toolLine = fmt.Sprintf("[tool] %s", name)
			}
			parts = append(parts, toolLine)
		}
		return strings.Join(parts, "\n")
	case "message_stop":
		if textBuf != nil && textBuf.Len() > 0 {
			s := strings.TrimSpace(textBuf.String())
			textBuf.Reset()
			if s != "" {
				return fmt.Sprintf("[agent] %s", truncateRunes(s, logStreamAgentTextMax))
			}
		}
		return ""
	case "content_block_stop":
		if textBuf != nil && textBuf.Len() > 0 {
			s := strings.TrimSpace(textBuf.String())
			textBuf.Reset()
			if s != "" {
				return fmt.Sprintf("[agent] %s", truncateRunes(s, logStreamAgentTextMax))
			}
		}
		return ""
	}
	delta, ok := inner["delta"].(map[string]any)
	if !ok {
		return ""
	}
	dt, _ := delta["type"].(string)
	switch dt {
	case "text_delta":
		txt, _ := delta["text"].(string)
		if textBuf != nil && txt != "" {
			textBuf.WriteString(txt)
		}
		return ""
	default:
		return ""
	}
}

func toolInputHint(v any) string {
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	for _, key := range []string{"command", "file_path", "path", "pattern", "query"} {
		if s, ok := m[key].(string); ok && strings.TrimSpace(s) != "" {
			return truncateRunes(strings.TrimSpace(s), logStreamAgentTextMax)
		}
	}
	return ""
}

func pickFloat(ev map[string]any, keys ...string) (float64, bool) {
	for _, k := range keys {
		switch x := ev[k].(type) {
		case float64:
			return x, true
		case int:
			return float64(x), true
		case int64:
			return float64(x), true
		}
	}
	return 0, false
}

func pickInt(ev map[string]any, keys ...string) (int, bool) {
	for _, k := range keys {
		switch x := ev[k].(type) {
		case float64:
			return int(x), true
		case int:
			return x, true
		case int64:
			return int(x), true
		}
	}
	return 0, false
}

func truncateRunes(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
