package config

import (
	"testing"
	"time"
)

func TestGetTimeout(t *testing.T) {
	tests := []struct {
		name      string
		timeoutMs int
		want      time.Duration
	}{
		{
			name:      "1 second timeout",
			timeoutMs: 1000,
			want:      time.Second,
		},
		{
			name:      "500ms timeout",
			timeoutMs: 500,
			want:      500 * time.Millisecond,
		},
		{
			name:      "5 second timeout",
			timeoutMs: 5000,
			want:      5 * time.Second,
		},
		{
			name:      "100ms timeout",
			timeoutMs: 100,
			want:      100 * time.Millisecond,
		},
		{
			name:      "zero timeout",
			timeoutMs: 0,
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{TimeoutMs: tt.timeoutMs}
			got := c.GetTimeout()
			if got != tt.want {
				t.Errorf("Config.GetTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := &Config{}

	// Test default values make sense
	if cfg.TimeoutMs < 0 {
		t.Error("Default timeout should not be negative")
	}

	timeout := cfg.GetTimeout()
	if timeout < 0 {
		t.Error("GetTimeout() should not return negative duration")
	}
}
