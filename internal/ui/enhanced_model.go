package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/pkg/config"
	"github.com/lucchesi-sec/portscan/pkg/theme"
)

type ViewMode int

const (
	ViewTable ViewMode = iota
	ViewDetails
	ViewStats
	ViewLogs
)

type EnhancedModel struct {
	// Core components
	config     *config.Config
	theme      theme.Theme
	results    []core.ResultEvent
	resultChan <-chan interface{}

	// View state
	viewMode ViewMode
	width    int
	height   int

	// Widgets
	table       table.Model
	list        list.Model
	viewport    viewport.Model
	progress    progress.Model
	spinner     spinner.Model
	stopwatch   stopwatch.Model
	filterInput textinput.Model
	help        help.Model
	keys        keyMap

	// State
	scanning         bool
	filtering        bool
	filterText       string
	selectedPort     *core.ResultEvent
	logs             []string
	stats            ScanStats
	showOnlyOpenPorts bool // Control whether to show only open ports or all results
}

type ScanStats struct {
	TotalPorts     int
	OpenPorts      int
	ClosedPorts    int
	FilteredPorts  int
	StartTime      time.Time
	EndTime        time.Time
	AverageLatency time.Duration
	FastestPort    uint16
	SlowestPort    uint16
}

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	Enter      key.Binding
	Tab        key.Binding
	Filter     key.Binding
	Details    key.Binding
	Stats      key.Binding
	Logs       key.Binding
	Export     key.Binding
	Help       key.Binding
	Quit       key.Binding
	ToggleView key.Binding // Toggle between showing all ports and only open ports
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Filter, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Enter, k.Tab, k.Filter, k.ToggleView},
		{k.Details, k.Stats, k.Logs},
		{k.Export, k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "right"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch view"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Details: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "details"),
	),
	Stats: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "stats"),
	),
	Logs: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "logs"),
	),
	Export: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "export"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	ToggleView: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "toggle all/open ports"),
	),
}

func NewEnhanced(cfg *config.Config, results <-chan interface{}) *EnhancedModel {
	// Get theme
	t := theme.GetTheme(cfg.UI.Theme)

	// Initialize table
	columns := []table.Column{
		{Title: "Port", Width: 8},
		{Title: "State", Width: 10},
		{Title: "Service", Width: 15},
		{Title: "Banner", Width: 30},
		{Title: "Latency", Width: 10},
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Table styling
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.Primary).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(t.Background).
		Background(t.Primary).
		Bold(false)
	tbl.SetStyles(s)

	// Initialize list for filtered results
	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Scan Results"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = l.Styles.Title.Foreground(t.Primary).Bold(true)

	// Initialize viewport for details/logs
	vp := viewport.New(50, 10)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary)

	// Initialize progress bar
	prog := progress.New(progress.WithDefaultGradient())

	// Initialize spinner
	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(t.Primary)

	// Initialize stopwatch
	sw := stopwatch.NewWithInterval(time.Second)

	// Initialize filter input
	fi := textinput.New()
	fi.Placeholder = "Filter ports..."
	fi.CharLimit = 50
	fi.Width = 30
	fi.Prompt = "🔍 "

	// Initialize help
	h := help.New()
	h.ShowAll = false

	return &EnhancedModel{
		config:           cfg,
		theme:            t,
		results:          []core.ResultEvent{},
		resultChan:       results,
		viewMode:         ViewTable,
		table:            tbl,
		list:             l,
		viewport:         vp,
		progress:         prog,
		spinner:          sp,
		stopwatch:        sw,
		filterInput:      fi,
		help:             h,
		keys:             keys,
		scanning:         true,
		logs:             []string{},
		showOnlyOpenPorts: true, // Default to showing only open ports
		stats: ScanStats{
			StartTime: time.Now(),
		},
	}
}

func (m *EnhancedModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.stopwatch.Init(),
		m.listenForResults(),
	)
}

func (m *EnhancedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()

	case tea.KeyMsg:
		if m.filtering {
			switch msg.String() {
			case "enter":
				m.filterText = m.filterInput.Value()
				m.applyFilter()
				m.filtering = false
				m.filterInput.Blur()
			case "esc":
				m.filtering = false
				m.filterInput.Blur()
				m.filterInput.SetValue("")
			default:
				var cmd tea.Cmd
				m.filterInput, cmd = m.filterInput.Update(msg)
				cmds = append(cmds, cmd)
			}
		} else {
			switch {
			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			case key.Matches(msg, m.keys.Tab):
				m.cycleView()
			case key.Matches(msg, m.keys.Filter):
				m.filtering = true
				m.filterInput.Focus()
				cmds = append(cmds, textinput.Blink)
			case key.Matches(msg, m.keys.Details):
				m.viewMode = ViewDetails
				m.updateViewport()
			case key.Matches(msg, m.keys.Stats):
				m.viewMode = ViewStats
				m.updateViewport()
			case key.Matches(msg, m.keys.Logs):
				m.viewMode = ViewLogs
				m.updateViewport()
			case key.Matches(msg, m.keys.Enter):
				m.selectCurrentItem()
			case key.Matches(msg, m.keys.ToggleView):
				m.showOnlyOpenPorts = !m.showOnlyOpenPorts
				m.updateTable()
			}
		}

	case scanResultMsg:
		m.handleScanResult(msg.result)

	case scanProgressMsg:
		cmd := m.progress.SetPercent(float64(msg.progress.Completed) / float64(msg.progress.Total))
		cmds = append(cmds, cmd)

	case scanCompleteMsg:
		m.scanning = false
		m.stats.EndTime = time.Now()
		m.calculateStats()

	case spinner.TickMsg:
		if m.scanning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case stopwatch.TickMsg:
		var cmd tea.Cmd
		m.stopwatch, cmd = m.stopwatch.Update(msg)
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	// Update sub-components based on view
	switch m.viewMode {
	case ViewTable:
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	case ViewDetails, ViewStats, ViewLogs:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.scanning {
		cmds = append(cmds, m.listenForResults())
	}

	return m, tea.Batch(cmds...)
}

func (m *EnhancedModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	var sections []string

	// Header
	header := m.renderHeader()
	sections = append(sections, header)

	// Progress bar (if scanning)
	if m.scanning {
		progressBar := m.renderProgressBar()
		sections = append(sections, progressBar)
	}

	// Filter input (if filtering)
	if m.filtering {
		sections = append(sections, m.filterInput.View())
	}

	// Main content based on view mode
	mainContent := m.renderMainContent()
	sections = append(sections, mainContent)

	// Status bar
	statusBar := m.renderStatusBar()
	sections = append(sections, statusBar)

	// Help
	helpView := m.help.View(m.keys)
	sections = append(sections, helpView)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *EnhancedModel) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		MarginBottom(1)

	title := "🔍 Port Scanner"
	if m.scanning {
		title += fmt.Sprintf(" %s %s", m.spinner.View(), "Scanning...")
	} else {
		title += " ✓ Complete"
	}

	timer := m.stopwatch.View()

	headerLeft := titleStyle.Render(title)
	headerRight := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Render(timer)

	gap := m.width - lipgloss.Width(headerLeft) - lipgloss.Width(headerRight)
	if gap < 0 {
		gap = 0
	}

	return headerLeft + strings.Repeat(" ", gap) + headerRight
}

func (m *EnhancedModel) renderProgressBar() string {
	return m.progress.View()
}

func (m *EnhancedModel) renderMainContent() string {
	switch m.viewMode {
	case ViewTable:
		return m.table.View()
	case ViewDetails:
		return m.viewport.View()
	case ViewStats:
		return m.renderStatsView()
	case ViewLogs:
		return m.viewport.View()
	default:
		return m.table.View()
	}
}

func (m *EnhancedModel) renderStatsView() string {
	statsStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary).
		Padding(1, 2)

	stats := fmt.Sprintf(
		"📊 Scan Statistics\n\n"+
			"Total Ports: %d\n"+
			"Open: %d\n"+
			"Closed: %d\n"+
			"Filtered: %d\n"+
			"Average Latency: %v\n"+
			"Fastest Port: %d\n"+
			"Slowest Port: %d\n",
		m.stats.TotalPorts,
		m.stats.OpenPorts,
		m.stats.ClosedPorts,
		m.stats.FilteredPorts,
		m.stats.AverageLatency,
		m.stats.FastestPort,
		m.stats.SlowestPort,
	)

	return statsStyle.Render(stats)
}

func (m *EnhancedModel) renderStatusBar() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		MarginTop(1)

	viewIndicator := fmt.Sprintf("[%s]", m.getViewName())
	openPorts := fmt.Sprintf("Open: %d", m.stats.OpenPorts)

	status := fmt.Sprintf("%s | %s | Tab: switch view | ?: help",
		viewIndicator, openPorts)

	return statusStyle.Render(status)
}

// Helper methods
func (m *EnhancedModel) updateLayout() {
	m.table.SetHeight(m.height - 8)
	m.table.SetWidth(m.width)
	m.viewport.Width = m.width - 4
	m.viewport.Height = m.height - 8
	m.list.SetSize(m.width, m.height-8)
}

func (m *EnhancedModel) cycleView() {
	m.viewMode = (m.viewMode + 1) % 4
	m.updateViewport()
}

func (m *EnhancedModel) getViewName() string {
	switch m.viewMode {
	case ViewTable:
		return "Table"
	case ViewDetails:
		return "Details"
	case ViewStats:
		return "Stats"
	case ViewLogs:
		return "Logs"
	default:
		return "Table"
	}
}

func (m *EnhancedModel) handleScanResult(result core.ResultEvent) {
	m.results = append(m.results, result)
	m.updateTable()
	m.updateStats(result)
	m.addLog(fmt.Sprintf("Port %d: %s (%v)", result.Port, result.State, result.Duration))
}

// updateTable populates the table with scan results.
// By default, only open ports are shown. This behavior can be changed by setting
// m.showOnlyOpenPorts to false, in which case all scanned ports (open, closed, filtered)
// will be displayed.
func (m *EnhancedModel) updateTable() {
	var rows []table.Row
	for _, r := range m.results {
		// Filter based on configuration
		if m.showOnlyOpenPorts && r.State != core.StateOpen {
			continue
		}
		
		service := getServiceName(r.Port)
		banner := r.Banner
		if len(banner) > 30 {
			banner = banner[:27] + "..."
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", r.Port),
			string(r.State),
			service,
			banner,
			fmt.Sprintf("%dms", r.Duration.Milliseconds()),
		})
	}
	m.table.SetRows(rows)
}

func (m *EnhancedModel) updateStats(result core.ResultEvent) {
	m.stats.TotalPorts++

	switch result.State {
	case core.StateOpen:
		m.stats.OpenPorts++
	case core.StateClosed:
		m.stats.ClosedPorts++
	case core.StateFiltered:
		m.stats.FilteredPorts++
	}

	// Update latency stats
	if result.State == core.StateOpen {
		if m.stats.FastestPort == 0 || result.Duration < m.stats.AverageLatency {
			m.stats.FastestPort = result.Port
		}
		if result.Duration > m.stats.AverageLatency {
			m.stats.SlowestPort = result.Port
		}
	}
}

func (m *EnhancedModel) calculateStats() {
	if len(m.results) == 0 {
		return
	}

	var totalLatency time.Duration
	openCount := 0

	for _, r := range m.results {
		if r.State == core.StateOpen {
			totalLatency += r.Duration
			openCount++
		}
	}

	if openCount > 0 {
		m.stats.AverageLatency = totalLatency / time.Duration(openCount)
	}
}

func (m *EnhancedModel) applyFilter() {
	// Filter logic here
}

func (m *EnhancedModel) selectCurrentItem() {
	// Selection logic here
}

func (m *EnhancedModel) updateViewport() {
	switch m.viewMode {
	case ViewDetails:
		if m.selectedPort != nil {
			details := fmt.Sprintf("Port Details:\n\nPort: %d\nState: %s\nBanner: %s\nLatency: %v",
				m.selectedPort.Port,
				m.selectedPort.State,
				m.selectedPort.Banner,
				m.selectedPort.Duration)
			m.viewport.SetContent(details)
		}
	case ViewLogs:
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
	}
}

func (m *EnhancedModel) addLog(log string) {
	timestamp := time.Now().Format("15:04:05")
	m.logs = append(m.logs, fmt.Sprintf("[%s] %s", timestamp, log))
	if len(m.logs) > 1000 {
		m.logs = m.logs[1:]
	}
}

func (m *EnhancedModel) listenForResults() tea.Cmd {
	return func() tea.Msg {
		select {
		case result, ok := <-m.resultChan:
			if !ok {
				return scanCompleteMsg{}
			}

			switch r := result.(type) {
			case core.ResultEvent:
				return scanResultMsg{result: r}
			case core.ProgressEvent:
				return scanProgressMsg{progress: r}
			}
		case <-time.After(100 * time.Millisecond):
			// Check again soon
		}
		return nil
	}
}

func (m *EnhancedModel) Run() error {
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
