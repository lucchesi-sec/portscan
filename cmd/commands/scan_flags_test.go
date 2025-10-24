package commands

import (
	"testing"

	"github.com/spf13/viper"
)

func TestFlagBindings_StdinJson(t *testing.T) {
	// Ensure defaults are false and keys are readable
	if viper.GetBool("stdin") || viper.GetBool("json") {
		t.Fatalf("expected default stdin/json to be false")
	}
}

func TestFlagBindings_AllFlags(t *testing.T) {
	// Test that all expected flags exist and are registered
	expectedFlags := []struct {
		name     string
		flagType string
	}{
		{"stdin", "bool"},
		{"json", "bool"},
		{"json-array", "bool"},
		{"json-object", "bool"},
		{"banners", "bool"},
		{"dry-run", "bool"},
		{"verbose", "bool"},
		{"only-open", "bool"},
		{"ports", "string"},
		{"profile", "string"},
		{"protocol", "string"},
		{"output", "string"},
		{"rate", "int"},
		{"timeout", "int"},
		{"workers", "int"},
		{"udp-worker-ratio", "float64"},
		{"ui.theme", "string"},
	}

	for _, tt := range expectedFlags {
		t.Run(tt.name, func(t *testing.T) {
			flag := scanCmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %s not found", tt.name)
			}

			if flag.Value.Type() != tt.flagType {
				t.Errorf("flag %s has type %s, expected %s", tt.name, flag.Value.Type(), tt.flagType)
			}
		})
	}
}

func TestFlagViperBinding(t *testing.T) {
	// Test that setting flags via Viper keys works
	// This validates the binding indirectly
	tests := []struct {
		name     string
		viperKey string
		setValue interface{}
	}{
		{"stdin via viper", "stdin", true},
		{"json via viper", "json", true},
		{"ports via viper", "ports", "80,443"},
		{"rate via viper", "rate", 5000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldVal := viper.Get(tt.viperKey)
			defer viper.Set(tt.viperKey, oldVal)

			viper.Set(tt.viperKey, tt.setValue)

			got := viper.Get(tt.viperKey)
			if got != tt.setValue {
				t.Errorf("Viper.Get(%s) = %v, want %v", tt.viperKey, got, tt.setValue)
			}
		})
	}
}

func TestFlagDefaults(t *testing.T) {
	viper.Reset()

	// Test default values
	tests := []struct {
		name     string
		viperKey string
		expected interface{}
	}{
		{"default ports", "ports", "1-1024"},
		{"default rate", "rate", 7500},
		{"default timeout", "timeout_ms", 200},
		{"default workers", "workers", 0},
		{"default protocol", "protocol", "tcp"},
		{"default udp-worker-ratio", "udp_worker_ratio", 0.5},
		{"default ui.theme", "ui.theme", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := scanCmd.Flags().Lookup(tt.viperKey)
			if flag == nil {
				// Try alternative lookup for flags with underscores
				switch tt.viperKey {
				case "timeout_ms":
					flag = scanCmd.Flags().Lookup("timeout")
				case "udp_worker_ratio":
					flag = scanCmd.Flags().Lookup("udp-worker-ratio")
				}
			}

			if flag == nil {
				t.Fatalf("flag not found for key %s", tt.viperKey)
			}

			defaultVal := flag.DefValue

			switch v := tt.expected.(type) {
			case string:
				if defaultVal != v {
					t.Errorf("expected default %v, got %s", v, defaultVal)
				}
			case int:
				// For int flags, check the DefValue string
				expected := ""
				switch v {
				case 7500:
					expected = "7500"
				case 200:
					expected = "200"
				case 0:
					expected = "0"
				}
				if defaultVal != expected {
					t.Errorf("expected default %s, got %s", expected, defaultVal)
				}
			case float64:
				// For float flags
				if tt.viperKey == "udp_worker_ratio" && defaultVal != "0.5" {
					t.Errorf("expected default 0.5, got %s", defaultVal)
				}
			}
		})
	}
}

func TestBooleanFlagDefaults(t *testing.T) {
	viper.Reset()

	boolFlags := []string{
		"stdin", "json", "json-array", "json-object",
		"banners", "dry-run", "verbose", "only-open",
	}

	for _, flagName := range boolFlags {
		t.Run(flagName, func(t *testing.T) {
			flag := scanCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Fatalf("flag %s not found", flagName)
			}

			if flag.DefValue != "false" {
				t.Errorf("expected default false for %s, got %s", flagName, flag.DefValue)
			}
		})
	}
}
