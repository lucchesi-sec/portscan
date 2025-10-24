package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the UI
func (m *ScanUI) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	if m.showHelp {
		return m.renderHelp()
	}

	if m.viewState == UIViewSortMenu {
		return m.renderSortMenu()
	}

	if m.viewState == UIViewFilterMenu {
		return m.renderFilterMenu()
	}

	return m.renderMain()
}

func (m *ScanUI) renderMain() string {
	// Check if dashboard view is enabled
	if m.showDashboard && m.width >= 120 {
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

	return style.Render(location)
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
	return m.progressBar.ViewAs(progress)
}

func (m *ScanUI) renderStatus() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(m.theme.Foreground)

	status := m.progressTrack.GetStatusLine()
	details := m.progressTrack.GetDetailedStats()

	return statusStyle.Render(status + "\n" + details)
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

func (m *ScanUI) renderSortMenu() string {
	menuStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary)

	content := `📊 SORT OPTIONS

1. Port (ascending)
2. Port (descending)
3. Host (A → Z)
4. State (Open → Closed → Filtered)
5. Service (alphabetical)
6. Latency (fastest first)
7. Latency (slowest first)
8. Discovery order (original)

Current: ` + m.sortState.GetModeString() + `

Press number to select or ESC to cancel`

	return menuStyle.Render(content)
}

func (m *ScanUI) renderFilterMenu() string {
	menuStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary)

	var current string
	switch m.filterState.StateFilter {
	case StateFilterAll:
		current = "All States"
	case StateFilterOpen:
		current = "Open Only"
	case StateFilterClosed:
		current = "Closed Only"
	case StateFilterFiltered:
		current = "Filtered Only"
	}

	content := `🔍 FILTER OPTIONS

1. Show All States
2. Show Open Ports Only
3. Show Closed Ports Only
4. Show Filtered Ports Only

Current: ` + current + `

Press number to select or ESC to cancel`

	return menuStyle.Render(content)
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
  /          Search banners

Filtering & Sorting:
  s          Sort menu
  f          Filter menu
  r          Reset filters
  o          Toggle open-only

View Controls:
  D          Toggle dashboard view
  p / Space  Pause/resume
  ?          Toggle help
  Ctrl+L     Clear screen
  q / Esc    Quit
`

	return helpStyle.Render(content)
}

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
	labelStyle := lipgloss.NewStyle().Width(10)

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
