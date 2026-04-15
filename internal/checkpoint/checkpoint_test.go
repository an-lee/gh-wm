package checkpoint

import (
	"strings"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()
	c := Checkpoint{
		Branch:       "feat",
		SHA:          "abc123",
		Step:         "done",
		FilesChanged: []string{"a.go"},
		Summary:      "hello",
		Timestamp:    "2020-01-01T00:00:00Z",
	}
	s := Encode(c)
	if !strings.HasPrefix(s, marker) {
		t.Fatalf("expected marker prefix: %q", s)
	}
	parsed, err := ParseLatest([]string{s})
	if err != nil {
		t.Fatal(err)
	}
	if parsed == nil || parsed.Summary != "hello" || parsed.Branch != "feat" {
		t.Fatalf("got %+v", parsed)
	}
}

func TestParseLatest_LastWins(t *testing.T) {
	t.Parallel()
	first := Encode(Checkpoint{Summary: "a"})
	second := Encode(Checkpoint{Summary: "b"})
	last, err := ParseLatest([]string{"x", first, "y", second})
	if err != nil {
		t.Fatal(err)
	}
	if last == nil || last.Summary != "b" {
		t.Fatalf("want last summary b, got %+v", last)
	}
}

func TestParseLatest_InvalidSkipped(t *testing.T) {
	t.Parallel()
	good := Encode(Checkpoint{Summary: "ok"})
	last, err := ParseLatest([]string{"<!-- wm-checkpoint: not json -->", good})
	if err != nil {
		t.Fatal(err)
	}
	if last == nil || last.Summary != "ok" {
		t.Fatalf("got %+v", last)
	}
}

func TestParseLatest_NoMarker(t *testing.T) {
	t.Parallel()
	last, err := ParseLatest([]string{"plain text", "no checkpoint"})
	if err != nil {
		t.Fatal(err)
	}
	if last != nil {
		t.Fatal("expected nil")
	}
}

func TestParseLatest_UnclosedMarker(t *testing.T) {
	t.Parallel()
	last, err := ParseLatest([]string{"<!-- wm-checkpoint: {\"foo\":1} "})
	if err != nil {
		t.Fatal(err)
	}
	if last != nil {
		t.Fatal("expected nil")
	}
}
