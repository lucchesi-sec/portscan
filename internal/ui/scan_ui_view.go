package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/truncate"
)

// View renders the UI
func (m *ScanUI) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	if m.showHelp {
		return m.renderHelp()
	}

	if m.modalState.IsActive {
		return m.renderModalOverlay()
	}

	return m.renderMain()
}

func (m *ScanUI) renderMain() string {
	// Check if dashboard view is enabled
	if m.showDashboard && m.width >= DashboardMinWidth {
		return m.renderDashboardView()
	}

	var b strings.Builder

	breadcrumb := m.renderBreadcrumb()
	b.WriteString(breadcrumb + "\n")

	header := m.renderHeader()
	b.WriteString(header + "\n")

	if m.scanning {
		progressView := m.renderProgress()
		b.WriteString(progressView + "\n")
	}

	status := m.renderStatus()
	b.WriteString(status + "\n")

	indicators := m.renderSortFilterIndicators()
	if indicators != "" {
		b.WriteString(indicators + "\n")
	}

	b.WriteString("\n")
	b.WriteString(m.table.View() + "\n")

	footer := m.renderFooter()
	b.WriteString(footer)

	return b.String()
}

func (m *ScanUI) renderBreadcrumb() string {
	style := lipgloss.NewStyle().
		Foreground(m.theme.Secondary).
		Bold(true)

	location := "Port Scanner"
	if m.isPaused {
		location += " › Paused"
	} else if m.scanning {
		location += " › Scanning"
	} else {
		location += " › Complete"
	}

	// Enhanced scan metrics
	if m.scanning {
		// Add scan duration and performance indicator with color coding
		duration := m.progressTrack.GetElapsedDuration()
		indicator := m.progressTrack.GetPerformanceIndicator()
		rate := m.progressTrack.GetFormattedRate() + " ports/sec"

		// Color code performance indicator
		indicatorStyle := m.getPerformanceIndicatorStyle()
		coloredIndicator := indicatorStyle.Render(indicator + " " + rate)

		// Add host count, percentage completion, and ETA
		hosts := m.progressTrack.GetHostProgress()
		percent := fmt.Sprintf("%.1f%%", m.progressTrack.GetProgress())
		eta := m.progressTrack.GetDetailedETA()

		// Build metrics string with color-coded status
		metrics := fmt.Sprintf(" • Running: %s • %s • Hosts: %s • Progress: %s complete • ETA: %s • %s",
			coloredIndicator,
			duration,
			hosts,
			percent,
			eta,
			m.getStatusMessage())

		return style.Render(location + metrics)
	} else {
		// Completed scan metrics
		total, open, closed, filtered := m.stats.Totals()
		duration := m.progressTrack.GetElapsedDuration()

		metrics := fmt.Sprintf(" • %d total • %d open • %d closed • %d filtered • Duration: %s • Complete ✓",
			total, open, closed, filtered, duration)

		return style.Render(location + metrics)
	}
}

// getPerformanceIndicatorStyle returns appropriate color based on performance trend
func (m *ScanUI) getPerformanceIndicatorStyle() lipgloss.Style {
	trend := m.progressTrack.PerformanceTrend
	style := lipgloss.NewStyle()

	switch trend {
	case TrendImproving:
		// Green for improving
		return style.Foreground(lipgloss.Color("#00FF00"))
	case TrendDegrading:
		// Red for degrading
		return style.Foreground(lipgloss.Color("#FF4444"))
	default:
		// Blue/gray for stable
		return style.Foreground(m.theme.Secondary)
	}
}

// getStatusMessage returns color-coded status message based on performance
func (m *ScanUI) getStatusMessage() string {
	trend := m.progressTrack.PerformanceTrend
	rate := m.progressTrack.CurrentRate

	var message string
	var color lipgloss.Color

	switch trend {
	case TrendImproving:
		message = "Performance improving"
		color = lipgloss.Color("#00FF00") // Green
	case TrendDegrading:
		message = "Performance degrading"
		color = lipgloss.Color("#FF4444") // Red
	default:
		if rate >= 5000 {
			message = "High performance"
			color = m.theme.Primary
		} else if rate >= 2000 {
			message = "Normal performance"
			color = m.theme.Info
		} else {
			message = "Low performance"
			color = m.theme.Warning
		}
	}

	style := lipgloss.NewStyle().Foreground(color)
	return style.Render(message)
}

func (m *ScanUI) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary)

	var icon string
	if m.scanning && !m.isPaused {
		icon = m.spinner.View() + " "
	} else if m.isPaused {
		icon = "⏸ "
	} else {
		icon = "✓ "
	}

	return titleStyle.Render(icon + "Port Scanner Results")
}

func (m *ScanUI) renderProgress() string {
	progress := m.progressTrack.GetProgress() / 100.0
	progressBar := m.progressBar.ViewAs(progress)

	// Add rate display
	rateStyle := lipgloss.NewStyle().Foreground(m.theme.Secondary)
	rateText := rateStyle.Render(fmt.Sprintf("  %0.1f pps • ETA: %s", m.currentRate, formatDuration(m.progressTrack.GetETA())))

	return progressBar + rateText
}

func (m *ScanUI) renderStatus() string {
	status := m.progressTrack.GetStatusLine()
	details := m.progressTrack.GetDetailedStats()

	// Color-code based on status
	statusStyle := lipgloss.NewStyle()
	if m.isPaused {
		statusStyle = statusStyle.Foreground(m.theme.Warning)
	} else if m.scanning {
		statusStyle = statusStyle.Foreground(m.theme.Info)
	} else {
		statusStyle = statusStyle.Foreground(m.theme.Success)
	}

	detailStyle := lipgloss.NewStyle().Foreground(m.theme.Secondary)

	// Add elapsed time to details
	elapsed := fmt.Sprintf(" • Elapsed: %s", formatDuration(m.progressTrack.GetActiveTime()))
	enhancedDetails := details + elapsed

	return statusStyle.Render(status) + "\n" + detailStyle.Render(enhancedDetails)
}

func (m *ScanUI) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(m.theme.Secondary)

	return footerStyle.Render(m.help.View(m.keys))
}

func (m *ScanUI) renderSortFilterIndicators() string {
	style := lipgloss.NewStyle().
		Foreground(m.theme.Secondary).
		Bold(true)

	var indicators []string

	if m.sortState.IsActive {
		indicators = append(indicators, m.sortState.GetSortDescription())
	}

	filterDesc := m.filterState.GetActiveFilterDescription()
	if filterDesc != "" {
		indicators = append(indicators, filterDesc)
	}

	if len(indicators) > 0 {
		return style.Render("▶ " + strings.Join(indicators, " | "))
	}
	return ""
}

func (m *ScanUI) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Padding(2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary)

	content := `
📖 KEYBOARD SHORTCUTS

Navigation:
  ↑/k        Move up
  ↓/j        Move down
  PgUp/PgDn  Page up/down
  g/G        Jump to top/bottom

Filtering & Sorting:
  s          Sort options (modal)
  r          Reset filters
  o          Toggle open-only

View Controls:
  D          Toggle dashboard view
  Enter      View details
  p / Space  Pause/resume
  ?          Toggle help
  Ctrl+L     Clear screen
  q / Esc    Quit/Close modal
`

	return helpStyle.Render(content)
}

// renderModalOverlay renders the semi-transparent background with modal on top
func (m *ScanUI) renderModalOverlay() string {
	mainView := m.renderMain()
	overlayStyle := m.theme.ModalOverlayStyle()
	dimmedBackground := overlayStyle.Render(mainView)

	availableWidth := max(1, m.width)
	availableHeight := max(1, m.height)

	modalWidth := max(ModalMinWidth, int(float64(availableWidth)*ModalWidthPercent))
	modalHeight := max(ModalMinHeight, int(float64(availableHeight)*ModalHeightPercent))
	modalWidth = min(modalWidth, availableWidth)
	modalHeight = min(modalHeight, availableHeight)

	var modalContent string
	switch m.modalState.Type {
	case ModalSort:
		modalContent = m.renderSortModal()
	case ModalDetails:
		modalContent = m.renderDetailsModal()
	default:
		modalContent = ""
	}

	modalStyle := m.theme.ModalStyle(ModalBorderPadding).
		Width(modalWidth).
		Height(modalHeight)
	styledModal := strings.TrimRight(modalStyle.Render(modalContent), "\n")
	modalLines := strings.Split(styledModal, "\n")
	if len(modalLines) == 0 {
		modalLines = []string{""}
	}

	actualModalHeight := len(modalLines)
	if actualModalHeight > availableHeight {
		modalLines = modalLines[:availableHeight]
		actualModalHeight = availableHeight
	}

	modalX := max((availableWidth-modalWidth)/2, 0)
	modalY := max((availableHeight-actualModalHeight)/2, 0)

	m.modalState.Position.X = modalX
	m.modalState.Position.Y = modalY
	m.modalState.Position.Width = modalWidth
	m.modalState.Position.Height = actualModalHeight

	lines := strings.Split(dimmedBackground, "\n")
	for len(lines) < availableHeight {
		lines = append(lines, overlayStyle.Render(""))
	}

	for i := range lines {
		lines[i] = padLine(lines[i], availableWidth)
	}

	for i, modalLine := range modalLines {
		row := modalY + i
		if row >= len(lines) {
			lines = append(lines, padLine("", availableWidth))
		}

		left := truncate.String(lines[row], uint(modalX))
		body := lipgloss.NewStyle().Width(modalWidth).Render(modalLine)
		right := sliceFromColumn(lines[row], modalX+modalWidth)

		lines[row] = left + body + right
	}

	if len(lines) > availableHeight {
		lines = lines[:availableHeight]
	}

	return strings.Join(lines, "\n")
}

func padLine(line string, width int) string {
	diff := width - lipgloss.Width(line)
	if diff <= 0 {
		return line
	}
	return line + strings.Repeat(" ", diff)
}

func sliceFromColumn(line string, x int) string {
	if x <= 0 {
		return line
	}

	var (
		ansiSequences strings.Builder
		posX          int
		inANSI        bool
	)

	for i, r := range line {
		if r == ansi.Marker {
			ansiSequences.WriteRune(r)
			inANSI = true
			continue
		}

		if inANSI {
			ansiSequences.WriteRune(r)
			if ansi.IsTerminator(r) {
				inANSI = false
			}
			continue
		}

		posX += runewidth.RuneWidth(r)
		if posX >= x {
			return ansiSequences.String() + line[i:]
		}
	}

	return ""
}

// renderSortModal renders the sort options modal
func (m *ScanUI) renderSortModal() string {
	var b strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Width(30).
		Render("📊 SORT OPTIONS")
	b.WriteString(title + "\n\n")

	// Options
	options := []string{
		"1. Port (ascending)",
		"2. Port (descending)",
		"3. Host (A → Z)",
		"4. State (Open → Closed → Filtered)",
		"5. Service (alphabetical)",
		"6. Latency (fastest first)",
		"7. Latency (slowest first)",
		"8. Discovery order (original)",
	}

	for i, option := range options {
		style := lipgloss.NewStyle()
		if i == m.modalState.Cursor {
			style = style.Background(m.theme.Primary).Foreground(m.theme.Background)
		}
		b.WriteString(style.Render(option) + "\n")
	}

	// Current selection
	b.WriteString("\n")
	current := lipgloss.NewStyle().
		Foreground(m.theme.Secondary).
		Render("Current: " + m.sortState.GetModeString())
	b.WriteString(current + "\n")

	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Render("↑/↓: Navigate • Enter: Select • ESC: Cancel")
	b.WriteString("\n" + instructions)

	return b.String()
}

// renderDetailsModal renders the details view for a selected result
func (m *ScanUI) renderDetailsModal() string {
	if len(m.displayResults) == 0 {
		return "No results to display"
	}

	selectedResult := m.displayResults[m.table.Cursor()]

	// Calculate available content area for scrolling
	availableHeight := maxModalContentHeight - 10 // Account for title/borders

	// Build full content
	var fullContent strings.Builder

	// Title with host and port
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Render(fmt.Sprintf("📋 %s:%d (%s)", selectedResult.Host, selectedResult.Port, selectedResult.State))
	fullContent.WriteString(title + "\n\n")

	// Host information
	section := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Secondary).
		Render("🌐 Host Information")
	fullContent.WriteString(section + "\n")
	service := getServiceName(selectedResult.Port)
	hostInfo := fmt.Sprintf("  Host: %s\n  Port: %d/%s\n  State: %s\n  Service: %s",
		selectedResult.Host, selectedResult.Port, selectedResult.Protocol,
		selectedResult.State, service)
	fullContent.WriteString(hostInfo + "\n\n")

	// Banner information (scrollable)
	if selectedResult.Banner != "" {
		section = lipgloss.NewStyle().
			Bold(true).
			Foreground(m.theme.Secondary).
			Render("🏷️  Service Banner")
		fullContent.WriteString(section + "\n")

		// Keep original formatting for banner (don't truncate)
		bannerLines := strings.Split(selectedResult.Banner, "\n")
		for _, line := range bannerLines {
			fullContent.WriteString("  " + strings.TrimSpace(line) + "\n")
		}
		fullContent.WriteString("\n")
	}

	// Performance information
	section = lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Secondary).
		Render("⚡ Performance Metrics")
	fullContent.WriteString(section + "\n")
	perfInfo := fmt.Sprintf("  Latency: %s\n  Protocol: %s\n  Scanned: %s ago",
		formatDuration(selectedResult.Duration),
		selectedResult.Protocol,
		selectedResult.Duration.Round(time.Second).String())
	fullContent.WriteString(perfInfo + "\n")

	// Network analysis
	section = lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Secondary).
		Render("🔍 Network Analysis")
	fullContent.WriteString(section + "\n")

	// Check if it's a common service port
	correctService := getServiceName(selectedResult.Port)
	serviceAnalysis := fmt.Sprintf("  Expected Service: %s", correctService)
	// Note: We can't check for service mismatch since ResultEvent doesn't contain detected service
	serviceAnalysis += " (expected)"

	// Categorize port state
	stateAnalysis := ""
	switch selectedResult.State {
	case core.StateOpen:
		stateAnalysis = "  🔓 Port is Open (listening for connections)"
	case core.StateClosed:
		stateAnalysis = "  🔒 Port is Closed (not accepting connections)"
	case core.StateFiltered:
		stateAnalysis = "  🚫 Port is Filtered (blocked by firewall)"
	}

	fullContent.WriteString(serviceAnalysis + "\n" + stateAnalysis + "\n")

	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Render("↑/↓: Scroll • ESC: Return to main view")
	fullContent.WriteString("\n" + instructions)

	// Track content height for scrolling
	contentLines := strings.Split(fullContent.String(), "\n")
	m.modalState.MaxScrollHeight = len(contentLines)

	// Apply scrolling
	if availableHeight > 0 && len(contentLines) > availableHeight {
		// Scrolling needed - show only visible portion
		start := m.modalState.ScrollPosition
		end := min(start+availableHeight, len(contentLines))
		visibleLines := contentLines[start:end]

		scrollIndicator := ""
		if start > 0 {
			scrollIndicator += "▲"
		} else {
			scrollIndicator += " "
		}
		if end < len(contentLines) {
			scrollIndicator += "▼"
		} else {
			scrollIndicator += " "
		}

		scrollStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)
		scrollBar := scrollStyle.Render(fmt.Sprintf("Lines %d-%d of %d %s",
			start+1, end, len(contentLines), scrollIndicator))

		return strings.Join(visibleLines, "\n") + "\n\n" + scrollBar
	}

	// No scrolling needed
	return fullContent.String()
}

// Helper function for modal content height calculation
func getMaxModalContentHeight() int {
	return 20 // Default maximum content lines for modal
}

// Add the constant at the top
const (
	maxModalContentHeight = 20
)

// renderDashboardView renders the split view dashboard
func (m *ScanUI) renderDashboardView() string {
	var b strings.Builder

	// Header
	breadcrumb := m.renderBreadcrumb()
	b.WriteString(breadcrumb + "\n")

	header := m.renderHeader()
	b.WriteString(header + "\n")

	if m.scanning {
		progressView := m.renderProgress()
		b.WriteString(progressView + "\n")
	}

	// Calculate split dimensions
	leftWidth := int(float64(m.width) * 0.65)
	rightWidth := m.width - leftWidth - 3 // -3 for spacing

	// Left side: Results table
	tableView := m.table.View()

	// Right side: Stats panel
	statsPanel := m.renderStatsPanel(rightWidth)

	// Join horizontally
	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(m.height - 8)

	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(m.height - 8).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary).
		Padding(1)

	leftContent := leftStyle.Render(tableView)
	rightContent := rightStyle.Render(statsPanel)

	dashboard := lipgloss.JoinHorizontal(lipgloss.Top, leftContent, " ", rightContent)
	b.WriteString(dashboard + "\n")

	// Footer
	footer := m.renderFooter()
	b.WriteString(footer)

	return b.String()
}

// renderStatsPanel renders the statistics panel content
func (m *ScanUI) renderStatsPanel(width int) string {
	if m.statsData == nil {
		m.statsData = m.computeStats()
	}

	stats := m.statsData
	var b strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Render("📊 Live Statistics")
	b.WriteString(title + "\n\n")

	// Port State Distribution
	b.WriteString(m.renderMiniBarChart() + "\n\n")

	// Top Services
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Secondary)

	b.WriteString(sectionStyle.Render("Top Services:") + "\n")
	if len(stats.TopServices) > 0 {
		for i, svc := range stats.TopServices {
			b.WriteString(fmt.Sprintf("  %d. %-12s %3d ports\n", i+1, svc.Name, svc.Count))
		}
	} else {
		b.WriteString("  No services detected yet\n")
	}
	b.WriteString("\n")

	// Performance Metrics
	b.WriteString(sectionStyle.Render("Performance:") + "\n")
	b.WriteString(fmt.Sprintf("  Current:  %7.0f pps\n", stats.CurrentRate))
	b.WriteString(fmt.Sprintf("  Average:  %7.0f pps\n", stats.AverageRate))
	b.WriteString("\n")

	// Sparklines
	b.WriteString(m.renderSparklines() + "\n")

	// Response Time Statistics
	if stats.AvgResponseTime > 0 {
		b.WriteString(sectionStyle.Render("Response Times:") + "\n")
		b.WriteString(fmt.Sprintf("  Min:  %s\n", formatDuration(stats.MinResponseTime)))
		b.WriteString(fmt.Sprintf("  Avg:  %s\n", formatDuration(stats.AvgResponseTime)))
		b.WriteString(fmt.Sprintf("  P95:  %s\n", formatDuration(stats.P95ResponseTime)))
		b.WriteString(fmt.Sprintf("  Max:  %s\n", formatDuration(stats.MaxResponseTime)))
		b.WriteString("\n")
	}

	// Network Overview
	b.WriteString(sectionStyle.Render("Network Overview:") + "\n")
	b.WriteString(fmt.Sprintf("  Hosts scanned:    %d\n", stats.UniqueHosts))
	b.WriteString(fmt.Sprintf("  Hosts with open:  %d\n", stats.HostsWithOpen))
	b.WriteString(fmt.Sprintf("  Unique services:  %d\n", len(stats.ServiceCounts)))

	return b.String()
}

// renderMiniBarChart renders an ASCII bar chart for port states
func (m *ScanUI) renderMiniBarChart() string {
	if m.statsData == nil {
		return ""
	}

	stats := m.statsData
	total := float64(stats.TotalResults)
	if total == 0 {
		total = 1
	}

	// Calculate bar lengths (max 30 chars)
	maxBarWidth := 30
	openBar := int((float64(stats.OpenCount) / total) * float64(maxBarWidth))
	closedBar := int((float64(stats.ClosedCount) / total) * float64(maxBarWidth))
	filteredBar := int((float64(stats.FilteredCount) / total) * float64(maxBarWidth))

	// Color styles
	openStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	closedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	filteredStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	labelStyle := lipgloss.NewStyle().Width(StatusBarLabelWidth)

	var b strings.Builder

	// Open ports
	openPct := getPercentage(stats.OpenCount, stats.TotalResults)
	b.WriteString(labelStyle.Render("Open:") + " ")
	b.WriteString(openStyle.Render(strings.Repeat("█", openBar)))
	b.WriteString(fmt.Sprintf(" %d (%.1f%%)\n", stats.OpenCount, openPct))

	// Closed ports
	closedPct := getPercentage(stats.ClosedCount, stats.TotalResults)
	b.WriteString(labelStyle.Render("Closed:") + " ")
	b.WriteString(closedStyle.Render(strings.Repeat("█", closedBar)))
	b.WriteString(fmt.Sprintf(" %d (%.1f%%)\n", stats.ClosedCount, closedPct))

	// Filtered ports
	filteredPct := getPercentage(stats.FilteredCount, stats.TotalResults)
	b.WriteString(labelStyle.Render("Filtered:") + " ")
	b.WriteString(filteredStyle.Render(strings.Repeat("█", filteredBar)))
	b.WriteString(fmt.Sprintf(" %d (%.1f%%)", stats.FilteredCount, filteredPct))

	return b.String()
}

// renderSparklines renders the sparkline charts for the dashboard
func (m *ScanUI) renderSparklines() string {
	if m.sparklineData == nil {
		return ""
	}

	var b strings.Builder

	// Scan Rate Sparkline
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Secondary)

	b.WriteString(sectionStyle.Render("Scan Rate (60s):") + "\n")
	scanRateSparkline := m.sparklineData.RenderSparkline(m.sparklineData.ScanRate, 20)
	summary := m.sparklineData.GetMetricSummary()

	sparklineStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
	b.WriteString("  " + sparklineStyle.Render(scanRateSparkline) + "\n")
	b.WriteString(fmt.Sprintf("  Cur: %0.1f • Avg: %0.1f • Peak: %0.1f pps\n",
		summary.CurrentScanRate, summary.AverageScanRate, summary.PeakScanRate))
	b.WriteString("\n")

	// Discovery Rate Sparkline (only show if we have data)
	if len(m.sparklineData.DiscoveryRate) > 0 {
		b.WriteString(sectionStyle.Render("Discovery Rate (60s):") + "\n")
		discoverySparkline := m.sparklineData.RenderSparkline(m.sparklineData.DiscoveryRate, 20)
		b.WriteString("  " + sparklineStyle.Render(discoverySparkline) + "\n")
		b.WriteString(fmt.Sprintf("  Cur: %0.1f • Avg: %0.1f pps\n",
			summary.CurrentDiscoveryRate, summary.AverageDiscoveryRate))
	}

	return b.String()
}
