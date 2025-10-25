package ui

import (
	"testing"
	"time"
)

func TestNewProgressTracker(t *testing.T) {
	totalPorts := 1000
	tracker := NewProgressTracker(totalPorts)

	if tracker.TotalPorts != totalPorts {
		t.Errorf("expected TotalPorts %d, got %d", totalPorts, tracker.TotalPorts)
	}

	if tracker.ScannedPorts != 0 {
		t.Errorf("expected ScannedPorts 0, got %d", tracker.ScannedPorts)
	}

	if tracker.OpenPorts != 0 || tracker.ClosedPorts != 0 || tracker.FilteredPorts != 0 {
		t.Error("expected all counts to be zero initially")
	}

	if tracker.IsPaused {
		t.Error("expected tracker to not be paused initially")
	}
}

func TestProgressTracker_Update(t *testing.T) {
	tracker := NewProgressTracker(1000)

	// Update with scan progress
	tracker.Update(100, 10, 80, 10, 500.0)

	if tracker.ScannedPorts != 100 {
		t.Errorf("expected ScannedPorts 100, got %d", tracker.ScannedPorts)
	}

	if tracker.OpenPorts != 10 {
		t.Errorf("expected OpenPorts 10, got %d", tracker.OpenPorts)
	}

	if tracker.ClosedPorts != 80 {
		t.Errorf("expected ClosedPorts 80, got %d", tracker.ClosedPorts)
	}

	if tracker.FilteredPorts != 10 {
		t.Errorf("expected FilteredPorts 10, got %d", tracker.FilteredPorts)
	}

	if tracker.CurrentRate != 500.0 {
		t.Errorf("expected CurrentRate 500.0, got %f", tracker.CurrentRate)
	}
}

func TestProgressTracker_GetProgress(t *testing.T) {
	tracker := NewProgressTracker(1000)

	// 0% complete
	if tracker.GetProgress() != 0.0 {
		t.Errorf("expected 0%% initially, got %f%%", tracker.GetProgress())
	}

	// 50% complete
	tracker.Update(500, 10, 480, 10, 500.0)
	expected := 50.0
	if tracker.GetProgress() != expected {
		t.Errorf("expected %f%%, got %f%%", expected, tracker.GetProgress())
	}

	// 100% complete
	tracker.Update(1000, 20, 960, 20, 500.0)
	expected = 100.0
	if tracker.GetProgress() != expected {
		t.Errorf("expected %f%%, got %f%%", expected, tracker.GetProgress())
	}
}

func TestProgressTracker_Pause(t *testing.T) {
	tracker := NewProgressTracker(1000)

	if tracker.IsPaused {
		t.Error("expected tracker to not be paused initially")
	}

	tracker.Pause()

	if !tracker.IsPaused {
		t.Error("expected tracker to be paused after Pause()")
	}

	// Pause again should be idempotent
	tracker.Pause()

	if !tracker.IsPaused {
		t.Error("expected tracker to still be paused")
	}
}

func TestProgressTracker_Resume(t *testing.T) {
	tracker := NewProgressTracker(1000)
	tracker.Pause()

	if !tracker.IsPaused {
		t.Error("expected tracker to be paused")
	}

	tracker.Resume()

	if tracker.IsPaused {
		t.Error("expected tracker to not be paused after Resume()")
	}

	// Resume again should be idempotent
	tracker.Resume()

	if tracker.IsPaused {
		t.Error("expected tracker to still be resumed")
	}
}

func TestProgressTracker_GetActiveTime(t *testing.T) {
	tracker := NewProgressTracker(1000)
	startTime := time.Now().Add(-5 * time.Second)
	tracker.StartTime = startTime

	elapsed := tracker.GetActiveTime()

	// Should be approximately 5 seconds (with some tolerance)
	if elapsed < 4*time.Second || elapsed > 6*time.Second {
		t.Errorf("expected elapsed time around 5s, got %v", elapsed)
	}
}

func TestProgressTracker_GetETA(t *testing.T) {
	tracker := NewProgressTracker(1000)
	tracker.StartTime = time.Now().Add(-10 * time.Second)

	// 50% complete, set average rate manually
	tracker.Update(500, 10, 480, 10, 50.0)
	tracker.AverageRate = 50.0

	eta := tracker.GetETA()

	// ETA should be approximately 10 seconds (500 remaining / 50 per sec)
	if eta < 5*time.Second || eta > 15*time.Second {
		t.Errorf("expected ETA around 10s, got %v", eta)
	}
}

func TestProgressTracker_GetETA_Complete(t *testing.T) {
	tracker := NewProgressTracker(1000)
	tracker.Update(1000, 20, 960, 20, 500.0)

	eta := tracker.GetETA()

	// When complete, ETA should be 0
	if eta != 0 {
		t.Errorf("expected ETA 0 when complete, got %v", eta)
	}
}

func TestGetStatusLine_NotPaused(t *testing.T) {
	tracker := NewProgressTracker(1000)
	tracker.Update(500, 10, 480, 10, 1000.0)

	status := tracker.GetStatusLine()

	if status == "" {
		t.Error("GetStatusLine returned empty string")
	}

	// Should contain key information
	if !contains(status, "50.0%") {
		t.Errorf("status line should contain progress percentage: %s", status)
	}

	if !contains(status, "500/1000") {
		t.Errorf("status line should contain scanned/total: %s", status)
	}

	if !contains(status, "pps") {
		t.Errorf("status line should contain speed: %s", status)
	}
}

func TestGetStatusLine_Paused(t *testing.T) {
	tracker := NewProgressTracker(1000)
	tracker.Update(500, 10, 480, 10, 1000.0)
	tracker.Pause()

	status := tracker.GetStatusLine()

	if !contains(status, "PAUSED") {
		t.Errorf("status line should indicate paused state: %s", status)
	}
}

func TestGetStatusLine_Complete(t *testing.T) {
	tracker := NewProgressTracker(100)
	tracker.Update(100, 10, 85, 5, 500.0)

	status := tracker.GetStatusLine()

	if !contains(status, "100.0%") {
		t.Errorf("status line should show 100%% completion: %s", status)
	}
}

func TestGetDetailedStats(t *testing.T) {
	tracker := NewProgressTracker(1000)
	tracker.Update(500, 25, 450, 25, 1000.0)

	stats := tracker.GetDetailedStats()

	if stats == "" {
		t.Error("GetDetailedStats returned empty string")
	}

	// Should contain counts
	if !contains(stats, "Open: 25") {
		t.Errorf("stats should contain open count: %s", stats)
	}

	if !contains(stats, "Closed: 450") {
		t.Errorf("stats should contain closed count: %s", stats)
	}

	if !contains(stats, "Filtered: 25") {
		t.Errorf("stats should contain filtered count: %s", stats)
	}

	if !contains(stats, "pps") {
		t.Errorf("stats should contain average speed: %s", stats)
	}
}

func TestGetDetailedStats_NoActivity(t *testing.T) {
	tracker := NewProgressTracker(1000)

	stats := tracker.GetDetailedStats()

	if stats == "" {
		t.Error("GetDetailedStats returned empty string")
	}

	// Should show zeros
	if !contains(stats, "Open: 0") {
		t.Errorf("stats should show zero open ports: %s", stats)
	}
}

func TestFormatDuration_Zero(t *testing.T) {
	formatted := formatDuration(0)
	if formatted != "--:--" {
		t.Errorf("formatDuration(0) = %q; want %q", formatted, "--:--")
	}
}

func TestFormatDuration_Negative(t *testing.T) {
	formatted := formatDuration(-5 * time.Second)
	if formatted != "--:--" {
		t.Errorf("formatDuration(-5s) = %q; want %q", formatted, "--:--")
	}
}

func TestFormatDuration_Seconds(t *testing.T) {
	formatted := formatDuration(45 * time.Second)
	if formatted != "00:45" {
		t.Errorf("formatDuration(45s) = %q; want %q", formatted, "00:45")
	}
}

func TestFormatDuration_Minutes(t *testing.T) {
	formatted := formatDuration(2*time.Minute + 30*time.Second)
	if formatted != "02:30" {
		t.Errorf("formatDuration(2m30s) = %q; want %q", formatted, "02:30")
	}
}

func TestFormatDuration_Hours(t *testing.T) {
	formatted := formatDuration(1*time.Hour + 15*time.Minute + 30*time.Second)
	if formatted != "01:15:30" {
		t.Errorf("formatDuration(1h15m30s) = %q; want %q", formatted, "01:15:30")
	}
}

func TestFormatDuration_MultipleHours(t *testing.T) {
	formatted := formatDuration(12*time.Hour + 5*time.Minute + 3*time.Second)
	if formatted != "12:05:03" {
		t.Errorf("formatDuration(12h5m3s) = %q; want %q", formatted, "12:05:03")
	}
}

func TestFormatDuration_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"1 millisecond", 1 * time.Millisecond, "00:00"},
		{"999 milliseconds", 999 * time.Millisecond, "00:00"},
		{"1 second", 1 * time.Second, "00:01"},
		{"59 seconds", 59 * time.Second, "00:59"},
		{"1 minute", 1 * time.Minute, "01:00"},
		{"59 minutes 59 seconds", 59*time.Minute + 59*time.Second, "59:59"},
		{"1 hour", 1 * time.Hour, "01:00:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := formatDuration(tt.duration)
			if formatted != tt.expected {
				t.Errorf("formatDuration(%v) = %q; want %q", tt.duration, formatted, tt.expected)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
