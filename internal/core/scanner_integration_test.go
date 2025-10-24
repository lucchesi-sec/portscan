package core

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestScannerEndToEnd(t *testing.T) {
	openLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer func() {
		_ = openLn.Close()
	}()

	openPort := uint16(openLn.Addr().(*net.TCPAddr).Port)

	closedLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate closed port: %v", err)
	}
	closedPort := uint16(closedLn.Addr().(*net.TCPAddr).Port)
	_ = closedLn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		for {
			conn, err := openLn.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	cfg := &Config{
		Workers:    4,
		Timeout:    200 * time.Millisecond,
		RateLimit:  0,
		BannerGrab: false,
		MaxRetries: 1,
	}

	scanner := NewScanner(cfg)
	results := scanner.Results()

	go scanner.ScanRange(ctx, "127.0.0.1", []uint16{openPort, closedPort})

	states := make(map[uint16]ScanState, 2)
	for event := range results {
		if event.Kind != EventKindResult {
			continue
		}
		if event.Result == nil {
			t.Fatal("received result event with nil Result")
		}
		states[event.Result.Port] = event.Result.State
	}

	if state, ok := states[openPort]; !ok {
		t.Fatalf("did not receive result for open port %d", openPort)
	} else if state != StateOpen {
		t.Fatalf("expected port %d to be reported open, got %s", openPort, state)
	}

	if state, ok := states[closedPort]; !ok {
		t.Fatalf("did not receive result for closed port %d", closedPort)
	} else if state != StateClosed {
		t.Fatalf("expected port %d to be reported closed, got %s", closedPort, state)
	}
}
