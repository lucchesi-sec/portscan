package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage scanner configuration",
	Long:  `Initialize, show, or validate the scanner configuration.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default configuration file",
	Long: `Create a default configuration file with documented settings.
The config file will be created at ~/.portscan.yaml`,
	RunE: runConfigInit,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long:  `Display the current configuration settings and their sources.`,
	RunE:  runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".portscan.yaml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Configuration file already exists at: %s\n", configPath)
		fmt.Println("To overwrite, please delete the existing file first.")
		return nil
	}

	// Default configuration with comments
	defaultConfig := `# Port Scanner Configuration
# This file configures default settings for the port scanner

# Performance settings
rate: 7500              # Packets per second (max safe: 15000)
workers: 0              # Concurrent workers (0 = auto-detect based on CPU)
timeout_ms: 200         # Connection timeout in milliseconds

# Default scan settings
ports: "1-1024"         # Default ports to scan
banners: false          # Grab service banners by default
output: ""              # Output format: json, csv, table, or empty for TUI

# UI preferences
ui:
  theme: default        # Options: default, dracula, monokai

# DNS settings
dns:
  timeout_ms: 1000      # DNS resolution timeout
  servers:              # Custom DNS servers (optional)
    - "1.1.1.1"
    - "8.8.8.8"

# Logging
quiet: false            # Suppress non-essential output
no_color: false         # Disable colored output
log_json: false         # Output logs in JSON format
verbose: false          # Enable verbose debug output

# Common port profiles (for reference)
# quick:    Top 100 most common ports
# web:      HTTP/HTTPS and web app ports
# database: Database service ports
# full:     All ports (1-65535)
`

	// Write config file
	err = os.WriteFile(configPath, []byte(defaultConfig), 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✓ Configuration file created at: %s\n", configPath)
	fmt.Println("\nYou can now edit this file to customize default settings.")
	fmt.Println("Use 'portscan config show' to view current configuration.")

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	fmt.Println("=== Current Configuration ===\n")

	// Show config file location
	configFile := viper.ConfigFileUsed()
	if configFile != "" {
		fmt.Printf("Config File: %s\n", configFile)
	} else {
		fmt.Println("Config File: (none - using defaults)")
	}

	fmt.Println("\n--- Settings ---")

	// Performance settings
	fmt.Println("\nPerformance:")
	fmt.Printf("  Rate:       %d pps\n", viper.GetInt("rate"))
	fmt.Printf("  Workers:    %d", viper.GetInt("workers"))
	if viper.GetInt("workers") == 0 {
		fmt.Print(" (auto-detect)")
	}
	fmt.Println()
	fmt.Printf("  Timeout:    %d ms\n", viper.GetInt("timeout_ms"))

	// Scan settings
	fmt.Println("\nScan Defaults:")
	fmt.Printf("  Ports:      %s\n", viper.GetString("ports"))
	fmt.Printf("  Banners:    %v\n", viper.GetBool("banners"))
	fmt.Printf("  Output:     %s", viper.GetString("output"))
	if viper.GetString("output") == "" {
		fmt.Print(" (TUI)")
	}
	fmt.Println()

	// UI settings
	fmt.Println("\nUI:")
	fmt.Printf("  Theme:      %s\n", viper.GetString("ui.theme"))

	// Output settings
	fmt.Println("\nOutput:")
	fmt.Printf("  Quiet:      %v\n", viper.GetBool("quiet"))
	fmt.Printf("  No Color:   %v\n", viper.GetBool("no_color"))
	fmt.Printf("  JSON Logs:  %v\n", viper.GetBool("log_json"))
	fmt.Printf("  Verbose:    %v\n", viper.GetBool("verbose"))

	// Environment variables
	fmt.Println("\n--- Environment Variables ---")
	fmt.Println("You can override any setting with PORTSCAN_ prefix:")
	fmt.Println("  PORTSCAN_RATE=10000")
	fmt.Println("  PORTSCAN_WORKERS=200")
	fmt.Println("  PORTSCAN_OUTPUT=json")

	return nil
}