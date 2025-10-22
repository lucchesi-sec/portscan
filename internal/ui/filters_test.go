package ui

import (
	"testing"
	"time"

	"github.com/lucchesi-sec/portscan/internal/core"
)

func TestNewFilterState(t *testing.T) {
	state := NewFilterState()

	if state.StateFilter != StateFilterAll {
		t.Errorf("expected default state filter to be StateFilterAll, got %v", state.StateFilter)
	}

	if state.ServiceFilter != "" {
		t.Errorf("expected empty service filter, got %s", state.ServiceFilter)
	}

	if state.PortRangeMin != 0 || state.PortRangeMax != 65535 {
		t.Errorf("expected full port range, got %d-%d", state.PortRangeMin, state.PortRangeMax)
	}

	if state.IsActive {
		t.Error("expected filter to be inactive initially")
	}
}

func TestFilterState_ApplyFilters_StateFilterAll(t *testing.T) {
	results := []core.ResultEvent{
		{Host: "host1", Port: 80, State: core.StateOpen},
		{Host: "host2", Port: 443, State: core.StateClosed},
		{Host: "host3", Port: 22, State: core.StateFiltered},
	}

	state := NewFilterState()
	state.SetStateFilter(StateFilterAll)

	filtered := state.ApplyFilters(results)

	if len(filtered) != 3 {
		t.Errorf("expected 3 results with StateFilterAll, got %d", len(filtered))
	}
}

func TestFilterState_ApplyFilters_StateFilterOpen(t *testing.T) {
	results := []core.ResultEvent{
		{Host: "host1", Port: 80, State: core.StateOpen},
		{Host: "host2", Port: 443, State: core.StateClosed},
		{Host: "host3", Port: 22, State: core.StateOpen},
		{Host: "host4", Port: 3306, State: core.StateFiltered},
	}

	state := NewFilterState()
	state.SetStateFilter(StateFilterOpen)

	filtered := state.ApplyFilters(results)

	if len(filtered) != 2 {
		t.Errorf("expected 2 open results, got %d", len(filtered))
	}

	for _, result := range filtered {
		if result.State != core.StateOpen {
			t.Errorf("expected only open ports, found %v", result.State)
		}
	}
}

func TestFilterState_ApplyFilters_StateFilterClosed(t *testing.T) {
	results := []core.ResultEvent{
		{Host: "host1", Port: 80, State: core.StateOpen},
		{Host: "host2", Port: 443, State: core.StateClosed},
		{Host: "host3", Port: 22, State: core.StateClosed},
		{Host: "host4", Port: 3306, State: core.StateFiltered},
	}

	state := NewFilterState()
	state.SetStateFilter(StateFilterClosed)

	filtered := state.ApplyFilters(results)

	if len(filtered) != 2 {
		t.Errorf("expected 2 closed results, got %d", len(filtered))
	}

	for _, result := range filtered {
		if result.State != core.StateClosed {
			t.Errorf("expected only closed ports, found %v", result.State)
		}
	}
}

func TestFilterState_ApplyFilters_StateFilterFiltered(t *testing.T) {
	results := []core.ResultEvent{
		{Host: "host1", Port: 80, State: core.StateOpen},
		{Host: "host2", Port: 443, State: core.StateFiltered},
		{Host: "host3", Port: 22, State: core.StateFiltered},
		{Host: "host4", Port: 3306, State: core.StateClosed},
	}

	state := NewFilterState()
	state.SetStateFilter(StateFilterFiltered)

	filtered := state.ApplyFilters(results)

	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered results, got %d", len(filtered))
	}

	for _, result := range filtered {
		if result.State != core.StateFiltered {
			t.Errorf("expected only filtered ports, found %v", result.State)
		}
	}
}

func TestFilterState_ApplyFilters_PortRange(t *testing.T) {
	results := []core.ResultEvent{
		{Host: "host1", Port: 80, State: core.StateOpen},
		{Host: "host2", Port: 443, State: core.StateOpen},
		{Host: "host3", Port: 22, State: core.StateOpen},
		{Host: "host4", Port: 8080, State: core.StateOpen},
	}

	state := NewFilterState()
	state.SetPortRange(80, 500)

	filtered := state.ApplyFilters(results)

	// Should match ports 80-500: 80, 443
	if len(filtered) != 2 {
		t.Errorf("expected 2 results in port range 80-500, got %d", len(filtered))
	}

	for _, result := range filtered {
		if result.Port < 80 || result.Port > 500 {
			t.Errorf("port %d outside expected range 80-500", result.Port)
		}
	}
}

func TestFilterState_ApplyFilters_BannerSearch(t *testing.T) {
	results := []core.ResultEvent{
		{Host: "host1", Port: 80, State: core.StateOpen, Banner: "Apache/2.4.41"},
		{Host: "host2", Port: 443, State: core.StateOpen, Banner: "nginx/1.18.0"},
		{Host: "host3", Port: 22, State: core.StateOpen, Banner: "SSH-2.0-OpenSSH_8.2"},
		{Host: "host4", Port: 3306, State: core.StateOpen, Banner: ""},
	}

	state := NewFilterState()
	state.SetBannerSearch("SSH")

	filtered := state.ApplyFilters(results)

	// Should match banner containing "SSH"
	if len(filtered) != 1 {
		t.Errorf("expected 1 result with SSH banner, got %d", len(filtered))
	}

	if len(filtered) > 0 && filtered[0].Port != 22 {
		t.Errorf("expected SSH on port 22, got port %d", filtered[0].Port)
	}
}

func TestFilterState_ApplyFilters_LatencyFilter(t *testing.T) {
	results := []core.ResultEvent{
		{Host: "host1", Port: 80, State: core.StateOpen, Duration: 50 * time.Millisecond},
		{Host: "host2", Port: 443, State: core.StateOpen, Duration: 200 * time.Millisecond},
		{Host: "host3", Port: 22, State: core.StateOpen, Duration: 10 * time.Millisecond},
		{Host: "host4", Port: 3306, State: core.StateOpen, Duration: 500 * time.Millisecond},
	}

	state := NewFilterState()
	state.SetLatencyFilter(100) // Max 100ms

	filtered := state.ApplyFilters(results)

	// Should match results with latency <= 100ms: 50ms and 10ms
	if len(filtered) != 2 {
		t.Errorf("expected 2 results with latency <= 100ms, got %d", len(filtered))
	}

	for _, result := range filtered {
		if result.Duration.Milliseconds() > 100 {
			t.Errorf("result has latency %dms, expected <= 100ms", result.Duration.Milliseconds())
		}
	}
}

func TestFilterState_ApplyFilters_MultipleFilters(t *testing.T) {
	results := []core.ResultEvent{
		{Host: "host1", Port: 80, State: core.StateOpen, Duration: 50 * time.Millisecond},
		{Host: "host2", Port: 80, State: core.StateClosed, Duration: 30 * time.Millisecond},
		{Host: "host3", Port: 443, State: core.StateOpen, Duration: 20 * time.Millisecond},
		{Host: "host4", Port: 8080, State: core.StateOpen, Duration: 10 * time.Millisecond},
	}

	state := NewFilterState()
	state.SetStateFilter(StateFilterOpen)
	state.SetPortRange(1, 500) // Exclude 8080

	filtered := state.ApplyFilters(results)

	// Should match: open AND port 1-500
	// Expected: host1 (80, open) and host3 (443, open)
	if len(filtered) != 2 {
		t.Errorf("expected 2 results matching both filters, got %d", len(filtered))
	}

	for _, result := range filtered {
		if result.State != core.StateOpen {
			t.Errorf("expected open state, got %v", result.State)
		}
		if result.Port > 500 {
			t.Errorf("expected port <= 500, got %d", result.Port)
		}
	}
}

func TestFilterState_ApplyFilters_EmptyResults(t *testing.T) {
	results := []core.ResultEvent{}

	state := NewFilterState()
	state.SetStateFilter(StateFilterOpen)

	filtered := state.ApplyFilters(results)

	if len(filtered) != 0 {
		t.Errorf("expected empty result, got %d items", len(filtered))
	}
}

func TestFilterState_ApplyFilters_NoMatches(t *testing.T) {
	results := []core.ResultEvent{
		{Host: "host1", Port: 80, State: core.StateClosed},
		{Host: "host2", Port: 443, State: core.StateClosed},
	}

	state := NewFilterState()
	state.SetStateFilter(StateFilterOpen)

	filtered := state.ApplyFilters(results)

	if len(filtered) != 0 {
		t.Errorf("expected no matches, got %d items", len(filtered))
	}
}

func TestFilterState_Reset(t *testing.T) {
	state := NewFilterState()
	state.SetStateFilter(StateFilterOpen)
	state.SetPortRange(80, 443)
	state.SetBannerSearch("Apache")
	state.SetLatencyFilter(100)

	state.Reset()

	if state.StateFilter != StateFilterAll {
		t.Errorf("expected StateFilterAll after reset, got %v", state.StateFilter)
	}

	if state.PortRangeMin != 0 || state.PortRangeMax != 65535 {
		t.Errorf("expected full port range after reset, got %d-%d", state.PortRangeMin, state.PortRangeMax)
	}

	if state.BannerSearch != "" {
		t.Errorf("expected empty banner search after reset, got %s", state.BannerSearch)
	}

	if state.LatencyMax != 0 {
		t.Errorf("expected zero latency max after reset, got %d", state.LatencyMax)
	}

	if state.IsActive {
		t.Error("expected filter to be inactive after reset")
	}
}

func TestFilterState_CaseInsensitiveBannerSearch(t *testing.T) {
	results := []core.ResultEvent{
		{Host: "host1", Port: 80, State: core.StateOpen, Banner: "Apache/2.4.41"},
	}

	state := NewFilterState()
	state.SetBannerSearch("apache") // lowercase

	filtered := state.ApplyFilters(results)

	// Banner search should be case-insensitive
	if len(filtered) != 1 {
		t.Errorf("expected case-insensitive banner match, got %d results", len(filtered))
	}
}

func TestFilterState_GetActiveFilterDescription(t *testing.T) {
	state := NewFilterState()

	// No filters active
	desc := state.GetActiveFilterDescription()
	if desc != "" {
		t.Errorf("expected empty description with no filters, got %s", desc)
	}

	// With state filter
	state.SetStateFilter(StateFilterOpen)
	desc = state.GetActiveFilterDescription()
	if desc == "" {
		t.Error("expected non-empty description with state filter")
	}

	// Reset and test port range filter
	state.Reset()
	state.SetPortRange(80, 443)
	desc = state.GetActiveFilterDescription()
	if desc == "" {
		t.Error("expected non-empty description with port range filter")
	}

	// Test banner search filter
	state.Reset()
	state.SetBannerSearch("Apache")
	desc = state.GetActiveFilterDescription()
	if desc == "" {
		t.Error("expected non-empty description with banner search filter")
	}
}
