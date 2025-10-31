package ui

import (
	"strings"
	"testing"

	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/pkg/config"
)

// TestScanUI_View_ShowHelp tests help view rendering
func TestScanUI_View_ShowHelp(t *testing.T) {
	results := make(chan core.Event, 10)
	close(results)

	cfg := &config.Config{}
	ui := NewScanUI(cfg, 100, results, false)
	ui.width = 80
	ui.height = 24
	ui.showHelp = true

	view := ui.View()

	if view == "" {
		t.Error("help view should not be empty")
	}

	// Check for help content
	expectedContent := []string{
		"KEYBOARD SHORTCUTS",
		"Navigation",
		"Filtering",
		"Sorting",
	}

	for _, content := range expectedContent {
		if !strings.Contains(view, content) {
			t.Errorf("help view should contain %q", content)
		}
	}
}

// TestScanUI_RenderHelp tests help rendering
func TestScanUI_RenderHelp(t *testing.T) {
	results := make(chan core.Event, 10)
	close(results)

	cfg := &config.Config{}
	ui := NewScanUI(cfg, 100, results, false)

	help := ui.renderHelp()

	if help == "" {
		t.Error("renderHelp should not return empty string")
	}

	// Check for key sections
	expectedSections := []string{
		"Navigation",
		"Filtering",
		"Sorting",
		"View Controls",
		"↑",
		"↓",
		"q",
		"?",
	}

	for _, section := range expectedSections {
		if !strings.Contains(help, section) {
			t.Errorf("help should contain %q", section)
		}
	}
}

// TestScanUI_RenderSortModal tests sort modal rendering
func TestScanUI_RenderSortModal(t *testing.T) {
	results := make(chan core.Event, 10)
	close(results)

	cfg := &config.Config{}
	ui := NewScanUI(cfg, 100, results, false)
	ui.modalState.IsActive = true
	ui.modalState.Type = ModalSort

	modal := ui.renderSortModal()

	if modal == "" {
		t.Error("renderSortModal should not return empty string")
	}

	// Check for sort options
	expectedOptions := []string{
		"SORT OPTIONS",
		"Port",
		"Host",
		"State",
		"Service",
		"Latency",
		"Discovery",
	}

	for _, option := range expectedOptions {
		if !strings.Contains(modal, option) {
			t.Errorf("sort modal should contain %q", option)
		}
	}
}

// TestScanUI_RenderFilterModal tests filter modal rendering
func TestScanUI_RenderFilterModal(t *testing.T) {
	results := make(chan core.Event, 10)
	close(results)

	cfg := &config.Config{}
	ui := NewScanUI(cfg, 100, results, false)
	ui.modalState.IsActive = true
	ui.modalState.Type = ModalFilter

	modal := ui.renderFilterModal()

	if modal == "" {
		t.Error("renderFilterModal should not return empty string")
	}

	// Check for filter options
	expectedOptions := []string{
		"FILTER OPTIONS",
		"Show All",
		"Open",
		"Closed",
		"Filtered",
	}

	for _, option := range expectedOptions {
		if !strings.Contains(modal, option) {
			t.Errorf("filter modal should contain %q", option)
		}
	}
}

// TestScanUI_RenderDashboardView tests dashboard view rendering
func TestScanUI_RenderDashboardView(t *testing.T) {
	results := make(chan core.Event, 10)

	// Add some sample results
	results <- core.NewResultEvent(core.ResultEvent{
		Host:     "localhost",
		Port:     80,
		State:    core.StateOpen,
		Protocol: "tcp",
		Banner:   "HTTP/1.1 200 OK",
	})
	close(results)

	cfg := &config.Config{}
	ui := NewScanUI(cfg, 100, results, false)
	ui.width = 120 // Wide enough for dashboard
	ui.height = 30
	ui.showDashboard = true

	view := ui.renderDashboardView()

	if view == "" {
		t.Error("dashboard view should not be empty")
	}

	// Dashboard should contain stats
	expectedContent := []string{
		"Port Scanner",
	}

	for _, content := range expectedContent {
		if !strings.Contains(view, content) {
			t.Errorf("dashboard view should contain %q", content)
		}
	}
}

// TestScanUI_RenderStatsPanel tests stats panel rendering
func TestScanUI_RenderStatsPanel(t *testing.T) {
	results := make(chan core.Event, 10)

	// Add results
	results <- core.NewResultEvent(core.ResultEvent{
		Host:     "localhost",
		Port:     80,
		State:    core.StateOpen,
		Protocol: "tcp",
		Banner:   "HTTP/1.1 200 OK",
	})
	results <- core.NewResultEvent(core.ResultEvent{
		Host:     "localhost",
		Port:     443,
		State:    core.StateOpen,
		Protocol: "tcp",
		Banner:   "HTTPS",
	})
	close(results)

	cfg := &config.Config{}
	ui := NewScanUI(cfg, 100, results, false)
	ui.width = 80
	ui.height = 24

	// Trigger stats computation
	ui.statsData = ui.computeStats()

	panel := ui.renderStatsPanel(40)

	if panel == "" {
		t.Error("stats panel should not be empty")
	}

	// Check for key stats sections
	expectedSections := []string{
		"Statistics",
		"Performance",
	}

	for _, section := range expectedSections {
		if !strings.Contains(panel, section) {
			t.Errorf("stats panel should contain %q", section)
		}
	}
}

// TestScanUI_RenderMiniBarChart tests bar chart rendering
func TestScanUI_RenderMiniBarChart(t *testing.T) {
	results := make(chan core.Event, 10)

	// Add mixed results
	results <- core.NewResultEvent(core.ResultEvent{
		Host:  "localhost",
		Port:  80,
		State: core.StateOpen,
	})
	results <- core.NewResultEvent(core.ResultEvent{
		Host:  "localhost",
		Port:  81,
		State: core.StateClosed,
	})
	results <- core.NewResultEvent(core.ResultEvent{
		Host:  "localhost",
		Port:  82,
		State: core.StateFiltered,
	})
	close(results)

	cfg := &config.Config{}
	ui := NewScanUI(cfg, 100, results, false)
	ui.statsData = ui.computeStats()

	chart := ui.renderMiniBarChart()

	if chart == "" {
		t.Error("mini bar chart should not be empty")
	}

	// Check for state labels
	expectedLabels := []string{
		"Open",
		"Closed",
		"Filtered",
	}

	for _, label := range expectedLabels {
		if !strings.Contains(chart, label) {
			t.Errorf("bar chart should contain %q label", label)
		}
	}
}

// TestScanUI_RenderBreadcrumb tests breadcrumb rendering
func TestScanUI_RenderBreadcrumb(t *testing.T) {
	results := make(chan core.Event, 10)
	close(results)

	cfg := &config.Config{}
	ui := NewScanUI(cfg, 100, results, false)

	tests := []struct {
		name     string
		scanning bool
		paused   bool
		expected string
	}{
		{"scanning", true, false, "Scanning"},
		{"paused", true, true, "Paused"},
		{"complete", false, false, "Complete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui.scanning = tt.scanning
			ui.isPaused = tt.paused

			breadcrumb := ui.renderBreadcrumb()

			if breadcrumb == "" {
				t.Error("breadcrumb should not be empty")
			}

			if !strings.Contains(breadcrumb, tt.expected) {
				t.Errorf("breadcrumb should contain %q for state %s", tt.expected, tt.name)
			}
		})
	}
}

// TestScanUI_RenderHeader tests header rendering
func TestScanUI_RenderHeader(t *testing.T) {
	results := make(chan core.Event, 10)
	close(results)

	cfg := &config.Config{}
	ui := NewScanUI(cfg, 100, results, false)

	tests := []struct {
		name     string
		scanning bool
		paused   bool
	}{
		{"scanning", true, false},
		{"paused", true, true},
		{"complete", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui.scanning = tt.scanning
			ui.isPaused = tt.paused

			header := ui.renderHeader()

			if header == "" {
				t.Error("header should not be empty")
			}

			if !strings.Contains(header, "Port Scanner") {
				t.Error("header should contain title")
			}
		})
	}
}

// TestScanUI_RenderProgress tests progress bar rendering
func TestScanUI_RenderProgress(t *testing.T) {
	results := make(chan core.Event, 10)
	close(results)

	cfg := &config.Config{}
	ui := NewScanUI(cfg, 100, results, false)

	progress := ui.renderProgress()

	// Progress bar should return some output (even if empty bar)
	// This test mainly ensures it doesn't panic
	if progress == "" {
		t.Log("progress bar is empty (expected for 0% progress)")
	}
}

// TestScanUI_RenderStatus tests status rendering
func TestScanUI_RenderStatus(t *testing.T) {
	results := make(chan core.Event, 10)
	close(results)

	cfg := &config.Config{}
	ui := NewScanUI(cfg, 100, results, false)

	status := ui.renderStatus()

	if status == "" {
		t.Error("status should not be empty")
	}
}

// TestScanUI_RenderFooter tests footer rendering
func TestScanUI_RenderFooter(t *testing.T) {
	results := make(chan core.Event, 10)
	close(results)

	cfg := &config.Config{}
	ui := NewScanUI(cfg, 100, results, false)

	footer := ui.renderFooter()

	// Footer contains help text, should not be empty
	if footer == "" {
		t.Error("footer should not be empty")
	}
}

// TestScanUI_RenderSortFilterIndicators tests indicator rendering
func TestScanUI_RenderSortFilterIndicators(t *testing.T) {
	results := make(chan core.Event, 10)
	close(results)

	cfg := &config.Config{}
	ui := NewScanUI(cfg, 100, results, false)

	tests := []struct {
		name           string
		setupUI        func(*ScanUI)
		expectNonEmpty bool
	}{
		{
			name: "no indicators",
			setupUI: func(ui *ScanUI) {
				ui.sortState = NewSortState()
				ui.sortState.IsActive = false // Explicitly disable sort
				ui.filterState = NewFilterState()
			},
			expectNonEmpty: false,
		},
		{
			name: "with sort",
			setupUI: func(ui *ScanUI) {
				ui.sortState = NewSortState()
				ui.sortState.SetMode(SortByPort)
				ui.filterState = NewFilterState()
			},
			expectNonEmpty: true,
		},
		{
			name: "with filter",
			setupUI: func(ui *ScanUI) {
				ui.sortState = NewSortState()
				ui.filterState = NewFilterState()
				ui.filterState.SetStateFilter(StateFilterOpen)
			},
			expectNonEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupUI(ui)

			indicators := ui.renderSortFilterIndicators()

			if tt.expectNonEmpty && indicators == "" {
				t.Error("indicators should not be empty")
			}

			if !tt.expectNonEmpty && indicators != "" {
				t.Error("indicators should be empty when no filters/sorts active")
			}
		})
	}
}

// TestScanUI_ViewStateTransitions tests different view states
func TestScanUI_ViewStateTransitions(t *testing.T) {
	results := make(chan core.Event, 10)
	close(results)

	cfg := &config.Config{}
	ui := NewScanUI(cfg, 100, results, false)
	ui.width = 80
	ui.height = 24

	tests := []struct {
		name      string
		viewState UIViewState
		checkFunc func(string) bool
	}{
		{
			name:      "main view",
			viewState: UIViewMain,
			checkFunc: func(v string) bool {
				return strings.Contains(v, "Port Scanner")
			},
		},
		{
			name:      "help view",
			viewState: UIViewHelp,
			checkFunc: func(v string) bool {
				return strings.Contains(v, "KEYBOARD SHORTCUTS")
			},
		},
		{
			name:      "sort modal",
			viewState: UIViewMain,
			checkFunc: func(v string) bool {
				// Activate modal and check for overlay
				ui.modalState.IsActive = true
				ui.modalState.Type = ModalSort
				modalView := ui.View()
				return strings.Contains(modalView, "SORT OPTIONS")
			},
		},
		{
			name:      "filter modal",
			viewState: UIViewMain,
			checkFunc: func(v string) bool {
				// Activate modal and check for overlay
				ui.modalState.IsActive = true
				ui.modalState.Type = ModalFilter
				modalView := ui.View()
				return strings.Contains(modalView, "FILTER OPTIONS")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui.viewState = tt.viewState
			if tt.viewState == UIViewHelp {
				ui.showHelp = true
			} else {
				ui.showHelp = false
			}

			view := ui.View()

			if !tt.checkFunc(view) {
				t.Errorf("view state %s failed validation", tt.name)
			}
		})
	}
}
