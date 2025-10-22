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
