package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/internal/ui"
	"github.com/lucchesi-sec/portscan/pkg/config"
	"github.com/lucchesi-sec/portscan/pkg/errors"
	"github.com/lucchesi-sec/portscan/pkg/exporter"
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

	monitorInterrupts(cancel)

	protocol := cfg.Protocol
	if protocol == "" {
		protocol = "tcp"
	}

	scannerCfg := buildScannerConfig(cfg)

	switch protocol {
	case "udp":
		return runProtocolScan(ctx, core.NewUDPScanner(scannerCfg), resolvedTargets, ports, cfg, "udp")
	case "both":
		if err := runProtocolScan(ctx, core.NewScanner(scannerCfg), resolvedTargets, ports, cfg, "tcp"); err != nil {
			return err
		}
		return runProtocolScan(ctx, core.NewUDPScanner(scannerCfg), resolvedTargets, ports, cfg, "udp")
	default:
		return runProtocolScan(ctx, core.NewScanner(scannerCfg), resolvedTargets, ports, cfg, "tcp")
	}
}

func runProtocolScan(ctx context.Context, scanner interface{}, hosts []string, ports []uint16, cfg *config.Config, _ string) error {
	if len(hosts) == 0 {
		return errors.NoTargetError()
	}

	scanTargets := make([]core.ScanTarget, len(hosts))
	for i, host := range hosts {
		scanTargets[i] = core.ScanTarget{Host: host, Ports: ports}
	}

	var events <-chan core.Event
	switch s := scanner.(type) {
	case *core.Scanner:
		events = s.Results()
		go s.ScanTargets(ctx, scanTargets)
	case *core.UDPScanner:
		events = s.Results()
		go s.ScanTargets(ctx, scanTargets)
	default:
		return fmt.Errorf("unsupported scanner type")
	}

	totalPorts := len(ports) * len(hosts)
	metadata := exporter.ScanMetadata{Targets: hosts, TotalPorts: totalPorts, Rate: cfg.Rate}

	switch {
	case viper.GetBool("json") || cfg.Output == "json":
		exporter := selectJSONExporter(metadata)
		return streamEvents(events, exporter.Export, exporter.Close)
	case cfg.Output == "csv":
		exporter := exporter.NewCSVExporter(os.Stdout)
		return streamEvents(events, exporter.Export, exporter.Close)
	default:
		onlyOpen := viper.GetBool("only_open")
		tui := ui.NewScanUI(cfg, totalPorts, events, onlyOpen)
		return tui.Run()
	}
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

func streamEvents(events <-chan core.Event, export func(<-chan core.Event), closeFn func() error) error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		export(events)
	}()
	wg.Wait()
	return closeFn()
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
	const maxSafeRate = 15000
	if rate > maxSafeRate {
		return errors.RateLimitError(rate, maxSafeRate)
	}
	return nil
}

func monitorInterrupts(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()
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
