package ui

import (
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

Controls:
  p / Space  Pause/resume
  ?          Toggle help
  Ctrl+L     Clear screen
  q / Esc    Quit
`

	return helpStyle.Render(content)
}
