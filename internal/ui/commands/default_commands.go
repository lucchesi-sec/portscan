package commands

import (
	tea "github.com/charmbracelet/bubbletea"
)

// DefaultCommands returns the set of default commands for the application
func DefaultCommands() []Command {
	return []Command{
		// Navigation commands
		{
			ID:          "nav-up",
			Name:        "Move Up",
			Description: "Move selection up in the results table",
			Alias:       "up",
			Keys:        []string{"↑", "k"},
			Category:    CommandTypeNavigation,
			Action: func() tea.Cmd {
				return nil // Will be handled through UIAction
			},
			IsActive: nil, // Always available
		},
		{
			ID:          "nav-down",
			Name:        "Move Down",
			Description: "Move selection down in the results table",
			Alias:       "down",
			Keys:        []string{"↓", "j"},
			Category:    CommandTypeNavigation,
			Action: func() tea.Cmd {
				return nil // Will be handled through UIAction
			},
			IsActive: nil, // Always available
		},
		{
			ID:          "nav-top",
			Name:        "Go to Top",
			Description: "Move to the top of the results table",
			Alias:       "top",
			Keys:        []string{"Home", "g"},
			Category:    CommandTypeNavigation,
			Action: func() tea.Cmd {
				return nil // Will be handled through UIAction
			},
			IsActive: nil,
		},
		{
			ID:          "nav-bottom",
			Name:        "Go to Bottom",
			Description: "Move to the bottom of the results table",
			Alias:       "bottom",
			Keys:        []string{"End", "G"},
			Category:    CommandTypeNavigation,
			Action: func() tea.Cmd {
				return nil // Will be handled through UIAction
			},
			IsActive: nil,
		},
		{
			ID:          "nav-page-up",
			Name:        "Page Up",
			Description: "Move up one page in the results table",
			Alias:       "page-up",
			Keys:        []string{"PgUp", "Ctrl+U"},
			Category:    CommandTypeNavigation,
			Action: func() tea.Cmd {
				return nil // Will be handled through UIAction
			},
			IsActive: nil,
		},
		{
			ID:          "nav-page-down",
			Name:        "Page Down",
			Description: "Move down one page in the results table",
			Alias:       "page-down",
			Keys:        []string{"PgDn", "Ctrl+D"},
			Category:    CommandTypeNavigation,
			Action: func() tea.Cmd {
				return nil // Will be handled through UIAction
			},
			IsActive: nil,
		},

		// Action commands
		{
			ID:          "action-pause",
			Name:        "Pause/Resume Scan",
			Description: "Pause or resume the active scan",
			Alias:       "pause",
			Keys:        []string{"p", "Space"},
			Category:    CommandTypeAction,
			Action: func() tea.Cmd {
				// This will be handled by the UI when the command is executed
				return nil // Placeholder - will be updated during execution
			},
			IsActive: nil, // Always available
		},
		{
			ID:          "action-sort",
			Name:        "Open Sort Modal",
			Description: "Open the sort options modal",
			Alias:       "sort",
			Keys:        []string{"s"},
			Category:    CommandTypeAction,
			Action: func() tea.Cmd {
				// This will be handled by the UI when the command is executed
				return nil // Placeholder - will be updated during execution
			},
			IsActive: nil,
		},
		{
			ID:          "action-filter",
			Name:        "Open Filter Modal",
			Description: "Open the filter options modal",
			Alias:       "filter",
			Keys:        []string{"f"},
			Category:    CommandTypeAction,
			Action: func() tea.Cmd {
				// This will be handled by the UI when the command is executed
				return nil // Placeholder - will be updated during execution
			},
			IsActive: nil,
		},
		{
			ID:          "action-reset-filters",
			Name:        "Reset Filters",
			Description: "Reset all applied filters",
			Alias:       "reset",
			Keys:        []string{"r"},
			Category:    CommandTypeAction,
			Action: func() tea.Cmd {
				// This will be handled by the UI when the command is executed
				return nil // Placeholder - will be updated during execution
			},
			IsActive: nil,
		},
		{
			ID:          "action-toggle-open-only",
			Name:        "Toggle Open Only",
			Description: "Show only open ports or all ports",
			Alias:       "open-only",
			Keys:        []string{"o"},
			Category:    CommandTypeAction,
			Action: func() tea.Cmd {
				// This will be handled by the UI when the command is executed
				return nil // Placeholder - will be updated during execution
			},
			IsActive: nil,
		},
		{
			ID:          "action-search",
			Name:        "Search Banners",
			Description: "Search through service banners",
			Alias:       "search",
			Keys:        []string{"/"},
			Category:    CommandTypeAction,
			Action: func() tea.Cmd {
				// This will be handled by the UI when the command is executed
				return nil // Placeholder - will be updated during execution
			},
			IsActive: nil,
		},
		{
			ID:          "action-toggle-dashboard",
			Name:        "Toggle Dashboard",
			Description: "Show/hide the dashboard view",
			Alias:       "dashboard",
			Keys:        []string{"D"},
			Category:    CommandTypeAction,
			Action: func() tea.Cmd {
				return nil // Will be handled through UIAction
			},
			IsActive: nil,
		},
		{
			ID:          "action-view-details",
			Name:        "View Selected Details",
			Description: "Show detailed information for selected port",
			Alias:       "details",
			Keys:        []string{"Enter"},
			Category:    CommandTypeAction,
			Action: func() tea.Cmd {
				return nil // Will be handled through UIAction
			},
			IsActive: func(model interface{}) bool {
				// Only available when there are results to view
				return true // The UI handles this internally
			},
		},

		// View commands
		{
			ID:          "view-help",
			Name:        "Toggle Help",
			Description: "Show/hide the help screen",
			Alias:       "help",
			Keys:        []string{"?"},
			Category:    CommandTypeView,
			Action: func() tea.Cmd {
				return nil // Will be handled through UIAction
			},
			IsActive: nil,
		},
		{
			ID:          "view-clear",
			Name:        "Clear Screen",
			Description: "Clear the terminal screen",
			Alias:       "clear",
			Keys:        []string{"Ctrl+L"},
			Category:    CommandTypeView,
			Action: func() tea.Cmd {
				return nil // Will be handled through UIAction
			},
			IsActive: nil,
		},
		{
			ID:          "view-quit",
			Name:        "Quit Application",
			Description: "Exit the port scanner",
			Alias:       "quit",
			Keys:        []string{"q", "Ctrl+C", "Esc"},
			Category:    CommandTypeView,
			Action: func() tea.Cmd {
				return nil // Will be handled through UIAction
			},
			IsActive: nil,
		},
	}
}
