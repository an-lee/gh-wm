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
