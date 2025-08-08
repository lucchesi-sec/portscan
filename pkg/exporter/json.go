package exporter

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/pkg/services"
)

type JSONExporter struct {
	writer     io.Writer
	encoder    *json.Encoder
	arrayMode  bool
	objectMode bool
	// metadata for object mode
	target     string
	totalPorts int
	scanRate   int
}

func NewJSONExporter(w io.Writer) *JSONExporter {
	return &JSONExporter{
		writer:  w,
		encoder: json.NewEncoder(w),
	}
}

// NewJSONExporterArray returns a JSON exporter that writes a single JSON array
// of result objects without buffering the entire result set in memory.
func NewJSONExporterArray(w io.Writer) *JSONExporter {
	return &JSONExporter{
		writer:    w,
		encoder:   json.NewEncoder(w),
		arrayMode: true,
	}
}

// NewJSONExporterObject returns a JSON exporter that writes a single JSON object
// with a results array and a scan_info metadata section, all streamed without
// buffering the entire result set in memory.
func NewJSONExporterObject(w io.Writer, target string, totalPorts int, scanRate int) *JSONExporter {
	return &JSONExporter{
		writer:     w,
		encoder:    json.NewEncoder(w),
		objectMode: true,
		target:     target,
		totalPorts: totalPorts,
		scanRate:   scanRate,
	}
}

func (e *JSONExporter) Export(results <-chan interface{}) {
	if e.objectMode {
		// Write opening object with results array first; scan_info appended at end.
		_, _ = e.writer.Write([]byte("{\n\"results\": ["))
		first := true
		startTime := time.Now()
		for result := range results {
			r, ok := result.(core.ResultEvent)
			if !ok {
				continue
			}
			dto := map[string]interface{}{
				"host":             r.Host,
				"port":             r.Port,
				"state":            string(r.State),
				"banner":           r.Banner,
				"response_time_ms": float64(r.Duration.Milliseconds()),
			}
			svc := strings.TrimSpace(r.Banner)
			if svc == "" {
				svc = services.GetName(r.Port)
			}
			dto["service"] = svc

			if !first {
				_, _ = e.writer.Write([]byte(","))
			}
			first = false
			b, err := json.Marshal(dto)
			if err == nil {
				_, _ = e.writer.Write(b)
			}
		}
		endTime := time.Now()
		_, _ = e.writer.Write([]byte("]"))
		// Append scan_info metadata
		info := map[string]interface{}{
			"target":      e.target,
			"start_time":  startTime.UTC().Format(time.RFC3339),
			"end_time":    endTime.UTC().Format(time.RFC3339),
			"total_ports": e.totalPorts,
			"scan_rate":   e.scanRate,
		}
		b, err := json.Marshal(info)
		if err == nil {
			_, _ = e.writer.Write([]byte(",\n\"scan_info\": "))
			_, _ = e.writer.Write(b)
		}
		_, _ = e.writer.Write([]byte("}\n"))
		return
	}

	if e.arrayMode {
		// Stream a JSON array: [obj1, obj2, ...]
		// We manually manage commas to avoid buffering.
		_, _ = e.writer.Write([]byte("["))
		first := true
		for result := range results {
			r, ok := result.(core.ResultEvent)
			if !ok {
				continue
			}
			dto := map[string]interface{}{
				"host":             r.Host,
				"port":             r.Port,
				"state":            string(r.State),
				"banner":           r.Banner,
				"response_time_ms": float64(r.Duration.Milliseconds()),
			}
			svc := strings.TrimSpace(r.Banner)
			if svc == "" {
				svc = services.GetName(r.Port)
			}
			dto["service"] = svc

			if !first {
				_, _ = e.writer.Write([]byte(","))
			}
			first = false
			// Marshal to control newline (Encoder.Encode adds a newline)
			b, err := json.Marshal(dto)
			if err == nil {
				_, _ = e.writer.Write(b)
			}
		}
		_, _ = e.writer.Write([]byte("]\n"))
		return
	}

	// Default: Stream each result as a single JSON object per line (NDJSON)
	for result := range results {
		if r, ok := result.(core.ResultEvent); ok {
			dto := map[string]interface{}{
				"host":             r.Host,
				"port":             r.Port,
				"state":            string(r.State),
				"banner":           r.Banner,
				"response_time_ms": float64(r.Duration.Milliseconds()),
			}
			// Derive service name: prefer banner-derived hint, else well-known port map
			svc := strings.TrimSpace(r.Banner)
			if svc == "" {
				svc = services.GetName(r.Port)
			}
			dto["service"] = svc

			// Best-effort encode; callers can check write errors on the underlying writer if needed.
			_ = e.encoder.Encode(dto)
		}
		// Ignore non-result events (e.g., progress) for JSON export
	}
}

func (e *JSONExporter) Close() error {
	return nil
}
