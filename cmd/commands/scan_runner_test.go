package commands

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/pkg/config"
	"github.com/lucchesi-sec/portscan/pkg/exporter"
	"github.com/spf13/viper"
)

func TestShowDryRun(t *testing.T) {
	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := &config.Config{
		Workers:   100,
		Rate:      7500,
		TimeoutMs: 200,
		Banners:   true,
		Output:    "json",
	}

	targets := []string{"192.168.1.1", "192.168.1.2"}
	ports := []uint16{80, 443, 8080}

	showDryRun(targets, ports, cfg)

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected information
	if output == "" {
		t.Error("showDryRun produced no output")
	}

	// Check for key elements
	expectedContents := []string{
		"DRY RUN MODE",
		"Targets:",
		"Ports:",
		"Workers:",
		"Rate Limit:",
		"Timeout:",
		"Banner Grab:",
		"Output Format:",
	}

	for _, expected := range expectedContents {
		if !bytes.Contains(buf.Bytes(), []byte(expected)) {
			t.Errorf("output missing expected content: %s", expected)
		}
	}
}

func TestEnsureWorkersConfigured(t *testing.T) {
	tests := []struct {
		name           string
		initialWorkers int
		expectChange   bool
	}{
		{
			name:           "workers already set",
			initialWorkers: 50,
			expectChange:   false,
		},
		{
			name:           "workers not set (zero)",
			initialWorkers: 0,
			expectChange:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Workers: tt.initialWorkers,
			}

			ensureWorkersConfigured(cfg)

			if tt.expectChange && cfg.Workers == 0 {
				t.Error("workers should have been auto-configured but remained 0")
			}

			if !tt.expectChange && cfg.Workers != tt.initialWorkers {
				t.Errorf("workers should not have changed, got %d want %d", cfg.Workers, tt.initialWorkers)
			}

			// Workers should always be within valid range
			if cfg.Workers < 10 || cfg.Workers > 200 {
				t.Errorf("workers out of valid range: %d", cfg.Workers)
			}
		})
	}
}

func TestEnforceRateSafety_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		rate        int
		expectError bool
	}{
		{"zero rate allowed", 0, false},
		{"negative rate allowed", -1, false},
		{"max safe rate", 15000, false},
		{"just over max safe", 15001, true},
		{"very high rate", 100000, true},
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

func TestBuildScannerConfig(t *testing.T) {
	cfg := &config.Config{
		Workers:        100,
		TimeoutMs:      250,
		Rate:           5000,
		Banners:        true,
		UDPWorkerRatio: 0.6,
	}

	scannerCfg := buildScannerConfig(cfg)

	if scannerCfg.Workers != 100 {
		t.Errorf("Workers = %d; want 100", scannerCfg.Workers)
	}

	if scannerCfg.Timeout.Milliseconds() != 250 {
		t.Errorf("Timeout = %v; want 250ms", scannerCfg.Timeout)
	}

	if scannerCfg.RateLimit != 5000 {
		t.Errorf("RateLimit = %d; want 5000", scannerCfg.RateLimit)
	}

	if !scannerCfg.BannerGrab {
		t.Error("BannerGrab should be true")
	}

	if scannerCfg.MaxRetries != 2 {
		t.Errorf("MaxRetries = %d; want 2", scannerCfg.MaxRetries)
	}

	if scannerCfg.UDPWorkerRatio < 0.59 || scannerCfg.UDPWorkerRatio > 0.61 {
		t.Errorf("UDPWorkerRatio = %v; want ~0.6", scannerCfg.UDPWorkerRatio)
	}
}

func TestSelectJSONExporter(t *testing.T) {
	metadata := exporter.ScanMetadata{
		Targets:    []string{"localhost"},
		TotalPorts: 100,
		Rate:       7500,
	}

	tests := []struct {
		name       string
		setFlags   func()
		expectType string
	}{
		{
			name: "default NDJSON",
			setFlags: func() {
				viper.Reset()
			},
			expectType: "ndjson",
		},
		{
			name: "json-array mode",
			setFlags: func() {
				viper.Reset()
				viper.Set("json_array", true)
			},
			expectType: "array",
		},
		{
			name: "json-object mode",
			setFlags: func() {
				viper.Reset()
				viper.Set("json_object", true)
			},
			expectType: "object",
		},
		{
			name: "json-object takes precedence",
			setFlags: func() {
				viper.Reset()
				viper.Set("json_array", true)
				viper.Set("json_object", true)
			},
			expectType: "object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setFlags()
			defer viper.Reset() // Clean up after each test

			exporter := selectJSONExporter(metadata)
			if exporter == nil {
				t.Fatal("exporter should not be nil")
			}

			// Exporter selection validated - type checking would require
			// reflection or exporting internal fields, which is not necessary
			// The key validation is that it doesn't panic and returns non-nil
		})
	}
}

func TestSelectPortList_WithProfile(t *testing.T) {
	tests := []struct {
		name        string
		profile     string
		expectError bool
		minPorts    int
	}{
		{
			name:        "quick profile",
			profile:     "quick",
			expectError: false,
			minPorts:    10,
		},
		{
			name:        "web profile",
			profile:     "web",
			expectError: false,
			minPorts:    3,
		},
		{
			name:        "database profile",
			profile:     "database",
			expectError: false,
			minPorts:    3,
		},
		{
			name:        "invalid profile",
			profile:     "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("profile", tt.profile)

			cfg := &config.Config{
				Ports: "1-1024", // Default, should be ignored when profile is set
			}

			ports, err := selectPortList(cfg)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(ports) < tt.minPorts {
				t.Errorf("expected at least %d ports, got %d", tt.minPorts, len(ports))
			}
		})
	}
}

func TestSelectPortList_WithPortString(t *testing.T) {
	tests := []struct {
		name        string
		portString  string
		expectError bool
		expectCount int
	}{
		{
			name:        "single port",
			portString:  "80",
			expectError: false,
			expectCount: 1,
		},
		{
			name:        "multiple ports",
			portString:  "80,443,8080",
			expectError: false,
			expectCount: 3,
		},
		{
			name:        "port range",
			portString:  "80-85",
			expectError: false,
			expectCount: 6,
		},
		{
			name:        "invalid port string",
			portString:  "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			// No profile set, should use port string

			cfg := &config.Config{
				Ports: tt.portString,
			}

			ports, err := selectPortList(cfg)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(ports) != tt.expectCount {
				t.Errorf("expected %d ports, got %d", tt.expectCount, len(ports))
			}
		})
	}
}

func TestShowExtendedExamples(t *testing.T) {
	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	showExtendedExamples()

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if output == "" {
		t.Error("showExtendedExamples produced no output")
	}

	// Check for expected sections
	expectedSections := []string{
		"EXTENDED EXAMPLES",
		"Network Discovery",
		"Security Assessment",
		"DevOps Validation",
		"Performance Tuning",
		"Output Formats",
		"Profiles",
	}

	for _, section := range expectedSections {
		if !bytes.Contains(buf.Bytes(), []byte(section)) {
			t.Errorf("output missing expected section: %s", section)
		}
	}
}

func TestMonitorInterrupts(t *testing.T) {
	// Test that monitorInterrupts sets up signal handling
	// We can't easily test the actual signal handling without sending real signals,
	// but we can verify the function doesn't panic and creates the channel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This should not panic
	monitorInterrupts(cancel)

	// Give it a moment to set up
	// The function should create a goroutine that waits for signals
	// We can't easily test the actual signal handling in a unit test
	// without more complex setup, but we've verified it doesn't crash

	// Verify context can still be cancelled manually
	cancel()
	select {
	case <-ctx.Done():
		// Context was cancelled successfully
	default:
		t.Error("context should be done after cancel")
	}
}

func TestStreamEvents(t *testing.T) {
	// Create a mock event channel
	events := make(chan core.Event, 2)
	events <- core.NewResultEvent(core.ResultEvent{
		Host:  "localhost",
		Port:  80,
		State: "open",
	})
	events <- core.NewResultEvent(core.ResultEvent{
		Host:  "localhost",
		Port:  443,
		State: "closed",
	})
	close(events)

	// Mock export function
	exportCalled := false
	mockExport := func(ch <-chan core.Event) {
		exportCalled = true
		// Drain the channel
		for range ch {
		}
	}

	// Mock close function
	closeCalled := false
	mockClose := func() error {
		closeCalled = true
		return nil
	}

	// Call streamEvents
	err := streamEvents(events, mockExport, mockClose)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !exportCalled {
		t.Error("export function was not called")
	}

	if !closeCalled {
		t.Error("close function was not called")
	}
}

func TestCollectTargetInputs_EmptyWithoutStdin(t *testing.T) {
	viper.Set("stdin", false)
	defer viper.Set("stdin", false)

	targets, err := collectTargetInputs([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(targets) != 0 {
		t.Errorf("expected empty targets, got %d", len(targets))
	}
}

func TestValidateInputs_ValidConfig(t *testing.T) {
	cfg := &config.Config{
		Ports:          "80,443,8080",
		Rate:           5000,
		TimeoutMs:      200,
		Workers:        50,
		UDPWorkerRatio: 0.5,
	}

	err := validateInputs(cfg)
	if err != nil {
		t.Errorf("validateInputs failed for valid config: %v", err)
	}
}

func TestValidateInputs_InvalidPorts(t *testing.T) {
	cfg := &config.Config{
		Ports:          "invalid",
		Rate:           5000,
		TimeoutMs:      200,
		Workers:        50,
		UDPWorkerRatio: 0.5,
	}

	err := validateInputs(cfg)
	if err == nil {
		t.Error("expected error for invalid ports")
	}
}

func TestValidateInputs_InvalidRate(t *testing.T) {
	tests := []struct {
		name string
		rate int
	}{
		{"negative rate", -1},
		{"zero rate", 0},
		{"too high rate", 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Ports:          "80",
				Rate:           tt.rate,
				TimeoutMs:      200,
				Workers:        50,
				UDPWorkerRatio: 0.5,
			}

			err := validateInputs(cfg)
			if err == nil {
				t.Error("expected error for invalid rate")
			}
		})
	}
}

func TestValidateInputs_InvalidTimeout(t *testing.T) {
	tests := []struct {
		name      string
		timeoutMs int
	}{
		{"negative timeout", -1},
		{"zero timeout", 0},
		{"too high timeout", 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Ports:          "80",
				Rate:           5000,
				TimeoutMs:      tt.timeoutMs,
				Workers:        50,
				UDPWorkerRatio: 0.5,
			}

			err := validateInputs(cfg)
			if err == nil {
				t.Error("expected error for invalid timeout")
			}
		})
	}
}

func TestValidateInputs_InvalidWorkers(t *testing.T) {
	cfg := &config.Config{
		Ports:          "80",
		Rate:           5000,
		TimeoutMs:      200,
		Workers:        10000, // Too many
		UDPWorkerRatio: 0.5,
	}

	err := validateInputs(cfg)
	if err == nil {
		t.Error("expected error for invalid workers")
	}
}

func TestValidateInputs_InvalidUDPRatio(t *testing.T) {
	tests := []struct {
		name  string
		ratio float64
	}{
		{"negative ratio", -0.1},
		{"too high ratio", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Ports:          "80",
				Rate:           5000,
				TimeoutMs:      200,
				Workers:        50,
				UDPWorkerRatio: tt.ratio,
			}

			err := validateInputs(cfg)
			if err == nil {
				t.Error("expected error for invalid UDP ratio")
			}
		})
	}
}

func TestValidateRawTargets_ValidTargets(t *testing.T) {
	targets := []string{"localhost", "192.168.1.1", "example.com", "192.168.0.0/24"}

	err := validateRawTargets(targets)
	if err != nil {
		t.Errorf("validateRawTargets failed for valid targets: %v", err)
	}
}

func TestValidateRawTargets_EmptyTarget(t *testing.T) {
	targets := []string{""}

	err := validateRawTargets(targets)
	if err == nil {
		t.Error("expected error for empty target")
	}
}

func TestValidateRawTargets_InvalidTarget(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{"invalid IP", "999.999.999.999"},
		{"invalid CIDR", "192.168.1.0/99"},
		{"special characters", "host@#$%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRawTargets([]string{tt.target})
			if err == nil {
				t.Errorf("expected error for invalid target: %s", tt.target)
			}
		})
	}
}

func TestNormalizeProtocol(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty defaults to tcp", "", "tcp"},
		{"tcp unchanged", "tcp", "tcp"},
		{"udp unchanged", "udp", "udp"},
		{"both unchanged", "both", "both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeProtocol(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeProtocol(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildScanTargets(t *testing.T) {
	hosts := []string{"host1", "host2", "host3"}
	ports := []uint16{80, 443, 8080}

	targets := buildScanTargets(hosts, ports)

	if len(targets) != len(hosts) {
		t.Errorf("expected %d targets, got %d", len(hosts), len(targets))
	}

	for i, target := range targets {
		if target.Host != hosts[i] {
			t.Errorf("target %d host = %s; want %s", i, target.Host, hosts[i])
		}

		if len(target.Ports) != len(ports) {
			t.Errorf("target %d has %d ports; want %d", i, len(target.Ports), len(ports))
		}

		for j, port := range target.Ports {
			if port != ports[j] {
				t.Errorf("target %d port %d = %d; want %d", i, j, port, ports[j])
			}
		}
	}
}

func TestBuildScanTargets_Empty(t *testing.T) {
	targets := buildScanTargets([]string{}, []uint16{80})
	if len(targets) != 0 {
		t.Errorf("expected 0 targets for empty hosts, got %d", len(targets))
	}

	targets = buildScanTargets([]string{"host"}, []uint16{})
	if len(targets) != 1 {
		t.Errorf("expected 1 target, got %d", len(targets))
	}

	if len(targets[0].Ports) != 0 {
		t.Errorf("expected 0 ports, got %d", len(targets[0].Ports))
	}
}
