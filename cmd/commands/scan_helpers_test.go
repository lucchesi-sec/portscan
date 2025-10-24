package commands

import (
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/spf13/viper"
)

func TestGetOptimalWorkerCount(t *testing.T) {
	count := getOptimalWorkerCount()

	// Should be between numCPU*50, with min 10 and max 200
	expected := runtime.NumCPU() * 50
	if expected > 200 {
		expected = 200
	}
	if expected < 10 {
		expected = 10
	}

	if count != expected {
		t.Errorf("expected worker count %d, got %d", expected, count)
	}

	// Should never be less than 10
	if count < 10 {
		t.Errorf("worker count should be at least 10, got %d", count)
	}

	// Should never exceed 200
	if count > 200 {
		t.Errorf("worker count should not exceed 200, got %d", count)
	}
}

func TestCollectTargetInputs_Args(t *testing.T) {
	// Reset viper
	viper.Set("stdin", false)

	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "single target",
			args:     []string{"192.168.1.1"},
			expected: []string{"192.168.1.1"},
		},
		{
			name:     "multiple targets",
			args:     []string{"192.168.1.1", "192.168.1.2", "example.com"},
			expected: []string{"192.168.1.1", "192.168.1.2", "example.com"},
		},
		{
			name:     "no targets",
			args:     []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := collectTargetInputs(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d targets, got %d", len(tt.expected), len(result))
			}

			for i, target := range result {
				if i < len(tt.expected) && target != tt.expected[i] {
					t.Errorf("at index %d: expected %s, got %s", i, tt.expected[i], target)
				}
			}
		})
	}
}

func TestCollectTargetInputs_Stdin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single target",
			input:    "192.168.1.1",
			expected: []string{"192.168.1.1"},
		},
		{
			name:     "multiple targets newline separated",
			input:    "192.168.1.1\n192.168.1.2\nexample.com",
			expected: []string{"192.168.1.1", "192.168.1.2", "example.com"},
		},
		{
			name:     "multiple targets space separated",
			input:    "192.168.1.1 192.168.1.2 example.com",
			expected: []string{"192.168.1.1", "192.168.1.2", "example.com"},
		},
		{
			name:     "mixed whitespace",
			input:    "192.168.1.1\n  192.168.1.2   \n\nexample.com  ",
			expected: []string{"192.168.1.1", "192.168.1.2", "example.com"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set stdin flag
			viper.Set("stdin", true)
			defer viper.Set("stdin", false)

			// Replace stdin
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			// Write test input
			go func() {
				defer w.Close()
				io.WriteString(w, tt.input)
			}()

			// Restore stdin after test
			defer func() { os.Stdin = oldStdin }()

			result, err := collectTargetInputs([]string{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d targets, got %d", len(tt.expected), len(result))
			}

			for i, target := range result {
				if i < len(tt.expected) && target != tt.expected[i] {
					t.Errorf("at index %d: expected %s, got %s", i, tt.expected[i], target)
				}
			}
		})
	}
}

func TestResolveTargetList(t *testing.T) {
	tests := []struct {
		name        string
		inputs      []string
		expectError bool
		minHosts    int // Minimum expected hosts (for CIDR)
	}{
		{
			name:        "single IP",
			inputs:      []string{"192.168.1.1"},
			expectError: false,
			minHosts:    1,
		},
		{
			name:        "single hostname",
			inputs:      []string{"localhost"},
			expectError: false,
			minHosts:    1,
		},
		{
			name:        "multiple hosts",
			inputs:      []string{"192.168.1.1", "192.168.1.2"},
			expectError: false,
			minHosts:    2,
		},
		{
			name:        "duplicate hosts removed",
			inputs:      []string{"192.168.1.1", "192.168.1.1", "192.168.1.2"},
			expectError: false,
			minHosts:    2,
		},
		{
			name:        "small CIDR",
			inputs:      []string{"192.168.1.0/30"},
			expectError: false,
			minHosts:    4, // /30 = 4 hosts
		},
		{
			name:        "empty list",
			inputs:      []string{},
			expectError: true,
		},
		{
			name:        "invalid target",
			inputs:      []string{"invalid!!!host"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveTargetList(tt.inputs)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) < tt.minHosts {
				t.Errorf("expected at least %d hosts, got %d", tt.minHosts, len(result))
			}
		})
	}
}

// Note: selectPortList and showDryRun tests would require full config.Config setup
// which is complex. These are better tested through integration tests.

func TestEnforceRateSafety(t *testing.T) {
	tests := []struct {
		name        string
		rate        int
		expectError bool
	}{
		{"safe rate 1000", 1000, false},
		{"safe rate 7500", 7500, false},
		{"max safe rate 15000", 15000, false},
		{"unsafe rate 16000", 16000, true},
		{"very high rate 50000", 50000, true},
		{"zero rate", 0, false},
		{"negative rate", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := enforceRateSafety(tt.rate)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
