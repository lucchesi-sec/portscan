package ui

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/pkg/config"
)

type Model struct {
	config     *config.Config
	results    []core.ResultEvent
	progress   core.ProgressEvent
	table      table.Model
	spinner    spinner.Model
	scanning   bool
	startTime  time.Time
	width      int
	height     int
	resultChan <-chan interface{}
}

type scanResultMsg struct {
	result core.ResultEvent
}

type scanProgressMsg struct {
	progress core.ProgressEvent
}

type scanCompleteMsg struct{}

func New(cfg *config.Config, results <-chan interface{}) *Model {
	columns := []table.Column{
		{Title: "Host", Width: 20},
		{Title: "Port", Width: 10},
		{Title: "State", Width: 10},
		{Title: "Service", Width: 20},
		{Title: "Latency", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &Model{
		config:     cfg,
		results:    []core.ResultEvent{},
		table:      t,
		spinner:    s,
		scanning:   true,
		startTime:  time.Now(),
		resultChan: results,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.listenForResults(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetHeight(m.height - 10)
		m.table.SetWidth(m.width)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			m.table.MoveDown(1)
		case "k", "up":
			m.table.MoveUp(1)
		case "g":
			m.table.GotoTop()
		case "G":
			m.table.GotoBottom()
		}

	case scanResultMsg:
		m.results = append(m.results, msg.result)
		m.updateTable()

	case scanProgressMsg:
		m.progress = msg.progress

	case scanCompleteMsg:
		m.scanning = false

	case spinner.TickMsg:
		if m.scanning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	if m.scanning {
		cmds = append(cmds, m.listenForResults())
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	var s string

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	s += headerStyle.Render("🔍 Port Scanner") + "\n"

	// Status line
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	elapsed := time.Since(m.startTime).Round(time.Second)
	status := fmt.Sprintf("Elapsed: %s | Open: %d | Rate: %.0f pps",
		elapsed, m.countOpen(), m.progress.Rate)

	if m.scanning {
		status = m.spinner.View() + " Scanning... " + status
	} else {
		status = "✓ Complete " + status
	}

	s += statusStyle.Render(status) + "\n\n"

	// Results table
	s += m.table.View() + "\n"

	// Footer
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1)

	footer := "j/k: navigate | g/G: top/bottom | q: quit"
	s += footerStyle.Render(footer)

	return s
}

func (m *Model) Run() error {
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m *Model) listenForResults() tea.Cmd {
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

func (m *Model) updateTable() {
	// Sort results by port
	sort.Slice(m.results, func(i, j int) bool {
		return m.results[i].Port < m.results[j].Port
	})

	// Convert to table rows
	var rows []table.Row
	for _, r := range m.results {
		if r.State == core.StateOpen {
			service := r.Banner
			if service == "" {
				service = getServiceName(r.Port)
			}

			rows = append(rows, table.Row{
				r.Host,
				fmt.Sprintf("%d", r.Port),
				string(r.State),
				service,
				fmt.Sprintf("%dms", r.Duration.Milliseconds()),
			})
		}
	}

	m.table.SetRows(rows)
}

func (m *Model) countOpen() int {
	count := 0
	for _, r := range m.results {
		if r.State == core.StateOpen {
			count++
		}
	}
	return count
}

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