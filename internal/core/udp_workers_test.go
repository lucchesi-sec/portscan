package core

import (
	"testing"
	"time"
)

// TestUDPWorkerCalculation tests that the UDP worker calculation respects user intent
func TestUDPWorkerCalculation(t *testing.T) {
	tests := []struct {
		name           string
		workers        int
		udpWorkerRatio float64
		expectedRatio  float64
	}{
		{
			name:           "Default ratio with 100 workers",
			workers:        100,
			udpWorkerRatio: 0, // Use default (0.5)
			expectedRatio:  0.5,
		},
		{
			name:           "Custom ratio with 100 workers",
			workers:        100,
			udpWorkerRatio: 0.3,
			expectedRatio:  0.3,
		},
		{
			name:           "Low workers with default ratio",
			workers:        1,
			udpWorkerRatio: 0, // Use default (0.5)
			expectedRatio:  0.5,
		},
		{
			name:           "Zero workers with ratio",
			workers:        0,
			udpWorkerRatio: 0.5,
			expectedRatio:  0.5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &Config{
				Workers:        test.workers,
				Timeout:        200 * time.Millisecond,
				UDPWorkerRatio: test.udpWorkerRatio,
			}
			udpScanner := NewUDPScanner(cfg)

			// Test that the configuration is properly stored
			if udpScanner.config.UDPWorkerRatio != test.expectedRatio {
				t.Errorf("Expected UDPWorkerRatio to be %f, got %f", test.expectedRatio, udpScanner.config.UDPWorkerRatio)
			}
		})
	}
}
