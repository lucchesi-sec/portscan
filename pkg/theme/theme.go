package theme

import "github.com/charmbracelet/lipgloss"

// Theme defines color scheme for the TUI.
type Theme struct {
	Name       string
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Danger     lipgloss.Color
	Info       lipgloss.Color
	Background lipgloss.Color
	Foreground lipgloss.Color
	Muted      lipgloss.Color
}

var (
	Default = Theme{
		Name:       "default",
		Primary:    lipgloss.Color("205"),
		Secondary:  lipgloss.Color("135"),
		Success:    lipgloss.Color("42"),
		Warning:    lipgloss.Color("214"),
		Danger:     lipgloss.Color("196"),
		Info:       lipgloss.Color("39"),
		Background: lipgloss.Color("0"),
		Foreground: lipgloss.Color("15"),
		Muted:      lipgloss.Color("240"),
	}

	Dracula = Theme{
		Name:       "dracula",
		Primary:    lipgloss.Color("#bd93f9"),
		Secondary:  lipgloss.Color("#ff79c6"),
		Success:    lipgloss.Color("#50fa7b"),
		Warning:    lipgloss.Color("#f1fa8c"),
		Danger:     lipgloss.Color("#ff5555"),
		Info:       lipgloss.Color("#8be9fd"),
		Background: lipgloss.Color("#282a36"),
		Foreground: lipgloss.Color("#f8f8f2"),
		Muted:      lipgloss.Color("#6272a4"),
	}

	Monokai = Theme{
		Name:       "monokai",
		Primary:    lipgloss.Color("#66d9ef"),
		Secondary:  lipgloss.Color("#f92672"),
		Success:    lipgloss.Color("#a6e22e"),
		Warning:    lipgloss.Color("#fd971f"),
		Danger:     lipgloss.Color("#f92672"),
		Info:       lipgloss.Color("#66d9ef"),
		Background: lipgloss.Color("#272822"),
		Foreground: lipgloss.Color("#f8f8f2"),
		Muted:      lipgloss.Color("#75715e"),
	}
)

// GetTheme returns the theme matching the given name.
// Defaults to the "default" theme if name is not recognized.
func GetTheme(name string) Theme {
	switch name {
	case "dracula":
		return Dracula
	case "monokai":
		return Monokai
	default:
		return Default
	}
}

// HeaderStyle returns the style for header text.
func (t Theme) HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary).
		MarginBottom(1)
}

// StatusStyle returns the style for status text.
func (t Theme) StatusStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Muted)
}

// FooterStyle returns the style for footer text.
func (t Theme) FooterStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Muted).
		MarginTop(1)
}

// SuccessStyle returns the style for success messages.
func (t Theme) SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Success)
}

// ErrorStyle returns the style for error messages.
func (t Theme) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Danger)
}

// WarningStyle returns the style for warning messages.
func (t Theme) WarningStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Warning)
}

// StateColors defines colors for different port states
type StateColors struct {
	Open     lipgloss.Color
	Closed   lipgloss.Color
	Filtered lipgloss.Color
}

// GetStateColors returns the color scheme for port states based on the theme
func (t Theme) GetStateColors() StateColors {
	return StateColors{
		Open:     lipgloss.Color("#00FF00"), // Green for open ports
		Closed:   lipgloss.Color("#FF0000"), // Red for closed ports
		Filtered: lipgloss.Color("#FFA500"), // Orange for filtered ports
	}
}

// TableHeaderStyle styles table headers.
func (t Theme) TableHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary).
		Padding(0, 1)
}

// TableCellStyle styles table cells.
func (t Theme) TableCellStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Foreground).
		Padding(0, 1)
}

// TableSelectedStyle styles the selected row in the table.
func (t Theme) TableSelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Background).
		Background(t.Primary).
		Bold(true).
		Padding(0, 1)
}

// TableContainerStyle styles the outer container of the table.
func (t Theme) TableContainerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Muted).
		Padding(0, 1)
}

// DashboardPanelStyle styles the dashboard side panel container.
func (t Theme) DashboardPanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(1)
}

// ModalStyle returns the style for modal dialogs (padding parameter to avoid import cycle)
func (t Theme) ModalStyle(padding int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Background(t.Background).
		Foreground(t.Foreground).
		Padding(padding)
}

// ModalOverlayStyle returns the style for the semi-transparent overlay
func (t Theme) ModalOverlayStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(t.Background).
		Foreground(lipgloss.Color("240")) // Dimmed overlay
}

// GetRowStyle returns the appropriate styling for a table row based on port state
func (t Theme) GetRowStyle(state string) lipgloss.Style {
	stateColors := t.GetStateColors()

	switch state {
	case "open":
		return lipgloss.NewStyle().
			Foreground(stateColors.Open).
			Background(lipgloss.Color("")) // Transparent background for contrast
	case "closed":
		return lipgloss.NewStyle().
			Foreground(stateColors.Closed).
			Background(lipgloss.Color("")) // Transparent background for contrast
	case "filtered":
		return lipgloss.NewStyle().
			Foreground(stateColors.Filtered).
			Background(lipgloss.Color("")) // Transparent background for contrast
	default:
		return lipgloss.NewStyle() // Default styling
	}
}
