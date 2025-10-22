package exporter

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/lucchesi-sec/portscan/internal/core"
)

func TestJSONExporterObjectMode(t *testing.T) {
	var buf bytes.Buffer
	exp := NewJSONExporterObjectWithMetadata(&buf, ScanMetadata{Targets: []string{"1.2.3.4"}, TotalPorts: 2, Rate: 7500})
	ch := make(chan core.Event, 3)

	ch <- core.Event{Type: core.EventTypeResult, Result: core.ResultEvent{Host: "1.2.3.4", Port: 22, State: core.StateOpen, Duration: 12 * time.Millisecond}}
	ch <- core.Event{Type: core.EventTypeProgress, Progress: core.ProgressEvent{Total: 2, Completed: 1, Rate: 100}}
	ch <- core.Event{Type: core.EventTypeResult, Result: core.ResultEvent{Host: "1.2.3.4", Port: 80, State: core.StateClosed, Duration: 5 * time.Millisecond}}
	close(ch)

	exp.Export(ch)
	_ = exp.Close()

	var obj struct {
		Results  []map[string]interface{} `json:"results"`
		ScanInfo map[string]interface{}   `json:"scan_info"`
	}
	if err := json.Unmarshal(buf.Bytes(), &obj); err != nil {
		t.Fatalf("object mode output not valid JSON object: %v\n%s", err, buf.String())
	}
	if len(obj.Results) != 2 {
		t.Fatalf("expected 2 results in object mode, got %d", len(obj.Results))
	}
	targets, ok := obj.ScanInfo["targets"].([]interface{})
	if !ok || len(targets) != 1 || targets[0].(string) != "1.2.3.4" {
		t.Errorf("unexpected targets: %v", obj.ScanInfo["targets"])
	}
	if int(obj.ScanInfo["total_ports"].(float64)) != 2 || int(obj.ScanInfo["scan_rate"].(float64)) != 7500 {
		t.Errorf("unexpected scan_info: %+v", obj.ScanInfo)
	}
}

func TestJSONExporterObjectModeEmptyResults(t *testing.T) {
	var buf bytes.Buffer
	exp := NewJSONExporterObjectWithMetadata(&buf, ScanMetadata{Targets: []string{"10.0.0.1"}, TotalPorts: 100, Rate: 5000})
	ch := make(chan core.Event)
	close(ch)

	exp.Export(ch)
	_ = exp.Close()

	var obj struct {
		Results  []map[string]interface{} `json:"results"`
		ScanInfo map[string]interface{}   `json:"scan_info"`
	}
	if err := json.Unmarshal(buf.Bytes(), &obj); err != nil {
		t.Fatalf("object mode output not valid JSON object: %v\n%s", err, buf.String())
	}
	if len(obj.Results) != 0 {
		t.Fatalf("expected empty results array, got %d results", len(obj.Results))
	}
	targets, ok := obj.ScanInfo["targets"].([]interface{})
	if !ok || len(targets) != 1 || targets[0].(string) != "10.0.0.1" {
		t.Errorf("unexpected targets: %v", obj.ScanInfo["targets"])
	}
	if int(obj.ScanInfo["total_ports"].(float64)) != 100 || int(obj.ScanInfo["scan_rate"].(float64)) != 5000 {
		t.Errorf("unexpected scan_info: %+v", obj.ScanInfo)
	}
}
