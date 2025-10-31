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
	"github.com/lucchesi-sec/portscan/internal/ui/commands"
	"github.com/lucchesi-sec/portscan/pkg/config"
	"github.com/lucchesi-sec/portscan/pkg/theme"
)

// UIViewState represents different views in the TUI
type UIViewState int

const (
	UIViewMain UIViewState = iota
	UIViewHelp
	UIViewDetails
	UIViewSortModal
	UIViewFilterModal
)

// ModalType represents different modal dialog types
type ModalType int

const (
	ModalSort ModalType = iota
	ModalFilter
	ModalDetails
	ModalCommandPalette
)

// Position represents screen coordinates and dimensions
type Position struct {
	X      int
	Y      int
	Width  int
	Height int
}

// ModalState represents the state of modal dialogs
type ModalState struct {
	IsActive        bool
	Type            ModalType
	Position        Position
	Cursor          int
	ScrollPosition  int // For details modal scrolling
	MaxScrollHeight int // Track content height for scrolling
}

// CommandPaletteState represents the state of the command palette
type CommandPaletteState struct {
	Query           string
	Cursor          int
	FilteredResults []commands.MatchResult
	CommandRegistry *commands.Registry
}

// Message types for communication with the UI
type scanResultMsg struct {
	result core.ResultEvent
}

type scanProgressMsg struct {
	progress core.ProgressEvent
}

type scanCompleteMsg struct{}

// Note: DefaultResultBufferSize is now defined in constants.go

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
		capacity = DefaultResultBufferSize
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
	viewState  UIViewState
	modalState ModalState
	width      int
	height     int

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
	stats             *ResultStats
	currentRate       float64
	previousOpenCount int

	// Sorting and Filtering
	sortState      *SortState
	filterState    *FilterState
	displayResults []core.ResultEvent // Filtered/sorted view of results

	// Command Palette
	commandPaletteState *CommandPaletteState

	// Dashboard
	showDashboard bool
	statsData     *StatsData
	sparklineData *SparklineData
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
	CommandPalette  key.Binding
	ToggleDashboard key.Binding
	Enter           key.Binding
	Escape          key.Binding
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
	CommandPalette: key.NewBinding(
		key.WithKeys("ctrl+k"),
		key.WithHelp("Ctrl+K", "command palette"),
	),
	ToggleDashboard: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("D", "toggle dashboard"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "confirm selection"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("ESC", "cancel/close"),
	),
}

// ShortHelp returns key bindings for the help bar
func (k KeyBindings) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.CommandPalette, k.Pause, k.Quit}
}

// FullHelp returns all key bindings for the help view
func (k KeyBindings) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Home, k.End, k.Clear},
		{k.Sort, k.Filter, k.Reset, k.OpenOnly},
		{k.Search, k.CommandPalette, k.Pause, k.Help, k.Quit},
	}
}

// NewScanUI creates a new scan UI model.
func NewScanUI(cfg *config.Config, totalPorts int, results <-chan core.Event, onlyOpen bool) *ScanUI {
	t := theme.GetTheme(cfg.UI.Theme)

	bufferSize := cfg.UI.ResultBufferSize
	if bufferSize <= 0 {
		bufferSize = DefaultResultBufferSize
	}

	resultBuffer := NewResultBuffer(bufferSize)
	stats := NewResultStats()

	columns := []table.Column{
		{Title: "Host", Width: ColumnWidthHost},
		{Title: "Port", Width: ColumnWidthPort},
		{Title: "Protocol", Width: ColumnWidthProtocol},
		{Title: "State", Width: ColumnWidthState},
		{Title: "Service", Width: ColumnWidthService},
		{Title: "Banner", Width: ColumnWidthBanner},
		{Title: "Latency", Width: ColumnWidthLatency},
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(TableDefaultHeight),
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
	sparklineData := NewSparklineData()
	commandRegistry := commands.NewRegistry()

	// Add default commands
	defaultCmds := commands.DefaultCommands()

	// Enhance commands with UI-specific actions
	enrichedCmds := enhanceCommandsWithUIActions(defaultCmds)

	commandRegistry.AddCommands(enrichedCmds)
	if onlyOpen {
		filterState.SetStateFilter(StateFilterOpen)
	}

	return &ScanUI{
		config:              cfg,
		theme:               t,
		results:             resultBuffer,
		resultChan:          results,
		bufferSize:          bufferSize,
		table:               tbl,
		progressBar:         prog,
		spinner:             spin,
		help:                helpModel,
		keys:                defaultKeys,
		progressTrack:       NewProgressTracker(totalPorts),
		scanning:            true,
		totalPorts:          totalPorts,
		viewState:           UIViewMain,
		showOnlyOpen:        onlyOpen,
		sortState:           sortState,
		filterState:         filterState,
		stats:               stats,
		displayResults:      []core.ResultEvent{},
		commandPaletteState: &CommandPaletteState{Query: "", Cursor: 0, FilteredResults: []commands.MatchResult{}, CommandRegistry: commandRegistry},
		sparklineData:       sparklineData,
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

// enhanceCommandsWithUIActions adds UI-specific actions to commands that need direct UI manipulation
func enhanceCommandsWithUIActions(cmds []commands.Command) []commands.Command {
	for i := range cmds {
		cmd := &cmds[i]

		switch cmd.ID {
		case "action-pause":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				if uiModel.scanning {
					uiModel.isPaused = !uiModel.isPaused
					if uiModel.isPaused {
						uiModel.progressTrack.Pause()
					} else {
						uiModel.progressTrack.Resume()
					}
				}
				return nil
			}
		case "action-sort":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				uiModel.openModal(ModalSort)
				return nil
			}
		case "action-filter":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				uiModel.openModal(ModalFilter)
				return nil
			}
		case "action-reset-filters":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				uiModel.filterState.Reset()
				uiModel.sortState = NewSortState()
				uiModel.updateTable()
				return nil
			}
		case "action-toggle-open-only":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				if uiModel.filterState.StateFilter == StateFilterOpen {
					uiModel.filterState.SetStateFilter(StateFilterAll)
				} else {
					uiModel.filterState.SetStateFilter(StateFilterOpen)
				}
				uiModel.updateTable()
				return nil
			}
		case "action-search":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				// For search, we want to open the command palette for search functionality
				uiModel.commandPaletteState.Query = ""
				uiModel.commandPaletteState.Cursor = 0
				activeCommands := uiModel.commandPaletteState.CommandRegistry.GetActiveCommands(uiModel)
				uiModel.commandPaletteState.FilteredResults = commands.FuzzySearch(activeCommands, "")
				uiModel.openModal(ModalCommandPalette)
				return nil
			}
		case "action-toggle-dashboard":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				uiModel.showDashboard = !uiModel.showDashboard
				if uiModel.showDashboard {
					uiModel.statsData = uiModel.computeStats()
				}
				return nil
			}
		case "action-view-details":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				if len(uiModel.displayResults) > 0 {
					uiModel.openModal(ModalDetails)
				}
				return nil
			}
		case "view-help":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				uiModel.showHelp = !uiModel.showHelp
				uiModel.help.ShowAll = uiModel.showHelp
				if uiModel.showHelp {
					uiModel.viewState = UIViewHelp
				} else {
					uiModel.viewState = UIViewMain
				}
				return nil
			}
		case "view-quit":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				return tea.Quit
			}
		case "view-clear":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				return tea.ClearScreen
			}
		case "nav-up":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				uiModel.table.MoveUp(1)
				return nil
			}
		case "nav-down":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				uiModel.table.MoveDown(1)
				return nil
			}
		case "nav-top":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				uiModel.table.GotoTop()
				return nil
			}
		case "nav-bottom":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				uiModel.table.GotoBottom()
				return nil
			}
		case "nav-page-up":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				uiModel.table.MoveUp(PageScrollLines)
				return nil
			}
		case "nav-page-down":
			cmd.UIAction = func(model interface{}) tea.Cmd {
				uiModel, ok := model.(*ScanUI)
				if !ok {
					return nil
				}
				uiModel.table.MoveDown(PageScrollLines)
				return nil
			}
		}
	}
	return cmds
}
