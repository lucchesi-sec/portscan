// Package config provides configuration management for the port scanner.
//
// This package implements hierarchical configuration loading using Viper,
// supporting multiple configuration sources with the following precedence
// (highest to lowest):
//
//   1. Command-line flags (highest priority)
//   2. Environment variables (PORTSCAN_*)
//   3. Configuration file (~/.portscan.yaml)
//   4. Default values (lowest priority)
//
// Example configuration file (~/.portscan.yaml):
//
//	rate: 7500
//	workers: 100
//	timeout_ms: 200
//	banners: true
//	ports: "1-1024,3306,5432,6379,8080,8443"
//	protocol: tcp
//	ui:
//	  theme: dracula
//	  result_buffer_size: 10000
//
// Usage:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Rate limit: %d pps\n", cfg.Rate)
//	timeout := cfg.GetTimeout() // Converts milliseconds to time.Duration
//
// Validation:
//
// All configuration values are validated using struct tags with
// go-playground/validator. Invalid values return descriptive errors:
//
//   - rate: 1-100,000 packets per second
//   - timeout_ms: 1-10,000 milliseconds
//   - workers: 0-1,000 (0 means auto-detect)
//   - output: json, csv, prometheus, table
//   - protocol: tcp, udp, both
//
// Environment Variables:
//
// All configuration keys can be set via environment variables:
//
//	PORTSCAN_RATE=10000
//	PORTSCAN_WORKERS=200
//	PORTSCAN_TIMEOUT_MS=500
//	PORTSCAN_UI_THEME=monokai
package config
