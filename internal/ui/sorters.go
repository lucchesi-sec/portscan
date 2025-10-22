package ui

import (
	"sort"
	"strings"

	"github.com/lucchesi-sec/portscan/internal/core"
)

// SortMode represents the current sorting configuration
type SortMode int

const (
	SortByPort SortMode = iota
	SortByPortDesc
	SortByHost
	SortByState
	SortByService
	SortByLatency
	SortByLatencyDesc
	SortByDiscovery // Original order
)

// SortState manages sorting configuration
type SortState struct {
	Mode     SortMode
	IsActive bool
}

// NewSortState creates a new sort state with default (port ascending)
func NewSortState() *SortState {
	return &SortState{
		Mode:     SortByPort,
		IsActive: true,
	}
}

// ApplySort sorts the results based on the current sort mode
func (s *SortState) ApplySort(results []core.ResultEvent) []core.ResultEvent {
	// Create a copy to avoid modifying the original
	sorted := make([]core.ResultEvent, len(results))
	copy(sorted, results)

	switch s.Mode {
	case SortByPort:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Port < sorted[j].Port
		})

	case SortByPortDesc:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Port > sorted[j].Port
		})

	case SortByHost:
		sort.Slice(sorted, func(i, j int) bool {
			if strings.EqualFold(sorted[i].Host, sorted[j].Host) {
				return sorted[i].Port < sorted[j].Port
			}
			return strings.ToLower(sorted[i].Host) < strings.ToLower(sorted[j].Host)
		})

	case SortByState:
		sort.Slice(sorted, func(i, j int) bool {
			// Open first, then Closed, then Filtered
			return stateOrder(sorted[i].State) < stateOrder(sorted[j].State)
		})

	case SortByService:
		sort.Slice(sorted, func(i, j int) bool {
			serviceI := getServiceName(sorted[i].Port)
			serviceJ := getServiceName(sorted[j].Port)
			// Sort by service name, then by port if services are equal
			if serviceI == serviceJ {
				return sorted[i].Port < sorted[j].Port
			}
			return strings.ToLower(serviceI) < strings.ToLower(serviceJ)
		})

	case SortByLatency:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Duration < sorted[j].Duration
		})

	case SortByLatencyDesc:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Duration > sorted[j].Duration
		})

	case SortByDiscovery:
		// Keep original order (no sorting needed)
	}

	return sorted
}

// stateOrder returns the sort order for states
func stateOrder(state core.ScanState) int {
	switch state {
	case core.StateOpen:
		return 0
	case core.StateClosed:
		return 1
	case core.StateFiltered:
		return 2
	default:
		return 3
	}
}

// SetMode sets the sort mode
func (s *SortState) SetMode(mode SortMode) {
	s.Mode = mode
	s.IsActive = true
}

// ToggleDirection toggles between ascending and descending for applicable modes
func (s *SortState) ToggleDirection() {
	switch s.Mode {
	case SortByPort:
		s.Mode = SortByPortDesc
	case SortByPortDesc:
		s.Mode = SortByPort
	case SortByLatency:
		s.Mode = SortByLatencyDesc
	case SortByLatencyDesc:
		s.Mode = SortByLatency
	}
}

// GetModeString returns a string representation of the current sort mode
func (s *SortState) GetModeString() string {
	switch s.Mode {
	case SortByPort:
		return "Port ↑"
	case SortByPortDesc:
		return "Port ↓"
	case SortByHost:
		return "Host"
	case SortByState:
		return "State"
	case SortByService:
		return "Service"
	case SortByLatency:
		return "Latency ↑"
	case SortByLatencyDesc:
		return "Latency ↓"
	case SortByDiscovery:
		return "Discovery"
	default:
		return "Unknown"
	}
}

// GetSortDescription returns a description of the current sort
func (s *SortState) GetSortDescription() string {
	if !s.IsActive {
		return ""
	}
	return "Sort: " + s.GetModeString()
}

// NextSortMode cycles to the next sort mode
func (s *SortState) NextSortMode() {
	switch s.Mode {
	case SortByPort:
		s.Mode = SortByPortDesc
	case SortByPortDesc:
		s.Mode = SortByHost
	case SortByHost:
		s.Mode = SortByState
	case SortByState:
		s.Mode = SortByService
	case SortByService:
		s.Mode = SortByLatency
	case SortByLatency:
		s.Mode = SortByLatencyDesc
	case SortByLatencyDesc:
		s.Mode = SortByDiscovery
	case SortByDiscovery:
		s.Mode = SortByPort
	}
}

// PreviousSortMode cycles to the previous sort mode
func (s *SortState) PreviousSortMode() {
	switch s.Mode {
	case SortByPort:
		s.Mode = SortByDiscovery
	case SortByPortDesc:
		s.Mode = SortByPort
	case SortByHost:
		s.Mode = SortByPortDesc
	case SortByState:
		s.Mode = SortByHost
	case SortByService:
		s.Mode = SortByState
	case SortByLatency:
		s.Mode = SortByService
	case SortByLatencyDesc:
		s.Mode = SortByLatency
	case SortByDiscovery:
		s.Mode = SortByLatencyDesc
	}
}
