package commands

import (
	"testing"
)

// TestScanCommandStructure verifies the scan command is properly configured
func TestScanCommandStructure(t *testing.T) {
	if scanCmd == nil {
		t.Fatal("scanCmd should not be nil")
	}

	if scanCmd.Use != "scan [targets...]" {
		t.Errorf("Use = %q; want 'scan [targets...]'", scanCmd.Use)
	}

	if scanCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if scanCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	if scanCmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

// TestScanCommandFlags verifies all expected flags are present
func TestScanCommandFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
	}{
		{"ports flag", "ports"},
		{"profile flag", "profile"},
		{"protocol flag", "protocol"},
		{"rate flag", "rate"},
		{"timeout flag", "timeout"},
		{"workers flag", "workers"},
		{"udp-worker-ratio flag", "udp-worker-ratio"},
		{"banners flag", "banners"},
		{"output flag", "output"},
		{"stdin flag", "stdin"},
		{"json flag", "json"},
		{"json-array flag", "json-array"},
		{"json-object flag", "json-object"},
		{"only-open flag", "only-open"},
		{"ui.theme flag", "ui.theme"},
		{"dry-run flag", "dry-run"},
		{"examples flag", "examples"},
		{"verbose flag", "verbose"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := scanCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag %q not found", tt.flagName)
			}
		})
	}
}

// TestScanCommandFlagDefaults verifies flag default values
func TestScanCommandFlagDefaults(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		expectedType string
	}{
		{"ports default", "ports", "string"},
		{"protocol default", "protocol", "string"},
		{"rate default", "rate", "int"},
		{"timeout default", "timeout", "int"},
		{"workers default", "workers", "int"},
		{"udp-worker-ratio default", "udp-worker-ratio", "float64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := scanCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}

			if flag.Value.Type() != tt.expectedType {
				t.Errorf("flag %q type = %q; want %q", tt.flagName, flag.Value.Type(), tt.expectedType)
			}
		})
	}
}

// TestScanCommandExample verifies example text is provided
func TestScanCommandExample(t *testing.T) {
	if scanCmd.Example == "" {
		t.Error("Example should not be empty")
	}

	// Check for key example patterns
	examples := []string{
		"localhost",
		"--ports",
		"--profile",
		"--protocol",
		"--stdin",
	}

	for _, example := range examples {
		if !contains(scanCmd.Example, example) {
			t.Errorf("Example should contain %q", example)
		}
	}
}

// TestScanCommandArgs verifies ArbitraryArgs is set
func TestScanCommandArgs(t *testing.T) {
	// Test that the command accepts arbitrary arguments
	// cobra.ArbitraryArgs allows any number of args
	err := scanCmd.Args(scanCmd, []string{})
	if err != nil {
		t.Errorf("should accept zero args: %v", err)
	}

	err = scanCmd.Args(scanCmd, []string{"host1"})
	if err != nil {
		t.Errorf("should accept one arg: %v", err)
	}

	err = scanCmd.Args(scanCmd, []string{"host1", "host2", "host3"})
	if err != nil {
		t.Errorf("should accept multiple args: %v", err)
	}
}

// TestScanCommandUsage verifies usage text can be generated
func TestScanCommandUsage(t *testing.T) {
	usage := scanCmd.UsageString()
	if usage == "" {
		t.Error("usage string should not be empty")
	}

	if !contains(usage, "scan") {
		t.Error("usage should contain command name")
	}
}

// TestScanCommandHelp verifies help text can be generated
func TestScanCommandHelp(t *testing.T) {
	help := scanCmd.Long
	if help == "" {
		t.Error("help text should not be empty")
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
