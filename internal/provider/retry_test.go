// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	control "github.com/ably/ably-control-go"
)

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "status code 0",
			err:      control.ErrorInfo{StatusCode: 0},
			expected: true,
		},
		{
			name:     "status code 429 (rate limit)",
			err:      control.ErrorInfo{StatusCode: 429},
			expected: true,
		},
		{
			name:     "status code 500 (server error)",
			err:      control.ErrorInfo{StatusCode: 500},
			expected: true,
		},
		{
			name:     "status code 503 (service unavailable)",
			err:      control.ErrorInfo{StatusCode: 503},
			expected: true,
		},
		{
			name:     "status code 404 (not found)",
			err:      control.ErrorInfo{StatusCode: 404},
			expected: false,
		},
		{
			name:     "status code 400 (bad request)",
			err:      control.ErrorInfo{StatusCode: 400},
			expected: false,
		},
		{
			name:     "url error",
			err:      &url.Error{Op: "Get", URL: "http://example.com", Err: errors.New("connection refused")},
			expected: true,
		},
		{
			name:     "net timeout error",
			err:      &netTimeoutError{},
			expected: true,
		},
		{
			name:     "generic error",
			err:      errors.New("something went wrong"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isRetryableError(%v) = %v; want %v", tt.err, result, tt.expected)
			}
		})
	}
}

// netTimeoutError is a test helper that implements net.Error
type netTimeoutError struct{}

func (e *netTimeoutError) Error() string   { return "timeout" }
func (e *netTimeoutError) Timeout() bool   { return true }
func (e *netTimeoutError) Temporary() bool { return true }

var _ net.Error = (*netTimeoutError)(nil)

func TestRetryWithBackoff(t *testing.T) {
	t.Run("succeeds on first attempt", func(t *testing.T) {
		ctx := context.Background()
		callCount := 0

		err := retryWithBackoff(ctx, "test", func() error {
			callCount++
			return nil
		})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if callCount != 1 {
			t.Errorf("expected 1 call, got %d", callCount)
		}
	})

	t.Run("retries on retryable error", func(t *testing.T) {
		ctx := context.Background()
		callCount := 0

		err := retryWithBackoff(ctx, "test", func() error {
			callCount++
			if callCount < 3 {
				return control.ErrorInfo{StatusCode: 0} // Connection error
			}
			return nil
		})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if callCount != 3 {
			t.Errorf("expected 3 calls, got %d", callCount)
		}
	})

	t.Run("does not retry on non-retryable error", func(t *testing.T) {
		ctx := context.Background()
		callCount := 0

		err := retryWithBackoff(ctx, "test", func() error {
			callCount++
			return control.ErrorInfo{StatusCode: 404}
		})

		if err == nil {
			t.Error("expected error, got nil")
		}
		if callCount != 1 {
			t.Errorf("expected 1 call, got %d", callCount)
		}
	})

	t.Run("respects max retries", func(t *testing.T) {
		ctx := context.Background()
		callCount := 0

		err := retryWithBackoff(ctx, "test", func() error {
			callCount++
			return control.ErrorInfo{StatusCode: 503} // Service unavailable
		})

		if err == nil {
			t.Error("expected error, got nil")
		}
		// Should be called 4 times: initial attempt + 3 retries
		if callCount != 4 {
			t.Errorf("expected 4 calls (initial + 3 retries), got %d", callCount)
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		callCount := 0

		err := retryWithBackoff(ctx, "test", func() error {
			callCount++
			// Cancel context after first call to deterministically test cancellation
			if callCount == 1 {
				cancel()
			}
			return control.ErrorInfo{StatusCode: 500} // Server error
		})

		if err == nil {
			t.Error("expected error, got nil")
		}
		// Should be called once, then cancelled - not the full 4 times
		if callCount >= 4 {
			t.Errorf("expected fewer than 4 calls due to context cancellation, got %d", callCount)
		}
		// Verify we got a context error
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
	})
}

func TestBackoff(t *testing.T) {
	// With equal jitter, backoff returns 50-100% of the base value
	// Base values: 1s, 4s, 16s, 64s (capped at 30s)
	tests := []struct {
		attempt int
		min     time.Duration
		max     time.Duration
	}{
		{0, 500 * time.Millisecond, 1 * time.Second}, // base 1s
		{1, 2 * time.Second, 4 * time.Second},        // base 4s
		{2, 8 * time.Second, 16 * time.Second},       // base 16s
		{3, 15 * time.Second, 30 * time.Second},      // base 64s, capped at 30s
		{10, 15 * time.Second, 30 * time.Second},     // capped at 30s
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			// Run multiple times to account for randomness
			for i := 0; i < 100; i++ {
				result := backoff(tt.attempt)
				if result < tt.min || result > tt.max {
					t.Errorf("backoff(%d) = %v; want between %v and %v", tt.attempt, result, tt.min, tt.max)
				}
			}
		})
	}
}
