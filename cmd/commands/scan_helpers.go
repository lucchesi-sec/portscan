package commands

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/lucchesi-sec/portscan/pkg/config"
	"github.com/lucchesi-sec/portscan/pkg/errors"
	"github.com/lucchesi-sec/portscan/pkg/parser"
	"github.com/lucchesi-sec/portscan/pkg/profiles"
	"github.com/lucchesi-sec/portscan/pkg/targets"
	"github.com/spf13/viper"
)

func collectTargetInputs(args []string) ([]string, error) {
	targets := append([]string{}, args...)

	if viper.GetBool("stdin") {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		text := strings.TrimSpace(string(data))
		if text != "" {
			targets = append(targets, strings.Fields(text)...)
		}
	}

	return targets, nil
}

func resolveTargetList(raw []string) ([]string, error) {
	return targets.Resolve(raw, targets.Options{})
}

func selectPortList(cfg *config.Config) ([]uint16, error) {
	if profile := viper.GetString("profile"); profile != "" {
		ports := profiles.GetProfile(profile)
		if ports == nil {
			return nil, fmt.Errorf("unknown profile '%s'. Available: quick, web, database, full", profile)
		}
		return ports, nil
	}

	ports, err := parser.ParsePorts(cfg.Ports)
	if err != nil {
		return nil, errors.InvalidPortError(cfg.Ports, err)
	}
	return ports, nil
}

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
	workers := cores * 50
	if workers > 200 {
		workers = 200
	}
	if workers < 10 {
		workers = 10
	}
	return workers
}

func showDryRun(targets []string, ports []uint16, cfg *config.Config) {
	fmt.Println("=== DRY RUN MODE ===")
	fmt.Printf("Targets:       %d\n", len(targets))
	if len(targets) > 0 && len(targets) <= 5 {
		fmt.Printf("Targets list: %v\n", targets)
	}
	fmt.Printf("Ports:         %d ports", len(ports))
	if len(ports) <= 10 {
		fmt.Printf(" %v", ports)
	}
	fmt.Println()
	fmt.Printf("Total sockets: %d\n", len(ports)*len(targets))
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
