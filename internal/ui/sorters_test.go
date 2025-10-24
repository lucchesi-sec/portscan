package ui

import (
	"testing"
	"time"

	"github.com/lucchesi-sec/portscan/internal/core"
)

func TestNewSortState(t *testing.T) {
	s := NewSortState()

	if s.Mode != SortByPort {
		t.Errorf("expected default mode SortByPort, got %v", s.Mode)
	}
	if !s.IsActive {
		t.Error("expected IsActive to be true by default")
	}
}

func TestSortState_SetMode(t *testing.T) {
	s := NewSortState()
	s.IsActive = false

	s.SetMode(SortByHost)

	if s.Mode != SortByHost {
		t.Errorf("expected mode SortByHost, got %v", s.Mode)
	}
	if !s.IsActive {
		t.Error("expected IsActive to be true after SetMode")
	}
}

func TestSortState_ToggleDirection(t *testing.T) {
	tests := []struct {
		name     string
		initial  SortMode
		expected SortMode
	}{
		{"Port asc to desc", SortByPort, SortByPortDesc},
		{"Port desc to asc", SortByPortDesc, SortByPort},
		{"Latency asc to desc", SortByLatency, SortByLatencyDesc},
		{"Latency desc to asc", SortByLatencyDesc, SortByLatency},
		{"Host unchanged", SortByHost, SortByHost},
		{"State unchanged", SortByState, SortByState},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SortState{Mode: tt.initial}
			s.ToggleDirection()
			if s.Mode != tt.expected {
				t.Errorf("expected mode %v, got %v", tt.expected, s.Mode)
			}
		})
	}
}

func TestSortState_GetModeString(t *testing.T) {
	tests := []struct {
		mode     SortMode
		expected string
	}{
		{SortByPort, "Port ↑"},
		{SortByPortDesc, "Port ↓"},
		{SortByHost, "Host"},
		{SortByState, "State"},
		{SortByService, "Service"},
		{SortByLatency, "Latency ↑"},
		{SortByLatencyDesc, "Latency ↓"},
		{SortByDiscovery, "Discovery"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			s := &SortState{Mode: tt.mode}
			result := s.GetModeString()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSortState_GetSortDescription(t *testing.T) {
	s := &SortState{Mode: SortByPort, IsActive: true}
	desc := s.GetSortDescription()
	expected := "Sort: Port ↑"
	if desc != expected {
		t.Errorf("expected %q, got %q", expected, desc)
	}

	s.IsActive = false
	desc = s.GetSortDescription()
	if desc != "" {
		t.Errorf("expected empty string when inactive, got %q", desc)
	}
}

func TestSortState_NextSortMode(t *testing.T) {
	tests := []struct {
		from SortMode
		to   SortMode
	}{
		{SortByPort, SortByPortDesc},
		{SortByPortDesc, SortByHost},
		{SortByHost, SortByState},
		{SortByState, SortByService},
		{SortByService, SortByLatency},
		{SortByLatency, SortByLatencyDesc},
		{SortByLatencyDesc, SortByDiscovery},
		{SortByDiscovery, SortByPort},
	}

	for _, tt := range tests {
		t.Run("next", func(t *testing.T) {
			s := &SortState{Mode: tt.from}
			s.NextSortMode()
			if s.Mode != tt.to {
				t.Errorf("NextSortMode from %v: expected %v, got %v", tt.from, tt.to, s.Mode)
			}
		})
	}
}

func TestSortState_PreviousSortMode(t *testing.T) {
	tests := []struct {
		from SortMode
		to   SortMode
	}{
		{SortByPort, SortByDiscovery},
		{SortByPortDesc, SortByPort},
		{SortByHost, SortByPortDesc},
		{SortByState, SortByHost},
		{SortByService, SortByState},
		{SortByLatency, SortByService},
		{SortByLatencyDesc, SortByLatency},
		{SortByDiscovery, SortByLatencyDesc},
	}

	for _, tt := range tests {
		t.Run("previous", func(t *testing.T) {
			s := &SortState{Mode: tt.from}
			s.PreviousSortMode()
			if s.Mode != tt.to {
				t.Errorf("PreviousSortMode from %v: expected %v, got %v", tt.from, tt.to, s.Mode)
			}
		})
	}
}

func TestSortState_ApplySort_Port(t *testing.T) {
	s := &SortState{Mode: SortByPort}
	results := []core.ResultEvent{
		{Port: 443},
		{Port: 80},
		{Port: 8080},
		{Port: 22},
	}

	sorted := s.ApplySort(results)

	expected := []uint16{22, 80, 443, 8080}
	for i, r := range sorted {
		if r.Port != expected[i] {
			t.Errorf("expected port %d at index %d, got %d", expected[i], i, r.Port)
		}
	}
}

func TestSortState_ApplySort_PortDesc(t *testing.T) {
	s := &SortState{Mode: SortByPortDesc}
	results := []core.ResultEvent{
		{Port: 443},
		{Port: 80},
		{Port: 8080},
		{Port: 22},
	}

	sorted := s.ApplySort(results)

	expected := []uint16{8080, 443, 80, 22}
	for i, r := range sorted {
		if r.Port != expected[i] {
			t.Errorf("expected port %d at index %d, got %d", expected[i], i, r.Port)
		}
	}
}

func TestSortState_ApplySort_Host(t *testing.T) {
	s := &SortState{Mode: SortByHost}
	results := []core.ResultEvent{
		{Host: "example.com", Port: 80},
		{Host: "beta.com", Port: 443},
		{Host: "alpha.com", Port: 22},
		{Host: "alpha.com", Port: 21},
	}

	sorted := s.ApplySort(results)

	if sorted[0].Host != "alpha.com" || sorted[0].Port != 21 {
		t.Errorf("expected alpha.com:21 first, got %s:%d", sorted[0].Host, sorted[0].Port)
	}
	if sorted[1].Host != "alpha.com" || sorted[1].Port != 22 {
		t.Errorf("expected alpha.com:22 second, got %s:%d", sorted[1].Host, sorted[1].Port)
	}
	if sorted[2].Host != "beta.com" {
		t.Errorf("expected beta.com third, got %s", sorted[2].Host)
	}
	if sorted[3].Host != "example.com" {
		t.Errorf("expected example.com last, got %s", sorted[3].Host)
	}
}

func TestSortState_ApplySort_State(t *testing.T) {
	s := &SortState{Mode: SortByState}
	results := []core.ResultEvent{
		{State: core.StateFiltered},
		{State: core.StateOpen},
		{State: core.StateClosed},
		{State: core.StateOpen},
	}

	sorted := s.ApplySort(results)

	if sorted[0].State != core.StateOpen {
		t.Errorf("expected StateOpen first, got %v", sorted[0].State)
	}
	if sorted[1].State != core.StateOpen {
		t.Errorf("expected StateOpen second, got %v", sorted[1].State)
	}
	if sorted[2].State != core.StateClosed {
		t.Errorf("expected StateClosed third, got %v", sorted[2].State)
	}
	if sorted[3].State != core.StateFiltered {
		t.Errorf("expected StateFiltered last, got %v", sorted[3].State)
	}
}

func TestSortState_ApplySort_Latency(t *testing.T) {
	s := &SortState{Mode: SortByLatency}
	results := []core.ResultEvent{
		{Duration: 100 * time.Millisecond},
		{Duration: 50 * time.Millisecond},
		{Duration: 200 * time.Millisecond},
		{Duration: 25 * time.Millisecond},
	}

	sorted := s.ApplySort(results)

	expected := []time.Duration{
		25 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		200 * time.Millisecond,
	}
	for i, r := range sorted {
		if r.Duration != expected[i] {
			t.Errorf("expected duration %v at index %d, got %v", expected[i], i, r.Duration)
		}
	}
}

func TestSortState_ApplySort_LatencyDesc(t *testing.T) {
	s := &SortState{Mode: SortByLatencyDesc}
	results := []core.ResultEvent{
		{Duration: 100 * time.Millisecond},
		{Duration: 50 * time.Millisecond},
		{Duration: 200 * time.Millisecond},
		{Duration: 25 * time.Millisecond},
	}

	sorted := s.ApplySort(results)

	expected := []time.Duration{
		200 * time.Millisecond,
		100 * time.Millisecond,
		50 * time.Millisecond,
		25 * time.Millisecond,
	}
	for i, r := range sorted {
		if r.Duration != expected[i] {
			t.Errorf("expected duration %v at index %d, got %v", expected[i], i, r.Duration)
		}
	}
}

func TestSortState_ApplySort_Discovery(t *testing.T) {
	s := &SortState{Mode: SortByDiscovery}
	results := []core.ResultEvent{
		{Port: 443},
		{Port: 80},
		{Port: 8080},
		{Port: 22},
	}

	sorted := s.ApplySort(results)

	// Discovery order should preserve original order
	expected := []uint16{443, 80, 8080, 22}
	for i, r := range sorted {
		if r.Port != expected[i] {
			t.Errorf("expected port %d at index %d, got %d", expected[i], i, r.Port)
		}
	}
}

func TestSortState_ApplySort_DoesNotModifyOriginal(t *testing.T) {
	s := &SortState{Mode: SortByPort}
	results := []core.ResultEvent{
		{Port: 443},
		{Port: 80},
		{Port: 8080},
	}

	originalFirst := results[0].Port
	sorted := s.ApplySort(results)

	// Original should not be modified
	if results[0].Port != originalFirst {
		t.Errorf("original array was modified")
	}
	// Sorted should be different
	if sorted[0].Port == originalFirst {
		t.Errorf("sorted array is same as original (expected sorted order)")
	}
}

func TestStateOrder(t *testing.T) {
	tests := []struct {
		state core.ScanState
		order int
	}{
		{core.StateOpen, 0},
		{core.StateClosed, 1},
		{core.StateFiltered, 2},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			order := stateOrder(tt.state)
			if order != tt.order {
				t.Errorf("expected order %d for state %v, got %d", tt.order, tt.state, order)
			}
		})
	}
}
