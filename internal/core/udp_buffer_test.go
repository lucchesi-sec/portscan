package core

import (
	"testing"
	"time"
)

// TestUDPBufferSizeConfiguration tests that the UDP buffer size is properly configured
func TestUDPBufferSizeConfiguration(t *testing.T) {
	// Test default behavior
	cfg := &Config{
		Workers: 100,
		Timeout: 200 * time.Millisecond,
		// UDPBufferSize not set, should default to 1024
	}
	scanner := NewScanner(cfg)

	if scanner.config.UDPBufferSize != 1024 {
		t.Errorf("Expected UDPBufferSize to default to 1024, got %d", scanner.config.UDPBufferSize)
	}

	// Test explicit setting
	customBufferSize := 2048
	cfg2 := &Config{
		Workers:       100,
		Timeout:       200 * time.Millisecond,
		UDPBufferSize: customBufferSize,
	}
	scanner2 := NewScanner(cfg2)

	if scanner2.config.UDPBufferSize != customBufferSize {
		t.Errorf("Expected UDPBufferSize to be %d, got %d", customBufferSize, scanner2.config.UDPBufferSize)
	}
}
