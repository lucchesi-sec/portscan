package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scanCmd = &cobra.Command{
	Use:   "scan [targets...]",
	Short: "Scan ports on target host(s)",
	Long: `Perform a port scan on the specified target(s) with real-time UI visualization.

The scanner supports single hosts, CIDR notation, and multiple targets via stdin.
It provides real-time progress updates and can export results in various formats.`,
	Example: `  # Scan common ports on localhost
  portscan scan localhost

  # Scan specific ports on a target
  portscan scan 192.168.1.1 --ports 22,80,443

  # Quick scan of top 100 ports
  portscan scan example.com --profile quick

  # Scan a network range with banner grabbing
  portscan scan 10.0.0.0/24 --ports 1-1024 --banners

  # High-speed scan with custom rate
  portscan scan target.com --ports 1-65535 --rate 10000

  # Web services scan profile
  portscan scan api.example.com --profile web

  # Export results to JSON
  portscan scan 192.168.1.1 --output json > results.json

  # Scan multiple targets from file
  cat targets.txt | portscan scan --stdin

  # UDP scanning
  portscan scan 192.168.1.1 --protocol udp --profile udp-common

  # Scan gateway with both TCP and UDP
  portscan scan 192.168.1.1 --protocol both --profile gateway

  # Scan for VoIP services
  portscan scan pbx.example.com --protocol udp --profile voip`,
	Args: cobra.ArbitraryArgs,
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringP("ports", "p", "1-1024", "ports to scan (e.g., '80,443,8080' or '1-1024')")
	scanCmd.Flags().StringP("profile", "P", "", "scan profile: quick, web, database, gateway, udp-common, voip, full")
	scanCmd.Flags().StringP("protocol", "u", "tcp", "protocol to scan: tcp (default), udp, or both")
	scanCmd.Flags().IntP("rate", "r", 7500, "packets per second rate limit")
	scanCmd.Flags().IntP("timeout", "t", 200, "connection timeout in milliseconds")
	scanCmd.Flags().IntP("workers", "w", 0, "number of concurrent workers (0=auto-detect)")
	scanCmd.Flags().Float64("udp-worker-ratio", 0.5, "ratio of workers to use for UDP scanning (0.0-1.0)")
	scanCmd.Flags().BoolP("banners", "b", false, "grab service banners")

	scanCmd.Flags().StringP("output", "o", "", "output format (json, csv, prometheus, table)")
	scanCmd.Flags().BoolP("stdin", "s", false, "read targets from stdin")
	scanCmd.Flags().Bool("json", false, "output results as JSON")
	scanCmd.Flags().Bool("json-array", false, "output JSON as a single array instead of NDJSON stream")
	scanCmd.Flags().Bool("json-object", false, "output a single JSON object with scan_info and results[]")
	scanCmd.Flags().Bool("only-open", false, "show only open ports in UI/table outputs")

	scanCmd.Flags().String("ui.theme", "default", "UI theme (default, dracula, monokai)")

	scanCmd.Flags().Bool("dry-run", false, "validate parameters without scanning")
	scanCmd.Flags().Bool("examples", false, "show extended examples and exit")
	scanCmd.Flags().Bool("verbose", false, "enable verbose output for debugging")

	_ = viper.BindPFlag("ports", scanCmd.Flags().Lookup("ports"))
	_ = viper.BindPFlag("profile", scanCmd.Flags().Lookup("profile"))
	_ = viper.BindPFlag("protocol", scanCmd.Flags().Lookup("protocol"))
	_ = viper.BindPFlag("rate", scanCmd.Flags().Lookup("rate"))
	_ = viper.BindPFlag("timeout_ms", scanCmd.Flags().Lookup("timeout"))
	_ = viper.BindPFlag("workers", scanCmd.Flags().Lookup("workers"))
	_ = viper.BindPFlag("udp_worker_ratio", scanCmd.Flags().Lookup("udp-worker-ratio"))
	_ = viper.BindPFlag("banners", scanCmd.Flags().Lookup("banners"))
	_ = viper.BindPFlag("output", scanCmd.Flags().Lookup("output"))
	_ = viper.BindPFlag("stdin", scanCmd.Flags().Lookup("stdin"))
	_ = viper.BindPFlag("json", scanCmd.Flags().Lookup("json"))
	_ = viper.BindPFlag("json_array", scanCmd.Flags().Lookup("json-array"))
	_ = viper.BindPFlag("json_object", scanCmd.Flags().Lookup("json-object"))
	_ = viper.BindPFlag("ui.theme", scanCmd.Flags().Lookup("ui.theme"))
	_ = viper.BindPFlag("dry_run", scanCmd.Flags().Lookup("dry-run"))
	_ = viper.BindPFlag("verbose", scanCmd.Flags().Lookup("verbose"))
	_ = viper.BindPFlag("only_open", scanCmd.Flags().Lookup("only-open"))
}
