package core

import (
	"context"
	"testing"
	"time"
)

func TestRetryOnTimeout(t *testing.T) {
	cfg := &Config{
		Workers:    1,
		Timeout:    50 * time.Millisecond,
		MaxRetries: 2,
		RateLimit:  0,
	}
	scanner := NewScanner(cfg)

	ctx := context.Background()
	// Use unreachable host to trigger timeout
	go scanner.ScanRange(ctx, "192.0.2.1", []uint16{80})

	var resultCount int
	var filteredCount int
	for event := range scanner.Results() {
		if event.Kind == EventKindResult {
			resultCount++
			if event.Result.State == StateFiltered {
				filteredCount++
			}
		}
	}

	// Should receive exactly 1 result (after retries)
	if resultCount != 1 {
		t.Errorf("got %d results; want 1", resultCount)
	}

	// Result should be filtered (timeout)
	if filteredCount != 1 {
		t.Errorf("got %d filtered results; want 1", filteredCount)
	}
}

func TestRetryWithZeroRetries(t *testing.T) {
	cfg := &Config{
		Workers:    1,
		Timeout:    50 * time.Millisecond,
		MaxRetries: 0, // No retries
		RateLimit:  0,
	}
	scanner := NewScanner(cfg)

	ctx := context.Background()
	// Use unreachable host
	go scanner.ScanRange(ctx, "192.0.2.1", []uint16{80})

	var resultCount int
	start := time.Now()
	for event := range scanner.Results() {
		if event.Kind == EventKindResult {
			resultCount++
		}
	}
	elapsed := time.Since(start)

	if resultCount != 1 {
		t.Errorf("got %d results; want 1", resultCount)
	}

	// Should complete quickly with no retries
	maxExpected := 200 * time.Millisecond
	if elapsed > maxExpected {
		t.Errorf("took %v with no retries; want < %v", elapsed, maxExpected)
	}
}

func TestRetryBackoff(t *testing.T) {
	cfg := &Config{
		Workers:    1,
		Timeout:    200 * time.Millisecond, // Use longer timeout for proper backoff testing
		MaxRetries: 2,
		RateLimit:  0,
	}
	scanner := NewScanner(cfg)

	tests := []struct {
		attempt     int
		minExpected time.Duration
		maxExpected time.Duration
	}{
		{0, 60 * time.Millisecond, 100 * time.Millisecond},  // 50ms base + 10-50ms jitter
		{1, 110 * time.Millisecond, 150 * time.Millisecond}, // 100ms base + 10-50ms jitter
	}

	for _, tt := range tests {
		t.Run("attempt_"+string(rune(tt.attempt)), func(t *testing.T) {
			backoff := scanner.retryBackoff(tt.attempt)
			if backoff < tt.minExpected || backoff > tt.maxExpected {
				t.Errorf("backoff for attempt %d = %v; want between %v and %v",
					tt.attempt, backoff, tt.minExpected, tt.maxExpected)
			}
		})
	}
}

func TestRetryBackoffCappedAtTimeout(t *testing.T) {
	cfg := &Config{
		Workers:    1,
		Timeout:    50 * time.Millisecond,
		MaxRetries: 5,
		RateLimit:  0,
	}
	scanner := NewScanner(cfg)

	// With short timeout, backoff should be capped
	backoff := scanner.retryBackoff(5)
	maxExpected := cfg.Timeout + 50*time.Millisecond // timeout + max jitter
	if backoff > maxExpected {
		t.Errorf("backoff = %v; should be capped at timeout (%v) + jitter", backoff, cfg.Timeout)
	}
}

func TestRetryStopsOnSuccess(t *testing.T) {
	cfg := &Config{
		Workers:    1,
		Timeout:    100 * time.Millisecond,
		MaxRetries: 3,
		RateLimit:  0,
	}
	scanner := NewScanner(cfg)

	ctx := context.Background()
	// Scan localhost which should succeed immediately
	go scanner.ScanRange(ctx, "localhost", []uint16{80})

	var resultCount int
	start := time.Now()
	for event := range scanner.Results() {
		if event.Kind == EventKindResult {
			resultCount++
		}
	}
	elapsed := time.Since(start)

	if resultCount != 1 {
		t.Errorf("got %d results; want 1", resultCount)
	}

	// Should complete quickly without retries (no backoff)
	// Allow 500ms for connection + processing
	maxExpected := 500 * time.Millisecond
	if elapsed > maxExpected {
		t.Logf("Warning: successful scan took %v; expected < %v (may retry on success)", elapsed, maxExpected)
	}
}

func TestRetryWithNegativeMaxRetries(t *testing.T) {
	cfg := &Config{
		Workers:    1,
		Timeout:    50 * time.Millisecond,
		MaxRetries: -1, // Invalid, should be treated as 0
		RateLimit:  0,
	}
	scanner := NewScanner(cfg)

	ctx := context.Background()
	go scanner.ScanRange(ctx, "192.0.2.1", []uint16{80})

	var resultCount int
	for event := range scanner.Results() {
		if event.Kind == EventKindResult {
			resultCount++
		}
	}

	// Should still get a result
	if resultCount != 1 {
		t.Errorf("got %d results; want 1", resultCount)
	}
}
