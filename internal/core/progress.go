package core

import (
	"context"
	"math"
	"sync/atomic"
	"time"
)

// ProgressReporter handles progress reporting for scanners.
type ProgressReporter struct {
	completed atomic.Uint64
	results   chan<- Event
}

// NewProgressReporter creates a new progress reporter.
func NewProgressReporter(results chan<- Event) *ProgressReporter {
	return &ProgressReporter{
		results: results,
	}
}

// IncrementCompleted atomically increments the completed counter.
func (p *ProgressReporter) IncrementCompleted() {
	p.completed.Add(1)
}

// GetCompleted returns the current completed count.
func (p *ProgressReporter) GetCompleted() uint64 {
	return p.completed.Load()
}

// SetCompleted sets the completed count (used for initialization).
func (p *ProgressReporter) SetCompleted(val uint64) {
	p.completed.Store(val)
}

// StartReporting starts the progress reporter in a background goroutine.
// Returns a channel that will be closed when reporting is complete.
func (p *ProgressReporter) StartReporting(ctx context.Context, total int) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		p.reportProgress(ctx, total)
		close(done)
	}()
	return done
}

// reportProgress periodically emits progress events until scanning is complete or context is cancelled.
func (p *ProgressReporter) reportProgress(ctx context.Context, total int) {
	ticker := time.NewTicker(ProgressReportInterval)
	defer ticker.Stop()
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			completedUint := p.completed.Load()
			// Safely convert to int with proper bounds checking
			var completed int
			if completedUint > math.MaxInt32 {
				completed = math.MaxInt32
			} else {
				completed = int(completedUint) // #nosec G115 - safe after bounds check
			}
			if completed > total {
				completed = total
			}
			elapsed := time.Since(startTime).Seconds()
			if elapsed <= 0 {
				elapsed = 0.001
			}
			rate := float64(completed) / elapsed

			progress := ProgressEvent{Total: total, Completed: completed, Rate: rate}
			select {
			case p.results <- NewProgressEvent(progress):
			case <-ctx.Done():
				return
			}

			if completed >= total {
				return
			}
		}
	}
}
