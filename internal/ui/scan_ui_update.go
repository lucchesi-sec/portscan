package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/pkg/theme"
	"github.com/muesli/reflow/truncate"
)

// Update handles messages
func (m *ScanUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	skipTableUpdate := false

	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleWindowSize(typed)

	case tea.KeyMsg:
		handled, skip, cmd := m.handleKeyMsg(typed)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if handled {
			skipTableUpdate = skip
			break
		}

	case scanResultMsg:
		m.handleScanResult(typed)
		skipTableUpdate = true

	case scanProgressMsg:
		m.handleScanProgress(typed)
		skipTableUpdate = true

	case scanCompleteMsg:
		m.scanning = false
		skipTableUpdate = true

	case spinner.TickMsg:
		if cmd := m.handleSpinnerTick(typed); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case progress.FrameMsg:
		if cmd := m.handleProgressFrame(typed); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if !skipTableUpdate {
		if cmd := m.updateTableModel(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if m.scanning {
		cmds = append(cmds, m.listenForResults())
	}

	return m, tea.Batch(cmds...)
}

func (m *ScanUI) handleWindowSize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	m.applyTableGeometry()
}

func (m *ScanUI) handleKeyMsg(msg tea.KeyMsg) (handled bool, skipTable bool, cmd tea.Cmd) {
	// Handle modal escape key globally if modal is active
	if m.modalState.IsActive && key.Matches(msg, m.keys.Escape) {
		m.modalState.IsActive = false
		m.modalState.Cursor = 0
		return true, true, nil
	}

	if key.Matches(msg, m.keys.Help) {
		m.showHelp = !m.showHelp
		m.help.ShowAll = m.showHelp
		if m.showHelp {
			m.viewState = UIViewHelp
		} else {
			m.viewState = UIViewMain
		}
		return true, true, nil
	}

	// Handle modal key events
	if m.modalState.IsActive {
		return m.handleModalKey(msg)
	}

	// Handle main view key events
	switch m.viewState {
	case UIViewHelp:
		return m.handleHelpKey(msg)
	case UIViewMain:
		return m.handleMainKey(msg)
	case UIViewSortModal:
		// Legacy support - convert to modal
		m.openModal(ModalSort)
		m.viewState = UIViewMain
		return true, true, nil
	default:
		return false, false, nil
	}
}

func (m *ScanUI) handleModalKey(msg tea.KeyMsg) (bool, bool, tea.Cmd) {
	switch m.modalState.Type {
	case ModalSort:
		return m.handleSortModalKey(msg)
	case ModalDetails:
		return m.handleDetailsModalKey(msg)
	default:
		return true, true, nil
	}
}

func (m *ScanUI) handleSortModalKey(msg tea.KeyMsg) (bool, bool, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.modalState.Cursor = max(0, m.modalState.Cursor-1)
		return true, true, nil
	case "down", "j":
		m.modalState.Cursor = min(7, m.modalState.Cursor+1)
		return true, true, nil
	case "enter":
		switch m.modalState.Cursor {
		case 0:
			m.sortState.SetMode(SortByPort)
		case 1:
			m.sortState.SetMode(SortByPortDesc)
		case 2:
			m.sortState.SetMode(SortByHost)
		case 3:
			m.sortState.SetMode(SortByState)
		case 4:
			m.sortState.SetMode(SortByService)
		case 5:
			m.sortState.SetMode(SortByLatency)
		case 6:
			m.sortState.SetMode(SortByLatencyDesc)
		case 7:
			m.sortState.SetMode(SortByDiscovery)
		}
		m.updateTable()
		m.modalState.IsActive = false
		m.modalState.Cursor = 0
		return true, true, nil
	default:
		return true, true, nil
	}
}

func (m *ScanUI) handleDetailsModalKey(msg tea.KeyMsg) (bool, bool, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		// Scroll up if content is scrollable
		if m.modalState.ScrollPosition > 0 {
			m.modalState.ScrollPosition--
		}
		return true, true, nil
	case "down", "j":
		// Scroll down if there's more content
		maxScroll := max(0, m.modalState.MaxScrollHeight-maxModalContentHeight)
		if m.modalState.ScrollPosition < maxScroll {
			m.modalState.ScrollPosition++
		}
		return true, true, nil
	default:
		return true, true, nil
	}
}

func (m *ScanUI) handleHelpKey(msg tea.KeyMsg) (bool, bool, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) || key.Matches(msg, m.keys.Help) {
		m.showHelp = false
		m.viewState = UIViewMain
	}
	return true, true, nil
}

// Helper functions
func (m *ScanUI) openModal(modalType ModalType) {
	m.modalState.IsActive = true
	m.modalState.Type = modalType
	m.modalState.Cursor = 0
	m.modalState.ScrollPosition = 0
	m.modalState.MaxScrollHeight = 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *ScanUI) handleMainKey(msg tea.KeyMsg) (bool, bool, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return true, true, tea.Quit
	case key.Matches(msg, m.keys.Pause):
		if m.scanning {
			m.isPaused = !m.isPaused
			if m.isPaused {
				m.progressTrack.Pause()
			} else {
				m.progressTrack.Resume()
			}
		}
		return true, true, nil
	case key.Matches(msg, m.keys.Clear):
		return true, true, tea.ClearScreen
	case key.Matches(msg, m.keys.Sort):
		m.openModal(ModalSort)
		return true, true, nil
	case key.Matches(msg, m.keys.Enter):
		if len(m.displayResults) > 0 {
			m.openModal(ModalDetails)
		}
		return true, true, nil
	case key.Matches(msg, m.keys.Reset):
		m.filterState.Reset()
		m.sortState = NewSortState()
		m.updateTable()
		return true, true, nil
	case key.Matches(msg, m.keys.OpenOnly):
		if m.filterState.StateFilter == StateFilterOpen {
			m.filterState.SetStateFilter(StateFilterAll)
		} else {
			m.filterState.SetStateFilter(StateFilterOpen)
		}
		m.updateTable()
		return true, true, nil
	case key.Matches(msg, m.keys.ToggleDashboard):
		m.showDashboard = !m.showDashboard
		m.applyTableGeometry()
		// Recompute stats when showing dashboard
		if m.showDashboard {
			m.statsData = m.computeStats()
		}
		return true, true, nil
	case key.Matches(msg, m.keys.Up):
		m.table.MoveUp(1)
		return true, true, nil
	case key.Matches(msg, m.keys.Down):
		m.table.MoveDown(1)
		return true, true, nil
	case key.Matches(msg, m.keys.PageUp):
		m.table.MoveUp(PageScrollLines)
		return true, true, nil
	case key.Matches(msg, m.keys.PageDown):
		m.table.MoveDown(PageScrollLines)
		return true, true, nil
	case key.Matches(msg, m.keys.Home):
		m.table.GotoTop()
		return true, true, nil
	case key.Matches(msg, m.keys.End):
		m.table.GotoBottom()
		return true, true, nil
	default:
		return false, false, nil
	}
}

func (m *ScanUI) handleScanResult(msg scanResultMsg) {
	m.results.Append(msg.result)
	m.stats.Add(msg.result)
	m.updateTable()
	total, open, closed, filtered := m.stats.Totals()

	// Calculate host metrics
	hostsScanned := m.calculateHostsScanned()

	// Update progress with host tracking
	m.progressTrack.Update(total, open, closed, filtered, m.currentRate)
	m.progressTrack.UpdateHosts(m.calculateTotalHosts(), hostsScanned)

	// Update dashboard stats if visible
	if m.showDashboard {
		m.statsData = m.computeStats()
	}
}

// calculateHostsScanned determines how many unique hosts have been scanned
func (m *ScanUI) calculateHostsScanned() int {
	hosts := make(map[string]bool)
	results := m.results.Items()

	for _, result := range results {
		hosts[result.Host] = true
	}

	return len(hosts)
}

// calculateTotalHosts returns the total number of unique hosts being scanned
// This is an estimation based on the configuration and known targets
func (m *ScanUI) calculateTotalHosts() int {
	hosts := make(map[string]bool)
	results := m.results.Items()

	for _, result := range results {
		hosts[result.Host] = true
	}

	return len(hosts)
}

func (m *ScanUI) handleScanProgress(msg scanProgressMsg) {
	m.currentRate = msg.progress.Rate
	if msg.progress.Total > 0 {
		m.progressTrack.TotalPorts = msg.progress.Total
	}

	total, open, closed, filtered := m.stats.Totals()
	scanned := msg.progress.Completed
	if scanned < total {
		scanned = total
	}

	// Calculate host metrics
	hostsScanned := m.calculateHostsScanned()
	totalHosts := m.calculateTotalHosts()

	m.progressTrack.Update(
		scanned,
		open,
		closed,
		filtered,
		m.currentRate,
	)
	m.progressTrack.UpdateHosts(totalHosts, hostsScanned)

	// Update sparkline data
	if m.sparklineData != nil {
		m.sparklineData.AddScanRate(msg.progress.Rate)

		// Calculate discovery rate (new open ports since last update)
		newlyOpen := open - m.previousOpenCount
		discoveryRate := 0.0

		// Estimate rate based on scan rate (assumes proportional discovery)
		if msg.progress.Rate > 0 && scanned > 0 {
			discoveryRate = float64(newlyOpen) * msg.progress.Rate / float64(scanned-open)
			if discoveryRate < 0 {
				discoveryRate = 0
			}
		}

		m.sparklineData.AddDiscoveryRate(discoveryRate)
		m.previousOpenCount = open

		// For now, error rate is 0 (we don't track errors explicitly yet)
		m.sparklineData.AddErrorRate(0.0)
	}
}

func (m *ScanUI) handleSpinnerTick(msg spinner.TickMsg) tea.Cmd {
	if !m.scanning || m.isPaused {
		return nil
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return cmd
}

func (m *ScanUI) handleProgressFrame(msg progress.FrameMsg) tea.Cmd {
	progressModel, cmd := m.progressBar.Update(msg)
	m.progressBar = progressModel.(progress.Model)
	return cmd
}

func (m *ScanUI) updateTableModel(msg tea.Msg) tea.Cmd {
	if m.isNavigationKey(msg) {
		return nil
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return cmd
}

func (m *ScanUI) isNavigationKey(msg tea.Msg) bool {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return false
	}

	return key.Matches(keyMsg, m.keys.Up) ||
		key.Matches(keyMsg, m.keys.Down) ||
		key.Matches(keyMsg, m.keys.PageUp) ||
		key.Matches(keyMsg, m.keys.PageDown) ||
		key.Matches(keyMsg, m.keys.Home) ||
		key.Matches(keyMsg, m.keys.End)
}

func (m *ScanUI) updateTable() {
	baseResults := m.results.Items()
	filtered := m.filterState.ApplyFilters(baseResults)
	m.displayResults = m.sortState.ApplySort(filtered)
	m.applyTableGeometry()

	stateColors := m.theme.GetStateColors()
	columns := m.table.Columns()
	if len(columns) != len(defaultColumnSpecs) {
		columns = calculateColumnWidths(m.tableViewportWidth())
	}

	widthFor := func(idx int) int {
		if idx < len(columns) {
			return columns[idx].Width
		}
		return 0
	}

	var rows []table.Row
	for _, r := range m.displayResults {
		rowStyle := m.theme.GetRowStyle(string(r.State))

		service := getServiceName(r.Port)
		banner := r.Banner
		stateDisplay := m.getRowStateDisplay(r, stateColors)

		protocol := r.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		protocol = strings.ToUpper(protocol)

		hostCell := rowStyle.Render(truncateToWidth(r.Host, widthFor(0)))
		portCell := rowStyle.Render(truncateToWidth(fmt.Sprintf("%d", r.Port), widthFor(1)))
		protocolCell := rowStyle.Render(truncateToWidth(protocol, widthFor(2)))
		stateCell := truncateStyled(stateDisplay, widthFor(3))
		serviceCell := rowStyle.Render(truncateToWidth(service, widthFor(4)))
		bannerCell := rowStyle.Render(truncateToWidth(banner, widthFor(5)))
		latencyCell := rowStyle.Render(truncateToWidth(fmt.Sprintf("%dms", r.Duration.Milliseconds()), widthFor(6)))

		row := table.Row{
			hostCell,
			portCell,
			protocolCell,
			stateCell,
			serviceCell,
			bannerCell,
			latencyCell,
		}

		rows = append(rows, row)
	}

	m.table.SetRows(rows)
}

func truncateToWidth(content string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(content) <= width {
		return content
	}
	return truncate.StringWithTail(content, uint(width), "…")
}

func truncateStyled(content string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(content) <= width {
		return content
	}
	return truncate.StringWithTail(content, uint(width), "…")
}

func (m *ScanUI) getRowStateDisplay(result core.ResultEvent, colors theme.StateColors) string {
	stateStyle := lipgloss.NewStyle()
	switch result.State {
	case core.StateOpen:
		stateStyle = stateStyle.Foreground(colors.Open)
	case core.StateClosed:
		stateStyle = stateStyle.Foreground(colors.Closed)
	case core.StateFiltered:
		stateStyle = stateStyle.Foreground(colors.Filtered)
	}
	return stateStyle.Render(string(result.State))
}

func (m *ScanUI) listenForResults() tea.Cmd {
	return func() tea.Msg {
		select {
		case event, ok := <-m.resultChan:
			if !ok {
				return scanCompleteMsg{}
			}

			switch event.Kind {
			case core.EventKindResult:
				return scanResultMsg{result: *event.Result}
			case core.EventKindProgress:
				return scanProgressMsg{progress: *event.Progress}
			case core.EventKindError:
				return scanCompleteMsg{}
			}
		case <-time.After(ResultPollTimeout):
		}
		return nil
	}
}
