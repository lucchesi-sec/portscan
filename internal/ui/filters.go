package ui

import (
	"fmt"
	"strings"

	"github.com/lucchesi-sec/portscan/internal/core"
)

// FilterState represents the current filter configuration
type FilterState struct {
	StateFilter    StateFilterType
	PortRangeMin   uint16
	PortRangeMax   uint16
	ServiceFilter  string
	LatencyMax     int // milliseconds, 0 = no filter
	BannerSearch   string
	IsActive       bool
}

// StateFilterType represents which states to show
type StateFilterType int

const (
	StateFilterAll StateFilterType = iota
	StateFilterOpen
	StateFilterClosed
	StateFilterFiltered
)

// NewFilterState creates a new filter state with defaults
func NewFilterState() *FilterState {
	return &FilterState{
		StateFilter:  StateFilterAll,
		PortRangeMin: 0,
		PortRangeMax: 65535,
		IsActive:     false,
	}
}

// ApplyFilters applies all active filters to the results
func (f *FilterState) ApplyFilters(results []core.ResultEvent) []core.ResultEvent {
	if !f.IsActive && f.StateFilter == StateFilterAll {
		return results
	}

	filtered := make([]core.ResultEvent, 0, len(results))

	for _, r := range results {
		if f.matchesFilters(r) {
			filtered = append(filtered, r)
		}
	}

	return filtered
}

// matchesFilters checks if a result matches all active filters
func (f *FilterState) matchesFilters(r core.ResultEvent) bool {
	// State filter
	if !f.matchesStateFilter(r) {
		return false
	}

	// Port range filter
	if r.Port < f.PortRangeMin || r.Port > f.PortRangeMax {
		return false
	}

	// Service filter
	if f.ServiceFilter != "" {
		service := getServiceName(r.Port)
		if !strings.Contains(strings.ToLower(service), strings.ToLower(f.ServiceFilter)) {
			return false
		}
	}

	// Latency filter
	if f.LatencyMax > 0 {
		if r.Duration.Milliseconds() > int64(f.LatencyMax) {
			return false
		}
	}

	// Banner search
	if f.BannerSearch != "" {
		if !strings.Contains(strings.ToLower(r.Banner), strings.ToLower(f.BannerSearch)) {
			return false
		}
	}

	return true
}

// matchesStateFilter checks if result matches the state filter
func (f *FilterState) matchesStateFilter(r core.ResultEvent) bool {
	switch f.StateFilter {
	case StateFilterAll:
		return true
	case StateFilterOpen:
		return r.State == core.StateOpen
	case StateFilterClosed:
		return r.State == core.StateClosed
	case StateFilterFiltered:
		return r.State == core.StateFiltered
	default:
		return true
	}
}

// SetPortRange sets the port range filter
func (f *FilterState) SetPortRange(min, max uint16) {
	f.PortRangeMin = min
	f.PortRangeMax = max
	f.IsActive = true
}

// SetStateFilter sets the state filter
func (f *FilterState) SetStateFilter(stateType StateFilterType) {
	f.StateFilter = stateType
	if stateType != StateFilterAll {
		f.IsActive = true
	}
}

// SetServiceFilter sets the service name filter
func (f *FilterState) SetServiceFilter(service string) {
	f.ServiceFilter = service
	f.IsActive = service != ""
}

// SetLatencyFilter sets the maximum latency filter
func (f *FilterState) SetLatencyFilter(maxMs int) {
	f.LatencyMax = maxMs
	f.IsActive = maxMs > 0
}

// SetBannerSearch sets the banner search filter
func (f *FilterState) SetBannerSearch(search string) {
	f.BannerSearch = search
	f.IsActive = search != ""
}

// Reset clears all filters
func (f *FilterState) Reset() {
	f.StateFilter = StateFilterAll
	f.PortRangeMin = 0
	f.PortRangeMax = 65535
	f.ServiceFilter = ""
	f.LatencyMax = 0
	f.BannerSearch = ""
	f.IsActive = false
}

// GetActiveFilterDescription returns a string describing active filters
func (f *FilterState) GetActiveFilterDescription() string {
	if !f.IsActive && f.StateFilter == StateFilterAll {
		return ""
	}

	var filters []string

	if f.StateFilter != StateFilterAll {
		switch f.StateFilter {
		case StateFilterOpen:
			filters = append(filters, "Open")
		case StateFilterClosed:
			filters = append(filters, "Closed")
		case StateFilterFiltered:
			filters = append(filters, "Filtered")
		}
	}

	if f.PortRangeMin > 0 || f.PortRangeMax < 65535 {
		filters = append(filters, fmt.Sprintf("Ports %d-%d", f.PortRangeMin, f.PortRangeMax))
	}

	if f.ServiceFilter != "" {
		filters = append(filters, "Service: "+f.ServiceFilter)
	}

	if f.LatencyMax > 0 {
		filters = append(filters, fmt.Sprintf("Latency <%dms", f.LatencyMax))
	}

	if f.BannerSearch != "" {
		filters = append(filters, "Banner: "+f.BannerSearch)
	}

	if len(filters) > 0 {
		return "Filters: " + strings.Join(filters, ", ")
	}
	return ""
}
