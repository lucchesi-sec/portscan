package core

import (
	"testing"
	"time"
)

// TestUDPJitterConfiguration tests that the UDP jitter is properly configured
func TestUDPJitterConfiguration(t *testing.T) {
	// Test default behavior
	cfg := &Config{
		Workers: 100,
		Timeout: 200 * time.Millisecond,
		// UDPJitterMaxMs not set, should default to 10
	}
	scanner := NewScanner(cfg)

	if scanner.config.UDPJitterMaxMs != 10 {
		t.Errorf("Expected UDPJitterMaxMs to default to 10, got %d", scanner.config.UDPJitterMaxMs)
	}

	// Test explicit setting
	customJitter := 20
	cfg2 := &Config{
		Workers:         100,
		Timeout:         200 * time.Millisecond,
		UDPJitterMaxMs:  customJitter,
	}
	scanner2 := NewScanner(cfg2)

	if scanner2.config.UDPJitterMaxMs != customJitter {
		t.Errorf("Expected UDPJitterMaxMs to be %d, got %d", customJitter, scanner2.config.UDPJitterMaxMs)
	}
}
