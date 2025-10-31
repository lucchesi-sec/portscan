package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

func TestRegistry_AddCommand(t *testing.T) {
	reg := NewRegistry()
	cmd := Command{
		ID:          "test",
		Name:        "Test Command",
		Description: "A test command",
		Category:    CommandTypeAction,
		Action:      func() tea.Cmd { return nil },
	}

	reg.AddCommand(cmd)
	commands := reg.GetCommands()

	if len(commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(commands))
	}
	if commands[0].ID != "test" {
		t.Errorf("Expected command ID 'test', got '%s'", commands[0].ID)
	}
}

func TestRegistry_AddCommands(t *testing.T) {
	reg := NewRegistry()
	cmds := []Command{
		{ID: "test1", Name: "Test 1", Category: CommandTypeAction},
		{ID: "test2", Name: "Test 2", Category: CommandTypeAction},
	}

	reg.AddCommands(cmds)
	commands := reg.GetCommands()

	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}
	if commands[0].ID != "test1" || commands[1].ID != "test2" {
		t.Errorf("Command IDs not preserved")
	}
}

func TestRegistry_GetCommandsByCategory(t *testing.T) {
	reg := NewRegistry()
	cmds := []Command{
		{ID: "nav", Name: "Nav", Category: CommandTypeNavigation},
		{ID: "action", Name: "Action", Category: CommandTypeAction},
		{ID: "nav2", Name: "Nav2", Category: CommandTypeNavigation},
	}
	reg.AddCommands(cmds)

	navCmds := reg.GetCommandsByCategory(CommandTypeNavigation)
	if len(navCmds) != 2 {
		t.Errorf("Expected 2 navigation commands, got %d", len(navCmds))
	}
}

func TestRegistry_GetCategories(t *testing.T) {
	reg := NewRegistry()
	cmds := []Command{
		{ID: "nav", Name: "Nav", Category: CommandTypeNavigation},
		{ID: "action", Name: "Action", Category: CommandTypeAction},
		{ID: "view", Name: "View", Category: CommandTypeView},
	}
	reg.AddCommands(cmds)

	categories := reg.GetCategories()
	if len(categories) != 3 {
		t.Errorf("Expected 3 categories, got %d", len(categories))
	}
}

func TestRegistry_GetCommandByID(t *testing.T) {
	reg := NewRegistry()
	cmd := Command{ID: "test", Name: "Test", Category: CommandTypeAction}
	reg.AddCommand(cmd)

	foundCmd, exists := reg.GetCommandByID("test")
	if !exists {
		t.Error("Expected command to exist")
	}
	if foundCmd.ID != "test" {
		t.Errorf("Expected command ID 'test', got '%s'", foundCmd.ID)
	}

	_, exists = reg.GetCommandByID("nonexistent")
	if exists {
		t.Error("Expected command not to exist")
	}
}

func TestRegistry_ExecuteCommand(t *testing.T) {
	executed := false
	reg := NewRegistry()
	cmd := Command{
		ID:       "execute-test",
		Name:     "Execute Test",
		Category: CommandTypeAction,
		Action: func() tea.Cmd {
			executed = true
			return nil
		},
	}
	reg.AddCommand(cmd)

	// Execute the command
	reg.ExecuteCommand("execute-test", nil)
	if !executed {
		t.Error("Expected command to be executed")
	}
}

func TestRegistry_GetActiveCommands(t *testing.T) {
	reg := NewRegistry()
	cmd1 := Command{
		ID:       "active",
		Name:     "Active",
		Category: CommandTypeAction,
		IsActive: func(interface{}) bool { return true },
	}
	cmd2 := Command{
		ID:       "inactive",
		Name:     "Inactive",
		Category: CommandTypeAction,
		IsActive: func(interface{}) bool { return false },
	}

	reg.AddCommands([]Command{cmd1, cmd2})
	active := reg.GetActiveCommands(nil)

	if len(active) != 1 {
		t.Errorf("Expected 1 active command, got %d", len(active))
	}
	if active[0].ID != "active" {
		t.Errorf("Expected active command ID 'active', got '%s'", active[0].ID)
	}
}
