// Package retry provides small helpers for retrying transient network/API failures.
package retry

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"time"
)

// IsTransientError reports whether err or supplemental stderr/output text suggests a retryable failure.
func IsTransientError(err error, stderr string) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	s := strings.ToLower(err.Error() + " " + stderr)
	switch {
	case strings.Contains(s, "eof"),
		strings.Contains(s, "connection reset"),
		strings.Contains(s, "broken pipe"),
		strings.Contains(s, "connection refused"),
		strings.Contains(s, "tls"),
		strings.Contains(s, "i/o timeout"),
		strings.Contains(s, "no such host"),
		strings.Contains(s, "server closed"),
		strings.Contains(s, "429"),
		strings.Contains(s, "500"),
		strings.Contains(s, "503"),
		strings.Contains(s, "502"),
		strings.Contains(s, "504"):
		return true
	default:
		return false
	}
}

// IsTransientHTTPStatus reports whether an HTTP status code should be retried.
func IsTransientHTTPStatus(code int) bool {
	switch code {
	case 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

// WithAttempts runs fn up to attempts times. On transient failure, waits 100ms before the 2nd try and 300ms before the 3rd.
// If fn returns nil, WithAttempts returns nil. Non-transient errors return immediately.
// If all attempts fail, returns the last error.
func WithAttempts(ctx context.Context, attempts int, fn func() error) error {
	if attempts < 1 {
		attempts = 1
	}
	var last error
	for attempt := 0; attempt < attempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		last = err
		if !IsTransientError(err, "") {
			return err
		}
		if attempt == attempts-1 {
			return last
		}
		var d time.Duration
		switch attempt {
		case 0:
			d = 100 * time.Millisecond
		case 1:
			d = 300 * time.Millisecond
		default:
			d = 300 * time.Millisecond
		}
		t := time.NewTimer(d)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
	}
	return last
}
