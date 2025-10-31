package commands

import (
	"strings"
)

// MatchResult represents the result of a fuzzy search match
type MatchResult struct {
	Command Command
	Score   float64 // Higher scores are better matches
	Matches []int   // Indices of matched characters in the search string (for highlighting)
}

// FuzzySearch searches commands based on a query string
func FuzzySearch(commands []Command, query string) []MatchResult {
	if query == "" {
		// Return all commands with default score if no query
		results := make([]MatchResult, len(commands))
		for i, cmd := range commands {
			results[i] = MatchResult{Command: cmd, Score: 1.0, Matches: []int{}}
		}
		return results
	}

	query = strings.ToLower(query)
	results := make([]MatchResult, 0, len(commands))

	for _, cmd := range commands {
		score, matches := calculateFuzzyScore(cmd, query)
		if score > 0 {
			results = append(results, MatchResult{
				Command: cmd,
				Score:   score,
				Matches: matches,
			})
		}
	}

	// Sort by score in descending order
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Score < results[j].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// calculateFuzzyScore calculates a fuzzy matching score for a command against a query
func calculateFuzzyScore(cmd Command, query string) (float64, []int) {
	// Build search text from multiple fields
	searchText := strings.ToLower(cmd.Name + " " + cmd.Description + " " + cmd.Alias)
	searchRunes := []rune(searchText)
	queryRunes := []rune(query)

	if len(queryRunes) == 0 {
		return 1.0, []int{}
	}

	// Check for substring match first (high score)
	substringIndex := strings.Index(searchText, query)
	if substringIndex != -1 {
		// Prefix matches get higher scores
		if substringIndex == 0 {
			// Perfect prefix match
			return 10.0, make([]int, len(queryRunes))
		}
		// Just substring match
		matches := make([]int, len(queryRunes))
		for i := range matches {
			matches[i] = substringIndex + i
		}
		return 8.0, matches
	}

	// Fuzzy matching algorithm
	bestScore := 0.0
	bestMatches := []int{}

	// Try matching query characters in sequence within the search text
	queryPos := 0
	matches := []int{}
	score := 0.0
	gaps := 0 // Count gaps between matched characters

	for i, char := range searchRunes {
		if queryPos < len(queryRunes) && char == queryRunes[queryPos] {
			matches = append(matches, i)
			queryPos++

			// Bonus for consecutive matches (acronyms, etc.)
			if queryPos > 1 && i > 0 && searchRunes[i-1] == queryRunes[queryPos-2] {
				score += 2.0
			} else {
				score += 1.0
			}
		} else if queryPos > 0 {
			// Count gaps between matches
			gaps++
		}
	}

	if queryPos == len(queryRunes) {
		// Apply penalty for gaps between matches
		gapPenalty := float64(gaps) * 0.1
		finalScore := score - gapPenalty

		// Bonus for matches closer to the start of the text
		if len(matches) > 0 {
			positionBonus := 5.0 / (float64(matches[0]) + 1.0)
			finalScore += positionBonus
		}

		bestScore = finalScore
		bestMatches = matches
	}

	return bestScore, bestMatches
}

// FuzzySearchByCategory performs fuzzy search and groups results by category
func FuzzySearchByCategory(registry *Registry, model interface{}, query string) map[CommandType][]MatchResult {
	activeCommands := registry.GetActiveCommands(model)
	results := FuzzySearch(activeCommands, query)

	grouped := make(map[CommandType][]MatchResult)
	for _, result := range results {
		grouped[result.Command.Category] = append(grouped[result.Command.Category], result)
	}

	return grouped
}
