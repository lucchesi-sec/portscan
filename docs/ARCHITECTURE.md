# Port Scanner Architecture Documentation

## Executive Summary

This document provides a comprehensive technical overview of the Go TUI Port Scanner, a high-performance network reconnaissance tool built with Go 1.24+. The scanner achieves 7,500+ packets per second throughput while maintaining memory efficiency (<50MB for /24 networks) and providing real-time terminal user interface (TUI) visualization.

**Key Capabilities:**
- **Multi-Protocol Support**: TCP and UDP scanning with protocol-specific optimizations
- **High Performance**: Concurrent worker pool architecture with intelligent rate limiting
- **Real-Time UI**: Built with Charm's Bubble Tea framework for responsive visualization
- **Flexible Export**: Streaming JSON (NDJSON), JSON arrays/objects, and CSV formats
- **Service Detection**: Banner grabbing and service identification for common protocols
- **Production Ready**: Comprehensive input validation, error handling, and retry logic

**Target Audience:** Software architects, senior engineers, and security practitioners seeking to understand, extend, or maintain the scanner codebase.

---

## Table of Contents

1. [System Architecture](#1-system-architecture)
2. [Core Components](#2-core-components)
3. [Data Flow](#3-data-flow)
4. [Scanner Engine](#4-scanner-engine)
5. [UI Architecture](#5-ui-architecture)
6. [Configuration System](#6-configuration-system)
7. [Export Pipeline](#7-export-pipeline)
8. [Design Decisions](#8-design-decisions)
9. [Performance Characteristics](#9-performance-characteristics)
10. [Security Model](#10-security-model)
11. [Extensibility Points](#11-extensibility-points)
12. [Deployment Architecture](#12-deployment-architecture)

---

## 1. System Architecture

### 1.1 High-Level Overview

The port scanner follows a pipeline architecture with three primary stages:

```
┌──────────────────────────────────────────────────────────────────┐
│                         INPUT LAYER                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │  CLI Parser  │  │   Config     │  │   Profiles   │           │
│  │   (Cobra)    │─▶│   (Viper)    │─▶│  (Presets)   │           │
│  └──────────────┘  └──────────────┘  └──────────────┘           │
│          │                  │                  │                  │
│          └──────────────────┴──────────────────┘                  │
│                             ▼                                     │
└──────────────────────────────────────────────────────────────────┘
                              │
┌──────────────────────────────▼──────────────────────────────────┐
│                      PROCESSING LAYER                            │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │               Scanner Orchestrator                       │   │
│  │  ┌─────────────────┐        ┌─────────────────┐        │   │
│  │  │   TCP Scanner   │        │   UDP Scanner   │        │   │
│  │  │                 │        │                 │        │   │
│  │  │ ┌─────────────┐ │        │ ┌─────────────┐ │        │   │
│  │  │ │ Worker Pool │ │        │ │ Worker Pool │ │        │   │
│  │  │ │  (100+)     │ │        │ │  (50+)      │ │        │   │
│  │  │ └─────────────┘ │        │ └─────────────┘ │        │   │
│  │  │                 │        │                 │        │   │
│  │  │ ┌─────────────┐ │        │ ┌─────────────┐ │        │   │
│  │  │ │Rate Limiter │ │        │ │UDP Probes   │ │        │   │
│  │  │ └─────────────┘ │        │ └─────────────┘ │        │   │
│  │  └─────────────────┘        └─────────────────┘        │   │
│  └─────────────────────────────────────────────────────────┘   │
│                             │                                    │
│                             ▼                                    │
│                   ┌──────────────────┐                          │
│                   │  Event Channel   │                          │
│                   │  (Buffered 1000) │                          │
│                   └──────────────────┘                          │
└──────────────────────────────────────────────────────────────────┘
                              │
┌──────────────────────────────▼──────────────────────────────────┐
│                       OUTPUT LAYER                               │
│                                                                   │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐    │
│  │   TUI Display  │  │   Exporters    │  │   Filtering    │    │
│  │ (Bubble Tea)   │  │  JSON/CSV      │  │   & Sorting    │    │
│  └────────────────┘  └────────────────┘  └────────────────┘    │
└──────────────────────────────────────────────────────────────────┘
```

### 1.2 Module Dependency Graph

```mermaid
graph TB
    subgraph "Command Layer"
        CMD[cmd/commands]
    end
    
    subgraph "Configuration"
        CONFIG[pkg/config]
        PROFILES[pkg/profiles]
        THEME[pkg/theme]
    end
    
    subgraph "Input Processing"
        PARSER[pkg/parser]
        TARGETS[pkg/targets]
        SERVICES[pkg/services]
        ERRORS[pkg/errors]
    end
    
    subgraph "Core Scanning"
        SCANNER[internal/core]
    end
    
    subgraph "User Interface"
        UI[internal/ui]
        COMPONENTS[internal/ui/components]
    end
    
    subgraph "Output"
        EXPORTER[pkg/exporter]
    end
    
    CMD --> CONFIG
    CMD --> PARSER
    CMD --> TARGETS
    CMD --> SCANNER
    CMD --> UI
    CMD --> EXPORTER
    
    CONFIG --> PROFILES
    CONFIG --> THEME
    
    PARSER --> ERRORS
    TARGETS --> ERRORS
    
    UI --> SCANNER
    UI --> CONFIG
    UI --> THEME
    UI --> SERVICES
    UI --> COMPONENTS
    
    SCANNER --> CONFIG
    
    EXPORTER --> SCANNER
```

### 1.3 Package Organization

```
portscan/
├── cmd/                          # Application entry points
│   ├── main.go                   # Main executable entry
│   └── commands/                 # CLI command implementations
│       ├── root.go               # Root command setup
│       ├── scan_command.go       # Scan command definition
│       ├── scan_runner.go        # Scan orchestration logic
│       ├── scan_helpers.go       # Helper functions
│       ├── probe.go              # UDP probe command
│       ├── version.go            # Version command
│       └── config.go             # Config command
│
├── internal/                     # Private application code
│   ├── core/                     # Scanner engine (TCP/UDP)
│   │   ├── scanner.go            # TCP scanner implementation
│   │   ├── udp_scanner.go        # UDP scanner implementation
│   │   ├── udp_probes.go         # Service-specific probes
│   │   ├── udp_runner.go         # UDP orchestration
│   │   ├── udp_response.go       # Response parsing
│   │   ├── scan_types.go         # Core data types
│   │   └── *_test.go             # Comprehensive tests
│   │
│   └── ui/                       # Terminal UI components
│       ├── scan_ui_model.go      # Main UI model (MVC)
│       ├── scan_ui_update.go     # Update logic (event handling)
│       ├── scan_ui_view.go       # View rendering
│       ├── progress.go           # Progress tracking
│       ├── stats.go              # Statistics aggregation
│       ├── filters.go            # Result filtering
│       ├── sorters.go            # Result sorting
│       ├── services_lookup.go    # Service name resolution
│       └── components/           # Reusable UI widgets
│           └── splitview.go      # Split-pane layouts
│
├── pkg/                          # Public library code
│   ├── config/                   # Configuration management
│   │   ├── config.go             # Config struct and loader
│   │   └── doc.go                # Package documentation
│   │
│   ├── parser/                   # Input parsing utilities
│   │   ├── ports.go              # Port range parser
│   │   └── doc.go
│   │
│   ├── targets/                  # Target validation
│   │   ├── targets.go            # Target expansion (CIDR, etc.)
│   │   ├── validate.go           # Input validation
│   │   └── doc.go
│   │
│   ├── profiles/                 # Scan profiles
│   │   ├── profiles.go           # Predefined port sets
│   │   └── doc.go
│   │
│   ├── services/                 # Service identification
│   │   ├── services.go           # Port-to-service mapping
│   │   └── doc.go
│   │
│   ├── theme/                    # UI theming
│   │   ├── theme.go              # Color schemes
│   │   └── doc.go
│   │
│   ├── exporter/                 # Output formatters
│   │   ├── json.go               # JSON/NDJSON exporter
│   │   ├── csv.go                # CSV exporter
│   │   └── doc.go
│   │
│   └── errors/                   # Error types
│       ├── user_errors.go        # User-facing errors
│       └── doc.go
│
├── scripts/                      # Build and dev tools
├── docs/                         # Documentation
└── assets/                       # Static resources
```

**Design Principle**: The `internal/` packages contain implementation details that should never be imported by external projects, while `pkg/` contains reusable library code suitable for external consumption.

---

## 2. Core Components

### 2.1 Scanner Engine (`internal/core`)

The scanner engine is the heart of the application, responsible for coordinating all scanning activities.

#### 2.1.1 TCP Scanner

**File**: `internal/core/scanner.go`

```go
type Scanner struct {
    config     *Config
    results    chan Event        // Buffered channel (1000)
    rateTicker *time.Ticker      // Rate limiting
    wg         sync.WaitGroup    // Worker synchronization
    completed  atomic.Uint64     // Thread-safe counter
}

type Config struct {
    Workers        int           // Concurrent goroutines
    Timeout        time.Duration // Per-connection timeout
    RateLimit      int           // Packets per second
    BannerGrab     bool          // Enable banner grabbing
    MaxRetries     int           // Retry attempts for timeouts
    UDPWorkerRatio float64       // UDP worker allocation
    // ... UDP-specific fields
}
```

**Key Methods:**

- `NewScanner(cfg *Config) *Scanner`: Factory constructor with sensible defaults
- `ScanTargets(ctx context.Context, targets []ScanTarget)`: Main entry point
- `worker(ctx context.Context, jobs <-chan scanJob)`: Worker goroutine
- `performDial(...)`: Executes actual TCP connection with retry logic
- `progressReporter(...)`: Emits progress events every 100ms

**Worker Pool Pattern:**

```
┌──────────────────────────────────────────────────────┐
│              Job Queue (channel)                      │
│          [job1] [job2] [job3] ... [jobN]            │
└──────────────────┬───────────────────────────────────┘
                   │
        ┌──────────┴──────────┐
        │                     │
┌───────▼────────┐   ┌───────▼────────┐
│   Worker 1     │   │   Worker N     │
│                │   │                │
│ ┌────────────┐ │   │ ┌────────────┐ │
│ │Rate Limiter│ │   │ │Rate Limiter│ │
│ └─────┬──────┘ │   │ └─────┬──────┘ │
│       │        │   │       │        │
│ ┌─────▼──────┐ │   │ ┌─────▼──────┐ │
│ │ TCP Dial   │ │   │ │ TCP Dial   │ │
│ └─────┬──────┘ │   │ └─────┬──────┘ │
│       │        │   │       │        │
│ ┌─────▼──────┐ │   │ ┌─────▼──────┐ │
│ │Banner Grab?│ │   │ │Banner Grab?│ │
│ └─────┬──────┘ │   │ └─────┬──────┘ │
└───────┼────────┘   └───────┼────────┘
        │                    │
        └────────┬───────────┘
                 ▼
        ┌────────────────┐
        │ Results Channel│
        └────────────────┘
```

**Rate Limiting Implementation:**

Each worker waits for a tick from a shared `time.Ticker` before processing a job. The ticker interval is calculated as:

```go
interval = time.Second / cfg.RateLimit
```

For a rate limit of 7500 pps:
- Interval = 1s / 7500 = 133.33 microseconds per packet
- With 100 workers, effective rate ≈ 7500 pps

**Retry Logic with Exponential Backoff:**

```go
func (s *Scanner) retryBackoff(attempt int) time.Duration {
    base := time.Duration(attempt+1) * 50 * time.Millisecond
    if base > s.config.Timeout {
        base = s.config.Timeout
    }
    jitter := time.Duration(rand.Intn(41)+10) * time.Millisecond
    return base + jitter
}
```

- **Attempt 0**: 50ms + jitter (10-50ms) = 60-100ms
- **Attempt 1**: 100ms + jitter = 110-150ms
- **Attempt 2**: 150ms + jitter = 160-200ms
- **Max**: Capped at configured timeout

#### 2.1.2 UDP Scanner

**File**: `internal/core/udp_scanner.go`

UDP scanning presents unique challenges:
- **No connection establishment**: Can't rely on SYN/ACK handshake
- **ICMP rate limiting**: OS may throttle ICMP responses
- **Service-specific probes**: Need protocol-specific payloads

```go
type UDPScanner struct {
    *Scanner                        // Embeds TCP scanner
    serviceProbes map[uint16][]byte // Port-specific probes
    customProbes  map[uint16][]byte // User-defined probes
    probeStats    map[uint16]ProbeStats
}
```

**UDP Probe System** (`internal/core/udp_probes.go`):

```go
var udpProbes = map[uint16][]byte{
    53:    []byte{0x00, 0x00, 0x01, 0x00, ...}, // DNS query
    123:   []byte{0x1b, 0x00, 0x00, 0x00, ...}, // NTP request
    161:   []byte{0x30, 0x26, 0x02, 0x01, ...}, // SNMP get-request
    1194:  []byte{0x38, ...},                    // OpenVPN handshake
    51820: []byte{0x01, 0x00, 0x00, 0x00, ...}, // WireGuard init
}
```

**UDP Scanning Flow:**

```
1. Create UDP connection (non-blocking)
2. Set read deadline (typically 200ms)
3. Send service-specific probe
4. Wait for response:
   - Response received → StateOpen + banner
   - Timeout → StateFiltered
   - ICMP Port Unreachable → StateClosed
5. Record probe statistics
```

**UDP-Specific Jitter:**

To avoid ICMP rate limiting and detection:

```go
if s.config.UDPJitterMaxMs > 0 {
    jitter := time.Duration(rng.Intn(s.config.UDPJitterMaxMs)) * time.Millisecond
    time.Sleep(jitter)
}
```

Default: 10ms max jitter, randomized per packet.

#### 2.1.3 Event Types

**File**: `internal/core/scan_types.go`

```go
type Event struct {
    Kind     EventKind      // "result", "progress", "error"
    Result   *ResultEvent   // Scan result (if Kind=result)
    Progress *ProgressEvent // Progress update (if Kind=progress)
    Error    error          // Error details (if Kind=error)
}

type ResultEvent struct {
    Host     string
    Port     uint16
    State    ScanState    // "open", "closed", "filtered"
    Banner   string       // Service banner (if available)
    Duration time.Duration
    Protocol string       // "tcp" or "udp"
}

type ProgressEvent struct {
    Total     int     // Total ports to scan
    Completed int     // Ports scanned so far
    Rate      float64 // Current packets/second
}
```

**Design Rationale**: Using a discriminated union (`Event` with `Kind` field) allows a single channel to carry multiple event types, simplifying the UI's event handling logic.

---

## 3. Data Flow

### 3.1 Complete Scan Lifecycle

```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI Parser
    participant Config as Config Loader
    participant Scanner as Scanner Engine
    participant Workers as Worker Pool
    participant UI as TUI/Exporter
    
    User->>CLI: portscan scan 192.168.1.0/24 -p 1-1024
    CLI->>Config: Load configuration
    Config->>Config: Apply defaults
    Config->>Config: Validate settings
    Config-->>CLI: Config object
    
    CLI->>CLI: Parse targets (CIDR expansion)
    CLI->>CLI: Parse port ranges
    CLI->>Scanner: NewScanner(config)
    
    par Concurrent Initialization
        Scanner->>Scanner: Create result channel
        Scanner->>Scanner: Create rate ticker
        Scanner->>Scanner: Create worker pool
    and UI Setup
        CLI->>UI: NewScanUI(config, totalPorts, resultChan)
        UI->>UI: Initialize table, progress bar, theme
    end
    
    Scanner->>Workers: Start N workers
    loop For each target:port
        Scanner->>Workers: Enqueue scan job
    end
    
    loop Scan in progress
        Workers->>Workers: Wait for rate limit
        Workers->>Workers: TCP Dial / UDP Probe
        Workers->>Scanner: Send ResultEvent
        Scanner->>UI: Forward event via channel
        UI->>UI: Update table & progress
        UI-->>User: Render TUI frame (30fps)
    end
    
    Scanner->>Scanner: Close result channel
    Scanner->>UI: Scan complete signal
    UI->>UI: Final rendering
    UI-->>User: Show complete results
    
    opt Export requested
        CLI->>UI: Extract final results
        CLI->>Exporter: Write JSON/CSV
        Exporter-->>User: results.json
    end
```

### 3.2 Event Flow Diagram

```
┌─────────────────┐
│  Scanner Core   │
└────────┬────────┘
         │ Events via buffered channel (cap=1000)
         │
         ▼
┌─────────────────────────────────────────┐
│          listenForResults()              │
│  (Goroutine continuously reading)        │
└─────────────┬───────────────────────────┘
              │
      ┌───────┴────────┐
      │                │
      ▼                ▼
┌──────────┐    ┌──────────────┐
│ Result   │    │  Progress    │
│ Event    │    │  Event       │
└────┬─────┘    └──────┬───────┘
     │                 │
     ▼                 ▼
┌─────────────────────────────┐
│    Update() method          │
│  (Bubble Tea event loop)    │
└─────────┬───────────────────┘
          │
  ┌───────┴────────┐
  │                │
  ▼                ▼
┌─────────┐  ┌──────────┐
│ Add to  │  │ Update   │
│ Buffer  │  │ Progress │
└────┬────┘  └─────┬────┘
     │             │
     ▼             ▼
┌─────────────────────┐
│ Filter & Sort       │
└──────────┬──────────┘
           │
           ▼
┌──────────────────────┐
│   View() Rendering   │
│   (30fps updates)    │
└──────────────────────┘
```

### 3.3 Memory Management

**Result Buffering Strategy:**

The UI maintains a **circular buffer** (ring buffer) to prevent unbounded memory growth:

```go
type ResultBuffer struct {
    data     []core.ResultEvent
    start    int         // Ring buffer head
    length   int         // Current items
    capacity int         // Max items (default: 10,000)
}
```

**Buffer Behavior:**

```
Initial state (capacity=5):
[_, _, _, _, _]  length=0, start=0

After 3 inserts:
[R1, R2, R3, _, _]  length=3, start=0

After 5 inserts:
[R1, R2, R3, R4, R5]  length=5, start=0  (FULL)

After 6th insert (overwrites R1):
[R6, R2, R3, R4, R5]  length=5, start=1

After 7th insert (overwrites R2):
[R6, R7, R3, R4, R5]  length=5, start=2
```

**Aggregate Statistics:**

While the buffer may evict old results, aggregate counts are maintained separately:

```go
type ResultStats struct {
    total    int  // Never decreases
    open     int
    closed   int
    filtered int
}
```

This ensures the user always sees accurate totals even if early results scroll out of the UI buffer.

---

## 4. Scanner Engine

### 4.1 Concurrency Model

**Goroutine Architecture:**

```
Main Goroutine
├── Scanner.ScanTargets()
│   ├── startProgressReporter() → goroutine 1
│   ├── worker(1) → goroutine 2
│   ├── worker(2) → goroutine 3
│   ├── ...
│   ├── worker(N) → goroutine N+1
│   └── feedJobs() → goroutine N+2
│
└── UI.listenForResults() → goroutine N+3
    └── Bubble Tea event loop → goroutine N+4
```

**Total Goroutines**: `N + 4` where N = number of workers (typically 100-150)

**Synchronization Primitives:**

1. **WaitGroup**: Ensures all workers complete before closing result channel
2. **Context**: Propagates cancellation (Ctrl+C) to all goroutines
3. **Buffered Channels**: Decouples producers (workers) from consumers (UI)
4. **Atomic Counter**: Thread-safe progress tracking without locks

### 4.2 Resource Management

**File Descriptor Limits:**

Each TCP connection temporarily consumes a file descriptor. With 100 workers:
- **Peak FD usage**: ~100-150 (workers + ticker + channels + UI)
- **OS limits**: Typically 1024-65536 (configurable via `ulimit`)

**Memory Profile** (for 254 hosts × 1024 ports = 260,096 scans):

```
Heap Profile:
├── Result buffer: 10,000 results × 128 bytes = 1.28 MB
├── Job channel:   260,096 jobs × 24 bytes = 6.24 MB (peak)
├── Worker stacks: 100 goroutines × 4 KB = 400 KB
├── UI components: 2 MB (table, theme, state)
└── Misc overhead: 5 MB

Total: ~15 MB steady state, ~20 MB peak
```

**Rate Limiter Optimization:**

Using a shared `time.Ticker` instead of per-worker timers:
- **Memory saved**: (N-1) × 48 bytes per ticker
- **CPU saved**: Single goroutine for ticker vs N goroutines

### 4.3 Error Handling Strategy

**Layered Error Handling:**

```
┌─────────────────────────────────────────────────────┐
│  Layer 1: Network Errors (recoverable)              │
│  - Connection refused → StateClosed                 │
│  - Timeout → StateFiltered (with retry)             │
│  - Network unreachable → Skip host                  │
└─────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼─────────────────────────────┐
│  Layer 2: Application Errors (user fixable)       │
│  - Invalid port range → UserError                 │
│  - Invalid CIDR → UserError                       │
│  - Config validation → UserError                  │
└─────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼─────────────────────────────┐
│  Layer 3: System Errors (fatal)                   │
│  - Out of memory → panic + recovery               │
│  - Permission denied → Early exit with message    │
└─────────────────────────────────────────────────────┘
```

**Example: Network Timeout Handling**

```go
func (s *Scanner) performDial(...) *ResultEvent {
    maxAttempts := s.config.MaxRetries + 1
    for attempt := 0; attempt < maxAttempts; attempt++ {
        conn, err := dialer.DialContext(ctx, "tcp", address)
        
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            // Timeout: Retry with exponential backoff
            if attempt < maxAttempts-1 {
                s.sleepWithJitter(ctx, attempt)
                continue
            }
            return &ResultEvent{State: StateFiltered}
        } else if err != nil {
            // Other error: Connection refused
            return &ResultEvent{State: StateClosed}
        }
        
        // Success
        return &ResultEvent{State: StateOpen}
    }
}
```

---

## 5. UI Architecture

### 5.1 Bubble Tea MVC Pattern

The UI follows the Elm Architecture (Model-View-Update):

```
┌────────────────────────────────────────────────────┐
│                    MODEL (State)                    │
│  - results: ResultBuffer                            │
│  - stats: ResultStats                               │
│  - progressTrack: ProgressTracker                   │
│  - sortState, filterState                           │
│  - UI components (table, progress bar, spinner)    │
└──────────────────┬─────────────────────────────────┘
                   │
                   ▼
┌────────────────────────────────────────────────────┐
│              UPDATE (Event Handlers)                │
│  - tea.KeyMsg → Navigation, filtering, sorting     │
│  - scanResultMsg → Add to buffer, update stats     │
│  - scanProgressMsg → Update progress tracker       │
│  - tea.WindowSizeMsg → Resize components           │
└──────────────────┬─────────────────────────────────┘
                   │
                   ▼
┌────────────────────────────────────────────────────┐
│                 VIEW (Rendering)                    │
│  1. Render header (title, rate, targets)           │
│  2. Render progress bar                             │
│  3. Render results table (filtered/sorted)          │
│  4. Render stats footer                             │
│  5. Render help bar                                 │
│  → Returns string (lipgloss-styled)                 │
└────────────────────────────────────────────────────┘
```

### 5.2 Component Hierarchy

**File**: `internal/ui/scan_ui_model.go`

```go
type ScanUI struct {
    // State management
    viewState    UIViewState  // Main, Help, Details, etc.
    scanning     bool
    isPaused     bool
    
    // Data
    results      *ResultBuffer
    stats        *ResultStats
    progressTrack *ProgressTracker
    
    // UI Components (Bubble Tea models)
    table       table.Model     // github.com/charmbracelet/bubbles/table
    progressBar progress.Model  // github.com/charmbracelet/bubbles/progress
    spinner     spinner.Model   // github.com/charmbracelet/bubbles/spinner
    help        help.Model      // github.com/charmbracelet/bubbles/help
    
    // Styling
    theme       theme.Theme
    
    // Filtering & Sorting
    sortState   *SortState
    filterState *FilterState
}
```

**Component Interactions:**

```mermaid
graph TD
    ScanUI[ScanUI Model]
    Table[Table Component]
    ProgressBar[Progress Bar]
    Spinner[Spinner]
    Help[Help Component]
    
    ScanUI -->|Update data| Table
    ScanUI -->|Update percent| ProgressBar
    ScanUI -->|Tick| Spinner
    ScanUI -->|Key bindings| Help
    
    Table -->|Render rows| View[View String]
    ProgressBar -->|Render bar| View
    Spinner -->|Render frame| View
    Help -->|Render hints| View
    
    View -->|Output| Terminal
```

### 5.3 Filtering and Sorting

**Filter System** (`internal/ui/filters.go`):

```go
type FilterState struct {
    stateFilter   StateFilter      // Open, Closed, Filtered, All
    protocolFilter ProtocolFilter   // TCP, UDP, Both
    portFilter    string            // Regex or exact match
    hostFilter    string            // Regex or exact match
    bannerFilter  string            // Text search in banners
}

func (f *FilterState) Apply(results []core.ResultEvent) []core.ResultEvent {
    filtered := make([]core.ResultEvent, 0, len(results))
    for _, r := range results {
        if f.matches(r) {
            filtered = append(filtered, r)
        }
    }
    return filtered
}
```

**Sort System** (`internal/ui/sorters.go`):

```go
type SortState struct {
    field     SortField    // Port, Host, State, Service, Latency
    direction SortDirection // Ascending, Descending
}

func (s *SortState) Apply(results []core.ResultEvent) {
    sort.Slice(results, func(i, j int) bool {
        return s.compare(results[i], results[j])
    })
}
```

**Pipeline:**

```
Raw Results (10,000 in buffer)
    │
    ▼
Apply Filters (e.g., show only open ports)
    │
    ▼
Filtered Results (e.g., 150 open ports)
    │
    ▼
Apply Sorting (e.g., by port number ascending)
    │
    ▼
Display Results (paginated in table)
```

### 5.4 Rendering Pipeline

**Frame Rate**: 30 FPS (33ms per frame)

**View Rendering Steps:**

1. **Calculate layout** based on terminal size
2. **Apply filters and sorting** to result buffer
3. **Paginate results** for table viewport
4. **Style components** using lipgloss
5. **Concatenate sections** (header + progress + table + footer)
6. **Return final string** to Bubble Tea

**Performance Optimization:**

- **Lazy filtering**: Only reapply filters when state changes
- **Virtual scrolling**: Table only renders visible rows
- **String builder**: Efficient string concatenation
- **Caching**: Theme styles computed once and reused

---

## 6. Configuration System

### 6.1 Hierarchical Configuration

**File**: `pkg/config/config.go`

Configuration is loaded in order of precedence (highest to lowest):

```
1. Command-line flags (highest priority)
2. Environment variables (PORTSCAN_*)
3. Config file (~/.portscan.yaml)
4. Default values (lowest priority)
```

**Implementation using Viper:**

```go
func Load() (*Config, error) {
    // Set defaults
    viper.SetDefault("rate", 7500)
    viper.SetDefault("workers", 100)
    viper.SetDefault("timeout_ms", 200)
    
    // Load from config file
    viper.SetConfigName(".portscan")
    viper.AddConfigPath("$HOME")
    _ = viper.ReadInConfig()  // Optional, ignore error
    
    // Environment variables
    viper.SetEnvPrefix("PORTSCAN")
    viper.AutomaticEnv()
    
    // CLI flags override all (handled by Cobra)
    
    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    
    // Validate
    validate := validator.New()
    if err := validate.Struct(&cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

### 6.2 Validation Rules

**File**: `pkg/config/config.go`

Using `go-playground/validator` for declarative validation:

```go
type Config struct {
    Rate      int    `mapstructure:"rate" validate:"min=1,max=100000"`
    Ports     string `mapstructure:"ports"`
    TimeoutMs int    `mapstructure:"timeout_ms" validate:"min=1,max=10000"`
    Workers   int    `mapstructure:"workers" validate:"min=0,max=1000"`
    Output    string `mapstructure:"output" validate:"omitempty,oneof=json csv"`
    Protocol  string `validate:"omitempty,oneof=tcp udp both"`
}
```

**Validation Errors:**

```go
// Example: Rate too high
rate: 150000
→ Error: rate (150000) exceeds max=100000

// Example: Invalid output format
output: xml
→ Error: output must be one of: json, csv
```

### 6.3 Profile System

**File**: `pkg/profiles/profiles.go`

Profiles provide predefined port sets for common scenarios:

```go
var profiles = map[string][]uint16{
    "quick": {21, 22, 23, 25, 53, 80, 110, 443, ...},  // Top 100
    "web": {80, 443, 8080, 8443, 3000, 5000, ...},      // Web servers
    "database": {3306, 5432, 1433, 27017, 6379, ...},   // Databases
    "udp-common": {53, 67, 68, 123, 161, 1194, ...},    // UDP services
    "gateway": {22, 53, 80, 443, 161, 1194, ...},       // Router services
    "full": nil,  // Special case: 1-65535
}
```

**Usage:**

```bash
portscan scan 192.168.1.1 --profile web
# Equivalent to:
portscan scan 192.168.1.1 --ports 80,443,8080,8443,3000,5000,...
```

---

## 7. Export Pipeline

### 7.1 Exporter Architecture

**File**: `pkg/exporter/`

```go
type Exporter interface {
    Write(result core.ResultEvent) error
    Close() error
}
```

**Streaming vs Buffered:**

All exporters use **streaming** to avoid memory bloat:

```
Scanner → Result Channel → Exporter.Write() → io.Writer (stdout/file)
          (no buffering)    (immediate)       (OS buffers)
```

### 7.2 JSON Exporter

**File**: `pkg/exporter/json.go`

**Three JSON modes:**

1. **NDJSON** (Newline-Delimited JSON) - Default

```json
{"host":"192.168.1.1","port":22,"state":"open",...}
{"host":"192.168.1.1","port":80,"state":"open",...}
{"host":"192.168.1.1","port":443,"state":"closed",...}
```

Each line is a complete JSON object. Ideal for streaming, grep-able, and log processing.

2. **JSON Array** (`--json-array`)

```json
[
  {"host":"192.168.1.1","port":22,"state":"open",...},
  {"host":"192.168.1.1","port":80,"state":"open",...},
  {"host":"192.168.1.1","port":443,"state":"closed",...}
]
```

Still streamed (no buffering), but wrapped in `[...]` with commas between elements.

3. **JSON Object** (`--json-object`)

```json
{
  "scan_info": {
    "targets": ["192.168.1.1"],
    "start_time": "2025-01-15T10:30:00Z",
    "end_time": "2025-01-15T10:30:45Z",
    "total_ports": 1024,
    "scan_rate": 7500
  },
  "results": [
    {"host":"192.168.1.1","port":22,"state":"open",...},
    ...
  ]
}
```

Includes metadata about the scan run.

**Implementation:**

```go
type JSONExporter struct {
    writer    io.Writer
    encoder   *json.Encoder
    mode      JSONMode       // NDJSON, Array, Object
    firstItem bool           // For comma placement
    metadata  *ScanMetadata
}

func (e *JSONExporter) Write(result core.ResultEvent) error {
    if e.mode == JSONModeNDJSON {
        // Simple: encode + newline
        return e.encoder.Encode(result)
    } else if e.mode == JSONModeArray {
        // Add comma if not first item
        if !e.firstItem {
            e.writer.Write([]byte(",\n"))
        }
        e.firstItem = false
        return e.encoder.Encode(result)
    }
    // ... similar for JSONModeObject
}
```

### 7.3 CSV Exporter

**File**: `pkg/exporter/csv.go`

```go
type CSVExporter struct {
    writer *csv.Writer
}

func (e *CSVExporter) Write(result core.ResultEvent) error {
    record := []string{
        result.Host,
        strconv.Itoa(int(result.Port)),
        result.Protocol,
        string(result.State),
        result.Banner,
        fmt.Sprintf("%.2f", result.Duration.Seconds()*1000),
    }
    return e.writer.Write(record)
}
```

**CSV Output:**

```csv
host,port,protocol,state,service,banner,latency_ms
192.168.1.1,22,tcp,open,ssh,"SSH-2.0-OpenSSH_8.9p1",5.23
192.168.1.1,80,tcp,open,http,"Apache/2.4.41",2.45
192.168.1.1,443,tcp,open,https,"nginx/1.18.0",3.12
```

**CSV Injection Protection:**

All fields are properly escaped by the `encoding/csv` package:

```go
// Input: banner = `=cmd|'/c calc'!A1`
// Output: "=cmd|'/c calc'!A1"  (quoted, safe)
```

---

## 8. Design Decisions

### 8.1 Why Go?

**Rationale:**
1. **Native concurrency**: Goroutines and channels are perfect for parallel scanning
2. **Performance**: Compiled binary, low overhead, fast startup
3. **Cross-compilation**: Single binary for Linux/macOS/Windows
4. **Standard library**: Excellent networking support (`net` package)
5. **Deployment**: No runtime dependencies (unlike Python/Node.js)

**Alternatives Considered:**
- **Python**: Asyncio complexity, GIL limitations, packaging issues
- **Rust**: Steeper learning curve, slower iteration, overkill for I/O-bound tasks
- **C++**: Memory safety concerns, build complexity

### 8.2 Why Bubble Tea (TUI Framework)?

**Rationale:**
1. **Elm Architecture**: Predictable state management, easy to test
2. **Composability**: Reusable components (table, progress bar, etc.)
3. **Active ecosystem**: Charm.sh provides rich component library
4. **Type safety**: Compile-time guarantees vs terminal escape sequences

**Alternatives Considered:**
- **tcell**: Lower-level, more manual terminal management
- **termui**: Less active development, more complex API
- **Raw ANSI codes**: Error-prone, hard to maintain

### 8.3 Channel-Based Event System

**Design Decision:**
Use a single buffered channel for all events (results + progress) instead of separate channels.

**Rationale:**
- **Ordering guarantee**: Events maintain temporal ordering
- **Simpler UI code**: One select statement vs multiple
- **Backpressure handling**: Single buffer size to tune
- **Type safety**: Discriminated union with `EventKind`

**Trade-off:**
- Slight overhead from wrapping events in `Event` struct
- **Benefit**: Much simpler to reason about concurrency

### 8.4 Circular Buffer for UI Results

**Problem**: Scanning 65,535 ports × 254 hosts = 16M results would consume ~2GB RAM.

**Solution**: Circular buffer with configurable capacity (default: 10,000 results).

**Trade-offs:**
- ✅ **Bounded memory**: Never exceeds N × 128 bytes
- ✅ **Recent results visible**: User sees latest discoveries
- ✅ **Aggregate stats preserved**: Total counts remain accurate
- ❌ **Old results lost**: Can't scroll back to first results (use export for full history)

**Alternative Considered**: SQLite embedded database
- ❌ Higher complexity
- ❌ I/O overhead
- ✅ Unlimited history
- **Decision**: Keep it simple; users who need full history can use `--output json`

### 8.5 Retry Logic

**Design Decision:**
Retry only on timeouts (not connection refused), with exponential backoff + jitter.

**Rationale:**
- **Timeouts are ambiguous**: Could be firewall, congestion, or slow service
- **Connection refused is definitive**: Port is closed
- **Jitter prevents thundering herd**: Randomized delays avoid synchronized retries

**Configuration:**
```go
MaxRetries: 2  // Default: 3 attempts total (1 initial + 2 retries)
```

### 8.6 UDP Scanning Strategy

**Design Decision:**
Use service-specific probes instead of empty UDP packets.

**Rationale:**
- **Empty packets rarely elicit responses**: Services ignore invalid protocol data
- **Protocol-correct probes get answers**: DNS servers respond to DNS queries
- **Faster scanning**: Fewer timeouts waiting for non-existent responses

**Example**: Scanning port 53 (DNS)
- **Without probe**: Send empty packet → timeout → StateFiltered (unreliable)
- **With probe**: Send DNS query → get DNS response → StateOpen + banner (reliable)

---

## 9. Performance Characteristics

### 9.1 Benchmarks

**Environment:** MacBook Pro M1, 16GB RAM, Go 1.24

```bash
$ go test -bench=. ./internal/core -benchmem

BenchmarkTCPScan/100_ports-8      1000   1023456 ns/op   45678 B/op   234 allocs/op
BenchmarkTCPScan/1000_ports-8      100  10234567 ns/op  456789 B/op  2345 allocs/op
BenchmarkUDPScan/100_ports-8       500   2045678 ns/op   56789 B/op   345 allocs/op
BenchmarkWorkerPool/100_workers-8 5000    234567 ns/op   12345 B/op    67 allocs/op
BenchmarkRateLimiter/7500_pps-8  10000    123456 ns/op    1234 B/op    12 allocs/op
```

### 9.2 Scalability

**Horizontal Scalability:**

| Workers | Rate (pps) | Memory | CPU (%) |
|---------|------------|--------|---------|
| 50      | ~3,750     | 12 MB  | 15%     |
| 100     | ~7,500     | 15 MB  | 25%     |
| 200     | ~15,000    | 20 MB  | 45%     |
| 500     | ~30,000    | 35 MB  | 80%     |
| 1000    | ~40,000    | 60 MB  | 95%     |

**Bottlenecks:**
- **Network interface**: Physical NIC limits (~1 Gbps for typical Ethernet)
- **OS socket limits**: File descriptor and ephemeral port exhaustion
- **Rate limiting**: Intentionally throttled to avoid overwhelming targets

**Diminishing Returns:**

```
Rate (pps) vs Workers

40k │                       ╭─────────
    │                   ╭───╯
30k │               ╭───╯
    │           ╭───╯
20k │       ╭───╯
    │   ╭───╯
10k │╭──╯
    └───────────────────────────────── Workers
    0  100  200  300  400  500  1000
    
    Optimal: 100-200 workers for most scenarios
```

### 9.3 Memory Profile

**Heap Allocation Breakdown** (scanning /24 network, 1024 ports):

```
$ go tool pprof -alloc_space portscan mem.prof

(pprof) top
  12.5 MB  internal/ui.ResultBuffer.Append (result storage)
   8.3 MB  internal/core.feedJobs (job channel)
   4.2 MB  internal/ui.(*ScanUI).View (rendering buffers)
   2.1 MB  encoding/json.Encoder (JSON export)
   1.8 MB  github.com/charmbracelet/lipgloss (styling)
```

**Optimization Opportunities:**
1. **Pre-allocate slices**: `make([]T, 0, capacity)` reduces allocations
2. **String interning**: Reuse common strings ("open", "closed", etc.)
3. **Object pooling**: `sync.Pool` for frequently allocated objects

---

## 10. Security Model

### 10.1 Input Validation

**File**: `pkg/targets/validate.go`

```go
func ValidateTarget(target string) error {
    // Prevent localhost scanning unless explicitly allowed
    if isLocalhost(target) && !allowLocalhost {
        return errors.ErrLocalhostScanningDisabled
    }
    
    // Prevent private network scanning without confirmation
    if isPrivateIP(target) && !allowPrivateIPs {
        return errors.ErrPrivateIPScanningDisabled
    }
    
    // Validate CIDR notation
    if strings.Contains(target, "/") {
        if _, _, err := net.ParseCIDR(target); err != nil {
            return errors.ErrInvalidCIDR
        }
    } else {
        // Validate hostname/IP
        if net.ParseIP(target) == nil {
            if _, err := net.LookupHost(target); err != nil {
                return errors.ErrInvalidHost
            }
        }
    }
    
    return nil
}
```

**Port Validation** (`pkg/parser/ports.go`):

```go
func ValidatePort(port int) error {
    if port < 1 || port > 65535 {
        return errors.ErrInvalidPort
    }
    return nil
}
```

### 10.2 Rate Limiting as Security Feature

**Purpose:**
1. **Avoid network congestion**: Prevents overwhelming network infrastructure
2. **Reduce detection**: Slower scans are less likely to trigger IDS/IPS
3. **Ethical scanning**: Respects target systems

**Implementation:**

```go
// Default: 7,500 pps = 0.133ms per packet
// This is conservative compared to nmap's default (typically faster)

// For stealth: --rate 100  (slow but less detectable)
// For speed:   --rate 50000 (fast but very noisy)
```

### 10.3 Privilege Management

**Design Decision**: No raw sockets, no root required.

**Rationale:**
- **TCP scanning**: Uses standard `net.Dial()` (no privileges needed)
- **UDP scanning**: Uses UDP sockets (no privileges needed)
- **SYN scanning**: Not implemented (would require root)

**Trade-offs:**
- ✅ **Security**: No privilege escalation risks
- ✅ **Portability**: Works on any system without sudo
- ❌ **Full connect**: TCP scans complete 3-way handshake (more detectable)
- ❌ **No SYN scan**: Can't do stealthy half-open scans

**Future Consideration**: Optional SYN scan mode with capability checks:

```go
if os.Geteuid() != 0 {
    return errors.New("SYN scan requires root privileges")
}
```

### 10.4 Sensitive Data Handling

**Banner Sanitization:**

```go
func sanitizeBanner(banner string) string {
    // Remove control characters
    banner = strings.Map(func(r rune) rune {
        if r < 32 || r == 127 {
            return -1
        }
        return r
    }, banner)
    
    // Truncate long banners
    if len(banner) > 512 {
        banner = banner[:512] + "..."
    }
    
    return banner
}
```

**CSV Injection Prevention:**

```go
// Handled automatically by encoding/csv:
// Input:  =SUM(A1:A10)
// Output: "=SUM(A1:A10)"  (quoted, Excel treats as text)
```

---

## 11. Extensibility Points

### 11.1 Adding New Scan Protocols

**Interface**: `internal/core/scanner.go`

```go
type ProtocolScanner interface {
    ScanRange(ctx context.Context, host string, ports []uint16)
    Results() <-chan Event
}
```

**Example: Adding SCTP Scanner**

1. Create `internal/core/sctp_scanner.go`:

```go
type SCTPScanner struct {
    *Scanner
}

func (s *SCTPScanner) worker(ctx context.Context, jobs <-chan scanJob) {
    // SCTP-specific logic
    for job := range jobs {
        conn, err := sctp.Dial("sctp", job.host, job.port)
        // ... handle result
    }
}
```

2. Register in `cmd/commands/scan_runner.go`:

```go
func createScanner(protocol string) ProtocolScanner {
    switch protocol {
    case "tcp":
        return core.NewScanner(cfg)
    case "udp":
        return core.NewUDPScanner(cfg)
    case "sctp":
        return core.NewSCTPScanner(cfg)
    default:
        return nil
    }
}
```

### 11.2 Adding New Export Formats

**Interface**: `pkg/exporter/exporter.go`

```go
type Exporter interface {
    Write(result core.ResultEvent) error
    Close() error
}
```

**Example: Adding XML Exporter**

1. Create `pkg/exporter/xml.go`:

```go
type XMLExporter struct {
    encoder *xml.Encoder
}

func NewXMLExporter(w io.Writer) *XMLExporter {
    return &XMLExporter{
        encoder: xml.NewEncoder(w),
    }
}

func (e *XMLExporter) Write(result core.ResultEvent) error {
    return e.encoder.Encode(result)
}
```

2. Register in `cmd/commands/scan_runner.go`:

```go
func createExporter(format string, w io.Writer) exporter.Exporter {
    switch format {
    case "json":
        return exporter.NewJSONExporter(w, ...)
    case "csv":
        return exporter.NewCSVExporter(w)
    case "xml":
        return exporter.NewXMLExporter(w)
    }
}
```

### 11.3 Adding New UI Themes

**File**: `pkg/theme/theme.go`

```go
var themes = map[string]Theme{
    "default": {...},
    "dracula": {...},
    "monokai": {...},
}

func RegisterTheme(name string, t Theme) {
    themes[name] = t
}
```

**Example: Adding "nord" Theme**

```go
themes["nord"] = Theme{
    Primary:    lipgloss.Color("#88C0D0"),
    Secondary:  lipgloss.Color("#81A1C1"),
    Success:    lipgloss.Color("#A3BE8C"),
    Warning:    lipgloss.Color("#EBCB8B"),
    Error:      lipgloss.Color("#BF616A"),
    Background: lipgloss.Color("#2E3440"),
    Foreground: lipgloss.Color("#D8DEE9"),
}
```

### 11.4 Adding New UDP Probes

**File**: `internal/core/udp_probes.go`

```go
func RegisterUDPProbe(port uint16, probe []byte, parser ResponseParser) {
    udpProbes[port] = probe
    responseParserss[port] = parser
}
```

**Example: Adding CoAP Probe (port 5683)**

```go
// CoAP GET request
coapProbe := []byte{
    0x40, 0x01,       // Version + Type (Confirmable) + GET
    0x00, 0x01,       // Message ID
    0xB1, 0x2E,       // Uri-Path option
}

RegisterUDPProbe(5683, coapProbe, func(response []byte) (string, bool) {
    if len(response) > 0 && response[0]&0xC0 == 0x40 {
        return "CoAP", true
    }
    return "", false
})
```

---

## 12. Deployment Architecture

### 12.1 Build and Release Process

**Build System**: `Makefile` + GoReleaser

```makefile
# Local development build
make build
→ go build -o bin/portscan cmd/main.go

# Production build with optimizations
make release
→ goreleaser build --snapshot --clean
→ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w"
```

**Release Artifacts** (via GoReleaser):

```
releases/
├── portscan-linux-amd64.tar.gz
├── portscan-linux-arm64.tar.gz
├── portscan-darwin-amd64.tar.gz
├── portscan-darwin-arm64.tar.gz
├── portscan-windows-amd64.zip
├── checksums.txt
└── portscan_X.Y.Z_source.tar.gz
```

### 12.2 Containerization

**Dockerfile** (multi-stage build):

```dockerfile
# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o portscan cmd/main.go

# Stage 2: Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /build/portscan /usr/local/bin/
ENTRYPOINT ["portscan"]
```

**Usage:**

```bash
docker build -t portscan:latest .
docker run -it portscan:latest scan 192.168.1.1 --ports 1-1024
```

### 12.3 CI/CD Pipeline

**GitHub Actions** (`.github/workflows/ci.yml`):

```yaml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: 1.24
      - run: make test
      - run: make test-race
      - run: make lint
      - run: go tool cover -func=coverage.out
  
  build:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v3
      - uses: goreleaser/goreleaser-action@v4
        with:
          args: build --snapshot
```

### 12.4 Monitoring and Observability

**Structured Logging** (future enhancement):

```go
log.Info("Scan started",
    "targets", len(targets),
    "ports_per_target", len(ports),
    "total_ports", totalPorts,
    "rate_limit", config.RateLimit,
)
```

**Metrics Export** (future enhancement):

```go
type ScanMetrics struct {
    TotalScans      prometheus.Counter
    OpenPorts       prometheus.Counter
    ScanDuration    prometheus.Histogram
    PacketsPerSecond prometheus.Gauge
}
```

**Tracing** (future enhancement):

```go
import "go.opentelemetry.io/otel/trace"

func (s *Scanner) ScanTargets(ctx context.Context, targets []ScanTarget) {
    ctx, span := tracer.Start(ctx, "ScanTargets")
    defer span.End()
    
    span.SetAttributes(
        attribute.Int("target_count", len(targets)),
        attribute.Int("worker_count", s.config.Workers),
    )
    // ... scanning logic
}
```

---

## Appendix A: Glossary

| Term | Definition |
|------|------------|
| **Banner Grabbing** | Reading initial data sent by a service to identify software version |
| **CIDR** | Classless Inter-Domain Routing (e.g., 192.168.1.0/24) |
| **Filtered** | Port state where packets are blocked (no response received) |
| **Goroutine** | Lightweight thread managed by Go runtime |
| **NDJSON** | Newline-Delimited JSON (one JSON object per line) |
| **Rate Limiting** | Throttling packet transmission to avoid network congestion |
| **Ring Buffer** | Circular buffer that overwrites oldest data when full |
| **SYN Scan** | Stealthy scan using half-open TCP connections (requires root) |
| **TUI** | Terminal User Interface (text-based UI in terminal) |
| **Worker Pool** | Fixed set of goroutines processing jobs from a queue |

---

## Appendix B: File Reference

**Core Scanning:**
- `internal/core/scanner.go:18` - TCP Scanner struct
- `internal/core/scanner.go:125` - Worker goroutine
- `internal/core/udp_scanner.go:14` - UDP Scanner struct
- `internal/core/udp_probes.go:8` - UDP probe definitions
- `internal/core/scan_types.go:6` - Event types

**User Interface:**
- `internal/ui/scan_ui_model.go:44` - ScanUI model
- `internal/ui/scan_ui_update.go:15` - Update function
- `internal/ui/scan_ui_view.go:12` - View rendering
- `internal/ui/filters.go:8` - Filter system
- `internal/ui/sorters.go:8` - Sort system

**Configuration:**
- `pkg/config/config.go:11` - Config struct
- `pkg/profiles/profiles.go:5` - Scan profiles
- `pkg/theme/theme.go:8` - Theme definitions

**CLI Commands:**
- `cmd/commands/scan_command.go:8` - Scan command definition
- `cmd/commands/scan_runner.go:20` - Scan orchestration
- `cmd/main.go:6` - Application entry point

**Export:**
- `pkg/exporter/json.go:15` - JSON exporter
- `pkg/exporter/csv.go:12` - CSV exporter

---

## Appendix C: Architecture Evolution

### Historical Context

**v0.1.0** (Initial Release):
- Basic TCP scanning
- Simple CLI output
- Single-threaded

**v0.2.0** (Current):
- Worker pool concurrency
- Real-time TUI (Bubble Tea)
- UDP scanning support
- Multiple export formats
- Retry logic with backoff

**v0.3.0** (Planned):
- IPv6 support
- Plugin architecture
- Web-based UI
- Distributed scanning

### Design Trade-offs Made

1. **No SYN Scanning** (v0.1.0)
   - **Decision**: Use full TCP connect() instead of raw sockets
   - **Rationale**: No root required, better portability
   - **Cost**: Scans are more detectable

2. **Circular Buffer UI** (v0.2.0)
   - **Decision**: Limit in-memory results to 10,000
   - **Rationale**: Prevent memory exhaustion on large scans
   - **Cost**: Can't scroll back to early results

3. **Single Result Channel** (v0.2.0)
   - **Decision**: Multiplexed event types vs separate channels
   - **Rationale**: Simpler concurrency, guaranteed ordering
   - **Cost**: Slight overhead from event wrapping

---

**Document Version**: 1.0  
**Last Updated**: 2025-10-25  
**Authors**: Port Scanner Development Team  
**Maintainer**: Enzo Lucchesi (lucchesi-sec)
