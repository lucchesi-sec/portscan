package ui

import (
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
	UIViewSortMenu
	UIViewFilterMenu
)

// Message types for communication with the UI
type scanResultMsg struct {
	result core.ResultEvent
}

type scanProgressMsg struct {
	progress core.ProgressEvent
}

type scanCompleteMsg struct{}

const defaultResultBufferSize = 10000

// ResultBuffer maintains a fixed-size circular buffer of recent scan results.
type ResultBuffer struct {
	data     []core.ResultEvent
	start    int
	length   int
	capacity int
}

// NewResultBuffer creates a new ring buffer with the provided capacity.
func NewResultBuffer(capacity int) *ResultBuffer {
	if capacity <= 0 {
		capacity = defaultResultBufferSize
	}

	return &ResultBuffer{
		data:     make([]core.ResultEvent, capacity),
		capacity: capacity,
	}
}

// Append inserts a result into the buffer, evicting the oldest when full.
func (b *ResultBuffer) Append(result core.ResultEvent) {
	if b.capacity == 0 {
		return
	}

	if b.length < b.capacity {
		idx := (b.start + b.length) % b.capacity
		b.data[idx] = result
		b.length++
		return
	}

	// Buffer full: overwrite the oldest and move start forward.
	b.data[b.start] = result
	b.start = (b.start + 1) % b.capacity
}

// Items returns the buffered results in discovery order (oldest to newest).
func (b *ResultBuffer) Items() []core.ResultEvent {
	if b.length == 0 {
		return nil
	}

	items := make([]core.ResultEvent, 0, b.length)
	for i := 0; i < b.length; i++ {
		idx := (b.start + i) % b.capacity
		items = append(items, b.data[idx])
	}
	return items
}

// Len reports the number of items currently stored.
func (b *ResultBuffer) Len() int {
	return b.length
}

// ResultStats tracks aggregate counts for scan results regardless of buffer eviction.
type ResultStats struct {
	total    int
	open     int
	closed   int
	filtered int
}

func NewResultStats() *ResultStats {
	return &ResultStats{}
}

func (s *ResultStats) Add(result core.ResultEvent) {
	s.total++

	switch result.State {
	case core.StateOpen:
		s.open++
	case core.StateClosed:
		s.closed++
	case core.StateFiltered:
		s.filtered++
	}
}

func (s *ResultStats) Totals() (total, open, closed, filtered int) {
	return s.total, s.open, s.closed, s.filtered
}

// ScanUI renders scan activity in the terminal.
type ScanUI struct {
	// Core
	config     *config.Config
	theme      theme.Theme
	results    *ResultBuffer
	resultChan <-chan core.Event
	bufferSize int

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
	stats       *ResultStats
	currentRate float64

	// Sorting and Filtering
	sortState      *SortState
	filterState    *FilterState
	displayResults []core.ResultEvent // Filtered/sorted view of results

	// Dashboard
	showDashboard bool
	statsData     *StatsData
}

// KeyBindings defines all keyboard shortcuts
type KeyBindings struct {
	Up              key.Binding
	Down            key.Binding
	PageUp          key.Binding
	PageDown        key.Binding
	Home            key.Binding
	End             key.Binding
	Help            key.Binding
	Pause           key.Binding
	Clear           key.Binding
	Quit            key.Binding
	Sort            key.Binding
	Filter          key.Binding
	Reset           key.Binding
	OpenOnly        key.Binding
	Search          key.Binding
	ToggleDashboard key.Binding
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
	Sort: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "sort results"),
	),
	Filter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "filter results"),
	),
	Reset: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reset filters"),
	),
	OpenOnly: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "toggle open only"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search banners"),
	),
	ToggleDashboard: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("D", "toggle dashboard"),
	),
}

// ShortHelp returns key bindings for the help bar
func (k KeyBindings) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Sort, k.Filter, k.Pause, k.Quit}
}

// FullHelp returns all key bindings for the help view
func (k KeyBindings) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Home, k.End, k.Clear},
		{k.Sort, k.Filter, k.Reset, k.OpenOnly},
		{k.Search, k.Pause, k.Help, k.Quit},
	}
}

// NewScanUI creates a new scan UI model.
func NewScanUI(cfg *config.Config, totalPorts int, results <-chan core.Event, onlyOpen bool) *ScanUI {
	t := theme.GetTheme(cfg.UI.Theme)

	bufferSize := cfg.UI.ResultBufferSize
	if bufferSize <= 0 {
		bufferSize = defaultResultBufferSize
	}

	resultBuffer := NewResultBuffer(bufferSize)
	stats := NewResultStats()

	columns := []table.Column{
		{Title: "Host", Width: 20},
		{Title: "Port", Width: 8},
		{Title: "Protocol", Width: 8},
		{Title: "State", Width: 10},
		{Title: "Service", Width: 15},
		{Title: "Banner", Width: 35},
		{Title: "Latency", Width: 10},
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.Primary).
		BorderBottom(true)
	styles.Selected = styles.Selected.
		Foreground(t.Background).
		Background(t.Primary)
	tbl.SetStyles(styles)

	prog := progress.New(progress.WithDefaultGradient())

	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(t.Primary)

	helpModel := help.New()
	helpModel.ShowAll = false

	sortState := NewSortState()
	filterState := NewFilterState()
	if onlyOpen {
		filterState.SetStateFilter(StateFilterOpen)
	}

	return &ScanUI{
		config:         cfg,
		theme:          t,
		results:        resultBuffer,
		resultChan:     results,
		bufferSize:     bufferSize,
		table:          tbl,
		progressBar:    prog,
		spinner:        spin,
		help:           helpModel,
		keys:           defaultKeys,
		progressTrack:  NewProgressTracker(totalPorts),
		scanning:       true,
		totalPorts:     totalPorts,
		viewState:      UIViewMain,
		showOnlyOpen:   onlyOpen,
		sortState:      sortState,
		filterState:    filterState,
		stats:          stats,
		displayResults: []core.ResultEvent{},
	}
}

// Init initializes the UI
func (m *ScanUI) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.listenForResults(),
	)
}

// Run starts the TUI program.
func (m *ScanUI) Run() error {
	program := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := program.Run()
	return err
}
