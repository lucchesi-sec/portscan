// Package exporter provides scan result export functionality in multiple formats.
//
// This package implements streaming exporters that write scan results without
// buffering, making it suitable for large scans that produce millions of results.
// All exporters implement a common interface for consistent usage.
//
// Supported Formats:
//
// 1. NDJSON (Newline-Delimited JSON) - Default
//
// Each result is a complete JSON object on its own line, ideal for streaming
// and log processing:
//
//	{"host":"192.168.1.1","port":22,"state":"open",...}
//	{"host":"192.168.1.1","port":80,"state":"open",...}
//
// 2. JSON Array
//
// Results wrapped in a JSON array with proper comma placement:
//
//	[
//	  {"host":"192.168.1.1","port":22,"state":"open",...},
//	  {"host":"192.168.1.1","port":80,"state":"open",...}
//	]
//
// 3. JSON Object with Metadata
//
// Complete scan results with metadata about the scan run:
//
//	{
//	  "scan_info": {
//	    "targets": ["192.168.1.1"],
//	    "start_time": "2025-01-15T10:30:00Z",
//	    "scan_rate": 7500
//	  },
//	  "results": [...]
//	}
//
// 4. CSV (Comma-Separated Values)
//
// Standard CSV format with headers, suitable for Excel/spreadsheets:
//
//	host,port,protocol,state,service,banner,latency_ms
//	192.168.1.1,22,tcp,open,ssh,"SSH-2.0-OpenSSH_8.9p1",5.23
//
// Example Usage:
//
//	// Create JSON exporter (NDJSON mode)
//	exporter := exporter.NewJSONExporter(os.Stdout, exporter.JSONModeNDJSON, nil)
//	defer exporter.Close()
//
//	// Write results as they arrive
//	for result := range resultChannel {
//	    if err := exporter.Write(result); err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
//	// Create CSV exporter
//	csvFile, _ := os.Create("results.csv")
//	csvExp := exporter.NewCSVExporter(csvFile)
//	defer csvExp.Close()
//
// Security:
//
// All exporters handle CSV injection attacks by properly escaping fields.
// Special characters in banners and service names are safely encoded.
//
// Performance:
//
// Exporters use streaming writes with OS-level buffering, avoiding memory
// accumulation. Suitable for scans producing gigabytes of results.
package exporter
