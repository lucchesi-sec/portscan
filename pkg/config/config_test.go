package config

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
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

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Rate:      7500,
				TimeoutMs: 200,
				Workers:   100,
				Protocol:  "tcp",
				UI: UIConfig{
					Theme:            "default",
					ResultBufferSize: 10000,
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with udp",
			config: Config{
				Rate:           7500,
				TimeoutMs:      200,
				Workers:        100,
				Protocol:       "udp",
				UDPWorkerRatio: 0.5,
				UI: UIConfig{
					Theme:            "dracula",
					ResultBufferSize: 5000,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid rate too high",
			config: Config{
				Rate:      200000,
				TimeoutMs: 200,
				Workers:   100,
				Protocol:  "tcp",
			},
			wantErr: true,
		},
		{
			name: "invalid rate too low",
			config: Config{
				Rate:      0,
				TimeoutMs: 200,
				Workers:   100,
				Protocol:  "tcp",
			},
			wantErr: true,
		},
		{
			name: "invalid timeout zero",
			config: Config{
				Rate:      7500,
				TimeoutMs: 0,
				Workers:   100,
				Protocol:  "tcp",
			},
			wantErr: true,
		},
		{
			name: "invalid timeout too high",
			config: Config{
				Rate:      7500,
				TimeoutMs: 70000,
				Workers:   100,
				Protocol:  "tcp",
			},
			wantErr: true,
		},
		{
			name: "invalid workers too many",
			config: Config{
				Rate:      7500,
				TimeoutMs: 200,
				Workers:   2000,
				Protocol:  "tcp",
			},
			wantErr: true,
		},
		{
			name: "invalid protocol",
			config: Config{
				Rate:      7500,
				TimeoutMs: 200,
				Workers:   100,
				Protocol:  "icmp",
			},
			wantErr: true,
		},
		{
			name: "invalid output format",
			config: Config{
				Rate:      7500,
				TimeoutMs: 200,
				Workers:   100,
				Output:    "xml",
				Protocol:  "tcp",
			},
			wantErr: true,
		},
		{
			name: "invalid theme",
			config: Config{
				Rate:      7500,
				TimeoutMs: 200,
				Workers:   100,
				Protocol:  "tcp",
				UI: UIConfig{
					Theme:            "cyberpunk",
					ResultBufferSize: 10000,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid buffer size negative",
			config: Config{
				Rate:      7500,
				TimeoutMs: 200,
				Workers:   100,
				Protocol:  "tcp",
				UI: UIConfig{
					Theme:            "default",
					ResultBufferSize: -1,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid buffer size too large",
			config: Config{
				Rate:      7500,
				TimeoutMs: 200,
				Workers:   100,
				Protocol:  "tcp",
				UI: UIConfig{
					Theme:            "default",
					ResultBufferSize: 2000000,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid udp worker ratio too high",
			config: Config{
				Rate:           7500,
				TimeoutMs:      200,
				Workers:        100,
				Protocol:       "tcp",
				UDPWorkerRatio: 1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid udp worker ratio too low",
			config: Config{
				Rate:           7500,
				TimeoutMs:      200,
				Workers:        100,
				Protocol:       "tcp",
				UDPWorkerRatio: -2.0,
			},
			wantErr: true,
		},
		{
			name: "valid config with both protocol",
			config: Config{
				Rate:           7500,
				TimeoutMs:      200,
				Workers:        100,
				Protocol:       "both",
				UDPWorkerRatio: 0.3,
				UI: UIConfig{
					Theme:            "monokai",
					ResultBufferSize: 15000,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := validator.New()
			err := validate.Struct(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Reset viper before each test
	viper.Reset()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify defaults are set correctly
	if cfg.Rate != 7500 {
		t.Errorf("Rate = %d; want 7500", cfg.Rate)
	}

	if cfg.Ports != "1-1024,3306,6379" {
		t.Errorf("Ports = %s; want 1-1024,3306,6379", cfg.Ports)
	}

	if cfg.TimeoutMs != 200 {
		t.Errorf("TimeoutMs = %d; want 200", cfg.TimeoutMs)
	}

	if cfg.Workers != 100 {
		t.Errorf("Workers = %d; want 100", cfg.Workers)
	}

	if cfg.Output != "" {
		t.Errorf("Output = %s; want empty string", cfg.Output)
	}

	if cfg.Banners != false {
		t.Errorf("Banners = %t; want false", cfg.Banners)
	}

	if cfg.Protocol != "tcp" {
		t.Errorf("Protocol = %s; want tcp", cfg.Protocol)
	}

	if cfg.UDPWorkerRatio != -1.0 {
		t.Errorf("UDPWorkerRatio = %f; want -1.0", cfg.UDPWorkerRatio)
	}

	if cfg.UI.Theme != "default" {
		t.Errorf("UI.Theme = %s; want default", cfg.UI.Theme)
	}

	if cfg.UI.ResultBufferSize != 10000 {
		t.Errorf("UI.ResultBufferSize = %d; want 10000", cfg.UI.ResultBufferSize)
	}
}

func TestLoadWithViperOverrides(t *testing.T) {
	// Reset viper
	viper.Reset()

	// Set custom values via Viper
	viper.Set("rate", 5000)
	viper.Set("timeout_ms", 300)
	viper.Set("workers", 50)
	viper.Set("protocol", "udp")
	viper.Set("ui.theme", "dracula")
	viper.Set("ui.result_buffer_size", 5000)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify overridden values
	if cfg.Rate != 5000 {
		t.Errorf("Rate = %d; want 5000", cfg.Rate)
	}

	if cfg.TimeoutMs != 300 {
		t.Errorf("TimeoutMs = %d; want 300", cfg.TimeoutMs)
	}

	if cfg.Workers != 50 {
		t.Errorf("Workers = %d; want 50", cfg.Workers)
	}

	if cfg.Protocol != "udp" {
		t.Errorf("Protocol = %s; want udp", cfg.Protocol)
	}

	if cfg.UI.Theme != "dracula" {
		t.Errorf("UI.Theme = %s; want dracula", cfg.UI.Theme)
	}

	if cfg.UI.ResultBufferSize != 5000 {
		t.Errorf("UI.ResultBufferSize = %d; want 5000", cfg.UI.ResultBufferSize)
	}
}

func TestLoadWithInvalidConfig(t *testing.T) {
	// Reset viper
	viper.Reset()

	// Set invalid values
	viper.Set("rate", 200000) // Too high
	viper.Set("timeout_ms", 200)
	viper.Set("workers", 100)

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid config")
	}
}

func TestGetTimeoutEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		timeoutMs int
		want      time.Duration
	}{
		{
			name:      "maximum valid timeout",
			timeoutMs: 10000,
			want:      10 * time.Second,
		},
		{
			name:      "minimum valid timeout",
			timeoutMs: 1,
			want:      time.Millisecond,
		},
		{
			name:      "200ms default",
			timeoutMs: 200,
			want:      200 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{TimeoutMs: tt.timeoutMs}
			got := cfg.GetTimeout()
			if got != tt.want {
				t.Errorf("GetTimeout() = %v; want %v", got, tt.want)
			}
		})
	}
}

func TestUIConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		uiConfig UIConfig
		wantErr  bool
	}{
		{
			name: "valid default theme",
			uiConfig: UIConfig{
				Theme:            "default",
				ResultBufferSize: 10000,
			},
			wantErr: false,
		},
		{
			name: "valid dracula theme",
			uiConfig: UIConfig{
				Theme:            "dracula",
				ResultBufferSize: 5000,
			},
			wantErr: false,
		},
		{
			name: "valid monokai theme",
			uiConfig: UIConfig{
				Theme:            "monokai",
				ResultBufferSize: 15000,
			},
			wantErr: false,
		},
		{
			name: "invalid theme",
			uiConfig: UIConfig{
				Theme:            "solarized",
				ResultBufferSize: 10000,
			},
			wantErr: true,
		},
		{
			name: "zero buffer size",
			uiConfig: UIConfig{
				Theme:            "default",
				ResultBufferSize: 0,
			},
			wantErr: false, // 0 is valid (gte=0)
		},
		{
			name: "max buffer size",
			uiConfig: UIConfig{
				Theme:            "default",
				ResultBufferSize: 1000000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := validator.New()
			err := validate.Struct(&tt.uiConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("UIConfig validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
