package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	tests := []struct {
		name        string
		cfg         Config
		attempts    int
		shouldError bool
	}{
		{
			name: "success_first_try",
			cfg: Config{
				Attempts:     3,
				InitialDelay: 10 * time.Millisecond,
				Factor:       2,
			},
			attempts:    1,
			shouldError: false,
		},
		{
			name: "success_after_retry",
			cfg: Config{
				Attempts:     3,
				InitialDelay: 10 * time.Millisecond,
				Factor:       2,
			},
			attempts:    2,
			shouldError: false,
		},
		{
			name: "failure_all_attempts",
			cfg: Config{
				Attempts:     3,
				InitialDelay: 10 * time.Millisecond,
				Factor:       2,
			},
			attempts:    3,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			op := func() error {
				count++
				if count < tt.attempts {
					return errors.New("temporary error")
				}
				if tt.shouldError {
					return errors.New("permanent error")
				}
				return nil
			}

			err := Do(context.Background(), op, tt.cfg)
			if tt.shouldError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
			if count != tt.attempts {
				t.Errorf("expected %d attempts, got %d", tt.attempts, count)
			}
		})
	}
}

func TestDoContext(t *testing.T) {
	cfg := Config{
		Attempts:     5,
		InitialDelay: 100 * time.Millisecond,
		Factor:       2,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	count := 0
	err := Do(ctx, func() error {
		count++
		return errors.New("temporary error")
	}, cfg)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded error, got: %v", err)
	}
	if count > 2 { // Given the timing, we expect 1-2 attempts before timeout
		t.Errorf("expected 1-2 attempts before context timeout, got: %d", count)
	}
}

func TestJitter(t *testing.T) {
	cfg := Config{
		Attempts:     3,
		InitialDelay: 20 * time.Millisecond,
		Factor:       2,
		Jitter:       true,
	}

	delays := make([]time.Duration, 0, 2)
	var lastTime time.Time

	err := Do(context.Background(), func() error {
		now := time.Now()
		if !lastTime.IsZero() {
			delays = append(delays, now.Sub(lastTime))
		}
		lastTime = now
		return errors.New("temporary error")
	}, cfg)

	if err == nil {
		t.Error("expected error but got nil")
	}

	// With jitter enabled, delays should be different
	if len(delays) == 2 {
		if delays[0] == delays[1] {
			t.Error("expected different delays with jitter enabled")
		}
	}
}

func TestMaxDelay(t *testing.T) {
	cfg := Config{
		Attempts:     4,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     15 * time.Millisecond,
		Factor:       2,
	}

	var lastTime time.Time
	var maxObservedDelay time.Duration

	err := Do(context.Background(), func() error {
		now := time.Now()
		if !lastTime.IsZero() {
			delay := now.Sub(lastTime)
			if delay > maxObservedDelay {
				maxObservedDelay = delay
			}
		}
		lastTime = now
		return errors.New("temporary error")
	}, cfg)

	if err == nil {
		t.Error("expected error but got nil")
	}

	// Allow for some timing variation but ensure we're not exceeding max delay by too much
	maxAllowedDelay := cfg.MaxDelay + (10 * time.Millisecond)
	if maxObservedDelay > maxAllowedDelay {
		t.Errorf("delay exceeded maximum: got %v, want <= %v", maxObservedDelay, maxAllowedDelay)
	}
}

func TestZeroAttempts(t *testing.T) {
	cfg := Config{
		Attempts:     0, // Should be corrected to 1
		InitialDelay: 10 * time.Millisecond,
	}

	count := 0
	err := Do(context.Background(), func() error {
		count++
		return nil
	}, cfg)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if count != 1 {
		t.Errorf("expected exactly 1 attempt, got: %d", count)
	}
}
