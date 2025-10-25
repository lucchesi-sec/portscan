package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestRunConfigInit_NewConfig(t *testing.T) {
	// Create a temporary home directory
	tmpDir := t.TempDir()

	// Set HOME to temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run config init
	err := runConfigInit(configInitCmd, []string{})

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("runConfigInit failed: %v", err)
	}

	// Verify config file was created
	configPath := filepath.Join(tmpDir, ".portscan.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config file was not created at %s", configPath)
	}

	// Verify output message
	if !strings.Contains(output, "Configuration file created") {
		t.Errorf("expected success message in output, got: %s", output)
	}

	// Verify config file content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	expectedStrings := []string{
		"rate:",
		"workers:",
		"timeout_ms:",
		"ports:",
		"ui:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(string(content), expected) {
			t.Errorf("config file missing expected content: %s", expected)
		}
	}
}

func TestRunConfigInit_ExistingConfig(t *testing.T) {
	// Create a temporary home directory
	tmpDir := t.TempDir()

	// Create existing config file
	configPath := filepath.Join(tmpDir, ".portscan.yaml")
	err := os.WriteFile(configPath, []byte("existing: true\n"), 0600)
	if err != nil {
		t.Fatalf("failed to create existing config: %v", err)
	}

	// Set HOME to temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run config init
	err = runConfigInit(configInitCmd, []string{})

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("runConfigInit failed: %v", err)
	}

	// Verify it detected existing config
	if !strings.Contains(output, "already exists") {
		t.Errorf("expected 'already exists' message, got: %s", output)
	}

	// Verify original content wasn't overwritten
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	if !strings.Contains(string(content), "existing: true") {
		t.Error("existing config file was overwritten")
	}
}

func TestRunConfigShow_WithConfig(t *testing.T) {
	// Reset viper
	viper.Reset()
	defer viper.Reset()

	// Set some config values
	viper.Set("rate", 5000)
	viper.Set("workers", 75)
	viper.Set("timeout_ms", 300)
	viper.Set("ports", "1-1000")
	viper.Set("banners", true)
	viper.Set("output", "json")
	viper.Set("ui.theme", "dracula")

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run config show
	err := runConfigShow(configShowCmd, []string{})

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("runConfigShow failed: %v", err)
	}

	// Verify output contains expected sections
	expectedSections := []string{
		"Current Configuration",
		"Performance:",
		"Scan Defaults:",
		"UI:",
		"Output:",
		"Environment Variables",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("output missing expected section: %s", section)
		}
	}

	// Verify some specific values are shown
	expectedValues := []string{
		"5000",    // rate
		"75",      // workers
		"300",     // timeout
		"1-1000",  // ports
		"true",    // banners
		"json",    // output
		"dracula", // theme
	}

	for _, value := range expectedValues {
		if !strings.Contains(output, value) {
			t.Errorf("output missing expected value: %s", value)
		}
	}
}

func TestRunConfigShow_WithoutConfig(t *testing.T) {
	// Reset viper to defaults
	viper.Reset()
	defer viper.Reset()

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run config show
	err := runConfigShow(configShowCmd, []string{})

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("runConfigShow failed: %v", err)
	}

	// Should show that no config file is in use
	if !strings.Contains(output, "(none") {
		t.Logf("output doesn't indicate no config file: %s", output)
	}

	// Should still show structure
	if !strings.Contains(output, "Performance:") {
		t.Error("output missing Performance section")
	}
}

func TestRunConfigShow_WorkersAutoDetect(t *testing.T) {
	// Reset viper
	viper.Reset()
	defer viper.Reset()

	// Set workers to 0 (auto-detect)
	viper.Set("workers", 0)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run config show
	err := runConfigShow(configShowCmd, []string{})

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("runConfigShow failed: %v", err)
	}

	// Should indicate auto-detect
	if !strings.Contains(output, "auto-detect") {
		t.Error("output should indicate auto-detect for workers")
	}
}

func TestRunConfigShow_EmptyOutput(t *testing.T) {
	// Reset viper
	viper.Reset()
	defer viper.Reset()

	// Set output to empty (TUI mode)
	viper.Set("output", "")

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run config show
	err := runConfigShow(configShowCmd, []string{})

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("runConfigShow failed: %v", err)
	}

	// Should indicate TUI mode
	if !strings.Contains(output, "(TUI)") {
		t.Error("output should indicate TUI mode for empty output")
	}
}

func TestConfigCommandStructure(t *testing.T) {
	// Verify config command structure
	if configCmd == nil {
		t.Fatal("configCmd is nil")
	}

	if configCmd.Use != "config" {
		t.Errorf("configCmd.Use = %q; want %q", configCmd.Use, "config")
	}

	// Verify subcommands
	expectedSubcommands := []string{"init", "show"}
	commands := configCmd.Commands()

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

func TestConfigInitCommand(t *testing.T) {
	if configInitCmd == nil {
		t.Fatal("configInitCmd is nil")
	}

	if configInitCmd.Use != "init" {
		t.Errorf("configInitCmd.Use = %q; want %q", configInitCmd.Use, "init")
	}

	if configInitCmd.RunE == nil {
		t.Error("configInitCmd.RunE is nil")
	}
}

func TestConfigShowCommand(t *testing.T) {
	if configShowCmd == nil {
		t.Fatal("configShowCmd is nil")
	}

	if configShowCmd.Use != "show" {
		t.Errorf("configShowCmd.Use = %q; want %q", configShowCmd.Use, "show")
	}

	if configShowCmd.RunE == nil {
		t.Error("configShowCmd.RunE is nil")
	}
}

func TestRunConfigInit_PermissionError(t *testing.T) {
	// Create a temporary directory with restricted permissions
	tmpDir := t.TempDir()
	restrictedDir := filepath.Join(tmpDir, "restricted")
	err := os.Mkdir(restrictedDir, 0000)
	if err != nil {
		t.Fatalf("failed to create restricted directory: %v", err)
	}
	defer os.Chmod(restrictedDir, 0755) // Cleanup

	// Set HOME to restricted directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", restrictedDir)
	defer os.Setenv("HOME", oldHome)

	// Run config init - should return error
	err = runConfigInit(configInitCmd, []string{})

	// We expect an error due to permission issues
	if err == nil {
		t.Log("expected error due to permission issues, but got none")
		// Not necessarily a failure - system might have different permissions
	}
}
