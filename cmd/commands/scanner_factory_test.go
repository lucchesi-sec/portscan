package commands

import (
	"testing"

	"github.com/lucchesi-sec/portscan/pkg/config"
)

// TestNewScannerFactory verifies factory creation
func TestNewScannerFactory(t *testing.T) {
	cfg := &config.Config{
		Workers:   100,
		TimeoutMs: 200,
		Rate:      5000,
		Banners:   true,
	}

	factory := NewScannerFactory(cfg)

	if factory == nil {
		t.Fatal("factory should not be nil")
	}

	// Factory stores config internally, we verify by creating a scanner
	scanner, err := factory.CreateScanner("tcp")
	if err != nil {
		t.Errorf("factory should be properly initialized: %v", err)
	}
	if scanner == nil {
		t.Error("factory should create valid scanner")
	}
}

// TestScannerFactory_CreateScanner_TCP verifies TCP scanner creation
func TestScannerFactory_CreateScanner_TCP(t *testing.T) {
	cfg := &config.Config{
		Workers:   50,
		TimeoutMs: 200,
		Rate:      5000,
		Banners:   false,
	}

	factory := NewScannerFactory(cfg)
	scanner, err := factory.CreateScanner("tcp")

	if err != nil {
		t.Fatalf("failed to create TCP scanner: %v", err)
	}

	if scanner == nil {
		t.Error("TCP scanner should not be nil")
	}
}

// TestScannerFactory_CreateScanner_UDP verifies UDP scanner creation
func TestScannerFactory_CreateScanner_UDP(t *testing.T) {
	cfg := &config.Config{
		Workers:        50,
		TimeoutMs:      500,
		Rate:           1000,
		Banners:        false,
		UDPWorkerRatio: 0.5,
	}

	factory := NewScannerFactory(cfg)
	scanner, err := factory.CreateScanner("udp")

	if err != nil {
		t.Fatalf("failed to create UDP scanner: %v", err)
	}

	if scanner == nil {
		t.Error("UDP scanner should not be nil")
	}
}

// TestScannerFactory_CreateScanner_Invalid verifies error on invalid protocol
func TestScannerFactory_CreateScanner_Invalid(t *testing.T) {
	cfg := &config.Config{
		Workers:   50,
		TimeoutMs: 200,
		Rate:      5000,
	}

	factory := NewScannerFactory(cfg)
	scanner, err := factory.CreateScanner("invalid")

	if err == nil {
		t.Error("expected error for invalid protocol")
	}

	if scanner != nil {
		t.Error("scanner should be nil for invalid protocol")
	}
}

// TestScannerFactory_CreateScanner_EmptyProtocol verifies error on empty protocol
func TestScannerFactory_CreateScanner_EmptyProtocol(t *testing.T) {
	cfg := &config.Config{
		Workers:   50,
		TimeoutMs: 200,
		Rate:      5000,
	}

	factory := NewScannerFactory(cfg)
	scanner, err := factory.CreateScanner("")

	if err == nil {
		t.Error("expected error for empty protocol")
	}

	if scanner != nil {
		t.Error("scanner should be nil for empty protocol")
	}
}

// TestScannerFactory_MultipleScanners verifies multiple scanner creation
func TestScannerFactory_MultipleScanners(t *testing.T) {
	cfg := &config.Config{
		Workers:        100,
		TimeoutMs:      200,
		Rate:           5000,
		UDPWorkerRatio: 0.6,
	}

	factory := NewScannerFactory(cfg)

	// Create TCP scanner
	tcpScanner, err := factory.CreateScanner("tcp")
	if err != nil {
		t.Fatalf("failed to create TCP scanner: %v", err)
	}

	// Create UDP scanner
	udpScanner, err := factory.CreateScanner("udp")
	if err != nil {
		t.Fatalf("failed to create UDP scanner: %v", err)
	}

	// Both scanners should be valid and different instances
	if tcpScanner == nil || udpScanner == nil {
		t.Error("scanners should not be nil")
	}

	// They should be different instances
	if tcpScanner == udpScanner {
		t.Error("TCP and UDP scanners should be different instances")
	}
}
