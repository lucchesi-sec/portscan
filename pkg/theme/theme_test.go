package theme

import (
	"testing"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Theme
	}{
		{
			name:     "default theme",
			input:    "default",
			expected: Default,
		},
		{
			name:     "dracula theme",
			input:    "dracula",
			expected: Dracula,
		},
		{
			name:     "monokai theme",
			input:    "monokai",
			expected: Monokai,
		},
		{
			name:     "unknown theme falls back to default",
			input:    "unknown",
			expected: Default,
		},
		{
			name:     "empty string falls back to default",
			input:    "",
			expected: Default,
		},
		{
			name:     "case sensitive - uppercase falls back to default",
			input:    "DRACULA",
			expected: Default,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTheme(tt.input)
			if got.Name != tt.expected.Name {
				t.Errorf("GetTheme(%q).Name = %q, want %q", tt.input, got.Name, tt.expected.Name)
			}
			if got.Primary != tt.expected.Primary {
				t.Errorf("GetTheme(%q).Primary = %q, want %q", tt.input, got.Primary, tt.expected.Primary)
			}
		})
	}
}

func TestThemeStyles(t *testing.T) {
	theme := Default

	t.Run("HeaderStyle", func(t *testing.T) {
		style := theme.HeaderStyle()
		// Just verify it returns a style without errors
		_ = style.Render("test")
	})

	t.Run("StatusStyle", func(t *testing.T) {
		style := theme.StatusStyle()
		_ = style.Render("test")
	})

	t.Run("SuccessStyle", func(t *testing.T) {
		style := theme.SuccessStyle()
		_ = style.Render("test")
	})

	t.Run("ErrorStyle", func(t *testing.T) {
		style := theme.ErrorStyle()
		_ = style.Render("test")
	})

	t.Run("WarningStyle", func(t *testing.T) {
		style := theme.WarningStyle()
		_ = style.Render("test")
	})
}

func TestThemeProperties(t *testing.T) {
	themes := []Theme{Default, Dracula, Monokai}

	for _, theme := range themes {
		t.Run(theme.Name, func(t *testing.T) {
			if theme.Name == "" {
				t.Error("Theme name should not be empty")
			}
			if theme.Primary == "" {
				t.Error("Primary color should not be empty")
			}
			if theme.Background == "" {
				t.Error("Background color should not be empty")
			}
			if theme.Foreground == "" {
				t.Error("Foreground color should not be empty")
			}
		})
	}
}
