package core

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func TestNewUDPScanner(t *testing.T) {
	cfg := &Config{
		Workers:        10,
		Timeout:        100 * time.Millisecond,
		RateLimit:      1000,
		BannerGrab:     true,
		UDPWorkerRatio: 0.5,
	}

	scanner := NewUDPScanner(cfg)
	if scanner == nil {
		t.Fatal("Expected scanner to be created")
	}

	if scanner.Scanner == nil {
		t.Fatal("Expected base scanner to be initialized")
	}

	if scanner.serviceProbes == nil {
		t.Fatal("Expected service probes to be initialized")
	}

	if scanner.customProbes == nil {
		t.Fatal("Expected custom probes to be initialized")
	}

	if scanner.probeStats == nil {
		t.Fatal("Expected probe stats to be initialized")
	}
}

func TestUDPProbes(t *testing.T) {
	probes := initUDPProbes()

	// Check that common UDP ports have probes
	expectedPorts := []uint16{53, 123, 161, 500, 1194}
	for _, port := range expectedPorts {
		if _, exists := probes[port]; !exists {
			t.Errorf("Expected probe for port %d", port)
		}
	}
}

func TestGetProbeForPort(t *testing.T) {
	scanner := &UDPScanner{
		serviceProbes: initUDPProbes(),
	}

	// Test known port
	probe := scanner.getProbeForPort(53)
	if len(probe) == 0 {
		t.Error("Expected non-empty probe for DNS port 53")
	}

	// Test unknown port - should return empty probe
	probe = scanner.getProbeForPort(12345)
	if len(probe) != 0 {
		t.Error("Expected empty probe for unknown port")
	}
}

func TestParseUDPResponse(t *testing.T) {
	scanner := &UDPScanner{}

	tests := []struct {
		port     uint16
		data     []byte
		contains string // Changed from exact match to contains check
	}{
		{53, []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, "DNS"},
		{123, make([]byte, 48), "NTP"},
		{161, []byte{0x30}, "SNMP"},
		{1194, []byte{0x38}, "OpenVPN"},
		{9999, []byte("Unknown service"), "Unknown service"},
	}

	for _, tt := range tests {
		result := scanner.parseUDPResponse(tt.port, tt.data)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("Port %d: expected result to contain %q, got %q", tt.port, tt.contains, result)
		}
	}
}

func TestUDPScannerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start a local UDP server for testing
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	port := uint16(conn.LocalAddr().(*net.UDPAddr).Port)

	// Handle incoming UDP packets
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, addr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				return
			}
			// Echo back the received data
			if n > 0 {
				_, _ = conn.WriteToUDP([]byte("ECHO"), addr)
			}
		}
	}()

	// Create scanner
	cfg := &Config{
		Workers:        1,
		Timeout:        500 * time.Millisecond,
		BannerGrab:     true,
		UDPWorkerRatio: 1.0, // Use all workers for test
	}

	scanner := NewUDPScanner(cfg)

	// Scan the test port
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resultChan := scanner.Results()
	go scanner.ScanRange(ctx, "127.0.0.1", []uint16{port})

	// Wait for result
	timeout := time.After(3 * time.Second)
	foundResult := false

	for !foundResult {
		select {
		case event := <-resultChan:
			if event.Type != EventTypeResult {
				continue
			}
			r := event.Result
			if r.Port != port {
				t.Errorf("Expected port %d, got %d", port, r.Port)
			}
			if r.Protocol != "udp" {
				t.Errorf("Expected protocol udp, got %s", r.Protocol)
			}
			foundResult = true
		case <-timeout:
			t.Error("Timeout waiting for scan result")
			return
		}
	}
}

func TestCustomProbes(t *testing.T) {
	scanner := NewUDPScanner(&Config{
		Workers:        10,
		Timeout:        100 * time.Millisecond,
		RateLimit:      0,
		BannerGrab:     false,
		UDPWorkerRatio: 0.5,
	})

	// Test adding a custom probe
	customProbe := []byte{0x01, 0x02, 0x03, 0x04}
	scanner.AddCustomProbe(12345, customProbe)

	// Test that the custom probe is returned
	probe := scanner.getProbeForPort(12345)
	if len(probe) != len(customProbe) {
		t.Errorf("Expected custom probe length %d, got %d", len(customProbe), len(probe))
	}

	// Test that service probes still work
	dnsProbe := scanner.getProbeForPort(53)
	if len(dnsProbe) == 0 {
		t.Error("Expected DNS probe to still be available")
	}

	// Test that custom probe takes precedence over service probe
	scanner.AddCustomProbe(53, customProbe)
	overrideProbe := scanner.getProbeForPort(53)
	if len(overrideProbe) != len(customProbe) {
		t.Error("Expected custom probe to override service probe")
	}
}

func TestBuildDNSProbe(t *testing.T) {
	probe := buildDNSProbe()
	if len(probe) == 0 {
		t.Error("DNS probe should not be empty")
	}

	// Check transaction ID and flags
	if probe[2] != 0x01 || probe[3] != 0x00 {
		t.Error("DNS probe has incorrect flags")
	}
}

func TestBuildNTPProbe(t *testing.T) {
	probe := buildNTPProbe()
	if len(probe) != 48 {
		t.Errorf("NTP probe should be 48 bytes, got %d", len(probe))
	}

	// Check NTP version and mode
	if probe[0] != 0x1b {
		t.Error("NTP probe has incorrect version/mode")
	}
}

func TestBuildSNMPProbe(t *testing.T) {
	probe := buildSNMPProbe()
	if len(probe) == 0 {
		t.Error("SNMP probe should not be empty")
	}

	// Check SNMP packet structure
	if probe[0] != 0x30 {
		t.Error("SNMP probe should start with SEQUENCE (0x30)")
	}
}

func TestUDPScannerContextCancellation(t *testing.T) {
	cfg := &Config{
		Workers:        1,
		Timeout:        100 * time.Millisecond,
		RateLimit:      0,
		BannerGrab:     false,
		UDPWorkerRatio: 1.0,
	}

	scanner := NewUDPScanner(cfg)

	// Create a context that we cancel immediately to test cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Run scan with cancelled context
	resultChan := scanner.Results()
	go scanner.ScanRange(ctx, "127.0.0.1", []uint16{53})

	// Should return quickly due to cancellation
	select {
	case _, ok := <-resultChan:
		// Channel might be closed or have no results due to immediate cancellation
		if !ok {
			// Channel closed, which is expected
			return
		}
		// If we get a result, it should be processed quickly
	case <-time.After(500 * time.Millisecond):
		t.Error("Test timed out - context cancellation not working properly")
	}
}

func TestUDPScannerRateLimiting(t *testing.T) {
	cfg := &Config{
		Workers:        2,
		Timeout:        100 * time.Millisecond,
		RateLimit:      10, // Very low rate limit
		BannerGrab:     false,
		UDPWorkerRatio: 1.0,
	}

	scanner := NewUDPScanner(cfg)

	// Test that scanner initializes with rate limiting
	if scanner.rateTicker == nil {
		t.Error("Expected rate ticker to be initialized")
	}
}

func BenchmarkUDPScanning(b *testing.B) {
	cfg := &Config{
		Workers:        10,
		Timeout:        100 * time.Millisecond,
		RateLimit:      1000,
		BannerGrab:     false,
		UDPWorkerRatio: 1.0,
	}

	ports := []uint16{53, 123, 161, 500, 1194}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a new scanner for each iteration
		scanner := NewUDPScanner(cfg)
		ctx := context.Background()
		results := scanner.Results()

		// Run scan in a separate goroutine
		done := make(chan struct{})
		go func() {
			defer close(done)
			scanner.ScanRange(ctx, "127.0.0.1", ports)
		}()

		// Drain results in this goroutine
		for range results {
			// Just consume
		}

		// Wait for scan to complete
		<-done
	}
}

func TestUDPWorkerRatio(t *testing.T) {
	// Create a mock UDP scanner to test worker calculation
	cfg := &Config{
		Workers:        100,
		Timeout:        100 * time.Millisecond,
		RateLimit:      0,
		BannerGrab:     false,
		UDPWorkerRatio: 0.3, // 30% of workers
	}

	scanner := NewUDPScanner(cfg)

	// Access the private method indirectly by examining behavior
	// We can't directly test the private worker calculation, but we can verify
	// the configuration is properly set
	if scanner.config.UDPWorkerRatio != 0.3 {
		t.Errorf("Expected UDPWorkerRatio to be 0.3, got %f", scanner.config.UDPWorkerRatio)
	}

	// Test with default ratio (0, should use 0.5)
	cfg2 := &Config{
		Workers:        100,
		Timeout:        100 * time.Millisecond,
		RateLimit:      0,
		BannerGrab:     false,
		UDPWorkerRatio: 0, // Should default to 0.5
	}

	scanner2 := NewUDPScanner(cfg2)
	if scanner2.config.UDPWorkerRatio != 0.5 {
		t.Errorf("Expected default UDPWorkerRatio to be 0.5, got %f", scanner2.config.UDPWorkerRatio)
	}
}

func TestProbeStats(t *testing.T) {
	scanner := NewUDPScanner(&Config{
		Workers:        10,
		Timeout:        100 * time.Millisecond,
		RateLimit:      0,
		BannerGrab:     false,
		UDPWorkerRatio: 0.5,
	})

	// Test recording probe attempts
	scanner.recordProbeAttempt(53, true)  // Success
	scanner.recordProbeAttempt(53, false) // Failure
	scanner.recordProbeAttempt(123, true) // Success

	// Check stats
	stats := scanner.GetProbeStats()
	if len(stats) != 2 {
		t.Errorf("Expected stats for 2 ports, got %d", len(stats))
	}

	dnsStats := stats[53]
	if dnsStats.Sent != 2 {
		t.Errorf("Expected 2 probes sent for DNS, got %d", dnsStats.Sent)
	}
	if dnsStats.Responses != 1 {
		t.Errorf("Expected 1 response for DNS, got %d", dnsStats.Responses)
	}
	if dnsStats.Successes != 1 {
		t.Errorf("Expected 1 success for DNS, got %d", dnsStats.Successes)
	}

	ntpStats := stats[123]
	if ntpStats.Sent != 1 {
		t.Errorf("Expected 1 probe sent for NTP, got %d", ntpStats.Sent)
	}
	if ntpStats.Responses != 1 {
		t.Errorf("Expected 1 response for NTP, got %d", ntpStats.Responses)
	}
	if ntpStats.Successes != 1 {
		t.Errorf("Expected 1 success for NTP, got %d", ntpStats.Successes)
	}
}
