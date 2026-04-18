package retry

import (
	"errors"
	"io"
	"testing"
)

func TestIsTransientError(t *testing.T) {
	t.Parallel()
	if !IsTransientError(errors.New(`Get "https://api.github.com/foo": EOF`), "") {
		t.Fatal("expected EOF to be transient")
	}
	if !IsTransientError(io.EOF, "") {
		t.Fatal("expected io.EOF to be transient")
	}
	if IsTransientError(errors.New("HTTP 404: Not Found"), "") {
		t.Fatal("404 message should not be transient by default")
	}
	if !IsTransientHTTPStatus(503) {
		t.Fatal("503 should be transient HTTP")
	}
	if IsTransientHTTPStatus(404) {
		t.Fatal("404 should not be transient HTTP")
	}
}

func TestWithAttempts_NonTransientStops(t *testing.T) {
	t.Parallel()
	n := 0
	err := WithAttempts(t.Context(), 3, func() error {
		n++
		return errors.New("HTTP 404: not found")
	})
	if err == nil || n != 1 {
		t.Fatalf("got n=%d err=%v", n, err)
	}
}
