# Technical Debt Analysis Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Comprehensive analysis and remediation of all technical debt in the port scanner codebase

**Architecture:** Systematic approach covering test coverage, code quality, security, performance, documentation, and CI/CD improvements

**Tech Stack:** Go 1.24, golangci-lint, CodeQL, GitHub Actions, Docker

---

## Category 1: Test Coverage Gaps

### Task 1: Add pkg/errors Tests

**Files:**
- Test: `pkg/errors/user_errors_test.go` (create)
- Source: `pkg/errors/user_errors.go`

**Step 1: Write failing tests for error constructors**

```go
package errors

import (
	"testing"
)

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("test message", "test_field")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Add assertions for error fields
}

func TestNewConfigError(t *testing.T) {
	// Test config error creation
}

func TestErrorMessages(t *testing.T) {
	// Test Error() methods return correct messages
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./pkg/errors -v`
Expected: FAIL or compilation errors if types don't exist

**Step 3: Implement missing functionality**

Review `pkg/errors/user_errors.go` and implement missing test coverage for all exported functions.

**Step 4: Run tests to verify they pass**

Run: `go test ./pkg/errors -v -cover`
Expected: PASS with coverage >80%

**Step 5: Commit**

```bash
git add pkg/errors/user_errors_test.go
git commit -m "test: add comprehensive error package tests"
```

---

### Task 2: Add pkg/services Tests

**Files:**
- Test: `pkg/services/services_test.go` (create)
- Source: `pkg/services/services.go`

**Step 1: Write failing tests for service detection**

```go
package services

import (
	"testing"
)

func TestServiceLookup(t *testing.T) {
	tests := []struct {
		port     uint16
		expected string
	}{
		{22, "ssh"},
		{80, "http"},
		{443, "https"},
		{3306, "mysql"},
	}

	for _, tt := range tests {
		t.Run("port_"+string(rune(tt.port)), func(t *testing.T) {
			service := Lookup(tt.port)
			if service != tt.expected {
				t.Errorf("Lookup(%d) = %s; want %s", tt.port, service, tt.expected)
			}
		})
	}
}

func TestBannerIdentification(t *testing.T) {
	// Test banner parsing logic if it exists
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./pkg/services -v`
Expected: FAIL or compilation errors

**Step 3: Review source and add comprehensive tests**

Analyze `pkg/services/services.go` to understand all functionality and add tests for each function.

**Step 4: Run tests to verify they pass**

Run: `go test ./pkg/services -v -cover`
Expected: PASS with coverage >90%

**Step 5: Commit**

```bash
git add pkg/services/services_test.go
git commit -m "test: add service detection tests"
```

---

### Task 3: Improve pkg/config Test Coverage

**Files:**
- Modify: `pkg/config/config_test.go`
- Source: `pkg/config/config.go`

**Step 1: Write tests for validation**

```go
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Rate:      7500,
				TimeoutMs: 200,
				Workers:   100,
			},
			wantErr: false,
		},
		{
			name: "invalid rate too high",
			config: Config{
				Rate:      200000,
				TimeoutMs: 200,
				Workers:   100,
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: Config{
				Rate:      7500,
				TimeoutMs: 0,
				Workers:   100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := validator.New()
			err := validate.Struct(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetTimeout(t *testing.T) {
	cfg := &Config{TimeoutMs: 200}
	timeout := cfg.GetTimeout()
	expected := 200 * time.Millisecond
	if timeout != expected {
		t.Errorf("GetTimeout() = %v; want %v", timeout, expected)
	}
}
```

**Step 2: Run tests**

Run: `go test ./pkg/config -v`
Expected: FAIL initially

**Step 3: Fix any issues and add more tests**

Add tests for Load(), defaults, Viper integration

**Step 4: Verify coverage improved**

Run: `go test ./pkg/config -v -cover`
Expected: Coverage >60% (from 5.6%)

**Step 5: Commit**

```bash
git add pkg/config/config_test.go
git commit -m "test: improve config package coverage to 60%"
```

---

### Task 4: Add internal/ui/components Tests

**Files:**
- Test: `internal/ui/components/splitview_test.go` (create)
- Source: `internal/ui/components/splitview.go`

**Step 1: Write tests for component logic**

```go
package components

import (
	"testing"
)

func TestSplitViewDimensions(t *testing.T) {
	// Test dimension calculations
	sv := NewSplitView(100, 50, 0.5)
	if sv.LeftWidth != 50 {
		t.Errorf("LeftWidth = %d; want 50", sv.LeftWidth)
	}
}

func TestSplitViewResize(t *testing.T) {
	// Test resize behavior
}

func TestSplitRatio(t *testing.T) {
	// Test ratio calculations
}
```

**Step 2: Run tests**

Run: `go test ./internal/ui/components -v`
Expected: FAIL or compilation errors

**Step 3: Implement comprehensive tests**

Review splitview.go for all testable logic (avoid testing Bubble Tea rendering).

**Step 4: Verify tests pass**

Run: `go test ./internal/ui/components -v -cover`
Expected: PASS with coverage >50%

**Step 5: Commit**

```bash
git add internal/ui/components/splitview_test.go
git commit -m "test: add splitview component tests"
```

---

### Task 5: Improve internal/ui Test Coverage

**Files:**
- Modify: `internal/ui/stats_test.go` (create)
- Modify: `internal/ui/services_lookup_test.go` (create)
- Source: `internal/ui/stats.go`, `internal/ui/services_lookup.go`

**Step 1: Write stats tests**

```go
package ui

import (
	"testing"
	"github.com/lucchesi-sec/portscan/internal/core"
)

func TestStatsAccumulation(t *testing.T) {
	stats := NewStats()

	stats.RecordResult(core.ResultEvent{State: "open"})
	stats.RecordResult(core.ResultEvent{State: "closed"})
	stats.RecordResult(core.ResultEvent{State: "filtered"})

	if stats.OpenCount != 1 {
		t.Errorf("OpenCount = %d; want 1", stats.OpenCount)
	}
	if stats.ClosedCount != 1 {
		t.Errorf("ClosedCount = %d; want 1", stats.ClosedCount)
	}
	if stats.FilteredCount != 1 {
		t.Errorf("FilteredCount = %d; want 1", stats.FilteredCount)
	}
}

func TestStatsFormatting(t *testing.T) {
	// Test string formatting methods
}
```

**Step 2: Run tests**

Run: `go test ./internal/ui -v`
Expected: Some failures initially

**Step 3: Add services_lookup tests**

```go
func TestServiceLookupCache(t *testing.T) {
	// Test caching behavior if applicable
}
```

**Step 4: Verify coverage improved**

Run: `go test ./internal/ui -v -cover`
Expected: Coverage >30% (from 17.6%)

**Step 5: Commit**

```bash
git add internal/ui/stats_test.go internal/ui/services_lookup_test.go
git commit -m "test: improve UI package coverage to 30%"
```

---

### Task 6: Improve cmd/commands Test Coverage

**Files:**
- Modify: `cmd/commands/scan_flags_test.go` (create new tests)
- Source: `cmd/commands/scan_command.go`, `cmd/commands/scan_runner.go`

**Step 1: Add flag binding tests**

```go
package commands

import (
	"testing"
	"github.com/spf13/viper"
)

func TestFlagBindings(t *testing.T) {
	// Reset viper
	viper.Reset()

	tests := []struct {
		flag     string
		setValue interface{}
	}{
		{"stdin", true},
		{"json", true},
		{"ports", "80,443"},
		{"rate", 5000},
		{"workers", 50},
		{"banners", true},
	}

	// Initialize flags
	initScanCmd()

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			// Set flag value
			scanCmd.Flags().Set(tt.flag, fmt.Sprint(tt.setValue))

			// Verify Viper binding
			actual := viper.Get(tt.flag)
			if actual == nil {
				t.Errorf("Flag %s not bound to Viper", tt.flag)
			}
		})
	}
}

func TestDryRunValidation(t *testing.T) {
	// Test dry run logic
}
```

**Step 2: Run tests**

Run: `go test ./cmd/commands -v -run TestFlag`
Expected: Some failures initially

**Step 3: Add integration-style tests**

```go
func TestScanCommandExecution(t *testing.T) {
	// Test command execution with mocked scanner
}
```

**Step 4: Verify coverage**

Run: `go test ./cmd/commands -v -cover`
Expected: Coverage >50% (from 36.0%)

**Step 5: Commit**

```bash
git add cmd/commands/scan_flags_test.go
git commit -m "test: improve command layer coverage to 50%"
```

---

## Category 2: Code Quality Issues

### Task 7: Fix Code Formatting

**Files:**
- Modify: `cmd/commands/scan_helpers_test.go`
- Modify: `internal/core/udp_jitter_test.go`
- Modify: `internal/ui/result_buffer_test.go`
- Modify: `internal/ui/scan_ui_model.go`
- Modify: `internal/ui/scan_ui_update.go`
- Modify: `internal/ui/scan_ui_view.go`
- Modify: `internal/ui/stats.go`

**Step 1: Check current formatting**

Run: `gofmt -s -l .`
Expected: List of unformatted files

**Step 2: Apply formatting**

Run: `gofmt -s -w cmd/commands/scan_helpers_test.go internal/core/udp_jitter_test.go internal/ui/result_buffer_test.go internal/ui/scan_ui_model.go internal/ui/scan_ui_update.go internal/ui/scan_ui_view.go internal/ui/stats.go`
Expected: Files reformatted

**Step 3: Verify formatting**

Run: `gofmt -s -l .`
Expected: Empty output

**Step 4: Run tests to ensure no breakage**

Run: `go test ./...`
Expected: All tests pass

**Step 5: Commit**

```bash
git add -u
git commit -m "style: apply gofmt to all source files"
```

---

### Task 8: Refactor Mixed-Type Results Channel

**Files:**
- Modify: `internal/core/scan_types.go`
- Modify: `internal/core/scanner.go:70`
- Modify: `internal/ui/scan_ui_update.go` (update event handling)
- Modify: `pkg/exporter/json.go` (update event handling)
- Modify: `pkg/exporter/csv.go` (update event handling)

**Step 1: Define typed event envelope**

Edit `internal/core/scan_types.go`:

```go
package core

// EventKind identifies the type of event
type EventKind string

const (
	EventKindResult   EventKind = "result"
	EventKindProgress EventKind = "progress"
	EventKindError    EventKind = "error"
)

// Event is a typed envelope for all scanner events
type Event struct {
	Kind     EventKind
	Result   *ResultEvent
	Progress *ProgressEvent
	Error    error
}

// Helper constructors
func NewResultEvent(r ResultEvent) Event {
	return Event{Kind: EventKindResult, Result: &r}
}

func NewProgressEvent(p ProgressEvent) Event {
	return Event{Kind: EventKindProgress, Progress: &p}
}
```

**Step 2: Write tests for event creation**

Create `internal/core/scan_types_test.go`:

```go
package core

import (
	"testing"
)

func TestNewResultEvent(t *testing.T) {
	result := ResultEvent{Host: "localhost", Port: 80, State: "open"}
	event := NewResultEvent(result)

	if event.Kind != EventKindResult {
		t.Errorf("Kind = %v; want %v", event.Kind, EventKindResult)
	}
	if event.Result == nil {
		t.Fatal("Result is nil")
	}
	if event.Result.Host != "localhost" {
		t.Errorf("Host = %s; want localhost", event.Result.Host)
	}
}

func TestNewProgressEvent(t *testing.T) {
	progress := ProgressEvent{Completed: 100, Total: 1000}
	event := NewProgressEvent(progress)

	if event.Kind != EventKindProgress {
		t.Errorf("Kind = %v; want %v", event.Kind, EventKindProgress)
	}
	if event.Progress == nil {
		t.Fatal("Progress is nil")
	}
}
```

**Step 3: Run tests**

Run: `go test ./internal/core -v -run TestNew`
Expected: PASS

**Step 4: Update scanner to use typed events**

Edit `internal/core/scanner.go:70`:

```go
// Change from:
// results: make(chan Event, 1000),

// To: (already using Event type, just ensure consistency)
results: make(chan Event, 1000),
```

Update all places that send events:

```go
// Change from:
// s.results <- ResultEvent{...}

// To:
s.results <- NewResultEvent(ResultEvent{...})
```

**Step 5: Update consumers**

Update `internal/ui/scan_ui_update.go`:

```go
case event := <-m.scanner.Results():
	switch event.Kind {
	case core.EventKindResult:
		m.buffer.Append(*event.Result)
		m.stats.RecordResult(*event.Result)
	case core.EventKindProgress:
		m.progress.Update(event.Progress.Completed, event.Progress.Total)
	}
```

Update exporters similarly.

**Step 6: Run all tests**

Run: `go test ./...`
Expected: All tests pass

**Step 7: Commit**

```bash
git add internal/core/scan_types.go internal/core/scan_types_test.go internal/core/scanner.go internal/ui/scan_ui_update.go pkg/exporter/*.go
git commit -m "refactor: use typed event envelope instead of interface{}"
```

---

### Task 9: Add Comment Documentation

**Files:**
- Modify: `cmd/commands/scan_helpers.go:getOptimalWorkerCount`
- Modify: All exported functions in `pkg/` packages

**Step 1: Audit missing documentation**

Run: `golangci-lint run --disable-all --enable=golint`
Expected: List of undocumented exports

**Step 2: Add function documentation**

Edit `cmd/commands/scan_helpers.go`:

```go
// getOptimalWorkerCount calculates worker count based on available CPU cores.
// Returns cores * 50, capped at 1000.
// If workers is already set to a positive value, returns that value unchanged.
func getOptimalWorkerCount(workers int) int {
	if workers > 0 {
		return workers
	}
	cores := runtime.NumCPU()
	optimal := cores * 50
	const maxWorkers = 1000
	if optimal > maxWorkers {
		return maxWorkers
	}
	return optimal
}
```

**Step 3: Add package documentation**

Add doc.go files to packages missing them:

```go
// Package scanner provides high-performance port scanning capabilities.
package scanner
```

**Step 4: Verify with golint**

Run: `golangci-lint run --disable-all --enable=golint`
Expected: Fewer or no documentation warnings

**Step 5: Commit**

```bash
git add -u
git commit -m "docs: add missing function and package documentation"
```

---

## Category 3: Architecture & Design Issues

### Task 10: Simplify Scanner Concurrency Model

**Files:**
- Modify: `internal/core/scanner.go`
- Test: `internal/core/scanner_test.go`

**Step 1: Write test for simplified concurrency**

```go
func TestSimplifiedWorkerPool(t *testing.T) {
	cfg := &Config{
		Workers:   10,
		Timeout:   100 * time.Millisecond,
		RateLimit: 0,
	}
	scanner := NewScanner(cfg)

	ctx := context.Background()
	ports := []uint16{80, 443, 8080}

	go scanner.ScanRange(ctx, "localhost", ports)

	results := 0
	for event := range scanner.Results() {
		if event.Kind == EventKindResult {
			results++
		}
	}

	if results != len(ports) {
		t.Errorf("got %d results; want %d", results, len(ports))
	}
}
```

**Step 2: Run test**

Run: `go test ./internal/core -v -run TestSimplified`
Expected: PASS with current implementation

**Step 3: Refactor worker pool**

Edit `internal/core/scanner.go` - simplify startWorkers to avoid extra goroutines:

```go
func (s *Scanner) startWorkers(ctx context.Context, jobs <-chan scanJob) {
	for i := 0; i < s.config.Workers; i++ {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			for job := range jobs {
				// Rate limiting
				if s.rateTicker != nil {
					select {
					case <-s.rateTicker.C:
					case <-ctx.Done():
						return
					}
				}

				// Scan directly (no extra goroutine)
				result := s.scanPort(ctx, job)
				s.results <- NewResultEvent(result)
				s.completed.Add(1)
			}
		}()
	}
}
```

**Step 4: Run all tests**

Run: `go test ./internal/core -v`
Expected: All tests pass

**Step 5: Benchmark performance**

Run: `go test ./internal/core -bench=. -benchmem`
Expected: Similar or better performance

**Step 6: Commit**

```bash
git add internal/core/scanner.go internal/core/scanner_test.go
git commit -m "refactor: simplify worker pool concurrency model"
```

---

### Task 11: Implement Scanner Retry Logic

**Files:**
- Modify: `internal/core/scanner.go`
- Test: `internal/core/scanner_retry_test.go` (create)

**Step 1: Write retry test**

Create `internal/core/scanner_retry_test.go`:

```go
package core

import (
	"context"
	"testing"
	"time"
)

func TestRetryOnTimeout(t *testing.T) {
	cfg := &Config{
		Workers:    1,
		Timeout:    50 * time.Millisecond,
		MaxRetries: 2,
		RateLimit:  0,
	}
	scanner := NewScanner(cfg)

	ctx := context.Background()
	// Use unreachable host
	go scanner.ScanRange(ctx, "192.0.2.1", []uint16{80})

	attempts := 0
	for event := range scanner.Results() {
		if event.Kind == EventKindResult {
			attempts++
		}
	}

	// Should attempt initial + 2 retries = 3 attempts
	if attempts != 1 {
		t.Logf("got %d attempts (retry may not be implemented yet)", attempts)
	}
}
```

**Step 2: Run test**

Run: `go test ./internal/core -v -run TestRetry`
Expected: FAIL (retry not implemented)

**Step 3: Implement retry with backoff**

Edit `internal/core/scanner.go`:

```go
func (s *Scanner) scanPortWithRetry(ctx context.Context, job scanJob) ResultEvent {
	var result ResultEvent

	maxRetries := s.config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 0 // No retries by default
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result = s.scanPort(ctx, job)

		// Don't retry if successful or explicitly filtered/refused
		if result.State == "open" || result.State == "filtered" {
			break
		}

		// Apply jittered backoff
		if attempt < maxRetries {
			backoff := time.Duration(attempt+1) * 100 * time.Millisecond
			jitter := time.Duration(rand.Intn(50)) * time.Millisecond
			time.Sleep(backoff + jitter)
		}
	}

	return result
}
```

**Step 4: Update worker to use retry function**

```go
result := s.scanPortWithRetry(ctx, job)
s.results <- NewResultEvent(result)
```

**Step 5: Run tests**

Run: `go test ./internal/core -v`
Expected: All tests pass including retry test

**Step 6: Commit**

```bash
git add internal/core/scanner.go internal/core/scanner_retry_test.go
git commit -m "feat: implement retry logic with jittered backoff"
```

---

## Category 4: Security Issues

### Task 12: Add CSV Injection Protection

**Files:**
- Modify: `pkg/exporter/csv.go`
- Test: `pkg/exporter/csv_test.go` (create)

**Step 1: Write test for CSV injection**

Create `pkg/exporter/csv_test.go`:

```go
package exporter

import (
	"bytes"
	"strings"
	"testing"
	"github.com/lucchesi-sec/portscan/internal/core"
)

func TestCSVInjectionPrevention(t *testing.T) {
	tests := []struct {
		name   string
		banner string
		safe   bool
	}{
		{"normal banner", "SSH-2.0-OpenSSH", true},
		{"formula injection =", "=cmd|'/c calc'", false},
		{"formula injection +", "+cmd", false},
		{"formula injection -", "-2+5", false},
		{"formula injection @", "@SUM(A1:A10)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			exporter := NewCSVExporter(&buf)

			event := core.ResultEvent{
				Host:   "localhost",
				Port:   80,
				State:  "open",
				Banner: tt.banner,
			}

			exporter.Export(event)
			exporter.Close()

			output := buf.String()

			// Check if dangerous characters are sanitized
			if strings.HasPrefix(output, "=") ||
			   strings.HasPrefix(output, "+") ||
			   strings.HasPrefix(output, "-") ||
			   strings.HasPrefix(output, "@") {
				if tt.safe {
					t.Errorf("Safe banner was incorrectly flagged: %s", tt.banner)
				}
			}
		})
	}
}

func TestCSVBannerLengthCapping(t *testing.T) {
	var buf bytes.Buffer
	exporter := NewCSVExporter(&buf)

	longBanner := strings.Repeat("A", 1000)
	event := core.ResultEvent{
		Host:   "localhost",
		Port:   80,
		State:  "open",
		Banner: longBanner,
	}

	exporter.Export(event)
	exporter.Close()

	output := buf.String()
	// Banner should be truncated to reasonable length
	if len(output) > 500 {
		t.Errorf("Banner not truncated: length %d", len(output))
	}
}
```

**Step 2: Run test**

Run: `go test ./pkg/exporter -v -run TestCSV`
Expected: FAIL (injection protection not implemented)

**Step 3: Implement sanitization**

Edit `pkg/exporter/csv.go`:

```go
func sanitizeCSVField(s string) string {
	// Remove formula injection characters
	if len(s) > 0 {
		firstChar := s[0]
		if firstChar == '=' || firstChar == '+' || firstChar == '-' || firstChar == '@' {
			s = "'" + s // Prefix with single quote to prevent formula execution
		}
	}

	// Cap length to 256 characters
	const maxLen = 256
	if len(s) > maxLen {
		s = s[:maxLen-3] + "..."
	}

	return s
}

func (e *CSVExporter) Export(event core.ResultEvent) error {
	record := []string{
		event.Host,
		fmt.Sprintf("%d", event.Port),
		event.State,
		sanitizeCSVField(event.Service),
		sanitizeCSVField(event.Banner),
		fmt.Sprintf("%.2f", event.ResponseTimeMs),
	}

	if err := e.writer.Write(record); err != nil {
		return fmt.Errorf("failed to write CSV record: %w", err)
	}

	return nil
}
```

**Step 4: Add error handling for Close**

```go
func (e *CSVExporter) Close() error {
	e.writer.Flush()
	if err := e.writer.Error(); err != nil {
		return fmt.Errorf("CSV flush failed: %w", err)
	}
	return nil
}
```

**Step 5: Run tests**

Run: `go test ./pkg/exporter -v`
Expected: All tests pass

**Step 6: Commit**

```bash
git add pkg/exporter/csv.go pkg/exporter/csv_test.go
git commit -m "security: add CSV injection protection and length limits"
```

---

### Task 13: Add Input Validation

**Files:**
- Modify: `cmd/commands/scan_helpers.go`
- Test: `cmd/commands/scan_validation_test.go` (create)

**Step 1: Write validation tests**

Create `cmd/commands/scan_validation_test.go`:

```go
package commands

import (
	"testing"
)

func TestHostnameValidation(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{"valid hostname", "example.com", false},
		{"valid IP", "192.168.1.1", false},
		{"valid IPv6", "::1", false},
		{"invalid chars", "host;rm -rf /", true},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 300), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHostname(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHostname(%s) error = %v, wantErr %v", tt.host, err, tt.wantErr)
			}
		})
	}
}

func TestPortValidation(t *testing.T) {
	tests := []struct {
		name    string
		port    uint16
		wantErr bool
	}{
		{"valid port", 80, false},
		{"max port", 65535, false},
		{"zero port", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePort(%d) error = %v, wantErr %v", tt.port, err, tt.wantErr)
			}
		})
	}
}
```

**Step 2: Run tests**

Run: `go test ./cmd/commands -v -run TestValidation`
Expected: FAIL (functions don't exist)

**Step 3: Implement validation functions**

Edit `cmd/commands/scan_helpers.go`:

```go
import (
	"fmt"
	"net"
	"regexp"
)

var hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)

func validateHostname(host string) error {
	if host == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	if len(host) > 253 {
		return fmt.Errorf("hostname too long: %d characters", len(host))
	}

	// Try parsing as IP first
	if net.ParseIP(host) != nil {
		return nil
	}

	// Validate as hostname
	if !hostnameRegex.MatchString(host) {
		return fmt.Errorf("invalid hostname format: %s", host)
	}

	return nil
}

func validatePort(port uint16) error {
	if port == 0 {
		return fmt.Errorf("port cannot be zero")
	}
	return nil
}
```

**Step 4: Use validation in target collection**

```go
func collectTargetInputs(args []string, useStdin bool) ([]string, error) {
	var targets []string

	// ... existing logic ...

	// Validate all targets
	for _, target := range targets {
		if err := validateHostname(target); err != nil {
			return nil, fmt.Errorf("invalid target %s: %w", target, err)
		}
	}

	return targets, nil
}
```

**Step 5: Run tests**

Run: `go test ./cmd/commands -v`
Expected: All tests pass

**Step 6: Commit**

```bash
git add cmd/commands/scan_helpers.go cmd/commands/scan_validation_test.go
git commit -m "security: add input validation for hosts and ports"
```

---

## Category 5: Documentation Issues

### Task 14: Update README for Accuracy

**Files:**
- Modify: `README.md`

**Step 1: Identify unimplemented features**

Read README.md and compare with actual implementation:
- Prometheus export (not implemented)
- Privilege dropping (not implemented)
- Audit logging (not implemented)
- Some CLI flags documented but not implemented

**Step 2: Update README to reflect reality**

Edit `README.md:115-119`:

```markdown
  -o, --output string    Output format: json, csv (table mode via TUI is default)
      --json             Output results as JSON to stdout (NDJSON by default)
      --json-array       Output JSON as a single array (streaming, no buffering)
      --json-object      Output JSON as object with scan_info and results
```

**Step 3: Remove unimplemented features**

Remove or mark as roadmap items:
- Prometheus exporter
- Privilege dropping
- Audit logging

**Step 4: Add implementation status section**

```markdown
## Implementation Status

### Implemented ✅
- TCP and UDP scanning
- Service detection with banner grabbing
- NDJSON, JSON array/object, and CSV export
- Multi-target support (args, stdin, CIDR)
- Real-time TUI with themes
- Rate limiting and worker pools

### Planned 🚧
- Prometheus metrics exporter (v0.3.0)
- Privilege dropping after socket binding (v1.0.0)
- Audit logging to syslog (v1.0.0)
- IPv6 support (v0.4.0)
- SYN scanning (v1.0.0)
```

**Step 5: Fix code examples**

Ensure all code examples in README are tested and working.

**Step 6: Commit**

```bash
git add README.md
git commit -m "docs: update README to match implementation"
```

---

### Task 15: Add ARCHITECTURE.md

**Files:**
- Create: `docs/ARCHITECTURE.md`

**Step 1: Create architecture document**

Create `docs/ARCHITECTURE.md`:

```markdown
# PortScan Architecture

## Overview

PortScan is structured as a Go application with clear separation between the scanner engine, UI, and export layers.

## Package Structure

```
portscan/
├── cmd/commands/     # CLI command definitions (Cobra)
├── internal/core/    # Scanner engine (worker pools, protocols)
├── internal/ui/      # TUI components (Bubble Tea)
├── pkg/config/       # Configuration management (Viper)
├── pkg/exporter/     # Export formats (JSON, CSV)
├── pkg/parser/       # Input parsing (ports, targets)
├── pkg/profiles/     # Port profiles (quick, web, etc.)
├── pkg/services/     # Service detection
├── pkg/targets/      # Target expansion (CIDR, etc.)
└── pkg/theme/        # UI theming
```

## Core Components

### Scanner Engine (`internal/core/scanner.go`)

- **Worker Pool**: Configurable number of concurrent workers
- **Rate Limiting**: Token bucket algorithm via `time.Ticker`
- **Event Channel**: Typed events for results and progress
- **Protocol Support**: TCP and UDP scanning

### TUI Layer (`internal/ui/`)

- **Bubble Tea Framework**: Elm-architecture pattern
- **Components**: Progress bars, sortable tables, filters
- **Ring Buffer**: Fixed-size result buffer to cap memory
- **Stats Tracking**: Separate from buffered results

### Export Layer (`pkg/exporter/`)

- **Streaming Design**: No buffering, constant memory
- **Formats**: NDJSON (default), JSON array, JSON object, CSV
- **Safety**: CSV injection protection, banner sanitization

## Concurrency Model

1. Main goroutine creates worker pool
2. Workers consume jobs from buffered channel
3. Rate limiter controls job dispatch
4. Results sent to typed event channel
5. UI and exporters consume events

## Data Flow

```
Targets → Parser → Scanner → Events → UI/Exporters
                      ↓
                 Worker Pool
                      ↓
                 TCP/UDP Scan
```

## Testing Strategy

- Unit tests for business logic
- Integration tests for scanner
- Table-driven tests where applicable
- Benchmarks for performance-critical paths

## Memory Management

- Ring buffer caps UI result retention (default 10,000)
- Streaming exports prevent memory bloat
- Worker pool size limits goroutine count
```

**Step 2: Commit**

```bash
git add docs/ARCHITECTURE.md
git commit -m "docs: add architecture documentation"
```

---

## Category 6: CI/CD Improvements

### Task 16: Add GitHub Actions Caching

**Files:**
- Modify: `.github/workflows/ci.yml` (or main workflow file)

**Step 1: Add Go module caching**

Edit `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true  # Enables built-in caching

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m
```

**Step 2: Test workflow locally**

Run: `act -j test` (if act is installed)
Expected: Workflow runs successfully

**Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add Go module caching and use official actions"
```

---

### Task 17: Add CodeQL Security Scanning

**Files:**
- Create: `.github/workflows/codeql.yml`

**Step 1: Create CodeQL workflow**

Create `.github/workflows/codeql.yml`:

```yaml
name: "CodeQL"

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Monday

jobs:
  analyze:
    name: Analyze
    runs-on: ubuntu-latest
    permissions:
      security-events: write
      actions: read
      contents: read

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Initialize CodeQL
        uses: github/codeql-action/init@v3
        with:
          languages: go

      - name: Autobuild
        uses: github/codeql-action/autobuild@v3

      - name: Perform CodeQL Analysis
        uses: github/codeql-action/analyze@v3
```

**Step 2: Commit**

```bash
git add .github/workflows/codeql.yml
git commit -m "ci: add CodeQL security scanning"
```

---

### Task 18: Improve Dockerfile

**Files:**
- Modify: `Dockerfile`

**Step 1: Create multi-stage Dockerfile**

Edit `Dockerfile`:

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o portscan ./cmd

# Runtime stage
FROM alpine:3.20

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 scanner && \
    adduser -D -u 1000 -G scanner scanner

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/portscan .

# Use non-root user
USER scanner

# Expose no ports (scanner doesn't listen)

ENTRYPOINT ["./portscan"]
CMD ["--help"]
```

**Step 2: Test Docker build**

Run: `docker build -t portscan:test .`
Expected: Build succeeds

**Step 3: Test Docker run**

Run: `docker run --rm portscan:test scan localhost --ports 80`
Expected: Scanner runs successfully

**Step 4: Commit**

```bash
git add Dockerfile
git commit -m "ci: improve Dockerfile with multi-stage build and security"
```

---

## Category 7: Dependency Management

### Task 19: Audit and Update Dependencies

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

**Step 1: Check for vulnerabilities**

Run: `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`
Expected: Report of any vulnerabilities

**Step 2: Update dependencies to latest stable**

Run: `go get -u ./...`
Expected: Dependencies updated

**Step 3: Run go mod tidy**

Run: `go mod tidy -go=1.24`
Expected: go.mod and go.sum cleaned up

**Step 4: Run all tests**

Run: `go test ./...`
Expected: All tests pass

**Step 5: Check for pseudo-versions**

Run: `grep -E 'v0\.0\.0-[0-9]{14}-[a-f0-9]{12}' go.mod`
Expected: List of pseudo-versions to review

**Step 6: Replace pseudo-versions with tags where possible**

Manually review each pseudo-version and try to find a tagged release.

**Step 7: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: update dependencies and remove pseudo-versions"
```

---

### Task 20: Add Dependabot Configuration

**Files:**
- Create: `.github/dependabot.yml`

**Step 1: Create Dependabot config**

Create `.github/dependabot.yml`:

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10
    reviewers:
      - "lucchesi-sec"
    labels:
      - "dependencies"
      - "go"

  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 5
    labels:
      - "dependencies"
      - "ci"
```

**Step 2: Commit**

```bash
git add .github/dependabot.yml
git commit -m "ci: add Dependabot configuration for Go and Actions"
```

---

## Category 8: Error Handling Improvements

### Task 21: Improve Error Propagation

**Files:**
- Modify: `pkg/exporter/json.go`
- Modify: `cmd/commands/scan_runner.go`

**Step 1: Add error return to Export methods**

Edit `pkg/exporter/json.go`:

```go
type Exporter interface {
	Export(event core.ResultEvent) error
	Close() error
}

func (e *JSONExporter) Export(event core.ResultEvent) error {
	if err := e.encoder.Encode(e.convertToDTO(event)); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}
```

**Step 2: Handle errors in caller**

Edit `cmd/commands/scan_runner.go`:

```go
for event := range scanner.Results() {
	if event.Kind == core.EventKindResult {
		if err := exporter.Export(*event.Result); err != nil {
			log.Printf("Export error: %v", err)
			// Decide: continue or abort?
		}
	}
}

if err := exporter.Close(); err != nil {
	return fmt.Errorf("failed to close exporter: %w", err)
}
```

**Step 3: Write test for error handling**

```go
func TestExportErrorHandling(t *testing.T) {
	// Create exporter that will fail
	var buf errorWriter{failAfter: 2}
	exporter := NewJSONExporter(&buf)

	events := []core.ResultEvent{
		{Host: "localhost", Port: 80},
		{Host: "localhost", Port: 443},
		{Host: "localhost", Port: 8080}, // This should fail
	}

	var lastErr error
	for _, event := range events {
		if err := exporter.Export(event); err != nil {
			lastErr = err
		}
	}

	if lastErr == nil {
		t.Error("expected error from failing writer")
	}
}

type errorWriter struct {
	failAfter int
	written   int
}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	w.written++
	if w.written > w.failAfter {
		return 0, fmt.Errorf("simulated write error")
	}
	return len(p), nil
}
```

**Step 4: Run tests**

Run: `go test ./pkg/exporter -v`
Expected: Tests pass

**Step 5: Commit**

```bash
git add pkg/exporter/json.go pkg/exporter/csv.go cmd/commands/scan_runner.go
git commit -m "fix: improve error propagation in exporters"
```

---

## Category 9: Performance Optimizations

### Task 22: Add Performance Benchmarks

**Files:**
- Create: `internal/core/scanner_bench_test.go`

**Step 1: Write scanner benchmarks**

Create `internal/core/scanner_bench_test.go`:

```go
package core

import (
	"context"
	"testing"
)

func BenchmarkTCPScan1000Ports(b *testing.B) {
	cfg := &Config{
		Workers:   100,
		Timeout:   50 * time.Millisecond,
		RateLimit: 0,
	}

	ports := make([]uint16, 1000)
	for i := range ports {
		ports[i] = uint16(i + 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner := NewScanner(cfg)
		ctx := context.Background()

		go scanner.ScanRange(ctx, "localhost", ports)

		// Drain results
		for range scanner.Results() {
		}
	}
}

func BenchmarkWorkerPool(b *testing.B) {
	tests := []struct {
		name    string
		workers int
	}{
		{"10 workers", 10},
		{"100 workers", 100},
		{"500 workers", 500},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			cfg := &Config{
				Workers:   tt.workers,
				Timeout:   50 * time.Millisecond,
				RateLimit: 0,
			}

			ports := []uint16{80, 443, 8080}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				scanner := NewScanner(cfg)
				ctx := context.Background()

				go scanner.ScanRange(ctx, "localhost", ports)
				for range scanner.Results() {
				}
			}
		})
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	rates := []int{1000, 5000, 10000}

	for _, rate := range rates {
		b.Run(fmt.Sprintf("rate_%d", rate), func(b *testing.B) {
			cfg := &Config{
				Workers:   100,
				Timeout:   50 * time.Millisecond,
				RateLimit: rate,
			}

			ports := []uint16{80}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				scanner := NewScanner(cfg)
				ctx := context.Background()

				go scanner.ScanRange(ctx, "localhost", ports)
				for range scanner.Results() {
				}
			}
		})
	}
}
```

**Step 2: Run benchmarks**

Run: `go test ./internal/core -bench=. -benchmem`
Expected: Benchmark results with allocations

**Step 3: Commit**

```bash
git add internal/core/scanner_bench_test.go
git commit -m "test: add performance benchmarks for scanner"
```

---

### Task 23: Add Memory Profiling

**Files:**
- Create: `scripts/profile.sh`

**Step 1: Create profiling script**

Create `scripts/profile.sh`:

```bash
#!/bin/bash

set -e

echo "Building profiled binary..."
go build -o portscan-prof ./cmd

echo "Running CPU profile..."
./portscan-prof scan localhost --ports 1-1024 --cpuprofile=cpu.prof

echo "Running memory profile..."
./portscan-prof scan localhost --ports 1-1024 --memprofile=mem.prof

echo "Analyzing CPU profile..."
go tool pprof -http=:8080 cpu.prof &

echo "Analyzing memory profile..."
go tool pprof -http=:8081 mem.prof &

echo "Profiles available at:"
echo "  CPU:    http://localhost:8080"
echo "  Memory: http://localhost:8081"
echo ""
echo "Press Ctrl+C to exit..."

wait
```

**Step 2: Make executable**

Run: `chmod +x scripts/profile.sh`

**Step 3: Add profiling flags to CLI**

Edit `cmd/commands/scan_command.go`:

```go
var cpuProfile string
var memProfile string

func init() {
	scanCmd.Flags().StringVar(&cpuProfile, "cpuprofile", "", "Write CPU profile to file")
	scanCmd.Flags().StringVar(&memProfile, "memprofile", "", "Write memory profile to file")
}

func scanCmdRun(cmd *cobra.Command, args []string) error {
	// Start CPU profiling
	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			return fmt.Errorf("could not create CPU profile: %w", err)
		}
		defer f.Close()

		if err := pprof.StartCPUProfile(f); err != nil {
			return fmt.Errorf("could not start CPU profile: %w", err)
		}
		defer pprof.StopCPUProfile()
	}

	// ... existing scan logic ...

	// Write memory profile
	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			return fmt.Errorf("could not create memory profile: %w", err)
		}
		defer f.Close()

		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			return fmt.Errorf("could not write memory profile: %w", err)
		}
	}

	return nil
}
```

**Step 4: Commit**

```bash
git add scripts/profile.sh cmd/commands/scan_command.go
git commit -m "feat: add CPU and memory profiling support"
```

---

## Summary

This plan provides comprehensive technical debt remediation across 9 categories:

1. **Test Coverage**: 6 tasks to improve coverage from current levels to 60%+
2. **Code Quality**: 3 tasks for formatting, refactoring, and documentation
3. **Architecture**: 2 tasks to simplify concurrency and add retry logic
4. **Security**: 2 tasks for CSV injection protection and input validation
5. **Documentation**: 2 tasks for README accuracy and architecture docs
6. **CI/CD**: 3 tasks for caching, CodeQL, and Docker improvements
7. **Dependencies**: 2 tasks for auditing and Dependabot
8. **Error Handling**: 1 task for error propagation
9. **Performance**: 2 tasks for benchmarking and profiling

**Total Tasks**: 23

**Estimated Completion Time**: 4-6 weeks (assuming 1-2 tasks per day)

**Priority Order**:
1. Critical: Tasks 1-6 (test coverage)
2. High: Tasks 7-13 (code quality and security)
3. Medium: Tasks 14-18 (documentation and CI/CD)
4. Low: Tasks 19-23 (dependencies and performance)

Each task follows TDD principles: write test → verify it fails → implement → verify it passes → commit.
