package commands

import (
	"context"
	stdErrors "errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/internal/ui"
	"github.com/lucchesi-sec/portscan/pkg/config"
	"github.com/lucchesi-sec/portscan/pkg/errors"
	"github.com/lucchesi-sec/portscan/pkg/exporter"
	"github.com/lucchesi-sec/portscan/pkg/targets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runScan(cmd *cobra.Command, args []string) error {
	if examples, _ := cmd.Flags().GetBool("examples"); examples {
		showExtendedExamples()
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return errors.ConfigLoadError(viper.ConfigFileUsed(), err)
	}

	// Validate all user inputs before processing
	if err := validateInputs(cfg); err != nil {
		return err
	}

	ensureWorkersConfigured(cfg)

	if err := enforceRateSafety(cfg.Rate); err != nil {
		return err
	}

	rawTargets, err := collectTargetInputs(args)
	if err != nil {
		return err
	}
	if len(rawTargets) == 0 {
		return errors.NoTargetError()
	}

	// Validate each raw target before resolution
	if err := validateRawTargets(rawTargets); err != nil {
		return err
	}

	resolvedTargets, err := resolveTargetList(rawTargets)
	if err != nil {
		return errors.InvalidTargetListError(err)
	}

	ports, err := selectPortList(cfg)
	if err != nil {
		return err
	}

	if viper.GetBool("dry_run") {
		showDryRun(resolvedTargets, ports, cfg)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cleanupInterrupts := monitorInterrupts(cancel)
	defer cleanupInterrupts()

	protocol := normalizeProtocol(cfg.Protocol)
	return executeScan(ctx, protocol, resolvedTargets, ports, cfg)
}

func runProtocolScan(ctx context.Context, scanner core.PortScanner, hosts []string, ports []uint16, cfg *config.Config, _ string) error {
	if len(hosts) == 0 {
		return errors.NoTargetError()
	}

	scanTargets := buildScanTargets(hosts, ports)
	events := scanner.Results()
	go scanner.ScanTargets(ctx, scanTargets)

	totalPorts := len(ports) * len(hosts)
	metadata := exporter.ScanMetadata{Targets: hosts, TotalPorts: totalPorts, Rate: cfg.Rate}

	return handleScanOutput(ctx, cfg, events, totalPorts, metadata)
}

func selectJSONExporter(meta exporter.ScanMetadata) *exporter.JSONExporter {
	switch {
	case viper.GetBool("json_object"):
		return exporter.NewJSONExporterObjectWithMetadata(os.Stdout, meta)
	case viper.GetBool("json_array"):
		return exporter.NewJSONExporterArray(os.Stdout)
	default:
		return exporter.NewJSONExporter(os.Stdout)
	}
}

func streamEvents(ctx context.Context, events <-chan core.Event, export func(<-chan core.Event), closeFn func() error) error {
	done := make(chan error, 1)
	go func() {
		export(events)
		done <- closeFn()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		err := <-done
		if err != nil {
			return err
		}
		if stdErrors.Is(ctx.Err(), context.Canceled) {
			return nil
		}
		return ctx.Err()
	}
}

func ensureWorkersConfigured(cfg *config.Config) {
	if cfg.Workers != 0 {
		return
	}
	cfg.Workers = getOptimalWorkerCount()
	if viper.GetBool("verbose") {
		fmt.Printf("Auto-detected optimal workers: %d (based on %d CPU cores)\n", cfg.Workers, runtime.NumCPU())
	}
}

func enforceRateSafety(rate int) error {
	if rate > core.MaxSafeRateLimit {
		return errors.RateLimitError(rate, core.MaxSafeRateLimit)
	}
	return nil
}

func monitorInterrupts(cancel context.CancelFunc) func() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	stop := make(chan struct{})

	go func() {
		for {
			select {
			case <-stop:
				return
			case <-sigChan:
				cancel()
			}
		}
	}()

	return func() {
		signal.Stop(sigChan)
		close(stop)
	}
}

func buildScannerConfig(cfg *config.Config) *core.Config {
	return &core.Config{
		Workers:        cfg.Workers,
		Timeout:        cfg.GetTimeout(),
		RateLimit:      cfg.Rate,
		BannerGrab:     cfg.Banners,
		MaxRetries:     2,
		UDPWorkerRatio: cfg.UDPWorkerRatio,
	}
}

// normalizeProtocol ensures the protocol string is valid and defaults to "tcp".
func normalizeProtocol(protocol string) string {
	if protocol == "" {
		return "tcp"
	}
	return protocol
}

// buildScanTargets creates a slice of ScanTarget from hosts and ports.
func buildScanTargets(hosts []string, ports []uint16) []core.ScanTarget {
	scanTargets := make([]core.ScanTarget, len(hosts))
	for i, host := range hosts {
		scanTargets[i] = core.ScanTarget{Host: host, Ports: ports}
	}
	return scanTargets
}

// executeScan executes the scan based on the protocol (tcp, udp, or both).
func executeScan(ctx context.Context, protocol string, hosts []string, ports []uint16, cfg *config.Config) error {
	factory := NewScannerFactory(cfg)

	switch protocol {
	case "udp":
		scanner, err := factory.CreateScanner("udp")
		if err != nil {
			return err
		}
		return runProtocolScan(ctx, scanner, hosts, ports, cfg, "udp")

	case "both":
		tcpScanner, err := factory.CreateScanner("tcp")
		if err != nil {
			return err
		}
		if err := runProtocolScan(ctx, tcpScanner, hosts, ports, cfg, "tcp"); err != nil {
			return err
		}

		udpScanner, err := factory.CreateScanner("udp")
		if err != nil {
			return err
		}
		return runProtocolScan(ctx, udpScanner, hosts, ports, cfg, "udp")

	default:
		scanner, err := factory.CreateScanner("tcp")
		if err != nil {
			return err
		}
		return runProtocolScan(ctx, scanner, hosts, ports, cfg, "tcp")
	}
}

// handleScanOutput routes scan results to the appropriate output handler (TUI, JSON, CSV).
func handleScanOutput(ctx context.Context, cfg *config.Config, events <-chan core.Event, totalPorts int, metadata exporter.ScanMetadata) error {
	switch {
	case viper.GetBool("json") || cfg.Output == "json":
		exporter := selectJSONExporter(metadata)
		return streamEvents(ctx, events, exporter.Export, exporter.Close)
	case cfg.Output == "csv":
		exporter := exporter.NewCSVExporter(os.Stdout)
		return streamEvents(ctx, events, exporter.Export, exporter.Close)
	default:
		onlyOpen := viper.GetBool("only_open")
		tui := ui.NewScanUI(cfg, totalPorts, events, onlyOpen)
		return tui.Run()
	}
}

// validateInputs validates all user-provided configuration values.
func validateInputs(cfg *config.Config) error {
	// Validate port specification
	if err := targets.ValidatePortRange(cfg.Ports); err != nil {
		return errors.InvalidPortError(cfg.Ports, err)
	}

	// Validate rate limit
	if err := targets.ValidateRateLimit(cfg.Rate); err != nil {
		return &errors.UserError{
			Code:       "INVALID_RATE",
			Message:    "Invalid rate limit",
			Details:    err.Error(),
			Suggestion: fmt.Sprintf("Use a rate between 1 and %d pps. Start with 5000-10000 for normal scans.", core.MaxSafeRateLimit),
		}
	}

	// Validate timeout
	if err := targets.ValidateTimeout(cfg.TimeoutMs); err != nil {
		return &errors.UserError{
			Code:       "INVALID_TIMEOUT",
			Message:    "Invalid timeout value",
			Details:    err.Error(),
			Suggestion: "Use a timeout between 1ms and 60000ms. Default is 200ms.",
		}
	}

	// Validate workers
	if err := targets.ValidateWorkers(cfg.Workers); err != nil {
		return &errors.UserError{
			Code:       "INVALID_WORKERS",
			Message:    "Invalid worker count",
			Details:    err.Error(),
			Suggestion: "Use 0 for auto-detect, or 1-1000 workers. Default auto-detection works well.",
		}
	}

	// Validate UDP worker ratio
	if err := targets.ValidateUDPWorkerRatio(cfg.UDPWorkerRatio); err != nil {
		return &errors.UserError{
			Code:       "INVALID_UDP_RATIO",
			Message:    "Invalid UDP worker ratio",
			Details:    err.Error(),
			Suggestion: "Use a ratio between 0.0 and 1.0. Default is 0.5 (half workers for UDP).",
		}
	}

	return nil
}

// validateRawTargets validates each target before resolution.
func validateRawTargets(rawTargets []string) error {
	for _, target := range rawTargets {
		if err := targets.ValidateHost(target); err != nil {
			return errors.InvalidTargetError(target, err)
		}
	}
	return nil
}
