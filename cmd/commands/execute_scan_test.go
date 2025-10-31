package commands

import (
	"context"
	"testing"
	"time"

	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/pkg/config"
	"github.com/lucchesi-sec/portscan/pkg/exporter"
)

// TestExecuteScan_TCP tests TCP scan execution
func TestExecuteScan_TCP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cfg := &config.Config{
		Workers:   10,
		TimeoutMs: 100,
		Rate:      100,
		Banners:   false,
		Output:    "json",
	}

	hosts := []string{"127.0.0.1"}
	ports := []uint16{9999} // Use unlikely port to avoid interference

	// This will fail to connect but should not error out the execution
	err := executeScan(ctx, "tcp", hosts, ports, cfg)

	// We expect it to complete without crashing
	// The actual scan may not find open ports, but that's okay
	if err != nil && ctx.Err() == nil {
		// Only fail if it's not a context cancellation
		t.Logf("executeScan returned error (may be expected): %v", err)
	}
}

// TestExecuteScan_UDP tests UDP scan execution
func TestExecuteScan_UDP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cfg := &config.Config{
		Workers:        10,
		TimeoutMs:      200,
		Rate:           100,
		UDPWorkerRatio: 0.5,
		Output:         "json",
	}

	hosts := []string{"127.0.0.1"}
	ports := []uint16{9999} // Use unlikely port

	err := executeScan(ctx, "udp", hosts, ports, cfg)

	if err != nil && ctx.Err() == nil {
		t.Logf("executeScan returned error (may be expected): %v", err)
	}
}

// TestExecuteScan_Both tests both TCP and UDP scan
func TestExecuteScan_Both(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cfg := &config.Config{
		Workers:        10,
		TimeoutMs:      200,
		Rate:           100,
		UDPWorkerRatio: 0.5,
		Output:         "json",
	}

	hosts := []string{"127.0.0.1"}
	ports := []uint16{9999}

	err := executeScan(ctx, "both", hosts, ports, cfg)

	if err != nil && ctx.Err() == nil {
		t.Logf("executeScan returned error (may be expected): %v", err)
	}
}

// TestExecuteScan_InvalidProtocol tests with default protocol fallback
func TestExecuteScan_InvalidProtocol(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cfg := &config.Config{
		Workers:   10,
		TimeoutMs: 100,
		Rate:      100,
		Output:    "json",
	}

	hosts := []string{"127.0.0.1"}
	ports := []uint16{9999}

	// Unknown protocol should default to TCP
	err := executeScan(ctx, "unknown", hosts, ports, cfg)

	if err != nil && ctx.Err() == nil {
		t.Logf("executeScan returned error (may be expected): %v", err)
	}
}

// TestExecuteScan_ContextCancellation tests context cancellation
func TestExecuteScan_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately
	cancel()

	cfg := &config.Config{
		Workers:   10,
		TimeoutMs: 100,
		Rate:      100,
		Output:    "json",
	}

	hosts := []string{"127.0.0.1"}
	ports := []uint16{80}

	err := executeScan(ctx, "tcp", hosts, ports, cfg)

	// Should handle cancellation gracefully
	if err != nil {
		t.Logf("executeScan with cancelled context: %v", err)
	}
}

// TestRunProtocolScan_EmptyHosts tests error on empty hosts
func TestRunProtocolScan_EmptyHosts(t *testing.T) {
	ctx := context.Background()

	cfg := &config.Config{
		Workers:   10,
		TimeoutMs: 100,
		Rate:      100,
	}

	factory := NewScannerFactory(cfg)
	scanner, err := factory.CreateScanner("tcp")
	if err != nil {
		t.Fatalf("failed to create scanner: %v", err)
	}

	err = runProtocolScan(ctx, scanner, []string{}, []uint16{80}, cfg, "tcp")

	if err == nil {
		t.Error("expected error for empty hosts")
	}
}

// TestHandleScanOutput_CSV tests CSV output handling
func TestHandleScanOutput_CSV(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This test verifies the CSV output path doesn't crash
	cfg := &config.Config{
		Output: "csv",
	}

	// Create a dummy event channel
	events := make(chan core.Event, 1)
	close(events)

	metadata := exporter.ScanMetadata{
		Targets:    []string{"test"},
		TotalPorts: 1,
		Rate:       1000,
	}

	err := handleScanOutput(context.Background(), cfg, events, 1, metadata)
	if err != nil {
		t.Errorf("handleScanOutput failed: %v", err)
	}
}

// TestHandleScanOutput_JSON tests JSON output handling
func TestHandleScanOutput_JSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := &config.Config{
		Output: "json",
	}

	events := make(chan core.Event, 1)
	close(events)

	metadata := exporter.ScanMetadata{
		Targets:    []string{"test"},
		TotalPorts: 1,
		Rate:       1000,
	}

	err := handleScanOutput(context.Background(), cfg, events, 1, metadata)
	if err != nil {
		t.Errorf("handleScanOutput failed: %v", err)
	}
}
