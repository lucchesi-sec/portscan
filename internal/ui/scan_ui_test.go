package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/pkg/config"
)

func TestNewScanUI(t *testing.T) {
	cfg := &config.Config{
		Workers:   50,
		TimeoutMs: 200,
		Rate:      5000,
	}

	events := make(chan core.Event, 10)
	totalPorts := 1000
	onlyOpen := false

	ui := NewScanUI(cfg, totalPorts, events, onlyOpen)

	if ui == nil {
		t.Fatal("NewScanUI returned nil")
	}

	if ui.results == nil {
		t.Error("results buffer not initialized")
	}

	if ui.progressTrack == nil {
		t.Error("progress tracker not initialized")
	}

	if ui.progressTrack.TotalPorts != totalPorts {
		t.Errorf("progress tracker total ports = %d; want %d", ui.progressTrack.TotalPorts, totalPorts)
	}

	if ui.showOnlyOpen != onlyOpen {
		t.Errorf("showOnlyOpen = %v; want %v", ui.showOnlyOpen, onlyOpen)
	}

	if ui.scanning != true {
		t.Error("scanning should be true initially")
	}
}

func TestNewScanUI_OnlyOpen(t *testing.T) {
	cfg := &config.Config{}
	events := make(chan core.Event, 10)

	ui := NewScanUI(cfg, 100, events, true)

	if !ui.showOnlyOpen {
		t.Error("showOnlyOpen should be true when specified")
	}
}

func TestScanUIKeyBindings_ShortHelp(t *testing.T) {
	kb := KeyBindings{}

	help := kb.ShortHelp()

	if len(help) == 0 {
		t.Error("ShortHelp returned empty slice")
	}
}

func TestScanUIKeyBindings_FullHelp(t *testing.T) {
	kb := KeyBindings{}

	help := kb.FullHelp()

	if len(help) == 0 {
		t.Error("FullHelp returned empty slice")
	}
}

func TestScanUI_Init(t *testing.T) {
	cfg := &config.Config{}
	events := make(chan core.Event, 10)

	ui := NewScanUI(cfg, 100, events, false)

	cmd := ui.Init()

	if cmd == nil {
		t.Error("Init should return a non-nil command")
	}
}

func TestScanUI_View_Uninitialized(t *testing.T) {
	cfg := &config.Config{}
	events := make(chan core.Event, 10)

	ui := NewScanUI(cfg, 100, events, false)

	// Before any size update
	view := ui.View()

	if view == "" {
		t.Error("View should return non-empty string even when uninitialized")
	}

	if view != "Initializing..." {
		t.Logf("View returned: %s", view)
	}
}

func TestScanUI_View_Initialized(t *testing.T) {
	cfg := &config.Config{}
	events := make(chan core.Event, 10)

	ui := NewScanUI(cfg, 100, events, false)

	// Set dimensions
	ui.width = 120
	ui.height = 40

	view := ui.View()

	if view == "" {
		t.Error("View should return non-empty string")
	}

	// Should contain some UI elements
	if len(view) < 10 {
		t.Error("View seems too short for a proper UI")
	}
}

func TestScanUI_Update_WindowSize(t *testing.T) {
	cfg := &config.Config{}
	events := make(chan core.Event, 10)

	ui := NewScanUI(cfg, 100, events, false)

	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}

	model, _ := ui.Update(msg)
	updatedUI := model.(*ScanUI)

	if updatedUI.width != 120 {
		t.Errorf("width = %d; want 120", updatedUI.width)
	}

	if updatedUI.height != 40 {
		t.Errorf("height = %d; want 40", updatedUI.height)
	}
}

func TestScanUI_Update_ScanResult(t *testing.T) {
	cfg := &config.Config{}
	events := make(chan core.Event, 10)

	ui := NewScanUI(cfg, 100, events, false)
	ui.width = 120
	ui.height = 40

	result := scanResultMsg{
		result: core.ResultEvent{
			Host:   "localhost",
			Port:   80,
			State:  core.StateOpen,
			Banner: "HTTP/1.1",
		},
	}

	model, _ := ui.Update(result)
	updatedUI := model.(*ScanUI)

	// Result should be added to buffer
	if updatedUI.results.Len() != 1 {
		t.Errorf("results buffer length = %d; want 1", updatedUI.results.Len())
	}
}

func TestScanUI_Update_ScanProgress(t *testing.T) {
	cfg := &config.Config{}
	events := make(chan core.Event, 10)

	ui := NewScanUI(cfg, 100, events, false)
	ui.width = 120
	ui.height = 40

	progress := scanProgressMsg{
		progress: core.ProgressEvent{
			Total:     100,
			Completed: 50,
			Rate:      1000.0,
		},
	}

	model, _ := ui.Update(progress)
	updatedUI := model.(*ScanUI)

	// Progress should be updated (exact fields depend on how it's processed)
	_ = updatedUI // Just ensure update succeeded without panic
}

func TestScanUI_Update_ScanComplete(t *testing.T) {
	cfg := &config.Config{}
	events := make(chan core.Event, 10)

	ui := NewScanUI(cfg, 100, events, false)
	ui.width = 120
	ui.height = 40
	ui.scanning = true

	msg := scanCompleteMsg{}

	model, _ := ui.Update(msg)
	updatedUI := model.(*ScanUI)

	// Scanning should be false after completion
	if updatedUI.scanning {
		t.Error("scanning should be false after scanCompleteMsg")
	}
}

func TestResultBuffer_Append_Capacity(t *testing.T) {
	// Test that buffer respects capacity limits
	capacity := 5
	buffer := NewResultBuffer(capacity)

	// Add more items than capacity
	for i := 0; i < 10; i++ {
		buffer.Append(core.ResultEvent{
			Host: "host",
			Port: uint16(80 + i),
		})
	}

	// Buffer should contain only last 'capacity' items
	if buffer.Len() != capacity {
		t.Errorf("buffer length = %d; want %d", buffer.Len(), capacity)
	}

	items := buffer.Items()
	if len(items) != capacity {
		t.Errorf("items length = %d; want %d", len(items), capacity)
	}

	// First item should be from index 5 (oldest kept)
	if items[0].Port != 85 {
		t.Errorf("first item port = %d; want 85", items[0].Port)
	}

	// Last item should be from index 9 (newest)
	if items[capacity-1].Port != 89 {
		t.Errorf("last item port = %d; want 89", items[capacity-1].Port)
	}
}

func TestResultBuffer_ZeroCapacity(t *testing.T) {
	buffer := NewResultBuffer(0)

	// Should default to some capacity
	if buffer.capacity <= 0 {
		t.Error("buffer should have positive capacity even when 0 requested")
	}
}

func TestResultBuffer_NegativeCapacity(t *testing.T) {
	buffer := NewResultBuffer(-10)

	// Should handle gracefully and use default
	if buffer.capacity <= 0 {
		t.Error("buffer should have positive capacity even when negative requested")
	}
}

func TestResultStats_AddResults(t *testing.T) {
	stats := NewResultStats()

	stats.Add(core.ResultEvent{State: core.StateOpen})
	stats.Add(core.ResultEvent{State: core.StateClosed})
	stats.Add(core.ResultEvent{State: core.StateFiltered})

	total, open, closed, filtered := stats.Totals()

	if total != 3 {
		t.Errorf("total = %d; want 3", total)
	}

	if open != 1 {
		t.Errorf("open = %d; want 1", open)
	}

	if closed != 1 {
		t.Errorf("closed = %d; want 1", closed)
	}

	if filtered != 1 {
		t.Errorf("filtered = %d; want 1", filtered)
	}
}

// TestResultStats_ServiceCounts removed as ResultStats doesn't track service counts

func TestScanUI_ComputeStats(t *testing.T) {
	cfg := &config.Config{}
	events := make(chan core.Event, 10)

	ui := NewScanUI(cfg, 1000, events, false)

	// Add some results
	ui.results.Append(core.ResultEvent{Host: "h1", Port: 80, State: core.StateOpen})
	ui.results.Append(core.ResultEvent{Host: "h1", Port: 81, State: core.StateOpen})
	ui.results.Append(core.ResultEvent{Host: "h1", Port: 22, State: core.StateClosed})
	ui.results.Append(core.ResultEvent{Host: "h2", Port: 443, State: core.StateFiltered})

	stats := ui.computeStats()

	if stats.TotalResults != 4 {
		t.Errorf("total results = %d; want 4", stats.TotalResults)
	}

	if stats.OpenCount != 2 {
		t.Errorf("open count = %d; want 2", stats.OpenCount)
	}

	if stats.ClosedCount != 1 {
		t.Errorf("closed count = %d; want 1", stats.ClosedCount)
	}

	if stats.FilteredCount != 1 {
		t.Errorf("filtered count = %d; want 1", stats.FilteredCount)
	}
}

func TestScanUI_OnlyOpenFilter(t *testing.T) {
	cfg := &config.Config{}
	events := make(chan core.Event, 10)

	ui := NewScanUI(cfg, 100, events, true)
	ui.width = 120
	ui.height = 40

	// Add mixed results
	ui.results.Append(core.ResultEvent{Host: "h1", Port: 80, State: core.StateOpen})
	ui.results.Append(core.ResultEvent{Host: "h1", Port: 81, State: core.StateClosed})
	ui.results.Append(core.ResultEvent{Host: "h1", Port: 82, State: core.StateOpen})

	// Update should filter results for display
	// The actual filtering happens in updateTable which is called during Update
	// For unit test, we verify the showOnlyOpen flag is respected
	if !ui.showOnlyOpen {
		t.Error("showOnlyOpen flag should be true")
	}
}
