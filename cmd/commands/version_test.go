package commands

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// TestVersionCommand verifies the version command structure
func TestVersionCommand(t *testing.T) {
	if versionCmd == nil {
		t.Fatal("versionCmd should not be nil")
	}

	if versionCmd.Use != "version" {
		t.Errorf("Use = %q; want 'version'", versionCmd.Use)
	}

	if versionCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if versionCmd.Run == nil {
		t.Error("Run should be set")
	}
}

// TestVersionCommandOutput verifies version command output
func TestVersionCommandOutput(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the version command
	versionCmd.Run(versionCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected information
	if output == "" {
		t.Error("version command produced no output")
	}

	expectedContents := []string{
		"portscan version",
		"commit:",
		"built:",
	}

	for _, expected := range expectedContents {
		if !bytes.Contains(buf.Bytes(), []byte(expected)) {
			t.Errorf("output missing expected content: %s", expected)
		}
	}
}

// TestVersionVariables verifies version variables are defined
func TestVersionVariables(t *testing.T) {
	// version, commit, and buildDate should be defined (even if "unknown")
	if version == "" {
		t.Error("version should not be empty")
	}

	if commit == "" {
		t.Error("commit should not be empty")
	}

	if buildDate == "" {
		t.Error("buildDate should not be empty")
	}
}

// TestVersionCommandExecution verifies the command can be executed
func TestVersionCommandExecution(t *testing.T) {
	// Create a buffer to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the command
	err := versionCmd.Execute()
	if err != nil {
		t.Fatalf("version command execution failed: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	if buf.Len() == 0 {
		t.Error("version command produced no output")
	}
}
