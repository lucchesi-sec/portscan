package exporter

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/lucchesi-sec/portscan/internal/core"
)

const (
	// maxFieldLength is the maximum allowed length for CSV fields to prevent abuse
	maxFieldLength = 256
)

// CSVExporter exports scan results to CSV format.
type CSVExporter struct {
	writer    io.Writer
	csvWriter *csv.Writer
	writeErr  error
}

// NewCSVExporter creates a new CSV exporter that writes to the given writer.
func NewCSVExporter(w io.Writer) *CSVExporter {
	csvWriter := csv.NewWriter(w)
	// Write header
	_ = csvWriter.Write([]string{"host", "port", "state", "banner", "latency_ms"})
	return &CSVExporter{
		writer:    w,
		csvWriter: csvWriter,
	}
}

// sanitizeCSVField sanitizes a CSV field to prevent formula injection attacks.
// It strips leading formula characters (=, +, -, @), caps field length,
// and escapes dangerous patterns that could be executed in spreadsheet applications.
func sanitizeCSVField(field string) string {
	if field == "" {
		return field
	}

	// Strip leading whitespace first, then formula characters
	field = strings.TrimSpace(field)
	field = strings.TrimLeft(field, "=+-@")

	// Cap field length to prevent abuse
	if len(field) > maxFieldLength {
		field = field[:maxFieldLength]
	}

	// If field starts with tab, carriage return, or newline after trimming, prefix with single quote
	if len(field) > 0 && (field[0] == '\t' || field[0] == '\r' || field[0] == '\n') {
		field = "'" + field
	}

	return field
}

// Export writes scan result events to CSV format.
func (e *CSVExporter) Export(events <-chan core.Event) {
	for event := range events {
		if event.Kind != core.EventKindResult {
			continue
		}

		r := *event.Result
		record := []string{
			sanitizeCSVField(r.Host),
			fmt.Sprintf("%d", r.Port),
			sanitizeCSVField(string(r.State)),
			sanitizeCSVField(r.Banner),
			fmt.Sprintf("%d", r.Duration.Milliseconds()),
		}
		if err := e.csvWriter.Write(record); err != nil {
			e.writeErr = err
			return
		}
	}
}

// Close flushes the CSV writer and returns any errors.
func (e *CSVExporter) Close() error {
	e.csvWriter.Flush()
	if err := e.csvWriter.Error(); err != nil {
		return err
	}
	return e.writeErr
}
