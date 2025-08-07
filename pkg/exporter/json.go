package exporter

import (
	"encoding/json"
	"io"

	"github.com/lucchesi-sec/portscan/internal/core"
)

type JSONExporter struct {
	writer  io.Writer
	encoder *json.Encoder
}

func NewJSONExporter(w io.Writer) *JSONExporter {
	return &JSONExporter{
		writer:  w,
		encoder: json.NewEncoder(w),
	}
}

func (e *JSONExporter) Export(results <-chan interface{}) {
	var allResults []core.ResultEvent

	for result := range results {
		if r, ok := result.(core.ResultEvent); ok {
			allResults = append(allResults, r)
		}
	}

	_ = e.encoder.Encode(allResults)
}

func (e *JSONExporter) Close() error {
	return nil
}
