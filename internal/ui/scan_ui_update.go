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
	m.table.SetHeight(m.height - 12)
	m.table.SetWidth(m.width)
}

func (m *ScanUI) handleKeyMsg(msg tea.KeyMsg) (handled bool, skipTable bool, cmd tea.Cmd) {
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

	switch m.viewState {
	case UIViewHelp:
		return m.handleHelpKey(msg)
	case UIViewMain:
		return m.handleMainKey(msg)
	case UIViewSortMenu:
		return m.handleSortMenuKey(msg)
	case UIViewFilterMenu:
		return m.handleFilterMenuKey(msg)
	default:
		return false, false, nil
	}
}

func (m *ScanUI) handleHelpKey(msg tea.KeyMsg) (bool, bool, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) || key.Matches(msg, m.keys.Help) {
		m.showHelp = false
		m.viewState = UIViewMain
	}
	return true, true, nil
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
		m.viewState = UIViewSortMenu
		return true, true, nil
	case key.Matches(msg, m.keys.Filter):
		m.viewState = UIViewFilterMenu
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
	case key.Matches(msg, m.keys.Up):
		m.table.MoveUp(1)
		return true, true, nil
	case key.Matches(msg, m.keys.Down):
		m.table.MoveDown(1)
		return true, true, nil
	case key.Matches(msg, m.keys.PageUp):
		m.table.MoveUp(10)
		return true, true, nil
	case key.Matches(msg, m.keys.PageDown):
		m.table.MoveDown(10)
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

func (m *ScanUI) handleSortMenuKey(msg tea.KeyMsg) (bool, bool, tea.Cmd) {
	switch msg.String() {
	case "1":
		m.sortState.SetMode(SortByPort)
	case "2":
		m.sortState.SetMode(SortByPortDesc)
	case "3":
		m.sortState.SetMode(SortByHost)
	case "4":
		m.sortState.SetMode(SortByState)
	case "5":
		m.sortState.SetMode(SortByService)
	case "6":
		m.sortState.SetMode(SortByLatency)
	case "7":
		m.sortState.SetMode(SortByLatencyDesc)
	case "8":
		m.sortState.SetMode(SortByDiscovery)
	case "q", "esc":
		m.viewState = UIViewMain
		return true, true, nil
	default:
		return true, true, nil
	}

	m.viewState = UIViewMain
	m.updateTable()
	return true, true, nil
}

func (m *ScanUI) handleFilterMenuKey(msg tea.KeyMsg) (bool, bool, tea.Cmd) {
	switch msg.String() {
	case "1":
		m.filterState.SetStateFilter(StateFilterAll)
	case "2":
		m.filterState.SetStateFilter(StateFilterOpen)
	case "3":
		m.filterState.SetStateFilter(StateFilterClosed)
	case "4":
		m.filterState.SetStateFilter(StateFilterFiltered)
	case "q", "esc":
		m.viewState = UIViewMain
		return true, true, nil
	default:
		return true, true, nil
	}

	m.viewState = UIViewMain
	m.updateTable()
	return true, true, nil
}

func (m *ScanUI) handleScanResult(msg scanResultMsg) {
	m.results = append(m.results, msg.result)
	m.updateTable()
	m.updateStats(msg.result)
}

func (m *ScanUI) handleScanProgress(msg scanProgressMsg) {
	m.currentRate = msg.progress.Rate
	m.progressTrack.Update(
		len(m.results),
		m.openCount,
		m.closedCount,
		m.filteredCount,
		m.currentRate,
	)
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
	filtered := m.filterState.ApplyFilters(m.results)
	m.displayResults = m.sortState.ApplySort(filtered)

	var rows []table.Row
	for _, r := range m.displayResults {
		service := getServiceName(r.Port)
		banner := r.Banner
		if len(banner) > 40 {
			banner = banner[:37] + "..."
		}

		stateDisplay := string(r.State)
		if r.State == core.StateOpen {
			stateDisplay = lipgloss.NewStyle().Foreground(m.theme.Success).Render(stateDisplay)
		}

		protocol := r.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		protocol = strings.ToUpper(protocol)

		rows = append(rows, table.Row{
			r.Host,
			fmt.Sprintf("%d", r.Port),
			protocol,
			stateDisplay,
			service,
			banner,
			fmt.Sprintf("%dms", r.Duration.Milliseconds()),
		})
	}
	m.table.SetRows(rows)
}

func (m *ScanUI) updateStats(result core.ResultEvent) {
	switch result.State {
	case core.StateOpen:
		m.openCount++
	case core.StateClosed:
		m.closedCount++
	case core.StateFiltered:
		m.filteredCount++
	}
}

func (m *ScanUI) listenForResults() tea.Cmd {
	return func() tea.Msg {
		select {
		case event, ok := <-m.resultChan:
			if !ok {
				return scanCompleteMsg{}
			}

			switch event.Type {
			case core.EventTypeResult:
				return scanResultMsg{result: event.Result}
			case core.EventTypeProgress:
				return scanProgressMsg{progress: event.Progress}
			case core.EventTypeError:
				return scanCompleteMsg{}
			}
		case <-time.After(100 * time.Millisecond):
		}
		return nil
	}
}
