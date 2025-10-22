package config

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Rate           int      `mapstructure:"rate" validate:"min=1,max=100000"`
	Ports          string   `mapstructure:"ports"`
	TimeoutMs      int      `mapstructure:"timeout_ms" validate:"min=1,max=10000"`
	Workers        int      `mapstructure:"workers" validate:"min=0,max=1000"` // 0 means auto-detect
	Output         string   `mapstructure:"output" validate:"omitempty,oneof=json csv prometheus table"`
	Banners        bool     `mapstructure:"banners"`
	Protocol       string   `mapstructure:"protocol" validate:"omitempty,oneof=tcp udp both"` // Scan protocol
	UDPWorkerRatio float64  `mapstructure:"udp_worker_ratio" validate:"min=-1.0,max=1.0"`     // Ratio of workers for UDP (-1=default, 0=disable, 0.1-1.0=ratio)
	UI             UIConfig `mapstructure:"ui"`
}

type UIConfig struct {
	Theme            string `mapstructure:"theme" validate:"oneof=default dracula monokai"`
	ResultBufferSize int    `mapstructure:"result_buffer_size" validate:"gte=0,lte=1000000"`
}

func Load() (*Config, error) {
	var cfg Config

	// Set defaults
	viper.SetDefault("rate", 7500)
	viper.SetDefault("ports", "1-1024,3306,6379")
	viper.SetDefault("timeout_ms", 200)
	viper.SetDefault("workers", 100)
	viper.SetDefault("output", "")
	viper.SetDefault("banners", false)
	viper.SetDefault("protocol", "tcp")
	viper.SetDefault("udp_worker_ratio", -1.0) // -1 means use default behavior (half of TCP workers)
	viper.SetDefault("ui.theme", "default")
	viper.SetDefault("ui.result_buffer_size", 10000)

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) GetTimeout() time.Duration {
	return time.Duration(c.TimeoutMs) * time.Millisecond
}
