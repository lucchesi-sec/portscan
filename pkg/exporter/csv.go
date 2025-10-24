package exporter

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/lucchesi-sec/portscan/internal/core"
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

// Export writes scan result events to CSV format.
func (e *CSVExporter) Export(events <-chan core.Event) {
	for event := range events {
		if event.Kind != core.EventKindResult {
			continue
		}

		r := *event.Result
		record := []string{
			r.Host,
			fmt.Sprintf("%d", r.Port),
			string(r.State),
			r.Banner,
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
