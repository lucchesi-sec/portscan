# Developer Guide

## Table of Contents

1. [Getting Started](#getting-started)
2. [Development Environment](#development-environment)
3. [Code Structure](#code-structure)
4. [Adding New Features](#adding-new-features)
5. [Testing Guidelines](#testing-guidelines)
6. [UI Development](#ui-development)
7. [Performance Optimization](#performance-optimization)
8. [Debugging](#debugging)
9. [Common Tasks](#common-tasks)

---

## 1. Getting Started

### Prerequisites

- **Go 1.24+**: Download from [golang.org](https://golang.org/dl/)
- **Git**: Version control
- **Make**: Build automation (pre-installed on macOS/Linux)
- **golangci-lint**: Linting (installed via `make dev-setup`)

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/lucchesi-sec/portscan.git
cd portscan

# Install development tools
make dev-setup

# Install dependencies
go mod download

# Verify installation
make test
make lint
```

### Quick Build and Run

```bash
# Build the binary
make build

# Run locally
./bin/portscan scan localhost --ports 22,80,443

# Run without building
go run cmd/main.go scan localhost --ports 22,80,443
```

---

## 2. Development Environment

### Recommended Tools

**IDE/Editor:**
- **VS Code** with Go extension (`golang.go`)
- **GoLand** by JetBrains
- **Vim/Neovim** with `vim-go` or `coc-go`

**Essential VS Code Extensions:**
```json
{
  "recommendations": [
    "golang.go",
    "eamodio.gitlens",
    "streetsidesoftware.code-spell-checker"
  ]
}
```

**VS Code Settings** (`.vscode/settings.json`):
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "gofumpt",
  "editor.formatOnSave": true,
  "go.testFlags": ["-v", "-race"],
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

### Makefile Targets

```bash
# Development workflow
make dev          # Start development server (with auto-reload)
make quick        # Fast build and test
make lint         # Run linters
make test         # Run tests
make test-race    # Run tests with race detector
make test-coverage # Generate coverage report

# Build targets
make build        # Build for current platform
make build-all    # Build for all platforms (using goreleaser)
make clean        # Remove build artifacts

# Code quality
make fmt          # Format code
make vet          # Run go vet
make staticcheck  # Run staticcheck

# Documentation
make docs         # Generate documentation
make godoc        # Start local godoc server

# CI simulation
make ci           # Run full CI pipeline locally
```

---

## 3. Code Structure

### Package Organization Principles

**`internal/` vs `pkg/`:**

- **`internal/`**: Implementation details, scanner engine, UI components
  - Cannot be imported by external projects
  - Free to change without breaking external dependencies
  
- **`pkg/`**: Public API, reusable libraries
  - Can be imported by other projects
  - Must maintain backward compatibility
  - Well-documented with godoc

### Key Packages

**Scanner Core** (`internal/core/`):
```
core/
├── scanner.go          # TCP scanner implementation
├── udp_scanner.go      # UDP scanner implementation
├── udp_probes.go       # Service-specific UDP probes
├── udp_runner.go       # UDP orchestration
├── udp_response.go     # Response parsing
├── scan_types.go       # Core data types
└── *_test.go           # Comprehensive tests
```

**User Interface** (`internal/ui/`):
```
ui/
├── scan_ui_model.go    # UI state (Model in MVC)
├── scan_ui_update.go   # Event handlers (Controller)
├── scan_ui_view.go     # Rendering (View)
├── progress.go         # Progress tracking
├── stats.go            # Statistics
├── filters.go          # Result filtering
├── sorters.go          # Result sorting
└── components/         # Reusable widgets
    └── splitview.go
```

**Public Libraries** (`pkg/`):
```
pkg/
├── config/      # Configuration management
├── parser/      # Port specification parsing
├── targets/     # Target resolution (CIDR, etc.)
├── profiles/    # Scan profiles
├── services/    # Port-to-service mapping
├── theme/       # UI themes
├── exporter/    # Output formatters
└── errors/      # User-friendly errors
```

### Import Guidelines

**Internal Package Imports:**
```go
package commands

import (
    // Standard library first
    "context"
    "fmt"
    "os"
    
    // Third-party packages (alphabetically)
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    
    // Local packages (grouped by purpose)
    "github.com/lucchesi-sec/portscan/internal/core"
    "github.com/lucchesi-sec/portscan/internal/ui"
    
    "github.com/lucchesi-sec/portscan/pkg/config"
    "github.com/lucchesi-sec/portscan/pkg/parser"
    "github.com/lucchesi-sec/portscan/pkg/targets"
)
```

---

## 4. Adding New Features

### Example: Adding a New Scanner Type

Let's add SCTP (Stream Control Transmission Protocol) scanning.

#### Step 1: Define the Interface

```go
// internal/core/sctp_scanner.go
package core

import (
    "context"
    "net"
    "time"
)

// SCTPScanner implements SCTP protocol scanning.
type SCTPScanner struct {
    *Scanner  // Embed base scanner
}

// NewSCTPScanner creates a new SCTP scanner instance.
func NewSCTPScanner(cfg *Config) *SCTPScanner {
    return &SCTPScanner{
        Scanner: NewScanner(cfg),
    }
}

// ScanRange implements the Scanner interface for SCTP scanning.
func (s *SCTPScanner) ScanRange(ctx context.Context, host string, ports []uint16) {
    s.ScanTargets(ctx, []ScanTarget{{Host: host, Ports: ports}})
}

// worker performs SCTP scanning of individual ports.
func (s *SCTPScanner) sctpWorker(ctx context.Context, jobs <-chan scanJob) {
    defer s.wg.Done()
    
    for {
        select {
        case <-ctx.Done():
            return
        case job, ok := <-jobs:
            if !ok {
                return
            }
            
            // Rate limiting
            if !s.waitForRate(ctx) {
                return
            }
            
            s.scanSCTPPort(ctx, job.host, job.port)
        }
    }
}

func (s *SCTPScanner) scanSCTPPort(ctx context.Context, host string, port uint16) {
    start := time.Now()
    
    // SCTP-specific connection logic
    // Note: Go's net package doesn't directly support SCTP,
    // would need external library or raw sockets
    
    result := ResultEvent{
        Host:     host,
        Port:     port,
        Protocol: "sctp",
        Duration: time.Since(start),
        State:    StateOpen, // or StateClosed/StateFiltered
    }
    
    s.emitResult(ctx, result)
}
```

#### Step 2: Add Tests

```go
// internal/core/sctp_scanner_test.go
package core

import (
    "context"
    "testing"
    "time"
)

func TestSCTPScanner_Basic(t *testing.T) {
    cfg := &Config{
        Workers:   10,
        Timeout:   200 * time.Millisecond,
        RateLimit: 1000,
    }
    
    scanner := NewSCTPScanner(cfg)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    go scanner.ScanRange(ctx, "127.0.0.1", []uint16{9899})
    
    results := make([]ResultEvent, 0)
    for event := range scanner.Results() {
        if event.Kind == EventKindResult {
            results = append(results, *event.Result)
        }
    }
    
    if len(results) == 0 {
        t.Error("Expected at least one result")
    }
}
```

#### Step 3: Integrate into CLI

```go
// cmd/commands/scan_runner.go
func createScanner(protocol string, cfg *core.Config) core.ProtocolScanner {
    switch protocol {
    case "tcp":
        return core.NewScanner(cfg)
    case "udp":
        return core.NewUDPScanner(cfg)
    case "sctp":
        return core.NewSCTPScanner(cfg)
    case "both":
        // Handle multi-protocol
        return core.NewMultiProtocolScanner(cfg, []string{"tcp", "udp"})
    default:
        return nil
    }
}
```

#### Step 4: Update CLI Flags

```go
// cmd/commands/scan_command.go
func init() {
    // Update protocol flag
    scanCmd.Flags().StringP("protocol", "u", "tcp", 
        "protocol to scan: tcp (default), udp, sctp, or both")
    
    // Update validation
    _ = viper.BindPFlag("protocol", scanCmd.Flags().Lookup("protocol"))
}
```

#### Step 5: Documentation

Update relevant documentation:
- `README.md`: Add SCTP examples
- `docs/ARCHITECTURE.md`: Document SCTP scanner
- `pkg/config/config.go`: Update validation to include "sctp"

---

### Example: Adding a New Export Format

Let's add XML export support.

#### Step 1: Implement Exporter

```go
// pkg/exporter/xml.go
package exporter

import (
    "encoding/xml"
    "io"
    
    "github.com/lucchesi-sec/portscan/internal/core"
)

// XMLExporter exports scan results in XML format.
type XMLExporter struct {
    writer  io.Writer
    encoder *xml.Encoder
    root    bool
}

// NewXMLExporter creates a new XML exporter.
func NewXMLExporter(w io.Writer) *XMLExporter {
    encoder := xml.NewEncoder(w)
    encoder.Indent("", "  ")
    
    return &XMLExporter{
        writer:  w,
        encoder: encoder,
        root:    false,
    }
}

// Write outputs a single result as XML.
func (e *XMLExporter) Write(result core.ResultEvent) error {
    if !e.root {
        // Write XML header and root element
        io.WriteString(e.writer, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
        io.WriteString(e.writer, "<scan_results>\n")
        e.root = true
    }
    
    type XMLResult struct {
        XMLName  xml.Name `xml:"result"`
        Host     string   `xml:"host"`
        Port     uint16   `xml:"port"`
        Protocol string   `xml:"protocol"`
        State    string   `xml:"state"`
        Banner   string   `xml:"banner,omitempty"`
        Latency  float64  `xml:"latency_ms"`
    }
    
    xmlResult := XMLResult{
        Host:     result.Host,
        Port:     result.Port,
        Protocol: result.Protocol,
        State:    string(result.State),
        Banner:   result.Banner,
        Latency:  result.Duration.Seconds() * 1000,
    }
    
    return e.encoder.Encode(xmlResult)
}

// Close finalizes the XML document.
func (e *XMLExporter) Close() error {
    if e.root {
        io.WriteString(e.writer, "</scan_results>\n")
    }
    return nil
}
```

#### Step 2: Add Tests

```go
// pkg/exporter/xml_test.go
package exporter

import (
    "bytes"
    "strings"
    "testing"
    "time"
    
    "github.com/lucchesi-sec/portscan/internal/core"
)

func TestXMLExporter(t *testing.T) {
    var buf bytes.Buffer
    exporter := NewXMLExporter(&buf)
    defer exporter.Close()
    
    result := core.ResultEvent{
        Host:     "192.168.1.1",
        Port:     80,
        Protocol: "tcp",
        State:    core.StateOpen,
        Banner:   "Apache/2.4.41",
        Duration: 5 * time.Millisecond,
    }
    
    if err := exporter.Write(result); err != nil {
        t.Fatalf("Write failed: %v", err)
    }
    
    exporter.Close()
    
    output := buf.String()
    if !strings.Contains(output, "<host>192.168.1.1</host>") {
        t.Error("Missing host in XML output")
    }
    if !strings.Contains(output, "<state>open</state>") {
        t.Error("Missing state in XML output")
    }
}
```

#### Step 3: Integrate

```go
// cmd/commands/scan_runner.go
func createExporter(format string, w io.Writer) (exporter.Exporter, error) {
    switch format {
    case "json":
        return exporter.NewJSONExporter(w, exporter.JSONModeNDJSON, nil), nil
    case "csv":
        return exporter.NewCSVExporter(w), nil
    case "xml":
        return exporter.NewXMLExporter(w), nil
    default:
        return nil, fmt.Errorf("unknown export format: %s", format)
    }
}
```

---

## 5. Testing Guidelines

### Test Structure

**Follow the Table-Driven Test Pattern:**

```go
func TestParsePortts(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected []uint16
        wantErr  bool
    }{
        {
            name:     "single port",
            input:    "80",
            expected: []uint16{80},
            wantErr:  false,
        },
        {
            name:     "port range",
            input:    "80-83",
            expected: []uint16{80, 81, 82, 83},
            wantErr:  false,
        },
        {
            name:     "mixed format",
            input:    "22,80,443,8000-8003",
            expected: []uint16{22, 80, 443, 8000, 8001, 8002, 8003},
            wantErr:  false,
        },
        {
            name:     "invalid port",
            input:    "70000",
            expected: nil,
            wantErr:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParsePorts(tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("ParsePorts() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
                t.Errorf("ParsePorts() = %v, want %v", got, tt.expected)
            }
        })
    }
}
```

### Test Categories

**1. Unit Tests**

Test individual functions in isolation:

```go
func TestRetryBackoff(t *testing.T) {
    scanner := NewScanner(&Config{Timeout: 200 * time.Millisecond})
    
    // Test exponential backoff
    attempt0 := scanner.retryBackoff(0)
    attempt1 := scanner.retryBackoff(1)
    attempt2 := scanner.retryBackoff(2)
    
    if attempt1 <= attempt0 {
        t.Error("Backoff should increase with attempts")
    }
    if attempt2 <= attempt1 {
        t.Error("Backoff should continue increasing")
    }
}
```

**2. Integration Tests**

Test component interactions:

```go
func TestScannerIntegration(t *testing.T) {
    cfg := &Config{
        Workers:   10,
        Timeout:   500 * time.Millisecond,
        RateLimit: 1000,
    }
    
    scanner := NewScanner(cfg)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Start scan
    go scanner.ScanRange(ctx, "127.0.0.1", []uint16{22, 80, 443})
    
    // Collect results
    resultCount := 0
    progressCount := 0
    
    for event := range scanner.Results() {
        switch event.Kind {
        case EventKindResult:
            resultCount++
        case EventKindProgress:
            progressCount++
        }
    }
    
    if resultCount != 3 {
        t.Errorf("Expected 3 results, got %d", resultCount)
    }
    if progressCount == 0 {
        t.Error("Expected progress events")
    }
}
```

**3. Benchmark Tests**

Measure performance:

```go
func BenchmarkPortParsing(b *testing.B) {
    input := "1-1024,3306,5432,6379,8080-9000"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = ParsePorts(input)
    }
}

func BenchmarkScanWorkerPool(b *testing.B) {
    cfg := &Config{
        Workers:   100,
        Timeout:   100 * time.Millisecond,
        RateLimit: 10000,
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        scanner := NewScanner(cfg)
        ctx := context.Background()
        
        go scanner.ScanRange(ctx, "127.0.0.1", []uint16{80})
        
        for range scanner.Results() {
            // Drain results
        }
    }
}
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/core

# Specific test
go test ./internal/core -run TestScannerBasic

# With coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# With race detector
go test -race ./...

# Verbose output
go test -v ./...

# Benchmarks
go test -bench=. ./internal/core
go test -bench=BenchmarkPortParsing -benchmem ./pkg/parser
```

### Test Best Practices

1. **Use descriptive test names**: `TestParsePortsHandlesInvalidInput`
2. **Test edge cases**: Empty input, max values, boundary conditions
3. **Mock external dependencies**: Network calls, file I/O
4. **Use `testing.T.Helper()`** for test utilities
5. **Clean up resources**: Use `defer` and `t.Cleanup()`
6. **Avoid flaky tests**: Use deterministic inputs, avoid time-based tests

---

## 6. UI Development

### Bubble Tea Architecture

The UI follows the **Elm Architecture** (Model-View-Update):

```
┌──────────────────────────────────────┐
│            User Input                │
│          (Keyboard/Mouse)            │
└──────────────┬───────────────────────┘
               │
               ▼
┌──────────────────────────────────────┐
│     Update(msg tea.Msg) tea.Cmd      │
│   ┌──────────────────────────────┐   │
│   │  Handle key presses          │   │
│   │  Process scan results        │   │
│   │  Update progress             │   │
│   │  Apply filters/sorts         │   │
│   └──────────────────────────────┘   │
└──────────────┬───────────────────────┘
               │
               ▼
┌──────────────────────────────────────┐
│              Model                   │
│   ┌──────────────────────────────┐   │
│   │  UI state                    │   │
│   │  Result buffer               │   │
│   │  Statistics                  │   │
│   │  Components (table, progress)│   │
│   └──────────────────────────────┘   │
└──────────────┬───────────────────────┘
               │
               ▼
┌──────────────────────────────────────┐
│       View() string                  │
│   ┌──────────────────────────────┐   │
│   │  Render header               │   │
│   │  Render progress bar         │   │
│   │  Render results table        │   │
│   │  Render footer/help          │   │
│   └──────────────────────────────┘   │
└──────────────┬───────────────────────┘
               │
               ▼
┌──────────────────────────────────────┐
│          Terminal Output             │
└──────────────────────────────────────┘
```

### Adding UI Components

**Example: Adding a Chart Component**

```go
// internal/ui/components/chart.go
package components

import (
    "strings"
    
    "github.com/charmbracelet/lipgloss"
)

// BarChart renders a simple horizontal bar chart.
type BarChart struct {
    data   map[string]int
    width  int
    height int
    style  lipgloss.Style
}

// NewBarChart creates a new bar chart.
func NewBarChart(width, height int) *BarChart {
    return &BarChart{
        data:   make(map[string]int),
        width:  width,
        height: height,
        style:  lipgloss.NewStyle(),
    }
}

// SetData updates the chart data.
func (c *BarChart) SetData(data map[string]int) {
    c.data = data
}

// Render returns the chart as a string.
func (c *BarChart) Render() string {
    if len(c.data) == 0 {
        return "No data"
    }
    
    // Find max value for scaling
    max := 0
    for _, v := range c.data {
        if v > max {
            max = v
        }
    }
    
    var rows []string
    for label, value := range c.data {
        barLength := int(float64(value) / float64(max) * float64(c.width-20))
        bar := strings.Repeat("█", barLength)
        row := fmt.Sprintf("%-15s %5d %s", label, value, bar)
        rows = append(rows, row)
    }
    
    return strings.Join(rows, "\n")
}
```

### Integrating Components

```go
// internal/ui/scan_ui_model.go
type ScanUI struct {
    // ... existing fields
    chart *components.BarChart
}

func NewScanUI(...) *ScanUI {
    // ... existing initialization
    
    chart := components.NewBarChart(40, 10)
    
    return &ScanUI{
        // ... existing fields
        chart: chart,
    }
}

// internal/ui/scan_ui_view.go
func (m *ScanUI) View() string {
    // ... existing rendering
    
    // Add chart to view
    chartData := map[string]int{
        "Open":     m.stats.open,
        "Closed":   m.stats.closed,
        "Filtered": m.stats.filtered,
    }
    m.chart.SetData(chartData)
    chartView := m.chart.Render()
    
    // Combine with other sections
    return lipgloss.JoinVertical(
        lipgloss.Left,
        header,
        progress,
        table,
        chartView,  // New chart section
        footer,
    )
}
```

---

## 7. Performance Optimization

### Profiling

**CPU Profiling:**

```bash
# Generate CPU profile
go test -cpuprofile=cpu.prof -bench=. ./internal/core

# Analyze profile
go tool pprof cpu.prof
(pprof) top
(pprof) list ScanRange
(pprof) web  # Open visualization in browser
```

**Memory Profiling:**

```bash
# Generate memory profile
go test -memprofile=mem.prof -bench=. ./internal/core

# Analyze profile
go tool pprof -alloc_space mem.prof
(pprof) top
(pprof) list ResultBuffer
```

**Tracing:**

```bash
# Generate trace
go test -trace=trace.out -bench=BenchmarkScannerIntegration ./internal/core

# View trace
go tool trace trace.out
```

### Optimization Techniques

**1. Pre-allocate Slices**

```go
// Bad: Growing slice repeatedly
results := []ResultEvent{}
for result := range ch {
    results = append(results, result)  // Multiple allocations
}

// Good: Pre-allocate capacity
results := make([]ResultEvent, 0, expectedCount)
for result := range ch {
    results = append(results, result)  // Single allocation
}
```

**2. Use sync.Pool for Frequent Allocations**

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

func scanUDPPort(...) {
    buffer := bufferPool.Get().([]byte)
    defer bufferPool.Put(buffer)
    
    // Use buffer
    n, err := conn.Read(buffer)
    // ...
}
```

**3. Avoid String Concatenation in Loops**

```go
// Bad: Multiple allocations
var result string
for _, line := range lines {
    result += line + "\n"
}

// Good: Use strings.Builder
var builder strings.Builder
for _, line := range lines {
    builder.WriteString(line)
    builder.WriteByte('\n')
}
result := builder.String()
```

**4. Use Buffered Channels Appropriately**

```go
// Balance between memory and throughput
jobs := make(chan scanJob, 1000)  // Buffer size = expected burst

// Monitor channel usage
select {
case jobs <- job:
    // Sent successfully
case <-time.After(100 * time.Millisecond):
    // Channel full, log warning
    log.Warn("Job queue is full")
}
```

---

## 8. Debugging

### Debug Logging

Add debug output without modifying production code:

```go
// Set environment variable
export PORTSCAN_DEBUG=1

// In code
if os.Getenv("PORTSCAN_DEBUG") != "" {
    log.Printf("DEBUG: Scanning host=%s port=%d\n", host, port)
}
```

### Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug tests
dlv test ./internal/core -- -test.run TestScannerBasic

# Debug application
dlv exec ./bin/portscan -- scan localhost

# Set breakpoints
(dlv) break scanner.go:125
(dlv) break ScanRange
(dlv) continue

# Inspect variables
(dlv) print scanner.config
(dlv) print result

# Stack trace
(dlv) stack
```

### Common Issues

**Issue: Race Conditions**

```bash
# Detect races
go test -race ./...

# Common fixes:
# 1. Use mutex for shared state
type Stats struct {
    mu    sync.Mutex
    count int
}

func (s *Stats) Increment() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.count++
}

# 2. Use atomic operations
var count atomic.Int64
count.Add(1)

# 3. Use channels for communication
results := make(chan ResultEvent, 100)
```

**Issue: Goroutine Leaks**

```bash
# Detect leaks
go test -parallel 1 ./...

# Check goroutine count
runtime.NumGoroutine()

# Ensure proper cleanup:
ctx, cancel := context.WithCancel(context.Background())
defer cancel()  # Always cancel context

go func() {
    for {
        select {
        case <-ctx.Done():
            return  # Exit goroutine
        case work := <-jobs:
            // Process work
        }
    }
}()
```

---

## 9. Common Tasks

### Adding a New Profile

```go
// pkg/profiles/profiles.go
var profiles = map[string][]uint16{
    // ... existing profiles
    
    "iot": {
        // IoT device ports
        80, 443,     // Web interfaces
        1883, 8883,  // MQTT
        5683,        // CoAP
        8080, 8081,  // Alternative HTTP
        9090,        // Prometheus metrics
    },
}
```

### Adding Service Detection

```go
// pkg/services/services.go
var tcpServices = map[uint16]string{
    // ... existing services
    
    9200: "elasticsearch",
    9300: "elasticsearch-transport",
    15672: "rabbitmq-management",
}

var udpServices = map[uint16]string{
    // ... existing services
    
    5683: "coap",
    47808: "bacnet",
}
```

### Adding a UDP Probe

```go
// internal/core/udp_probes.go
func initUDPProbes() map[uint16][]byte {
    return map[uint16][]byte{
        // ... existing probes
        
        // CoAP GET request
        5683: {
            0x40, 0x01,       // Version + Type + GET
            0x00, 0x01,       // Message ID
            0xB1, 0x2E,       // Uri-Path option
        },
    }
}
```

### Extending Configuration

```go
// pkg/config/config.go
type Config struct {
    // ... existing fields
    
    // New field
    MaxHostsPerScan int `mapstructure:"max_hosts_per_scan" validate:"min=1,max=100000"`
}

// Set default
viper.SetDefault("max_hosts_per_scan", 65536)
```

---

## Summary

This guide covers:
✅ Setting up your development environment  
✅ Understanding the code structure  
✅ Adding new features (scanners, exporters)  
✅ Writing comprehensive tests  
✅ Developing UI components  
✅ Optimizing performance  
✅ Debugging techniques  
✅ Common development tasks  

**Next Steps:**
1. Read [ARCHITECTURE.md](ARCHITECTURE.md) for system design
2. Review [CONTRIBUTING.md](../CONTRIBUTING.md) for contribution guidelines
3. Check [MAINTENANCE.md](MAINTENANCE.md) for operational procedures

**Questions?**
- Open a [GitHub Discussion](https://github.com/lucchesi-sec/portscan/discussions)
- Check existing [Issues](https://github.com/lucchesi-sec/portscan/issues)
- Review the [godoc documentation](https://pkg.go.dev/github.com/lucchesi-sec/portscan)
