// Package retry provides a simple, configurable mechanism to retry operations
// with exponential backoff and optional jitter.
package retry

import (
	"context"
	"math/rand"
	"time"
)

// Config defines the settings for retrying an operation.
type Config struct {
	// Attempts specifies the maximum number of attempts (including the initial try).
	Attempts int

	// InitialDelay is the duration to wait before the second attempt.
	InitialDelay time.Duration

	// MaxDelay is the maximum delay allowed between attempts.
	// If zero, no upper limit is applied.
	MaxDelay time.Duration

	// Factor is the multiplier used to increase the delay after each attempt.
	// For example, a factor of 2 will double the delay each time.
	Factor float64

	// Jitter, if true, adds randomness to the delay (up to the current delay duration)
	// to help prevent thundering herd issues.
	Jitter bool
}

// Do executes the provided operation function and retries it on error
// according to the provided configuration. It stops retrying if:
//   - The operation returns nil (success),
//   - The maximum number of attempts is reached,
//   - The provided context is cancelled.
//
// Do returns nil on success or the last error encountered.
func Do(ctx context.Context, op func() error, cfg Config) error {
	// Ensure at least one attempt.
	if cfg.Attempts < 1 {
		cfg.Attempts = 1
	}

	// Use the initial delay for the first retry (if needed).
	delay := cfg.InitialDelay
	var err error

	for attempt := 1; attempt <= cfg.Attempts; attempt++ {
		// Attempt the operation.
		if err = op(); err == nil {
			return nil
		}

		// If this was the final attempt, break out of the loop.
		if attempt == cfg.Attempts {
			break
		}

		// Determine how long to wait before the next attempt.
		sleepDuration := delay
		if cfg.Jitter {
			// Apply jitter: randomize the delay between 0 and the calculated delay.
			sleepDuration = time.Duration(rand.Int63n(int64(delay)))
		}

		// Respect the maximum delay if it's set.
		if cfg.MaxDelay > 0 && sleepDuration > cfg.MaxDelay {
			sleepDuration = cfg.MaxDelay
		}

		// Wait for either the delay period or context cancellation.
		select {
		case <-time.After(sleepDuration):
			// Increase delay for the next attempt using the exponential backoff factor.
			delay = time.Duration(float64(delay) * cfg.Factor)
			if cfg.MaxDelay > 0 && delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Return the last error if all attempts fail.
	return err
}
