package ui

import (
	"fmt"
	"math"
	"time"
)

// PerformanceTrend indicates the current performance trend
type PerformanceTrend int

const (
	TrendStable PerformanceTrend = iota
	TrendImproving
	TrendDegrading
)

// ProgressTracker tracks scan progress and calculates ETA
type ProgressTracker struct {
	TotalPorts     int
	ScannedPorts   int
	OpenPorts      int
	ClosedPorts    int
	FilteredPorts  int
	StartTime      time.Time
	LastUpdate     time.Time
	CurrentRate    float64
	AverageRate    float64
	IsPaused       bool
	PausedDuration time.Duration
	pauseStart     time.Time
	
	// Enhanced metrics for breadcrumb
	TotalHosts       int
	ScannedHosts     int
	PreviousRate     float64
	PerformanceTrend PerformanceTrend
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(totalPorts int) *ProgressTracker {
	now := time.Now()
	return &ProgressTracker{
		TotalPorts: totalPorts,
		StartTime:  now,
		LastUpdate: now,
	}
}

// Update updates the progress with new scan results
func (p *ProgressTracker) Update(scanned, open, closed, filtered int, currentRate float64) {
	p.ScannedPorts = scanned
	p.OpenPorts = open
	p.ClosedPorts = closed
	p.FilteredPorts = filtered
	p.PreviousRate = p.CurrentRate
	p.CurrentRate = currentRate
	p.LastUpdate = time.Now()

	// Calculate performance trend
	p.calculatePerformanceTrend()

	// Calculate average rate
	elapsed := p.GetActiveTime()
	if elapsed.Seconds() > 0 && p.ScannedPorts > 0 {
		p.AverageRate = float64(p.ScannedPorts) / elapsed.Seconds()
	}
}

// UpdateHosts updates the host tracking metrics
func (p *ProgressTracker) UpdateHosts(totalHosts, scannedHosts int) {
	p.TotalHosts = totalHosts
	p.ScannedHosts = scannedHosts
}

// calculatePerformanceTrend calculates the current performance trend
func (p *ProgressTracker) calculatePerformanceTrend() {
	// Calculate relative change
	if p.PreviousRate > 0 && p.CurrentRate > 0 {
		change := (p.CurrentRate - p.PreviousRate) / p.PreviousRate
		// Consider significant change if > 5%
		if math.Abs(change) > 0.05 {
			if change > 0 {
				p.PerformanceTrend = TrendImproving
			} else {
				p.PerformanceTrend = TrendDegrading
			}
		} else {
			p.PerformanceTrend = TrendStable
		}
	} else {
		p.PerformanceTrend = TrendStable
	}
}

// GetProgress returns the completion percentage
func (p *ProgressTracker) GetProgress() float64 {
	if p.TotalPorts == 0 {
		return 0
	}
	return float64(p.ScannedPorts) / float64(p.TotalPorts) * 100
}

// GetETA calculates the estimated time of arrival
func (p *ProgressTracker) GetETA() time.Duration {
	if p.IsPaused || p.AverageRate <= 0 {
		return 0
	}

	remaining := p.TotalPorts - p.ScannedPorts
	if remaining <= 0 {
		return 0
	}

	secondsRemaining := float64(remaining) / p.AverageRate
	return time.Duration(secondsRemaining * float64(time.Second))
}

// GetActiveTime returns the time spent actively scanning (excluding pauses)
func (p *ProgressTracker) GetActiveTime() time.Duration {
	total := time.Since(p.StartTime)
	if p.IsPaused {
		// Currently paused, add current pause duration
		total -= time.Since(p.pauseStart)
	}
	return total - p.PausedDuration
}

// Pause pauses the progress tracking
func (p *ProgressTracker) Pause() {
	if !p.IsPaused {
		p.IsPaused = true
		p.pauseStart = time.Now()
	}
}

// Resume resumes the progress tracking
func (p *ProgressTracker) Resume() {
	if p.IsPaused {
		p.PausedDuration += time.Since(p.pauseStart)
		p.IsPaused = false
	}
}

// GetStatusLine returns a formatted status line
func (p *ProgressTracker) GetStatusLine() string {
	progress := p.GetProgress()
	eta := p.GetETA()
	activeTime := p.GetActiveTime()

	status := fmt.Sprintf(
		"Progress: %.1f%% (%d/%d) | ETA: %s | Speed: %.0f pps | Time: %s",
		progress,
		p.ScannedPorts,
		p.TotalPorts,
		formatDuration(eta),
		p.CurrentRate,
		formatDuration(activeTime),
	)

	if p.IsPaused {
		status = "⏸ PAUSED | " + status
	}

	return status
}

// GetDetailedStats returns detailed statistics
func (p *ProgressTracker) GetDetailedStats() string {
	return fmt.Sprintf(
		"Open: %d | Closed: %d | Filtered: %d | Avg Speed: %.0f pps",
		p.OpenPorts,
		p.ClosedPorts,
		p.FilteredPorts,
		p.AverageRate,
	)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "--:--"
	}

	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// GetElapsedDuration returns formatted elapsed time
func (p *ProgressTracker) GetElapsedDuration() string {
	return formatDuration(p.GetActiveTime())
}

// GetPerformanceIndicator returns the performance trend indicator
func (p *ProgressTracker) GetPerformanceIndicator() string {
	switch p.PerformanceTrend {
	case TrendImproving:
		return "↑"
	case TrendDegrading:
		return "↓"
	default:
		return "→"
	}
}

// GetFormattedRate returns current rate with formatting
func (p *ProgressTracker) GetFormattedRate() string {
	if p.CurrentRate >= 1000 {
		return fmt.Sprintf("%.0fK", p.CurrentRate/1000)
	}
	return fmt.Sprintf("%.0f", p.CurrentRate)
}

// GetHostProgress returns host scanning progress string
func (p *ProgressTracker) GetHostProgress() string {
	if p.TotalHosts > 0 {
		return fmt.Sprintf("%d/%d", p.ScannedHosts, p.TotalHosts)
	}
	return fmt.Sprintf("%d", p.ScannedHosts)
}

// GetDetailedETA returns formatted ETA with "remaining" suffix
func (p *ProgressTracker) GetDetailedETA() string {
	eta := p.GetETA()
	if eta <= 0 {
		return "complete"
	}
	return formatDuration(eta) + " remaining"
}
