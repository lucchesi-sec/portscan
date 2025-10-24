package core

import (
	"errors"
	"testing"
)

func TestNewResultEvent(t *testing.T) {
	result := ResultEvent{Host: "localhost", Port: 80, State: StateOpen}
	event := NewResultEvent(result)

	if event.Kind != EventKindResult {
		t.Errorf("Kind = %v; want %v", event.Kind, EventKindResult)
	}
	if event.Result == nil {
		t.Fatal("Result is nil")
	}
	if event.Result.Host != "localhost" {
		t.Errorf("Host = %s; want localhost", event.Result.Host)
	}
	if event.Result.Port != 80 {
		t.Errorf("Port = %d; want 80", event.Result.Port)
	}
	if event.Result.State != StateOpen {
		t.Errorf("State = %v; want %v", event.Result.State, StateOpen)
	}
}

func TestNewProgressEvent(t *testing.T) {
	progress := ProgressEvent{Completed: 100, Total: 1000}
	event := NewProgressEvent(progress)

	if event.Kind != EventKindProgress {
		t.Errorf("Kind = %v; want %v", event.Kind, EventKindProgress)
	}
	if event.Progress == nil {
		t.Fatal("Progress is nil")
	}
	if event.Progress.Completed != 100 {
		t.Errorf("Completed = %d; want 100", event.Progress.Completed)
	}
	if event.Progress.Total != 1000 {
		t.Errorf("Total = %d; want 1000", event.Progress.Total)
	}
}

func TestNewErrorEvent(t *testing.T) {
	err := errors.New("test error")
	event := NewErrorEvent(err)

	if event.Kind != EventKindError {
		t.Errorf("Kind = %v; want %v", event.Kind, EventKindError)
	}
	if event.Error == nil {
		t.Fatal("Error is nil")
	}
	if event.Error.Error() != "test error" {
		t.Errorf("Error = %v; want 'test error'", event.Error)
	}
}

func TestNilSafetyResultEvent(t *testing.T) {
	// Test that manually created event with nil Result doesn't panic
	event := Event{
		Kind:   EventKindResult,
		Result: nil,
	}

	if event.Kind != EventKindResult {
		t.Errorf("Kind = %v; want %v", event.Kind, EventKindResult)
	}

	// This would panic without nil checks in consuming code
	if event.Result != nil {
		t.Error("Expected nil Result pointer")
	}
}

func TestNilSafetyProgressEvent(t *testing.T) {
	// Test that manually created event with nil Progress doesn't panic
	event := Event{
		Kind:     EventKindProgress,
		Progress: nil,
	}

	if event.Kind != EventKindProgress {
		t.Errorf("Kind = %v; want %v", event.Kind, EventKindProgress)
	}

	// This would panic without nil checks in consuming code
	if event.Progress != nil {
		t.Error("Expected nil Progress pointer")
	}
}

func TestNilSafetyErrorEvent(t *testing.T) {
	// Test that manually created event with nil Error doesn't panic
	event := Event{
		Kind:  EventKindError,
		Error: nil,
	}

	if event.Kind != EventKindError {
		t.Errorf("Kind = %v; want %v", event.Kind, EventKindError)
	}

	// This would panic without nil checks in consuming code
	if event.Error != nil {
		t.Error("Expected nil Error")
	}
}
