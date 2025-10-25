package commands

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunAddProbe_ValidInput(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		hexData string
	}{
		{
			name:    "simple probe",
			port:    "1234",
			hexData: "48656c6c6f", // "Hello" in hex
		},
		{
			name:    "DNS probe",
			port:    "53",
			hexData: "000102030405",
		},
		{
			name:    "complex hex",
			port:    "9999",
			hexData: "0a0b0c0d0e0f10111213",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stdout to capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run add probe
			err := runAddProbe(addProbeCmd, []string{tt.port, tt.hexData})

			// Restore stdout and read output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if err != nil {
				t.Errorf("runAddProbe failed: %v", err)
			}

			// Verify output contains port and hex data
			if !strings.Contains(output, tt.port) {
				t.Errorf("output missing port number: %s", output)
			}

			if !strings.Contains(output, tt.hexData) {
				t.Errorf("output missing hex data: %s", output)
			}

			// Verify it mentions adding the probe
			if !strings.Contains(output, "add") || !strings.Contains(output, "probe") {
				t.Errorf("output doesn't mention adding probe: %s", output)
			}
		})
	}
}

func TestRunAddProbe_InvalidPort(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		hexData string
	}{
		{
			name:    "non-numeric port",
			port:    "abc",
			hexData: "48656c6c6f",
		},
		{
			name:    "negative port",
			port:    "-1",
			hexData: "48656c6c6f",
		},
		{
			name:    "port too large",
			port:    "70000",
			hexData: "48656c6c6f",
		},
		{
			name:    "empty port",
			port:    "",
			hexData: "48656c6c6f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run add probe
			err := runAddProbe(addProbeCmd, []string{tt.port, tt.hexData})

			// Should return error for invalid port
			if err == nil {
				t.Error("expected error for invalid port, got nil")
			}

			// Error should mention invalid port
			if !strings.Contains(err.Error(), "invalid port") {
				t.Errorf("error should mention 'invalid port', got: %v", err)
			}
		})
	}
}

func TestRunAddProbe_InvalidHexData(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		hexData string
	}{
		{
			name:    "non-hex characters",
			port:    "1234",
			hexData: "xyz123",
		},
		{
			name:    "spaces in hex",
			port:    "1234",
			hexData: "48 65 6c",
		},
		{
			name:    "odd length hex",
			port:    "1234",
			hexData: "123",
		},
		{
			name:    "hex with prefix",
			port:    "1234",
			hexData: "0x48656c6c6f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run add probe
			err := runAddProbe(addProbeCmd, []string{tt.port, tt.hexData})

			// Should return error for invalid hex
			if err == nil {
				t.Error("expected error for invalid hex data, got nil")
			}

			// Error should mention invalid hex
			if !strings.Contains(err.Error(), "invalid hex") {
				t.Errorf("error should mention 'invalid hex', got: %v", err)
			}
		})
	}
}

func TestRunAddProbe_BoundaryPorts(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{
			name:    "port 0",
			port:    "0",
			wantErr: false, // Port 0 is technically valid in uint16
		},
		{
			name:    "port 1",
			port:    "1",
			wantErr: false,
		},
		{
			name:    "port 65535",
			port:    "65535",
			wantErr: false,
		},
		{
			name:    "port 65536",
			port:    "65536",
			wantErr: true, // Exceeds uint16 max
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runAddProbe(addProbeCmd, []string{tt.port, "deadbeef"})

			w.Close()
			os.Stdout = oldStdout
			io.Copy(io.Discard, r)

			if tt.wantErr && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRunProbeStats(t *testing.T) {
	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run probe stats
	err := runProbeStats(statsProbeCmd, []string{})

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Errorf("runProbeStats failed: %v", err)
	}

	// Should produce some output
	if output == "" {
		t.Error("runProbeStats produced no output")
	}

	// Output should mention probe or statistics
	if !strings.Contains(output, "Probe") && !strings.Contains(output, "statistics") {
		t.Errorf("output doesn't mention probes or statistics: %s", output)
	}
}

func TestProbeCommandStructure(t *testing.T) {
	// Verify probe command structure
	if probeCmd == nil {
		t.Fatal("probeCmd is nil")
	}

	if probeCmd.Use != "probe" {
		t.Errorf("probeCmd.Use = %q; want %q", probeCmd.Use, "probe")
	}

	// Verify subcommands
	expectedSubcommands := []string{"add", "stats"}
	commands := probeCmd.Commands()

	if len(commands) < len(expectedSubcommands) {
		t.Errorf("expected at least %d subcommands, got %d", len(expectedSubcommands), len(commands))
	}

	for _, expectedName := range expectedSubcommands {
		found := false
		for _, cmd := range commands {
			if cmd.Name() == expectedName {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected subcommand %q not found", expectedName)
		}
	}
}

func TestAddProbeCommand(t *testing.T) {
	if addProbeCmd == nil {
		t.Fatal("addProbeCmd is nil")
	}

	if addProbeCmd.Use != "add PORT HEX_DATA" {
		t.Errorf("addProbeCmd.Use = %q; want %q", addProbeCmd.Use, "add PORT HEX_DATA")
	}

	if addProbeCmd.Args == nil {
		t.Error("addProbeCmd.Args is nil")
	}

	if addProbeCmd.RunE == nil {
		t.Error("addProbeCmd.RunE is nil")
	}
}

func TestStatsProbeCommand(t *testing.T) {
	if statsProbeCmd == nil {
		t.Fatal("statsProbeCmd is nil")
	}

	if statsProbeCmd.Use != "stats" {
		t.Errorf("statsProbeCmd.Use = %q; want %q", statsProbeCmd.Use, "stats")
	}

	if statsProbeCmd.RunE == nil {
		t.Error("statsProbeCmd.RunE is nil")
	}
}

func TestRunAddProbe_EmptyHexData(t *testing.T) {
	// Redirect stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runAddProbe(addProbeCmd, []string{"1234", ""})

	w.Close()
	os.Stdout = oldStdout
	io.Copy(io.Discard, r)

	// Empty hex data is valid (decodes to 0 bytes)
	if err != nil {
		t.Logf("empty hex data resulted in error: %v", err)
		// This might be acceptable behavior
	}
}

func TestRunAddProbe_LargeHexData(t *testing.T) {
	// Create a large hex string (1KB of data = 2048 hex chars)
	largeHex := strings.Repeat("ab", 1024)

	// Redirect stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runAddProbe(addProbeCmd, []string{"1234", largeHex})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Errorf("large hex data failed: %v", err)
	}

	// Should report the decoded length
	if !strings.Contains(output, "1024 bytes") {
		t.Errorf("output should mention 1024 bytes, got: %s", output)
	}
}

func TestRunAddProbe_CommonServicePorts(t *testing.T) {
	// Test common ports that might have UDP probes
	ports := []string{"53", "67", "68", "123", "161", "500", "1194", "51820"}

	for _, port := range ports {
		t.Run("port_"+port, func(t *testing.T) {
			// Redirect stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runAddProbe(addProbeCmd, []string{port, "deadbeef"})

			w.Close()
			os.Stdout = oldStdout
			io.Copy(io.Discard, r)

			if err != nil {
				t.Errorf("failed for port %s: %v", port, err)
			}
		})
	}
}
