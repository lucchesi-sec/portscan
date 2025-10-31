package commands

import (
	tea "github.com/charmbracelet/bubbletea"
)

// CommandType defines categories for commands
type CommandType string

const (
	CommandTypeNavigation CommandType = "navigation"
	CommandTypeAction     CommandType = "action"
	CommandTypeView       CommandType = "view"
	CommandTypeFilter     CommandType = "filter"
	CommandTypeSort       CommandType = "sort"
)

// Command represents an executable command in the command palette
type Command struct {
	ID          string      // Unique identifier for the command
	Name        string      // Display name
	Description string      // Description of what the command does
	Alias       string      // Optional alias for the command
	Keys        []string    // Keyboard shortcuts (for display only)
	Category    CommandType // Category for grouping
	// Action is a function that returns a tea.Cmd to execute the command
	Action func() tea.Cmd
	// UIAction is a function that executes a command with direct access to the UI model
	UIAction func(model interface{}) tea.Cmd
	// IsActive returns whether the command should be available in the current context
	IsActive func(model interface{}) bool
}

// Registry holds all available commands
type Registry struct {
	commands []Command
}

// NewRegistry creates a new command registry with default commands
func NewRegistry() *Registry {
	return &Registry{
		commands: []Command{},
	}
}

// AddCommand adds a command to the registry
func (r *Registry) AddCommand(cmd Command) {
	r.commands = append(r.commands, cmd)
}

// AddCommands adds multiple commands to the registry
func (r *Registry) AddCommands(cmds []Command) {
	r.commands = append(r.commands, cmds...)
}

// GetCommands returns all registered commands
func (r *Registry) GetCommands() []Command {
	return r.commands
}

// GetCommandsByCategory returns commands filtered by category
func (r *Registry) GetCommandsByCategory(category CommandType) []Command {
	var filtered []Command
	for _, cmd := range r.commands {
		if cmd.Category == category {
			filtered = append(filtered, cmd)
		}
	}
	return filtered
}

// GetCategories returns all unique command categories
func (r *Registry) GetCategories() []CommandType {
	seen := make(map[CommandType]bool)
	var categories []CommandType
	for _, cmd := range r.commands {
		if !seen[cmd.Category] {
			seen[cmd.Category] = true
			categories = append(categories, cmd.Category)
		}
	}
	return categories
}

// ExecuteCommand finds a command by ID and executes it if available
func (r *Registry) ExecuteCommand(id string, model interface{}) tea.Cmd {
	for _, cmd := range r.commands {
		if cmd.ID == id && (cmd.IsActive == nil || cmd.IsActive(model)) {
			if cmd.UIAction != nil {
				return cmd.UIAction(model)
			}
			return cmd.Action()
		}
	}
	return nil
}

// GetActiveCommands returns all commands that are currently active in the given model context
func (r *Registry) GetActiveCommands(model interface{}) []Command {
	var active []Command
	for _, cmd := range r.commands {
		if cmd.IsActive == nil || cmd.IsActive(model) {
			active = append(active, cmd)
		}
	}
	return active
}

// GetCommandByID finds a command by its ID
func (r *Registry) GetCommandByID(id string) (*Command, bool) {
	for _, cmd := range r.commands {
		if cmd.ID == id {
			return &cmd, true
		}
	}
	return nil, false
}
