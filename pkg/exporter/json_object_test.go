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
	exp := NewJSONExporterObject(&buf, "1.2.3.4", 2, 7500)
	ch := make(chan interface{}, 3)

	ch <- core.ResultEvent{Host: "1.2.3.4", Port: 22, State: core.StateOpen, Duration: 12 * time.Millisecond}
	ch <- core.ProgressEvent{Total: 2, Completed: 1, Rate: 100}
	ch <- core.ResultEvent{Host: "1.2.3.4", Port: 80, State: core.StateClosed, Duration: 5 * time.Millisecond}
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
	if obj.ScanInfo["target"].(string) != "1.2.3.4" {
		t.Errorf("unexpected target: %v", obj.ScanInfo["target"])
	}
	if int(obj.ScanInfo["total_ports"].(float64)) != 2 || int(obj.ScanInfo["scan_rate"].(float64)) != 7500 {
		t.Errorf("unexpected scan_info: %+v", obj.ScanInfo)
	}
}
