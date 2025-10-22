package exporter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/lucchesi-sec/portscan/internal/core"
)

type resultDTO struct {
	Host           string  `json:"host"`
	Port           uint16  `json:"port"`
	State          string  `json:"state"`
	Service        string  `json:"service"`
	Banner         string  `json:"banner"`
	ResponseTimeMS float64 `json:"response_time_ms"`
}

func TestJSONExporterStreamsNDJSON(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	exporter := NewJSONExporter(w)
	ch := make(chan core.Event, 3)

	// Send two results and a progress event in between
	ch <- core.Event{Type: core.EventTypeResult, Result: core.ResultEvent{Host: "127.0.0.1", Port: 80, State: core.StateOpen, Banner: "HTTP", Duration: 150 * time.Millisecond}}
	ch <- core.Event{Type: core.EventTypeProgress, Progress: core.ProgressEvent{Total: 2, Completed: 1, Rate: 10}}
	ch <- core.Event{Type: core.EventTypeResult, Result: core.ResultEvent{Host: "127.0.0.1", Port: 22, State: core.StateClosed, Banner: "", Duration: 20 * time.Millisecond}}
	close(ch)

	exporter.Export(ch)
	_ = exporter.Close()
	// Ensure buffered data is flushed before reading
	_ = w.Flush()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 NDJSON lines, got %d: %q", len(lines), buf.String())
	}

	var r1, r2 resultDTO
	if err := json.Unmarshal([]byte(lines[0]), &r1); err != nil {
		t.Fatalf("first line invalid JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &r2); err != nil {
		t.Fatalf("second line invalid JSON: %v", err)
	}

	if r1.Port != 80 || r1.State != string(core.StateOpen) || r1.Service == "" || r1.ResponseTimeMS <= 0 {
		t.Errorf("unexpected first record: %+v", r1)
	}
	if r2.Port != 22 || r2.State != string(core.StateClosed) {
		t.Errorf("unexpected second record: %+v", r2)
	}
}

func TestJSONExporterEmptyInputNDJSON(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	exporter := NewJSONExporter(w)
	ch := make(chan core.Event)
	close(ch)

	exporter.Export(ch)
	_ = exporter.Close()
	_ = w.Flush()

	output := strings.TrimSpace(buf.String())
	if output != "" {
		t.Errorf("expected empty output for empty input channel, got: %q", output)
	}
}
