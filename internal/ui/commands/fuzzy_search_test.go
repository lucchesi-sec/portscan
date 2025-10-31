package commands

import (
	"testing"
)

func TestFuzzySearch_EmptyQuery(t *testing.T) {
	commands := []Command{
		{Name: "Test Command", Description: "A test command"},
		{Name: "Another Command", Description: "Another test command"},
	}

	results := FuzzySearch(commands, "")

	if len(results) != 2 {
		t.Errorf("Expected 2 results for empty query, got %d", len(results))
	}

	for _, result := range results {
		if result.Score != 1.0 {
			t.Errorf("Expected score of 1.0 for empty query, got %f", result.Score)
		}
	}
}

func TestFuzzySearch_ExactMatch(t *testing.T) {
	commands := []Command{
		{ID: "test", Name: "Test Command", Description: "A test command"},
		{ID: "other", Name: "Other Command", Description: "Another command"},
	}

	results := FuzzySearch(commands, "test")

	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'test' query, got %d", len(results))
	} else if results[0].Command.ID != "test" {
		t.Errorf("Expected command 'test', got '%s'", results[0].Command.ID)
	}
}

func TestFuzzySearch_SubstringMatch(t *testing.T) {
	commands := []Command{
		{ID: "test", Name: "Test Command", Description: "A test command"},
		{ID: "demo", Name: "Demo Command", Description: "A demo command"},
	}

	results := FuzzySearch(commands, "Com")

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'Com' query, got %d", len(results))
	}

	// Both commands should contain "Com" in "Command"
	for _, result := range results {
		if result.Score <= 0 {
			t.Errorf("Expected positive score for substring match, got %f", result.Score)
		}
	}
}

func TestFuzzySearch_CaseInsensitive(t *testing.T) {
	commands := []Command{
		{ID: "test", Name: "Test Command", Description: "A TEST command"},
	}

	results := FuzzySearch(commands, "test")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for case-insensitive 'test' query, got %d", len(results))
	}

	results = FuzzySearch(commands, "TEST")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for case-insensitive 'TEST' query, got %d", len(results))
	}
}

func TestFuzzySearchByCategory(t *testing.T) {
	model := &struct{}{} // Mock model
	registry := NewRegistry()

	cmds := []Command{
		{ID: "nav", Name: "Navigate", Category: CommandTypeNavigation},
		{ID: "action", Name: "Action", Category: CommandTypeAction},
	}
	registry.AddCommands(cmds)

	grouped := FuzzySearchByCategory(registry, model, "")

	if len(grouped) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(grouped))
	}
	if len(grouped[CommandTypeNavigation]) != 1 {
		t.Errorf("Expected 1 navigation command, got %d", len(grouped[CommandTypeNavigation]))
	}
	if len(grouped[CommandTypeAction]) != 1 {
		t.Errorf("Expected 1 action command, got %d", len(grouped[CommandTypeAction]))
	}
}

func TestCalculateFuzzyScore_Basic(t *testing.T) {
	cmd := Command{
		Name:        "Test Command",
		Description: "A test description",
	}

	score, _ := calculateFuzzyScore(cmd, "test")
	if score <= 0 {
		t.Errorf("Expected positive score for 'test' in 'Test Command', got %f", score)
	}
}

func TestCalculateFuzzyScore_PrefixMatch(t *testing.T) {
	cmd := Command{
		Name: "Test Command",
	}

	score1, _ := calculateFuzzyScore(cmd, "tes") // Prefix match
	score2, _ := calculateFuzzyScore(cmd, "omm") // Non-prefix match

	// Prefix matches should have higher scores than non-prefix matches
	if score1 <= score2 {
		t.Errorf("Prefix match should score higher than non-prefix match: %f vs %f", score1, score2)
	}
}

func TestCalculateFuzzyScore_NoMatch(t *testing.T) {
	cmd := Command{
		Name: "Test Command",
	}

	score, _ := calculateFuzzyScore(cmd, "xyz")
	if score > 0 {
		t.Errorf("Expected 0 score for no match, got %f", score)
	}
}
