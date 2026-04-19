package engine

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatStreamJSONEvent_Result(t *testing.T) {
	t.Parallel()
	ev := map[string]any{
		"type":           "result",
		"subtype":        "success",
		"total_cost_usd": 0.042,
		"num_turns":      12.0,
	}
	got := formatStreamJSONEvent(ev, nil)
	if !strings.Contains(got, "[result]") || !strings.Contains(got, "success") {
		t.Fatalf("got %q", got)
	}
	if !strings.Contains(got, "cost=") {
		t.Fatalf("expected cost: %q", got)
	}
	if !strings.Contains(got, "turns=12") {
		t.Fatalf("expected turns: %q", got)
	}
}

func TestFormatStreamJSONEvent_SystemInit(t *testing.T) {
	t.Parallel()
	got := formatStreamJSONEvent(map[string]any{"type": "system", "subtype": "init"}, nil)
	if got != "[session] started" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatStreamJSONEvent_AssistantToolUse(t *testing.T) {
	t.Parallel()
	ev := map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"content": []any{
				map[string]any{
					"type": "tool_use",
					"name": "Bash",
					"input": map[string]any{
						"command": "git log --oneline -5",
					},
				},
			},
		},
	}
	got := formatStreamJSONEvent(ev, nil)
	if !strings.HasPrefix(got, "[tool] Bash") || !strings.Contains(got, "git log") {
		t.Fatalf("got %q", got)
	}
}

func TestFormatStreamJSONEvent_StreamEventTextDeltaAndStop(t *testing.T) {
	t.Parallel()
	var textBuf strings.Builder
	got1 := formatStreamJSONEvent(map[string]any{
		"type": "stream_event",
		"event": map[string]any{
			"delta": map[string]any{"type": "text_delta", "text": "Hello "},
		},
	}, &textBuf)
	if got1 != "" {
		t.Fatalf("delta should accumulate: got %q", got1)
	}
	got2 := formatStreamJSONEvent(map[string]any{
		"type": "stream_event",
		"event": map[string]any{
			"type": "message_stop",
		},
	}, &textBuf)
	if !strings.Contains(got2, "[agent]") || !strings.Contains(got2, "Hello") {
		t.Fatalf("got %q buf=%q", got2, textBuf.String())
	}
}

func TestFormatStreamJSONEvent_UnknownType(t *testing.T) {
	t.Parallel()
	got := formatStreamJSONEvent(map[string]any{"type": "custom_kind"}, nil)
	if got != "[event] type=custom_kind" {
		t.Fatalf("got %q", got)
	}
}

func TestLogStreamWriter_ChunksAndInvalidJSON(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	w := newLogStreamWriter(&buf)
	line := `{"type":"result","subtype":"success"}` + "\n"
	for i := 0; i < len(line); i++ {
		if _, err := w.Write([]byte{line[i]}); err != nil {
			t.Fatal(err)
		}
	}
	if !strings.Contains(buf.String(), "[result]") {
		t.Fatalf("got %q", buf.String())
	}
	if _, err := w.Write([]byte("not json at all\n")); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "not json") {
		t.Fatalf("expected fallback line: %q", buf.String())
	}
}

func TestTruncateRunes(t *testing.T) {
	t.Parallel()
	out := truncateRunes("abcdefghijklmnopqrstuvwxyz", 10)
	want := "abcdefghij…"
	if out != want {
		t.Fatalf("got %q want %q", out, want)
	}
}

func TestFormatUserEvent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ev   map[string]any
		want string
	}{
		{
			name: "nil_map",
			ev:   nil,
			want: "",
		},
		{
			name: "missing_message",
			ev:   map[string]any{"type": "user"},
			want: "",
		},
		{
			name: "message_not_map",
			ev:   map[string]any{"message": "not a map"},
			want: "",
		},
		{
			name: "missing_content",
			ev:   map[string]any{"message": map[string]any{}},
			want: "",
		},
		{
			name: "content_not_slice",
			ev: map[string]any{
				"message": map[string]any{"content": "not a slice"},
			},
			want: "",
		},
		{
			name: "valid_tool_result_with_id",
			ev: map[string]any{
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "tool_result", "tool_use_id": "tool_12345678"},
					},
				},
			},
			want: "[tool_result] id=tool_12345678",
		},
		{
			name: "valid_tool_result_no_id",
			ev: map[string]any{
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "tool_result"},
					},
				},
			},
			want: "[tool_result]",
		},
		{
			name: "multiple_tool_results",
			ev: map[string]any{
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "tool_result", "tool_use_id": "tool_a"},
						map[string]any{"type": "text", "text": "hello"},
						map[string]any{"type": "tool_result", "tool_use_id": "tool_b"},
					},
				},
			},
			want: "[tool_result] id=tool_a\n[tool_result] id=tool_b",
		},
		{
			name: "empty_content_slice",
			ev: map[string]any{
				"message": map[string]any{
					"content": []any{},
				},
			},
			want: "",
		},
		{
			name: "non_tool_result_blocks_skipped",
			ev: map[string]any{
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text", "text": "hello"},
						map[string]any{"type": "input_message", "content": "world"},
					},
				},
			},
			want: "",
		},
		{
			name: "tool_use_id_preserved_full",
			ev: map[string]any{
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "tool_result", "tool_use_id": "tool_12345678901234567890123456789012345"},
					},
				},
			},
			want: "[tool_result] id=tool_12345678901234567890123456789012345",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := formatUserEvent(tc.ev)
			if got != tc.want {
				t.Fatalf("formatUserEvent(%v): got %q, want %q", tc.ev, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// formatStreamEvent additional edge cases
// ---------------------------------------------------------------------------

func TestFormatStreamEvent_ContentBlockStart_MissingEvent(t *testing.T) {
	t.Parallel()
	ev := map[string]any{"type": "stream_event"}
	got := formatStreamJSONEvent(ev, nil)
	if got != "" {
		t.Fatalf("expected empty string for missing event key, got %q", got)
	}
}

func TestFormatStreamEvent_ContentBlockStart_MissingContentBlock(t *testing.T) {
	t.Parallel()
	ev := map[string]any{
		"type": "stream_event",
		"event": map[string]any{
			"type": "content_block_start",
		},
	}
	got := formatStreamJSONEvent(ev, nil)
	if got != "" {
		t.Fatalf("expected empty string when content_block missing, got %q", got)
	}
}

func TestFormatStreamEvent_ContentBlockStart_TextBlock(t *testing.T) {
	t.Parallel()
	ev := map[string]any{
		"type": "stream_event",
		"event": map[string]any{
			"type":          "content_block_start",
			"content_block": map[string]any{"type": "text", "text": "hello"},
		},
	}
	got := formatStreamJSONEvent(ev, nil)
	if got != "" {
		t.Fatalf("expected empty string for text content_block, got %q", got)
	}
}

func TestFormatStreamEvent_ContentBlockStart_ThinkingBlock(t *testing.T) {
	t.Parallel()
	ev := map[string]any{
		"type": "stream_event",
		"event": map[string]any{
			"type":          "content_block_start",
			"content_block": map[string]any{"type": "thinking"},
		},
	}
	got := formatStreamJSONEvent(ev, nil)
	if got != "" {
		t.Fatalf("expected empty string for thinking content_block, got %q", got)
	}
}

func TestFormatStreamEvent_ContentBlockStart_ToolUseWithAccumulatedText(t *testing.T) {
	t.Parallel()
	var textBuf strings.Builder
	textBuf.WriteString("Thinking silently...")

	ev := map[string]any{
		"type": "stream_event",
		"event": map[string]any{
			"type": "content_block_start",
			"content_block": map[string]any{
				"type": "tool_use",
				"name": "Bash",
				"input": map[string]any{
					"command": "echo hi",
				},
			},
		},
	}
	got := formatStreamJSONEvent(ev, &textBuf)
	if !strings.Contains(got, "[agent]") {
		t.Fatalf("expected buffered text to be flushed: %q", got)
	}
	if !strings.Contains(got, "[tool] Bash") {
		t.Fatalf("expected tool line: %q", got)
	}
}

func TestFormatStreamEvent_MessageStop_EmptyBuf(t *testing.T) {
	t.Parallel()
	var textBuf strings.Builder
	ev := map[string]any{
		"type": "stream_event",
		"event": map[string]any{
			"type": "message_stop",
		},
	}
	got := formatStreamJSONEvent(ev, &textBuf)
	if got != "" {
		t.Fatalf("expected empty string when textBuf is empty, got %q", got)
	}
}

func TestFormatStreamEvent_ContentBlockStop_EmptyBuf(t *testing.T) {
	t.Parallel()
	var textBuf strings.Builder
	ev := map[string]any{
		"type": "stream_event",
		"event": map[string]any{
			"type": "content_block_stop",
		},
	}
	got := formatStreamJSONEvent(ev, &textBuf)
	if got != "" {
		t.Fatalf("expected empty string when textBuf is empty, got %q", got)
	}
}

func TestFormatStreamEvent_ContentBlockStop_FlushesText(t *testing.T) {
	t.Parallel()
	var textBuf strings.Builder
	textBuf.WriteString("Final thought before block ends")

	ev := map[string]any{
		"type": "stream_event",
		"event": map[string]any{
			"type": "content_block_stop",
		},
	}
	got := formatStreamJSONEvent(ev, &textBuf)
	if !strings.Contains(got, "[agent] Final thought") {
		t.Fatalf("expected flushed text: %q", got)
	}
	if textBuf.Len() != 0 {
		t.Fatalf("textBuf should be cleared after flush, got %q", textBuf.String())
	}
}

func TestFormatStreamEvent_UnknownStreamType(t *testing.T) {
	t.Parallel()
	ev := map[string]any{
		"type": "stream_event",
		"event": map[string]any{
			"type": "ping",
		},
	}
	got := formatStreamJSONEvent(ev, nil)
	if got != "" {
		t.Fatalf("expected empty string for unknown stream type, got %q", got)
	}
}

func TestFormatStreamEvent_NonTextDelta(t *testing.T) {
	t.Parallel()
	var textBuf strings.Builder
	ev := map[string]any{
		"type": "stream_event",
		"event": map[string]any{
			"delta": map[string]any{"type": "content_block_delta", "text": "hi"},
		},
	}
	got := formatStreamJSONEvent(ev, &textBuf)
	if got != "" {
		t.Fatalf("expected empty for non-text delta, got %q", got)
	}
	if textBuf.Len() != 0 {
		t.Fatalf("textBuf should not be written for non-text delta, got %q", textBuf.String())
	}
}

func TestFormatAssistantEvent_MissingMessage(t *testing.T) {
	t.Parallel()
	ev := map[string]any{"type": "assistant"}
	got := formatStreamJSONEvent(ev, nil)
	if got != "" {
		t.Fatalf("expected empty for missing message, got %q", got)
	}
}

func TestFormatAssistantEvent_MessageNotMap(t *testing.T) {
	t.Parallel()
	ev := map[string]any{"type": "assistant", "message": "not a map"}
	got := formatStreamJSONEvent(ev, nil)
	if got != "" {
		t.Fatalf("expected empty for non-map message, got %q", got)
	}
}

func TestFormatAssistantEvent_ContentNotSlice(t *testing.T) {
	t.Parallel()
	ev := map[string]any{
		"type":    "assistant",
		"message": map[string]any{"content": "not a slice"},
	}
	got := formatStreamJSONEvent(ev, nil)
	if got != "" {
		t.Fatalf("expected empty for non-slice content, got %q", got)
	}
}

func TestFormatAssistantEvent_ToolUseWithEmptyInput(t *testing.T) {
	t.Parallel()
	ev := map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"content": []any{
				map[string]any{
					"type":  "tool_use",
					"name":  "Bash",
					"input": nil,
				},
			},
		},
	}
	got := formatStreamJSONEvent(ev, nil)
	if !strings.Contains(got, "[tool] Bash") {
		t.Fatalf("expected tool line with no hint: %q", got)
	}
}

func TestFormatResultEvent_NilSubtype(t *testing.T) {
	t.Parallel()
	ev := map[string]any{"type": "result"}
	got := formatStreamJSONEvent(ev, nil)
	if !strings.Contains(got, "[result]") {
		t.Fatalf("expected [result] in output, got %q", got)
	}
	if !strings.Contains(got, "done") {
		t.Fatalf("expected 'done' subtype for nil, got %q", got)
	}
}

func TestFormatResultEvent_WithCostAndTurns(t *testing.T) {
	t.Parallel()
	ev := map[string]any{
		"type":           "result",
		"total_cost_usd": 1.23,
		"num_turns":      5,
	}
	got := formatStreamJSONEvent(ev, nil)
	if !strings.Contains(got, "cost=$1.2300") {
		t.Fatalf("expected cost=$1.2300, got %q", got)
	}
	if !strings.Contains(got, "turns=5") {
		t.Fatalf("expected turns=5, got %q", got)
	}
}

func TestFormatResultEvent_IntCost(t *testing.T) {
	t.Parallel()
	ev := map[string]any{
		"type":    "result",
		"cost_usd": 42,
	}
	got := formatStreamJSONEvent(ev, nil)
	if !strings.Contains(got, "cost=$42") {
		t.Fatalf("expected cost=$42 for int, got %q", got)
	}
}

func TestFormatResultEvent_Int64Cost(t *testing.T) {
	t.Parallel()
	ev := map[string]any{
		"type":    "result",
		"cost_usd": int64(99),
	}
	got := formatStreamJSONEvent(ev, nil)
	if !strings.Contains(got, "cost=$99") {
		t.Fatalf("expected cost=$99 for int64, got %q", got)
	}
}

func TestFormatResultEvent_IntTurns(t *testing.T) {
	t.Parallel()
	ev := map[string]any{
		"type":      "result",
		"num_turns": 7,
	}
	got := formatStreamJSONEvent(ev, nil)
	if !strings.Contains(got, "turns=7") {
		t.Fatalf("expected turns=7 for int, got %q", got)
	}
}

func TestFormatResultEvent_Int64Turns(t *testing.T) {
	t.Parallel()
	ev := map[string]any{
		"type":      "result",
		"num_turns": int64(11),
	}
	got := formatStreamJSONEvent(ev, nil)
	if !strings.Contains(got, "turns=11") {
		t.Fatalf("expected turns=11 for int64, got %q", got)
	}
}

func TestFormatSystemEvent_DefaultSubtype(t *testing.T) {
	t.Parallel()
	ev := map[string]any{"type": "system"}
	got := formatStreamJSONEvent(ev, nil)
	if !strings.Contains(got, "[system]") {
		t.Fatalf("expected [system] in output, got %q", got)
	}
}

func TestFormatSystemEvent_UnknownSubtype(t *testing.T) {
	t.Parallel()
	ev := map[string]any{"type": "system", "subtype": "custom"}
	got := formatStreamJSONEvent(ev, nil)
	if !strings.Contains(got, "[system] custom") {
		t.Fatalf("expected [system] custom, got %q", got)
	}
}

func TestFormatUserEvent_ContentBlockNotMap(t *testing.T) {
	t.Parallel()
	ev := map[string]any{
		"message": map[string]any{
			"content": []any{"not a map"},
		},
	}
	got := formatStreamJSONEvent(ev, nil)
	if got != "" {
		t.Fatalf("expected empty for non-map content block, got %q", got)
	}
}

func TestToolInputHint_UnknownKeys(t *testing.T) {
	t.Parallel()
	m := map[string]any{"foo": "bar", "baz": "qux"}
	got := toolInputHint(m)
	if got != "" {
		t.Fatalf("expected empty for unknown keys, got %q", got)
	}
}

func TestToolInputHint_MultipleKeys(t *testing.T) {
	t.Parallel()
	m := map[string]any{
		"file_path": "/path/to/file",
		"pattern":   "*.go",
		"query":    "something",
	}
	got := toolInputHint(m)
	if !strings.HasPrefix(got, "/path/to/file") {
		t.Fatalf("expected file_path hint, got %q", got)
	}
}

func TestToolInputHint_WhitespaceOnly(t *testing.T) {
	t.Parallel()
	m := map[string]any{"command": "   "}
	got := toolInputHint(m)
	if got != "" {
		t.Fatalf("expected empty for whitespace-only command, got %q", got)
	}
}

func TestPickFloat_NoMatch(t *testing.T) {
	t.Parallel()
	_, ok := pickFloat(map[string]any{"foo": "bar"}, "cost", "total")
	if ok {
		t.Fatal("expected false for non-numeric value")
	}
}

func TestPickInt_NoMatch(t *testing.T) {
	t.Parallel()
	_, ok := pickInt(map[string]any{"foo": "bar"}, "turns", "num_turns"
	if ok {
		t.Fatal("expected false for non-numeric value")
	}
}

func TestTruncateRunes_ExactlyMax(t *testing.T) {
	t.Parallel()
	s := "abcdefghij" // 10 chars, exactly max
	got := truncateRunes(s, 10)
	if got != s {
		t.Fatalf("got %q, want %q", got, s)
	}
}

func TestTruncateRunes_SingleRune(t *testing.T) {
	t.Parallel()
	got := truncateRunes("a", 1)
	if got != "a" {
		t.Fatalf("got %q", got)
	}
}

func TestTruncateRunes_ZeroMax(t *testing.T) {
	t.Parallel()
	got := truncateRunes("abc", 0)
	if got != "abc" {
		t.Fatalf("got %q", got)
	}
}

func TestTruncateRunes_NegativeMax(t *testing.T) {
	t.Parallel()
	got := truncateRunes("abc", -1)
	if got != "abc" {
		t.Fatalf("got %q", got)
	}
}

func TestLogStreamWriter_NilWriter(t *testing.T) {
	t.Parallel()
	w := newLogStreamWriter(nil)
	n, err := w.Write([]byte("hello\n"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("n=%d, want 5", n)
	}
}

func TestLogStreamWriter_EmptyWrite(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	w := newLogStreamWriter(&buf)
	n, err := w.Write([]byte{})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("n=%d, want 0", n)
	}
}

func TestLogStreamWriter_MultipleLines(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	w := newLogStreamWriter(&buf)
	_, err := w.Write([]byte(`{"type":"result","subtype":"success"}
{"type":"system","subtype":"init"}
`))
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "[result]") {
		t.Fatalf("expected [result]: %q", out)
	}
	if !strings.Contains(out, "[session]") {
		t.Fatalf("expected [session]: %q", out)
	}
}