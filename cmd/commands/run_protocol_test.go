package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/pkg/config"
	"github.com/spf13/viper"
)

func TestRunProtocolScanJSONOutput(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	openPort := uint16(ln.Addr().(*net.TCPAddr).Port)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(cancel)

	var acceptWg sync.WaitGroup
	acceptWg.Add(1)
	go func() {
		defer acceptWg.Done()
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	cfg := &config.Config{
		Rate:           1000,
		Workers:        4,
		TimeoutMs:      200,
		Banners:        false,
		UDPWorkerRatio: 0.5,
	}

	scannerCfg := &core.Config{
		Workers:        cfg.Workers,
		Timeout:        cfg.GetTimeout(),
		RateLimit:      cfg.Rate,
		BannerGrab:     cfg.Banners,
		MaxRetries:     1,
		UDPWorkerRatio: cfg.UDPWorkerRatio,
	}

	scanner := core.NewScanner(scannerCfg)

	viper.Set("json", true)
	viper.Set("json_object", true)
	viper.Set("json_array", false)

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer func() {
		_ = r.Close()
	}()
	os.Stdout = w
	defer func() {
		os.Stdout = origStdout
	}()
	var buf bytes.Buffer
	readDone := make(chan struct{})
	go func() {
		_, _ = buf.ReadFrom(r)
		close(readDone)
	}()

	err = runProtocolScan(ctx, scanner, []string{"127.0.0.1"}, []uint16{openPort}, cfg, "tcp")
	if err != nil {
		t.Fatalf("runProtocolScan returned error: %v", err)
	}

	_ = w.Close()
	<-readDone

	_ = ln.Close()
	acceptWg.Wait()

	output := bytes.TrimSpace(buf.Bytes())
	if len(output) == 0 {
		t.Fatalf("expected JSON output, got empty string")
	}

	type jsonResult struct {
		Host  string `json:"host"`
		Port  uint16 `json:"port"`
		State string `json:"state"`
	}

	type jsonEnvelope struct {
		Results  []jsonResult `json:"results"`
		ScanInfo struct {
			Targets    []string `json:"targets"`
			TotalPorts int      `json:"total_ports"`
		} `json:"scan_info"`
	}

	var parsed jsonEnvelope
	if err := json.Unmarshal(output, &parsed); err != nil {
		t.Fatalf("failed to decode JSON output: %v\n%s", err, output)
	}

	if len(parsed.Results) != 1 {
		t.Fatalf("expected 1 result entry, got %d", len(parsed.Results))
	}

	res := parsed.Results[0]
	if res.Host != "127.0.0.1" || res.Port != openPort || res.State != string(core.StateOpen) {
		t.Fatalf("unexpected result entry: %+v", res)
	}

	if len(parsed.ScanInfo.Targets) != 1 || parsed.ScanInfo.Targets[0] != "127.0.0.1" {
		t.Fatalf("unexpected targets metadata: %+v", parsed.ScanInfo.Targets)
	}

	if parsed.ScanInfo.TotalPorts != 1 {
		t.Fatalf("expected total_ports to be 1, got %d", parsed.ScanInfo.TotalPorts)
	}
}
