package exporter

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/lucchesi-sec/portscan/internal/core"
	"github.com/lucchesi-sec/portscan/pkg/services"
)

// JSONExporter exports scan results in JSON format (NDJSON, array, or object).
type JSONExporter struct {
	writer     io.Writer
	encoder    *json.Encoder
	arrayMode  bool
	objectMode bool
	// metadata for object mode
	metadata ScanMetadata
}

// ScanMetadata holds metadata about a scan for inclusion in JSON export.
type ScanMetadata struct {
	Targets    []string
	TotalPorts int
	Rate       int
}

// buildResultDTO creates a consistent DTO from a ResultEvent
func buildResultDTO(r core.ResultEvent) map[string]interface{} {
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

	return dto
}

// NewJSONExporter creates a new NDJSON exporter that writes one JSON object per line.
func NewJSONExporter(w io.Writer) *JSONExporter {
	return &JSONExporter{
		writer:   w,
		encoder:  json.NewEncoder(w),
		metadata: ScanMetadata{},
	}
}

// NewJSONExporterArray returns a JSON exporter that writes a single JSON array
// of result objects without buffering the entire result set in memory.
func NewJSONExporterArray(w io.Writer) *JSONExporter {
	return &JSONExporter{
		writer:    w,
		encoder:   json.NewEncoder(w),
		arrayMode: true,
		metadata:  ScanMetadata{},
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
		metadata: ScanMetadata{
			Targets:    []string{target},
			TotalPorts: totalPorts,
			Rate:       scanRate,
		},
	}
}

// NewJSONExporterObjectWithMetadata creates a JSON object exporter with custom metadata.
func NewJSONExporterObjectWithMetadata(w io.Writer, meta ScanMetadata) *JSONExporter {
	copyTargets := make([]string, len(meta.Targets))
	copy(copyTargets, meta.Targets)
	return &JSONExporter{
		writer:     w,
		encoder:    json.NewEncoder(w),
		objectMode: true,
		metadata: ScanMetadata{
			Targets:    copyTargets,
			TotalPorts: meta.TotalPorts,
			Rate:       meta.Rate,
		},
	}
}

// Export writes scan result events in the configured JSON format.
func (e *JSONExporter) Export(events <-chan core.Event) {
	if e.objectMode {
		// Write opening object with results array first; scan_info appended at end.
		_, _ = e.writer.Write([]byte("{\n\"results\": ["))
		first := true
		startTime := time.Now()
		for event := range events {
			if event.Kind != core.EventKindResult {
				continue
			}
			r := *event.Result
			dto := buildResultDTO(r)

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
			"targets":     e.metadata.Targets,
			"start_time":  startTime.UTC().Format(time.RFC3339),
			"end_time":    endTime.UTC().Format(time.RFC3339),
			"total_ports": e.metadata.TotalPorts,
			"scan_rate":   e.metadata.Rate,
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
		for event := range events {
			if event.Kind != core.EventKindResult {
				continue
			}
			r := *event.Result
			dto := buildResultDTO(r)

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
	for event := range events {
		if event.Kind != core.EventKindResult {
			continue
		}
		dto := buildResultDTO(*event.Result)

		// Best-effort encode; callers can check write errors on the underlying writer if needed.
		_ = e.encoder.Encode(dto)
	}
}

// Close is a no-op for JSON exporters.
func (e *JSONExporter) Close() error {
	return nil
}
