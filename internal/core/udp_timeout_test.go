package core

import (
	"testing"
	"time"
)

// TestUDPReadTimeoutConfiguration tests that the UDP read timeout is properly configured
func TestUDPReadTimeoutConfiguration(t *testing.T) {
	// Test default behavior
	cfg := &Config{
		Workers: 100,
		Timeout: 200 * time.Millisecond,
		// UDPReadTimeout not set, should default to Timeout
	}
	scanner := NewScanner(cfg)

	if scanner.config.UDPReadTimeout != scanner.config.Timeout {
		t.Errorf("Expected UDPReadTimeout to default to Timeout, got %v, expected %v",
			scanner.config.UDPReadTimeout, scanner.config.Timeout)
	}

	// Test explicit setting
	customTimeout := 500 * time.Millisecond
	cfg2 := &Config{
		Workers:        100,
		Timeout:        200 * time.Millisecond,
		UDPReadTimeout: customTimeout,
	}
	scanner2 := NewScanner(cfg2)

	if scanner2.config.UDPReadTimeout != customTimeout {
		t.Errorf("Expected UDPReadTimeout to be %v, got %v",
			customTimeout, scanner2.config.UDPReadTimeout)
	}
}
