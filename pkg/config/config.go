package config

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Rate      int      `mapstructure:"rate" validate:"min=1,max=100000"`
	Ports     string   `mapstructure:"ports"`
	TimeoutMs int      `mapstructure:"timeout_ms" validate:"min=1,max=10000"`
	Workers   int      `mapstructure:"workers" validate:"min=1,max=1000"`
	Output    string   `mapstructure:"output" validate:"omitempty,oneof=json csv prometheus"`
	Banners   bool     `mapstructure:"banners"`
	UI        UIConfig `mapstructure:"ui"`
}

type UIConfig struct {
	Theme string `mapstructure:"theme" validate:"oneof=default dracula monokai"`
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
	viper.SetDefault("ui.theme", "default")

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
