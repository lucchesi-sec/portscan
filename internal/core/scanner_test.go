package core

import (
	"context"
	"testing"
	"time"
)

func TestNewScanner(t *testing.T) {
	cfg := &Config{
		Workers:    10,
		Timeout:    time.Second,
		RateLimit:  1000,
		BannerGrab: true,
		MaxRetries: 3,
	}

	scanner := NewScanner(cfg)

	if scanner == nil {
		t.Fatal("NewScanner() returned nil")
	}

	if scanner.config.Workers != cfg.Workers {
		t.Errorf("Expected %d workers, got %d", cfg.Workers, scanner.config.Workers)
	}

	if scanner.config.Timeout != cfg.Timeout {
		t.Errorf("Expected %v timeout, got %v", cfg.Timeout, scanner.config.Timeout)
	}
}

func TestScannerChannels(t *testing.T) {
	cfg := &Config{
		Workers:   1,
		Timeout:   100 * time.Millisecond,
		RateLimit: 100,
	}

	scanner := NewScanner(cfg)

	// Test that channels are properly initialized
	results := scanner.Results()
	if results == nil {
		t.Error("Results() returned nil channel")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		valid  bool
	}{
		{
			name: "valid config",
			config: &Config{
				Workers:   10,
				Timeout:   time.Second,
				RateLimit: 1000,
			},
			valid: true,
		},
		{
			name: "zero workers",
			config: &Config{
				Workers:   0,
				Timeout:   time.Second,
				RateLimit: 1000,
			},
			valid: false,
		},
		{
			name: "zero timeout",
			config: &Config{
				Workers:   10,
				Timeout:   0,
				RateLimit: 1000,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner(tt.config)
			if tt.valid && scanner == nil {
				t.Error("Expected valid scanner, got nil")
			}
			if !tt.valid && tt.config.Workers == 0 {
				// Scanner should handle zero workers gracefully
				if scanner != nil && scanner.config.Workers <= 0 {
					t.Error("Scanner should not allow zero or negative workers")
				}
			}
		})
	}
}

func TestSimplifiedWorkerPool(t *testing.T) {
	cfg := &Config{
		Workers:   10,
		Timeout:   100 * time.Millisecond,
		RateLimit: 0,
	}
	scanner := NewScanner(cfg)

	ctx := context.Background()
	ports := []uint16{80, 443, 8080}

	go scanner.ScanRange(ctx, "localhost", ports)

	results := 0
	progressEvents := 0
	for event := range scanner.Results() {
		switch event.Kind {
		case EventKindResult:
			results++
		case EventKindProgress:
			progressEvents++
		}
	}

	if results != len(ports) {
		t.Errorf("got %d results; want %d", results, len(ports))
	}

	if progressEvents == 0 {
		t.Error("expected at least one progress event")
	}
}

// BenchmarkWorkerPool benchmarks worker pool performance
func BenchmarkWorkerPool(b *testing.B) {
	cfg := &Config{
		Workers:   100,
		Timeout:   50 * time.Millisecond,
		RateLimit: 0,
	}

	ports := []uint16{80, 443, 8080, 22, 3306}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner := NewScanner(cfg)
		ctx := context.Background()

		go scanner.ScanRange(ctx, "localhost", ports)

		// Drain results
		for range scanner.Results() {
		}
	}
}

// BenchmarkWorkerPoolScaling benchmarks worker pool scaling
func BenchmarkWorkerPoolScaling(b *testing.B) {
	tests := []struct {
		name    string
		workers int
	}{
		{"10 workers", 10},
		{"50 workers", 50},
		{"100 workers", 100},
		{"200 workers", 200},
	}

	ports := []uint16{80, 443, 8080}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			cfg := &Config{
				Workers:   tt.workers,
				Timeout:   50 * time.Millisecond,
				RateLimit: 0,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				scanner := NewScanner(cfg)
				ctx := context.Background()

				go scanner.ScanRange(ctx, "localhost", ports)
				for range scanner.Results() {
				}
			}
		})
	}
}

// BenchmarkScannerCreation benchmarks scanner creation
func BenchmarkScannerCreation(b *testing.B) {
	cfg := &Config{
		Workers:   10,
		Timeout:   time.Second,
		RateLimit: 1000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner := NewScanner(cfg)
		_ = scanner
	}
}
