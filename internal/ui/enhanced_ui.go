package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/pkg/config"
	"github.com/lucchesi-sec/portscan/pkg/theme"
)

// UIViewState represents different views in the TUI
type UIViewState int

const (
	UIViewMain UIViewState = iota
	UIViewHelp
	UIViewDetails
)

// Message types for communication with the UI
type scanResultMsg struct {
	result core.ResultEvent
}

type scanProgressMsg struct {
	progress core.ProgressEvent
}

type scanCompleteMsg struct{}

// EnhancedUI is the improved TUI model with better navigation and feedback
type EnhancedUI struct {
	// Core
	config     *config.Config
	theme      theme.Theme
	results    []core.ResultEvent
	resultChan <-chan interface{}

	// View state
	viewState UIViewState
	width     int
	height    int

	// Components
	table       table.Model
	progressBar progress.Model
	spinner     spinner.Model
	help        help.Model
	keys        KeyBindings

	// Progress tracking
	progressTrack *ProgressTracker

	// State
	scanning     bool
	isPaused     bool
	showHelp     bool
	totalPorts   int
	showOnlyOpen bool

	// Stats
	openCount     int
	closedCount   int
	filteredCount int
	currentRate   float64
}

// KeyBindings defines all keyboard shortcuts
type KeyBindings struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding
	Help     key.Binding
	Pause    key.Binding
	Clear    key.Binding
	Quit     key.Binding
}

var defaultKeys = KeyBindings{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup", "ctrl+u"),
		key.WithHelp("PgUp/Ctrl+u", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown", "ctrl+d"),
		key.WithHelp("PgDn/Ctrl+d", "page down"),
	),
	Home: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("Home/g", "go to top"),
	),
	End: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("End/G", "go to bottom"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Pause: key.NewBinding(
		key.WithKeys("p", " "),
		key.WithHelp("p/Space", "pause/resume"),
	),
	Clear: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("Ctrl+L", "clear screen"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c", "esc"),
		key.WithHelp("q/Esc", "quit"),
	),
}

// ShortHelp returns key bindings for the help bar
func (k KeyBindings) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Pause, k.Quit}
}

// FullHelp returns all key bindings for the help view
func (k KeyBindings) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Home, k.End, k.Clear},
		{k.Pause, k.Help, k.Quit},
	}
}

// NewEnhancedUI creates a new enhanced UI model
func NewEnhancedUI(cfg *config.Config, totalPorts int, results <-chan interface{}, onlyOpen bool) *EnhancedUI {
	// Get theme
	t := theme.GetTheme(cfg.UI.Theme)

	// Initialize table
	columns := []table.Column{
		{Title: "Port", Width: 8},
		{Title: "State", Width: 10},
		{Title: "Service", Width: 15},
		{Title: "Banner", Width: 40},
		{Title: "Latency", Width: 10},
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.Primary).
		BorderBottom(true)
	s.Selected = s.Selected.
		Foreground(t.Background).
		Background(t.Primary)
	tbl.SetStyles(s)

	// Initialize progress bar
	prog := progress.New(progress.WithDefaultGradient())

	// Initialize spinner
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(t.Primary)

	// Initialize help
	h := help.New()
	h.ShowAll = false

	return &EnhancedUI{
		config:        cfg,
		theme:         t,
		results:       []core.ResultEvent{},
		resultChan:    results,
		table:         tbl,
		progressBar:   prog,
		spinner:       spin,
		help:          h,
		keys:          defaultKeys,
		progressTrack: NewProgressTracker(totalPorts),
		scanning:      true,
		totalPorts:    totalPorts,
		viewState:     UIViewMain,
		showOnlyOpen:  onlyOpen,
	}
}

// Init initializes the UI
func (m *EnhancedUI) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.listenForResults(),
	)
}

// Update handles messages
func (m *EnhancedUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetHeight(m.height - 12) // Leave space for header and footer
		m.table.SetWidth(m.width)

	case tea.KeyMsg:
		// Handle help toggle first
		if key.Matches(msg, m.keys.Help) {
			m.showHelp = !m.showHelp
			m.help.ShowAll = m.showHelp
			if m.showHelp {
				m.viewState = UIViewHelp
			} else {
				m.viewState = UIViewMain
			}
			return m, nil
		}

		// Handle other keys based on view state
		switch m.viewState {
		case UIViewHelp:
			if key.Matches(msg, m.keys.Quit) || key.Matches(msg, m.keys.Help) {
				m.showHelp = false
				m.viewState = UIViewMain
			}

		case UIViewMain:
			switch {
			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit

			case key.Matches(msg, m.keys.Pause):
				if m.scanning {
					m.isPaused = !m.isPaused
					if m.isPaused {
						m.progressTrack.Pause()
					} else {
						m.progressTrack.Resume()
					}
				}

			case key.Matches(msg, m.keys.Clear):
				return m, tea.ClearScreen

			case key.Matches(msg, m.keys.Up):
				m.table.MoveUp(1)

			case key.Matches(msg, m.keys.Down):
				m.table.MoveDown(1)

			case key.Matches(msg, m.keys.PageUp):
				m.table.MoveUp(10)

			case key.Matches(msg, m.keys.PageDown):
				m.table.MoveDown(10)

			case key.Matches(msg, m.keys.Home):
				m.table.GotoTop()

			case key.Matches(msg, m.keys.End):
				m.table.GotoBottom()
			}
		}

	case scanResultMsg:
		m.results = append(m.results, msg.result)
		m.updateTable()
		m.updateStats(msg.result)

	case scanProgressMsg:
		m.currentRate = msg.progress.Rate
		m.progressTrack.Update(
			len(m.results),
			m.openCount,
			m.closedCount,
			m.filteredCount,
			m.currentRate,
		)

	case scanCompleteMsg:
		m.scanning = false

	case spinner.TickMsg:
		if m.scanning && !m.isPaused {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case progress.FrameMsg:
		progressModel, cmd := m.progressBar.Update(msg)
		m.progressBar = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	// Update table
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	// Continue listening for results
	if m.scanning {
		cmds = append(cmds, m.listenForResults())
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m *EnhancedUI) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Show help overlay if active
	if m.showHelp {
		return m.renderHelp()
	}

	// Main view
	return m.renderMain()
}

// renderMain renders the main scanning view
func (m *EnhancedUI) renderMain() string {
	var b strings.Builder

	// Breadcrumb
	breadcrumb := m.renderBreadcrumb()
	b.WriteString(breadcrumb + "\n")

	// Header
	header := m.renderHeader()
	b.WriteString(header + "\n")

	// Progress bar
	if m.scanning {
		progressView := m.renderProgress()
		b.WriteString(progressView + "\n")
	}

	// Status line
	status := m.renderStatus()
	b.WriteString(status + "\n\n")

	// Results table
	b.WriteString(m.table.View() + "\n")

	// Footer with help
	footer := m.renderFooter()
	b.WriteString(footer)

	return b.String()
}

// renderBreadcrumb shows the current location/mode
func (m *EnhancedUI) renderBreadcrumb() string {
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

// renderHeader renders the header section
func (m *EnhancedUI) renderHeader() string {
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

// renderProgress renders the progress bar
func (m *EnhancedUI) renderProgress() string {
	progress := m.progressTrack.GetProgress() / 100.0
	return m.progressBar.ViewAs(progress)
}

// renderStatus renders the status line
func (m *EnhancedUI) renderStatus() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(m.theme.Foreground)

	status := m.progressTrack.GetStatusLine()
	details := m.progressTrack.GetDetailedStats()

	return statusStyle.Render(status + "\n" + details)
}

// renderFooter renders the footer with shortcuts
func (m *EnhancedUI) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(m.theme.Secondary)

	return footerStyle.Render(m.help.View(m.keys))
}

// renderHelp renders the help overlay
func (m *EnhancedUI) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Padding(2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary)

	content := `
📖 KEYBOARD SHORTCUTS

Navigation:
  ↑/k        Move up
  ↓/j        Move down
  PgUp/Ctrl+u  Page up
  PgDn/Ctrl+d  Page down
  Home/g     Go to top
  End/G      Go to bottom

Control:
  p/Space    Pause/Resume scan
  Ctrl+L     Clear screen
  ?          Toggle this help
  q/Esc      Quit

Tips:
  • Results update in real-time
  • ETA adjusts based on scan speed
  • Open ports are highlighted
  • Service names are auto-detected

Press ? or Esc to close help`

	return helpStyle.Render(content)
}

// Helper methods

func (m *EnhancedUI) updateTable() {
	var rows []table.Row
	for _, r := range m.results {
		if m.showOnlyOpen && r.State != core.StateOpen {
			continue
		}
		service := getServiceName(r.Port)
		banner := r.Banner
		if len(banner) > 40 {
			banner = banner[:37] + "..."
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

func (m *EnhancedUI) updateStats(result core.ResultEvent) {
	switch result.State {
	case core.StateOpen:
		m.openCount++
	case core.StateClosed:
		m.closedCount++
	case core.StateFiltered:
		m.filteredCount++
	}
}

func (m *EnhancedUI) listenForResults() tea.Cmd {
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

// Run starts the TUI
func (m *EnhancedUI) Run() error {
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

// getServiceName returns a human-friendly service name for well-known ports
func getServiceName(port uint16) string {
	services := map[uint16]string{
		21:    "FTP",
		22:    "SSH",
		23:    "Telnet",
		25:    "SMTP",
		53:    "DNS",
		80:    "HTTP",
		110:   "POP3",
		143:   "IMAP",
		443:   "HTTPS",
		445:   "SMB",
		3306:  "MySQL",
		3389:  "RDP",
		5432:  "PostgreSQL",
		6379:  "Redis",
		8080:  "HTTP-Alt",
		8443:  "HTTPS-Alt",
		27017: "MongoDB",
	}

	if service, ok := services[port]; ok {
		return service
	}
	return "Unknown"
}
