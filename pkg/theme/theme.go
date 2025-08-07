package theme

import "github.com/charmbracelet/lipgloss"

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

func (t Theme) HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary).
		MarginBottom(1)
}

func (t Theme) StatusStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Muted)
}

func (t Theme) FooterStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Muted).
		MarginTop(1)
}

func (t Theme) SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Success)
}

func (t Theme) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Danger)
}

func (t Theme) WarningStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Warning)
}
