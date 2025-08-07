Go Port Scanner — Architecture Guide

Purpose: This architecture.md document is the authoritative reference for contributors building and maintaining the Go TUI Port-Scanner. It lays out goals, guiding principles, component boundaries, project structure, and operational concerns so the codebase stays coherent as the feature-set grows.

⸻

1 — Vision & Goals

Goal	Description
Speed	Scan thousands of ports per second on commodity laptops without exhausting local ephemeral ports.
Insightful UI	Real-time progress bars, sortable tables, latency graphs, and summary dashboards in a single-binary terminal app.
Modular Design	Core scanner reusable as a library or micro-service; UI and exporters are pluggable layers.
Cross-platform	First-class support for Linux, macOS, and Windows (ConPTY) with 24-bit colour.
Dev Ergonomics	Clear directory layout, strict linting, one-command release pipeline, reproducible builds.


⸻

2 — High-Level Architecture

┌────────────────────────────┐        scan results      ┌────────────────────────────┐
│  internal/core (scanner)   ├──────►  internal/ui (TUI) │
│                            │  JSON / struct events    │   Bubble Tea model + view  │
└─────▲───────────────┬──────┘                         └────────────▲──────────────┘
      │ ctx.Cancel()  │ errors                                     │ render()
      │               │                                           │
┌─────┴───────────────▼──────┐                         ┌───────────┴──────────────┐
│        cmd/ (CLI)          │  config & flags        │ pkg/exporter (JSON/CSV)  │
└────────────────────────────┘                         └───────────────────────────┘

Communication contract: core publishes scan.Event messages on a buffered channel. Consumers (UI, JSON encoder, etc.) own their own goroutine and never block the scanner.

⸻

3 — Project Layout

portscanner/
├── cmd/                # Cobra commands (root, scan, version)
├── internal/
│   ├── core/           # Worker pool, banner grabber, rate limiter
│   └── ui/             # Bubble Tea model, views, theme tokens
├── pkg/
│   ├── config/         # Typed config structs & Viper loader
│   ├── exporter/       # JSON, CSV, Prometheus, etc.
│   └── theme/          # Lip Gloss colour palette & helpers
├── scripts/            # Dev helpers (lint, benchmark)
├── .github/workflows/  # CI: test → lint → build → release
└── go.mod

Follow Go 1.22 module workspaces if future sub-projects (e.g., scanner-lib) emerge.

⸻

4 — Core Scanner Design (internal/core)

4.1 Event Types

// ResultEvent is emitted when a port is conclusively open/closed.
type ResultEvent struct {
    Host     string
    Port     uint16
    State    ScanState // open, closed, filtered
    Banner   string    // optional first-line service banner
    Duration time.Duration
}

// ProgressEvent periodically reports totals for UI progress bars.
//   Emitted every 100 ms or on Ctrl-C.

4.2 Worker Pool
	•	Dispatcher feeds ports into jobs <- uint16.
	•	Workers (N = GOMAXPROCS*2 default) dial with net.DialTimeout.
	•	Timeout Strategy: Exponential back-off for retransmissions; default 200 ms cap.
	•	Rate Limiting: Token-bucket (100 kpps default) with time.Ticker.
	•	Context Cancellation: Root context.Context aborts immediately on SIGINT.

4.3 Banner Grabbing

Optional TCP read up to 512 B after connect for protocols exposing banners (SSH, FTP, SMTP). Controlled via --banners flag.

⸻

5 — Terminal UI (internal/ui)

Concern	Implementation
Framework	Bubble Tea state-machine (Init, Update, View).
Styling	Lip Gloss tokens in pkg/theme, e.g., var Danger = lipgloss.Color("#FF5555").
Widgets	Bubbles (table, spinner), progressbar/v3 for multi-line bars, asciigraph for latency graphs, termui gauges for post-scan summaries.
Input	Bubble Tea keymaps (j/k navigation, g/G jump, q quit).
Layout	Flex layout: sidebar (stats) + main pane (table/graph).

The UI subscribes to events via chan interface{} and updates model state; all heavy rendering is throttled to 30 fps.

⸻

6 — Configuration (pkg/config)
	•	Hierarchy: CLI flags → Env Vars → ~/.portscan.yaml → Defaults.
	•	Backend: Viper auto-binds.
	•	Validation: github.com/go-playground/validator/v10 in Init().
	•	Sample Config:

rate: 7500       # packets per second
ports: "1-1024,3306,6379"
timeout_ms: 200
output: json
ui:
  theme: dracula


⸻

7 — Command-line Interface (cmd/)

$ portscan scan 10.0.0.0/24 --ports 1-1024 --banners --rate 8000
$ portscan scan --stdin --json > results.json
$ portscan version   # git tag + commit + build date

	•	Root Flags: --config, --quiet, --no-color.
	•	Hidden Flags: --profile (pprof), --trace.

⸻

8 — Observability
	•	p prof endpoints on localhost:6060 when PORTSCAN_PPROF=1.
	•	Structured Logs: Zap logger in JSON mode if --log-json; human mode otherwise.
	•	Metrics: Prometheus textfile exporter via pkg/exporter.

⸻

9 — Build & Release Pipeline
	1.	CI: GitHub Actions matrix - lints (golangci-lint), unit tests, race detector.
	2.	Release: GoReleaser auto-tags binaries for linux_amd64, darwin_arm64, windows_amd64; generates Homebrew tap & Docker image.
	3.	Checksums: SHA-256 and SBOM via CycloneDX.

⸻

10 — Testing Strategy

Layer	Approach
Unit	Dialer mocked with local net.Listener; 100% event path coverage.
Integration	Spin up ephemeral docker containers (nginx, redis) to verify banner detection.
Fuzzing	go test -fuzz on input parsers (CIDR, port ranges).
Benchmarks	go test -bench ./internal/core -run=^$ under varying rates.


⸻

11 — Performance Considerations
	•	Ephemeral Port Exhaustion: Use raw sockets (future) or reuseaddr knob.
	•	Adaptive Timeouts: Dynamic per-host RTT sampling.
	•	Batch DNS: Resolve in parallel with net.Resolver custom dialer.

⸻

12 — Security Notes
	•	Drop root after raw-socket bind (if capability allowed).
	•	Harden TLS for banner grabs; verify cert chain.
	•	Dependency scanning via github-actions/setup-go-v4 + govulncheck.

⸻

13 — Extensibility Roadmap

Milestone	Feature
v0.3	Export to SQLite + Web UI (React + SQLite WASM).
v0.4	Wish middleware to serve TUI over SSH (ssh server :0).
v1.0	UDP scanning, SYN scan with raw packets, plugin interface.


⸻

14 — Coding & Style Guidelines
	•	go fumpt + goimports enforced pre-commit.
	•	golangci-lint with govet, staticcheck, errcheck, revive.
	•	Conventional Commits (feat:, fix:, docs:) — required for auto-changelog.

⸻

15 — Appendix A: Sample .goreleaser.yaml

project_name: portscan
builds:
  - env: ["CGO_ENABLED=0"]
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
release:
  github:
    owner: you
    name: portscan
brew:
  github:
    owner: you
    name: homebrew-tap


⸻

16 — Appendix B: Key Dependencies

Library	Purpose
Bubble Tea	Core TUI framework
Lip Gloss	Styling/theme layer
Bubbles	Ready-made widgets
tcell	Terminal backend (via Bubble Tea)
Cobra	CLI parsing
Viper	Config management
progressbar/v3	Thread-safe progress bars
asciigraph	ASCII graphs for latency trends
termui	Post-scan dashboards
Wish	SSH TUI server (future)
goreleaser	Binary packaging & release


⸻

End of architecture.md
