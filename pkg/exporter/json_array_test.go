package exporter

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/lucchesi-sec/portscan/internal/core"
)

func TestJSONExporterArrayMode(t *testing.T) {
	var buf bytes.Buffer
	exp := NewJSONExporterArray(&buf)
	ch := make(chan interface{}, 3)

	ch <- core.ResultEvent{Host: "h", Port: 1, State: core.StateOpen, Duration: 10 * time.Millisecond}
	ch <- core.ProgressEvent{Total: 2, Completed: 1, Rate: 10}
	ch <- core.ResultEvent{Host: "h", Port: 2, State: core.StateClosed, Duration: 5 * time.Millisecond}
	close(ch)

	exp.Export(ch)
	_ = exp.Close()

	var arr []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &arr); err != nil {
		t.Fatalf("array mode output not valid JSON array: %v\n%s", err, buf.String())
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 elements in array, got %d", len(arr))
	}
	if int(arr[0]["port"].(float64)) != 1 || arr[0]["state"].(string) != string(core.StateOpen) {
		t.Errorf("unexpected first element: %+v", arr[0])
	}
}

func TestJSONExporterArrayModeEmptyInput(t *testing.T) {
	var buf bytes.Buffer
	exp := NewJSONExporterArray(&buf)
	ch := make(chan interface{})
	close(ch)

	exp.Export(ch)
	_ = exp.Close()

	var arr []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &arr); err != nil {
		t.Fatalf("array mode output not valid JSON array: %v\n%s", err, buf.String())
	}
	if len(arr) != 0 {
		t.Fatalf("expected empty array, got %d elements", len(arr))
	}
}
