// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"errors"
	"math/rand/v2"
	"net"
	"net/url"
	"time"

	control "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	maxRetries     = 3
	initialBackoff = 1 * time.Second
	maxBackoff     = 30 * time.Second
)

// isRetryableError reports whether err is transient and should be retried.
// It returns true for network or URL-level errors (types satisfying net.Error or *url.Error),
// and for Ably control API errors (control.ErrorInfo) with StatusCode 0, 429, or any 5xx.
// For all other errors, including 4xx Ably errors other than 429, it returns false.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// Check for URL errors (connection failures)
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}

	// Check for Ably control API errors
	var controlErr control.ErrorInfo
	if errors.As(err, &controlErr) {
		// Retry on:
		// - Status code 0 (connection error)
		// - Status code 429 (rate limit)
		// - Status codes 5xx (server errors)
		// Don't retry on 4xx client errors (except 429)
		if controlErr.StatusCode == 0 || controlErr.StatusCode == 429 {
			return true
		}
		if controlErr.StatusCode >= 500 && controlErr.StatusCode < 600 {
			return true
		}
		return false
	}

	return false
}

// backoff returns the backoff duration for a given attempt (0-indexed)
// between 50% and 100% of that capped backoff.
func backoff(attempt int) time.Duration {
	b := initialBackoff << (attempt * 2) // equivalent to initialBackoff * 4^attempt
	b = min(b, maxBackoff)
	return b/2 + rand.N(b/2) // Equal jitter: 50-100% of calculated backoff
}

// retryWithBackoff retries the provided operation using exponential backoff until it succeeds,
// a non-retryable error occurs, the context is cancelled, or the maximum number of retries is reached.
// 
// The function invokes fn repeatedly and returns nil on success. If fn returns a non-retryable error
// the error is returned immediately. If retries are exhausted, the last error returned by fn is returned.
// The call respects ctx cancellation and will return ctx.Err() if the context is cancelled.
func retryWithBackoff(ctx context.Context, operation string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			wait := backoff(attempt - 1)
			tflog.Warn(ctx, "Retrying operation after error",
				map[string]interface{}{
					"operation": operation,
					"attempt":   attempt,
					"backoff":   wait.String(),
					"error":     lastErr.Error(),
				})

			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		lastErr = fn()

		if lastErr == nil {
			if attempt > 0 {
				tflog.Info(ctx, "Operation succeeded after retry",
					map[string]interface{}{
						"operation": operation,
						"attempt":   attempt,
					})
			}
			return nil
		}

		if !isRetryableError(lastErr) {
			tflog.Debug(ctx, "Error is not retryable",
				map[string]interface{}{
					"operation": operation,
					"error":     lastErr.Error(),
				})
			return lastErr
		}
	}

	tflog.Error(ctx, "Operation failed after max retries",
		map[string]interface{}{
			"operation":   operation,
			"max_retries": maxRetries,
			"error":       lastErr.Error(),
		})

	return lastErr
}