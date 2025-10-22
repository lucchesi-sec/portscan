# Port Scanner Improvement Tasks

## Critical Issues (High Priority)

### ~~1. Flag Binding Bug~~
**Issue**: `--stdin` and `--json` flags aren't bound to Viper, causing them to always return false
**Location**: `cmd/commands/scan.go:82-91`
**Fix**: Add missing bindings:
```go
_ = viper.BindPFlag("stdin", scanCmd.Flags().Lookup("stdin"))
_ = viper.BindPFlag("json", scanCmd.Flags().Lookup("json"))
```
Status: Fixed in Phase 1 (bound flags in cmd/commands/scan.go; added unit test)

### ~~2. Version Variables Mismatch~~
**Issue**: Makefile sets LDFLAGS for `main.version/commit/buildTime` but variables live in `cmd/commands`
**Location**: `Makefile` and `cmd/commands/version.go`
**Fix**: Align LDFLAGS to:
```makefile
-X github.com/lucchesi-sec/portscan/cmd/commands.version=$(VERSION)
-X github.com/lucchesi-sec/portscan/cmd/commands.commit=$(COMMIT)
-X github.com/lucchesi-sec/portscan/cmd/commands.buildDate=$(BUILD_TIME)
```
Status: Fixed in Phase 1 (updated Makefile LDFLAGS)

### ~~3. Memory-Intensive JSON Export~~
**Issue**: JSON exporter collects all results in memory before encoding
**Location**: `pkg/exporter/json.go`
**Fix**: Replaced with NDJSON (newline-delimited JSON) streaming; added `--json-array` and `--json-object` modes for array and scan_info wrapper without buffering.
Status: Done (NDJSON default, array/object modes, tests added)

### ~~3a. JSON Schema Mismatch With README~~
**Issue**: Exported JSON used Go struct defaults.
**Location**: `pkg/exporter/json.go`, README examples
**Fix**: Implemented DTO with fields: `host`, `port`, `state`, `service`, `banner`, `response_time_ms`; added `scan_info` in object mode; updated README; added tests.
Status: Done (follow-up: propagate write errors explicitly)

## Scanner Engine Issues (Medium Priority)

### 4. Inaccurate Progress Reporting
**Issue**: Progress estimated from elapsed time × rate limit, not actual completions
**Location**: `internal/core/scanner.go` - progressReporter
**Fix**: Track atomic completion counter and compute rate over sliding windows

### 5. Overly Complex Concurrency Model
**Issue**: Workers spawn extra goroutines per job while also using workerPool
**Location**: `internal/core/scanner.go`
**Fix**: Simplify to N workers consuming jobs inline

### 6. Unused Retry Configuration
**Issue**: `Config.MaxRetries` exists but isn't implemented
**Location**: `internal/core/scanner.go`
**Fix**: Implement bounded retry with jittered backoff on timeouts

### 6a. Mixed-Type Results Channel
**Issue**: `Scanner.results` is `chan interface{}` carrying both `ResultEvent` and `ProgressEvent`, complicating consumers and interleaving progress with exports.
**Location**: `internal/core/scanner.go`, UI consumers, exporters
**Fix**: Use separate channels (results/progress) or a typed envelope (`type Event struct { Kind string; ... }`) to avoid interface{} and make exporters/UI simpler and testable.

## Target Parsing Issues

### 7. CIDR Not Expanded
**Issue**: `validateTarget` is a stub; CIDR input fails in net.Dial
**Location**: `cmd/commands/scan.go:274-281`
**Fix**: Implement proper CIDR expansion and multi-host scanning

### 8. No stdin List Support
**Issue**: stdin currently reads single target, not list
**Location**: `cmd/commands/scan.go:124-137`
**Fix**: Parse stdin as newline-delimited targets and deduplicate

### 8a. CLI Flag Duplication and Semantics
**Issue**: `--quiet` exists as both a root persistent flag (suppress logs) and a scan flag ("only show open ports"), creating ambiguity and potential conflicts.
**Location**: `cmd/commands/root.go`, `cmd/commands/scan.go`
**Fix**: Rename scan-level flag to `--open-only` (or `--hide-closed`) and bind via Viper; plumb into UI to filter rows. Keep root `--quiet` for logging only.

## Data Quality Issues

### 9. Duplicate Ports in Profiles
**Issue**: Profiles contain duplicate ports (e.g., 11211 appears multiple times)
**Location**: `pkg/profiles/profiles.go`
**Fix**: Deduplicate ports at selection time and update tests

### 10. CSV Export Error Handling
**Issue**: No error checking on csvWriter.Write or Flush
**Location**: `pkg/exporter/csv.go`
**Fix**: Add error handling and periodic flush with sanitization
**Also**: Return and handle `Close()` errors in callers; sanitize or truncate banners to avoid CSV injection.

## UI/UX Issues

### 11. Multiple UI Models
**Issue**: Three different UI models exist (Model, ScanUI, EnhancedModel)
**Location**: `internal/ui/`
**Fix**: Consolidate to single model with all features

### 12. Hidden Closed Ports
**Issue**: Table only shows open ports; toggle exists but not wired
**Location**: `internal/ui/scan_ui_view.go`
**Fix**: Wire 'v' toggle to show all port states
 
### 12a. Unbounded Result Growth in TUI
**Issue**: UIs store all `ResultEvent`s in memory; large scans can bloat memory.
**Location**: `internal/ui/*`
**Fix**: Cap retained history (e.g., ring buffer) and keep a compact aggregate for stats; optionally stream to disk when exporting.

## Configuration Issues

### ~~13. Invalid go.mod Directive~~
**Issue**: `go 1.24.5` should be `go 1.24` (patch in toolchain directive)
**Location**: `go.mod`
**Fix**: Change to `go 1.24` and optionally add `toolchain go1.24.5`
Status: Fixed in Phase 1 (go.mod updated to go 1.24 + toolchain go1.24.5)

### 13a. Unused/Undocumented Config Keys
**Issue**: README and `config init` mention `dns` and `log_json`, but code does not consume `dns` settings and only partially uses logging flags.
**Location**: `cmd/commands/config.go` (init content), `pkg/config`, README
**Fix**: Either implement DNS resolver options and structured logging (honor `log_json`, `no_color`, `quiet`) or remove from config/docs to avoid confusion.

### 13b. Dependency Versions
**Issue**: Some `go.mod` deps use pseudo-versions tracking commits; this can hurt reproducibility.
**Location**: `go.mod`
**Fix**: Prefer stable tagged releases where possible; run `go mod tidy -go=1.24` and pin to tags to stabilize builds.

## CI/CD Improvements

### 14. Optimize GitHub Actions
**Issue**: No caching, manual golangci-lint installation
**Location**: `.github/workflows/`
**Fix**: 
- Add `actions/cache` for Go modules
- Replace curl install with `golangci/golangci-lint-action`

### 14a. Add CodeQL and Caching
**Issue**: No code scanning and limited module caching.
**Location**: `.github/workflows/*.yml`
**Fix**: Add GitHub CodeQL workflow; use `actions/cache` for `~/go/pkg/mod` and build cache across jobs/matrix.

### 15. Multi-stage Dockerfile
**Issue**: Current Dockerfile expects prebuilt binary
**Location**: `Dockerfile`
**Fix**: Implement multi-stage build for dev-friendly container
**Also**: Pin base image (e.g., `alpine:3.20`), add non-root user, and set a read-only filesystem with necessary capabilities.

### 15a. Repo Hygiene
**Issue**: Build artifacts and local tool data are committed (e.g., `coverage.out`, `.crush/`).
**Location**: repo root
**Fix**: Remove tracked artifacts from VCS and extend `.gitignore` as needed; enforce via CI (fail on dirty tree after build/test).

## Testing Gaps

### 16. Missing Critical Tests
**Areas needing tests**:
- NDJSON line-by-line export
- CSV flushing and error handling
- Flag wiring with --dry-run
- Profile deduplication
- Scanner retry/backoff behavior
- Cancellation handling
 - JSON schema conformance (fields, units, metadata)
 - UI table filtering toggles and memory caps

## Documentation

### 17. README Misalignment
**Issue**: README promises unimplemented features (CIDR, Prometheus, audit logging, privilege dropping)
**Location**: `README.md`
**Fix**: Either implement features or update documentation to match reality

### 17a. CLI Options Misalignment
**Issue**: `--output prometheus` and non-interactive `table` are documented but not implemented.
**Location**: README, `cmd/commands/scan.go`
**Fix**: Implement exporters or remove from docs; unify `--json` boolean with `--output json` (deprecate one and document).

### 17b. Minor Comment Drift
**Issue**: `getOptimalWorkerCount` comment says “2x CPU cores” but code uses `cores * 50` (capped).
**Location**: `cmd/commands/scan.go`
**Fix**: Align implementation and comment; document reasoning and caps.

## Advanced Features (Low Priority)

### 18. Banner Grabbing Enhancement
**Current**: Simple 1s deadline read
**Improvement**: Protocol-specific probes (HTTP GET, TLS handshake, SSH banner)

### 19. Rate Limiter Optimization
**Note**: time.Tick is GC-safe in Go 1.23+, but NewTicker with Stop is clearer

### 20. Security Enhancements
- Randomize port scan order (detection evasion)
- Sanitize and cap banner length
- Add ephemeral port exhaustion protection
- Document ulimit guidance

## Implementation Order

### Phase 1: Critical Fixes (Immediate)
1. Fix flag bindings (#1)
2. Fix Makefile LDFLAGS (#2)
3. Update go.mod directive (#13)

### Phase 2: Core Functionality (Week 1)
4. Replace JSON with NDJSON (#3)
5. Fix progress accuracy (#4)
6. Implement CIDR expansion (#7)
7. Deduplicate profile ports (#9)

### Phase 3: Robustness (Week 2)
8. Simplify concurrency model (#5)
9. Implement retry logic (#6)
10. Add CSV error handling (#10)
11. Fix stdin list support (#8)

### Phase 4: Polish (Week 3)
12. Consolidate UI models (#11)
13. Wire port state toggle (#12)
14. Update CI/CD (#14, #15)
15. Add comprehensive tests (#16)

### Phase 5: Documentation & Features
16. Update README (#17)
17. Enhance banner grabbing (#18)
18. Add security features (#20)

## Notes

- Go 1.23+ makes time.Tick GC-safe (no leak concern)
- Use Viper's BindPFlag for consistent flag handling
- NDJSON is idiomatic for streaming large JSON outputs
- Consider CodeQL addition for security scanning
