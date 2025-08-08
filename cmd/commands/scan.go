package commands

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/internal/ui"
	"github.com/lucchesi-sec/portscan/pkg/config"
	"github.com/lucchesi-sec/portscan/pkg/errors"
	"github.com/lucchesi-sec/portscan/pkg/exporter"
	"github.com/lucchesi-sec/portscan/pkg/parser"
	"github.com/lucchesi-sec/portscan/pkg/profiles"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scanCmd = &cobra.Command{
	Use:   "scan [target]",
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
  cat targets.txt | portscan scan --stdin`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Core scanning flags
	scanCmd.Flags().StringP("ports", "p", "1-1024", "ports to scan (e.g., '80,443,8080' or '1-1024')")
	scanCmd.Flags().StringP("profile", "P", "", "scan profile: quick (top 100), web (HTTP/HTTPS), database (DB ports), full (1-65535)")
	scanCmd.Flags().IntP("rate", "r", 7500, "packets per second rate limit")
	scanCmd.Flags().IntP("timeout", "t", 200, "connection timeout in milliseconds")
	scanCmd.Flags().IntP("workers", "w", 0, "number of concurrent workers (0=auto-detect)")
	scanCmd.Flags().BoolP("banners", "b", false, "grab service banners")

	// Output flags
	scanCmd.Flags().StringP("output", "o", "", "output format (json, csv, prometheus, table)")
	scanCmd.Flags().BoolP("stdin", "s", false, "read targets from stdin")
	scanCmd.Flags().Bool("json", false, "output results as JSON")
	scanCmd.Flags().Bool("json-array", false, "output JSON as a single array instead of NDJSON stream")
	scanCmd.Flags().Bool("json-object", false, "output a single JSON object with scan_info and results[]")
	scanCmd.Flags().Bool("only-open", false, "show only open ports in UI/table outputs")

	// UI flags
	scanCmd.Flags().String("ui.theme", "default", "UI theme (default, dracula, monokai)")

	// Validation flags
	scanCmd.Flags().Bool("dry-run", false, "validate parameters without scanning")
	scanCmd.Flags().Bool("examples", false, "show extended examples and exit")
	scanCmd.Flags().Bool("verbose", false, "enable verbose output for debugging")

	_ = viper.BindPFlag("ports", scanCmd.Flags().Lookup("ports"))
	_ = viper.BindPFlag("profile", scanCmd.Flags().Lookup("profile"))
	_ = viper.BindPFlag("rate", scanCmd.Flags().Lookup("rate"))
	_ = viper.BindPFlag("timeout_ms", scanCmd.Flags().Lookup("timeout"))
	_ = viper.BindPFlag("workers", scanCmd.Flags().Lookup("workers"))
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

func runScan(cmd *cobra.Command, args []string) error {
	// Handle --examples flag first (before checking for target)
	examples, _ := cmd.Flags().GetBool("examples")
	if examples {
		showExtendedExamples()
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return errors.ConfigLoadError(viper.ConfigFileUsed(), err)
	}

	// Auto-detect optimal worker count if not specified
	if cfg.Workers == 0 {
		cfg.Workers = getOptimalWorkerCount()
		if viper.GetBool("verbose") {
			fmt.Printf("Auto-detected optimal workers: %d (based on %d CPU cores)\n", cfg.Workers, runtime.NumCPU())
		}
	}

	// Check rate limit safety
	const maxSafeRate = 15000
	if cfg.Rate > maxSafeRate {
		return errors.RateLimitError(cfg.Rate, maxSafeRate)
	}

	var target string
	if len(args) > 0 {
		target = args[0]
	} else if viper.GetBool("stdin") {
		// Read from stdin and select the first non-empty token (line or whitespace separated)
		data, _ := io.ReadAll(os.Stdin)
		text := strings.TrimSpace(string(data))
		if text == "" {
			return errors.NoTargetError()
		}
		// Split by any whitespace/newlines
		fields := strings.Fields(text)
		if len(fields) == 0 {
			return errors.NoTargetError()
		}
		target = fields[0]
		// TODO: Support multi-target stdin and CIDR expansion with a global scanner
	} else {
		return errors.NoTargetError()
	}

	// Apply profile if specified
	profile := viper.GetString("profile")
	var ports []uint16
	if profile != "" {
		ports = profiles.GetProfile(profile)
		if ports == nil {
			return fmt.Errorf("unknown profile '%s'. Available: quick, web, database, full", profile)
		}
	} else {
		ports, err = parser.ParsePorts(cfg.Ports)
		if err != nil {
			return errors.InvalidPortError(cfg.Ports, err)
		}
	}

	// Pre-validate target
	if err := validateTarget(target); err != nil {
		return errors.InvalidTargetError(target, err)
	}

	// Handle --dry-run flag
	if viper.GetBool("dry_run") {
		showDryRun(target, ports, cfg)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	scannerCfg := &core.Config{
		Workers:    cfg.Workers,
		Timeout:    cfg.GetTimeout(),
		RateLimit:  cfg.Rate,
		BannerGrab: cfg.Banners,
		MaxRetries: 2,
	}

	scanner := core.NewScanner(scannerCfg)

	if viper.GetBool("json") || cfg.Output == "json" {
		var exp *exporter.JSONExporter
		switch {
		case viper.GetBool("json_object"):
			exp = exporter.NewJSONExporterObject(os.Stdout, target, len(ports), cfg.Rate)
		case viper.GetBool("json_array"):
			exp = exporter.NewJSONExporterArray(os.Stdout)
		default:
			exp = exporter.NewJSONExporter(os.Stdout)
		}
		go exp.Export(scanner.Results())
		scanner.ScanRange(ctx, target, ports)
		_ = exp.Close()
	} else if cfg.Output == "csv" {
		exp := exporter.NewCSVExporter(os.Stdout)
		go exp.Export(scanner.Results())
		scanner.ScanRange(ctx, target, ports)
		_ = exp.Close()
	} else {
		// Use Enhanced TUI
		onlyOpen := viper.GetBool("only_open")
		tui := ui.NewEnhancedUI(cfg, len(ports), scanner.Results(), onlyOpen)
		go scanner.ScanRange(ctx, target, ports)
		return tui.Run()
	}

	return nil
}

// Helper functions

func showExtendedExamples() {
	examples := `
EXTENDED EXAMPLES:

Network Discovery:
  # Find all hosts with web servers in a subnet
  portscan scan 192.168.1.0/24 --ports 80,443,8080,8443 --banners

  # Quick discovery of common services
  portscan scan 10.0.0.0/16 --profile quick --rate 5000

Security Assessment:
  # Comprehensive scan with service detection
  portscan scan target.com --ports 1-65535 --banners --output json

  # Check for database exposure
  portscan scan dmz.company.com --profile database --banners

DevOps Validation:
  # Verify deployed services
  echo "api.prod.com db.prod.com cache.prod.com" | tr ' ' '\n' | portscan scan --stdin --ports 443,5432,6379

  # Container network validation
  portscan scan 172.17.0.0/24 --ports 8080,9090 --timeout 100

Performance Tuning:
  # High-speed scan with custom workers
  portscan scan 10.0.0.0/8 --ports 22,80,443 --workers 200 --rate 10000

  # Conservative scan to avoid detection
  portscan scan sensitive.target --ports 1-1024 --rate 100 --timeout 500

Output Formats:
  # JSON for automation
  portscan scan target.com --output json | jq '.results[] | select(.state=="open")'

  # CSV for reporting
  portscan scan 192.168.1.0/24 --output csv > scan_results.csv

  # Table format without TUI
  portscan scan localhost --output table --quiet

Profiles:
  quick    - Top 100 most common ports
  web      - HTTP/HTTPS and common web app ports (80,443,8080,8443,3000,5000)
  database - Database ports (3306,5432,27017,6379,9042,9200)
  full     - All ports (1-65535)
`
	fmt.Println(examples)
}

func getOptimalWorkerCount() int {
	cores := runtime.NumCPU()
	// Use 2x CPU cores for I/O bound operations, max 200
	workers := cores * 50
	if workers > 200 {
		workers = 200
	}
	if workers < 10 {
		workers = 10
	}
	return workers
}

func validateTarget(target string) error {
	// Check for empty string
	if target == "" {
		return fmt.Errorf("empty target")
	}

	// Try to parse as IP address
	if net.ParseIP(target) != nil {
		return nil // Valid IP
	}

	// Try to parse as CIDR notation
	if _, _, err := net.ParseCIDR(target); err == nil {
		return nil // Valid CIDR
	}

	// Validate as hostname/domain
	// Must contain only valid chars: a-z, A-Z, 0-9, -, .
	// Cannot start/end with - or .
	// Cannot have consecutive dots
	if isValidHostname(target) {
		return nil
	}

	return fmt.Errorf("invalid target: must be IP, CIDR, or valid hostname")
}

func isValidHostname(hostname string) bool {
	// Basic hostname validation
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	// Cannot start or end with dot or hyphen
	if hostname[0] == '.' || hostname[0] == '-' ||
		hostname[len(hostname)-1] == '.' || hostname[len(hostname)-1] == '-' {
		return false
	}

	// Check for consecutive dots
	if strings.Contains(hostname, "..") {
		return false
	}

	// Split into labels and validate each
	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}

		// Label cannot start or end with hyphen
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}

		// Check all characters are valid (alphanumeric or hyphen)
		for _, ch := range label {
			if (ch < 'a' || ch > 'z') &&
				(ch < 'A' || ch > 'Z') &&
				(ch < '0' || ch > '9') &&
				ch != '-' {
				return false
			}
		}
	}

	return true
}

func showDryRun(target string, ports []uint16, cfg *config.Config) {
	fmt.Println("=== DRY RUN MODE ===")
	fmt.Printf("Target:        %s\n", target)
	fmt.Printf("Ports:         %d ports", len(ports))
	if len(ports) <= 10 {
		fmt.Printf(" %v", ports)
	}
	fmt.Println()
	fmt.Printf("Workers:       %d\n", cfg.Workers)
	fmt.Printf("Rate Limit:    %d pps\n", cfg.Rate)
	fmt.Printf("Timeout:       %dms\n", cfg.TimeoutMs)
	fmt.Printf("Banner Grab:   %v\n", cfg.Banners)
	fmt.Printf("Output Format: %s\n", cfg.Output)
	if cfg.Output == "" {
		fmt.Print(" (TUI)")
	}
	fmt.Println("\n\nScan would proceed with these parameters. Remove --dry-run to execute.")
}
