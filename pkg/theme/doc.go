// Package theme provides color schemes and styling for the TUI.
//
// This package defines visual themes for the terminal user interface, including
// color palettes and styling functions. Themes ensure consistent visual appearance
// across all UI components (tables, progress bars, status text, etc.).
//
// Built-in Themes:
//
//   - default: Clean light/dark theme suitable for most terminals
//   - dracula: Popular dark theme with vibrant colors
//   - monokai: Dark theme inspired by Monokai color scheme
//
// Example usage:
//
//	// Get a theme by name
//	theme := theme.GetTheme("dracula")
//
//	// Use theme colors in lipgloss styles
//	headerStyle := lipgloss.NewStyle().
//	    Foreground(theme.Primary).
//	    Background(theme.Background).
//	    Bold(true)
//
//	// Render styled text
//	fmt.Println(headerStyle.Render("Port Scanner"))
//
// Theme Structure:
//
// Each theme defines semantic colors:
//   - Primary: Main brand color, highlights
//   - Secondary: Supporting color, secondary highlights
//   - Success: Success states (open ports)
//   - Warning: Warning states (filtered ports)
//   - Error: Error states (closed ports, errors)
//   - Background: Background color
//   - Foreground: Default text color
//   - Muted: Subdued text (help text, timestamps)
//
// Custom Themes:
//
// Applications can register custom themes:
//
//	customTheme := theme.Theme{
//	    Primary:    lipgloss.Color("#FF6188"),
//	    Secondary:  lipgloss.Color("#FC9867"),
//	    Success:    lipgloss.Color("#A9DC76"),
//	    Warning:    lipgloss.Color("#FFD866"),
//	    Error:      lipgloss.Color("#FF6188"),
//	    Background: lipgloss.Color("#2D2A2E"),
//	    Foreground: lipgloss.Color("#FCFCFA"),
//	    Muted:      lipgloss.Color("#727072"),
//	}
//	theme.Register("custom", customTheme)
//
// Color Format:
//
// Colors use lipgloss.Color which supports:
//   - Hex colors: "#FF6188"
//   - Named colors: "red", "blue"
//   - ANSI colors: "1", "2", etc.
//
// The theme system automatically adapts to terminal capabilities.
package theme
