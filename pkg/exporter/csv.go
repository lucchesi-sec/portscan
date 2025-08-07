package exporter

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/lucchesi-sec/portscan/internal/core"
)

type CSVExporter struct {
	writer    io.Writer
	csvWriter *csv.Writer
}

func NewCSVExporter(w io.Writer) *CSVExporter {
	csvWriter := csv.NewWriter(w)
	// Write header
	csvWriter.Write([]string{"Host", "Port", "State", "Banner", "Latency(ms)"})
	return &CSVExporter{
		writer:    w,
		csvWriter: csvWriter,
	}
}

func (e *CSVExporter) Export(results <-chan interface{}) {
	for result := range results {
		if r, ok := result.(core.ResultEvent); ok {
			record := []string{
				r.Host,
				fmt.Sprintf("%d", r.Port),
				string(r.State),
				r.Banner,
				fmt.Sprintf("%d", r.Duration.Milliseconds()),
			}
			e.csvWriter.Write(record)
		}
	}
}

func (e *CSVExporter) Close() error {
	e.csvWriter.Flush()
	return e.csvWriter.Error()
}