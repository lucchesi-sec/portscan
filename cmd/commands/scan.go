package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/internal/ui"
	"github.com/lucchesi-sec/portscan/pkg/config"
	"github.com/lucchesi-sec/portscan/pkg/exporter"
	"github.com/lucchesi-sec/portscan/pkg/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scanCmd = &cobra.Command{
	Use:   "scan [target]",
	Short: "Scan ports on target host(s)",
	Long:  `Perform a port scan on the specified target(s) with real-time UI visualization.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringP("ports", "p", "1-1024", "ports to scan (e.g., '80,443,8080' or '1-1024')")
	scanCmd.Flags().IntP("rate", "r", 7500, "packets per second rate limit")
	scanCmd.Flags().IntP("timeout", "t", 200, "connection timeout in milliseconds")
	scanCmd.Flags().IntP("workers", "w", 100, "number of concurrent workers")
	scanCmd.Flags().BoolP("banners", "b", false, "grab service banners")
	scanCmd.Flags().StringP("output", "o", "", "output format (json, csv, prometheus)")
	scanCmd.Flags().BoolP("stdin", "s", false, "read targets from stdin")
	scanCmd.Flags().Bool("json", false, "output results as JSON")
	scanCmd.Flags().String("ui.theme", "default", "UI theme (default, dracula, monokai)")

	_ = viper.BindPFlag("ports", scanCmd.Flags().Lookup("ports"))
	_ = viper.BindPFlag("rate", scanCmd.Flags().Lookup("rate"))
	_ = viper.BindPFlag("timeout_ms", scanCmd.Flags().Lookup("timeout"))
	_ = viper.BindPFlag("workers", scanCmd.Flags().Lookup("workers"))
	_ = viper.BindPFlag("banners", scanCmd.Flags().Lookup("banners"))
	_ = viper.BindPFlag("output", scanCmd.Flags().Lookup("output"))
	_ = viper.BindPFlag("ui.theme", scanCmd.Flags().Lookup("ui.theme"))
}

func runScan(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var target string
	if len(args) > 0 {
		target = args[0]
	} else if viper.GetBool("stdin") {
		// Read from stdin
		var builder strings.Builder
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if n > 0 {
				builder.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		target = strings.TrimSpace(builder.String())
	} else {
		return fmt.Errorf("no target specified")
	}

	ports, err := parser.ParsePorts(cfg.Ports)
	if err != nil {
		return fmt.Errorf("invalid port specification: %w", err)
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
		exp := exporter.NewJSONExporter(os.Stdout)
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
		tui := ui.NewEnhanced(cfg, scanner.Results())
		go scanner.ScanRange(ctx, target, ports)
		return tui.Run()
	}

	return nil
}
