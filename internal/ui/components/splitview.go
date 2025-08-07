package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type SplitView struct {
	Left        string
	Right       string
	Width       int
	Height      int
	SplitRatio  float64 // 0.0 to 1.0
	Border      bool
	BorderStyle lipgloss.Border
	BorderColor lipgloss.Color
}

func NewSplitView(width, height int) *SplitView {
	return &SplitView{
		Width:       width,
		Height:      height,
		SplitRatio:  0.5,
		Border:      true,
		BorderStyle: lipgloss.RoundedBorder(),
		BorderColor: lipgloss.Color("240"),
	}
}

func (s *SplitView) SetContent(left, right string) {
	s.Left = left
	s.Right = right
}

func (s *SplitView) Render() string {
	leftWidth := int(float64(s.Width) * s.SplitRatio)
	rightWidth := s.Width - leftWidth - 1 // -1 for separator

	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(s.Height).
		MaxWidth(leftWidth)

	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(s.Height).
		MaxWidth(rightWidth)

	if s.Border {
		leftStyle = leftStyle.
			Border(s.BorderStyle).
			BorderForeground(s.BorderColor)
		rightStyle = rightStyle.
			Border(s.BorderStyle).
			BorderForeground(s.BorderColor)
	}

	leftContent := leftStyle.Render(s.Left)
	rightContent := rightStyle.Render(s.Right)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent)
}

type TabView struct {
	Tabs          []string
	ActiveTab     int
	Content       map[string]string
	Width         int
	ActiveColor   lipgloss.Color
	InactiveColor lipgloss.Color
}

func NewTabView(tabs []string, width int) *TabView {
	return &TabView{
		Tabs:          tabs,
		ActiveTab:     0,
		Content:       make(map[string]string),
		Width:         width,
		ActiveColor:   lipgloss.Color("205"),
		InactiveColor: lipgloss.Color("240"),
	}
}

func (t *TabView) SetContent(tab string, content string) {
	t.Content[tab] = content
}

func (t *TabView) NextTab() {
	t.ActiveTab = (t.ActiveTab + 1) % len(t.Tabs)
}

func (t *TabView) PrevTab() {
	t.ActiveTab--
	if t.ActiveTab < 0 {
		t.ActiveTab = len(t.Tabs) - 1
	}
}

func (t *TabView) Render() string {
	var tabBar []string

	for i, tab := range t.Tabs {
		style := lipgloss.NewStyle().
			Padding(0, 2)

		if i == t.ActiveTab {
			style = style.
				Foreground(t.ActiveColor).
				Bold(true).
				Underline(true)
		} else {
			style = style.
				Foreground(t.InactiveColor)
		}

		tabBar = append(tabBar, style.Render(tab))
	}

	tabBarStr := lipgloss.JoinHorizontal(lipgloss.Top, tabBar...)

	// Render active tab content
	activeTabName := t.Tabs[t.ActiveTab]
	content := t.Content[activeTabName]

	contentStyle := lipgloss.NewStyle().
		Width(t.Width).
		MarginTop(1)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabBarStr,
		contentStyle.Render(content),
	)
}

type StatusBar struct {
	Items      []StatusItem
	Width      int
	Background lipgloss.Color
	Foreground lipgloss.Color
}

type StatusItem struct {
	Label string
	Value string
	Icon  string
}

func NewStatusBar(width int) *StatusBar {
	return &StatusBar{
		Width:      width,
		Background: lipgloss.Color("235"),
		Foreground: lipgloss.Color("252"),
	}
}

func (s *StatusBar) AddItem(icon, label, value string) {
	s.Items = append(s.Items, StatusItem{
		Icon:  icon,
		Label: label,
		Value: value,
	})
}

func (s *StatusBar) Render() string {
	style := lipgloss.NewStyle().
		Background(s.Background).
		Foreground(s.Foreground).
		Width(s.Width).
		Padding(0, 1)

	var items []string
	for _, item := range s.Items {
		itemStr := fmt.Sprintf("%s %s: %s", item.Icon, item.Label, item.Value)
		items = append(items, itemStr)
	}

	content := strings.Join(items, " │ ")

	// Pad to width
	if len(content) < s.Width-2 {
		content += strings.Repeat(" ", s.Width-2-len(content))
	}

	return style.Render(content)
}
