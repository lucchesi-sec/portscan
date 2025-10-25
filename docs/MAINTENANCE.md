# Maintenance Procedures

## Table of Contents

1. [Dependency Management](#1-dependency-management)
2. [Vulnerability Scanning](#2-vulnerability-scanning)
3. [Release Process](#3-release-process)
4. [Version Compatibility](#4-version-compatibility)
5. [Monitoring and Health Checks](#5-monitoring-and-health-checks)
6. [Backup and Recovery](#6-backup-and-recovery)
7. [Performance Tuning](#7-performance-tuning)
8. [Troubleshooting](#8-troubleshooting)

---

## 1. Dependency Management

### Current Dependencies

The project uses Go modules for dependency management. Key dependencies:

```go
require (
    github.com/charmbracelet/bubbles v0.21.0      // TUI components
    github.com/charmbracelet/bubbletea v1.3.6     // TUI framework
    github.com/charmbracelet/lipgloss v1.1.0      // Terminal styling
    github.com/go-playground/validator/v10 v10.27.0 // Config validation
    github.com/spf13/cobra v1.9.1                 // CLI framework
    github.com/spf13/viper v1.20.1                // Configuration
)
```

### Updating Dependencies

#### Routine Updates (Monthly)

```bash
# Check for available updates
go list -u -m all

# Update all dependencies to latest minor/patch versions
go get -u ./...

# Update specific dependency
go get -u github.com/spf13/cobra@latest

# Tidy up go.mod and go.sum
go mod tidy

# Verify updates
go mod verify
```

#### Major Version Updates (Quarterly)

```bash
# Update to specific major version
go get github.com/charmbracelet/bubbletea@v2

# Check breaking changes
git diff go.mod

# Run full test suite
make test
make test-race
make test-integration

# Update code for breaking changes if needed
```

### Dependency Update Checklist

- [ ] Review CHANGELOG for breaking changes
- [ ] Update code for API changes
- [ ] Run full test suite
- [ ] Run benchmarks to check for performance regressions
- [ ] Test on all supported platforms (Linux, macOS, Windows)
- [ ] Update documentation if public API changed
- [ ] Create PR with dependency updates
- [ ] Monitor CI for issues

### Dependency Pinning

For critical dependencies, pin to specific versions:

```go
// go.mod
require (
    github.com/charmbracelet/bubbletea v1.3.6  // Pinned for stability
)
```

### Vendoring (Optional)

For reproducible builds:

```bash
# Enable vendoring
go mod vendor

# Verify vendor directory
go mod verify

# Build using vendor
go build -mod=vendor cmd/main.go
```

---

## 2. Vulnerability Scanning

### Automated Scanning

#### govulncheck

```bash
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan for vulnerabilities
govulncheck ./...

# Example output:
# Vulnerability #1: GO-2024-1234
#   Package: golang.org/x/net
#   Versions: < v0.17.0
#   Details: HTTP/2 DoS vulnerability
#   Fix: Update to v0.17.0 or later
```

#### GitHub Dependabot

Configured in `.github/dependabot.yml`:

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10
    labels:
      - "dependencies"
      - "security"
```

### Manual Vulnerability Checks

```bash
# Check Go version for security updates
go version

# Check for security advisories
curl -s https://golang.org/security/ | grep -i "go1.24"

# Review NIST database
# Visit: https://nvd.nist.gov/
```

### Vulnerability Response Process

**1. Detection**
- Automated scan finds vulnerability
- Dependabot creates PR
- Security alert from GitHub

**2. Assessment**
```bash
# Check if vulnerability affects our usage
govulncheck -json ./... | jq '.Vulns[] | {package, symbol, message}'

# Review affected code paths
grep -r "vulnerable_function" .
```

**3. Remediation**
```bash
# Update affected dependency
go get -u vulnerable/package@safe-version

# Test fixes
make test
make test-race

# Verify vulnerability is resolved
govulncheck ./...
```

**4. Communication**
- Create security advisory (if needed)
- Update CHANGELOG
- Notify users via GitHub release
- Update documentation

### Vulnerability Scanning Schedule

| Activity | Frequency | Tool |
|----------|-----------|------|
| Automated scan | Daily (CI) | govulncheck |
| Dependency updates | Weekly | Dependabot |
| Manual review | Monthly | Manual + NIST |
| Security audit | Quarterly | External review |

---

## 3. Release Process

### Semantic Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (v1.0.0 → v2.0.0): Breaking changes
- **MINOR** (v1.0.0 → v1.1.0): New features, backward compatible
- **PATCH** (v1.0.0 → v1.0.1): Bug fixes, backward compatible

### Release Checklist

#### Pre-Release (1 week before)

- [ ] Review open issues and PRs
- [ ] Update dependencies
- [ ] Run full test suite on all platforms
- [ ] Run vulnerability scan
- [ ] Update CHANGELOG.md
- [ ] Update version numbers
- [ ] Review documentation for accuracy

#### Release Day

**1. Prepare Release Branch**

```bash
# Create release branch
git checkout -b release/v0.3.0

# Update version in code
# Update README.md badges
# Update CHANGELOG.md

# Commit changes
git add .
git commit -m "chore: prepare v0.3.0 release"

# Push branch
git push origin release/v0.3.0
```

**2. Create Release PR**

```bash
# Create PR to main
gh pr create \
  --title "Release v0.3.0" \
  --body "Release notes:

## New Features
- Feature 1
- Feature 2

## Bug Fixes
- Fix 1
- Fix 2

## Breaking Changes
- None

## Upgrade Notes
- No special steps required
"
```

**3. Tag Release**

```bash
# After PR is merged
git checkout main
git pull origin main

# Create annotated tag
git tag -a v0.3.0 -m "Release v0.3.0

New Features:
- Feature 1
- Feature 2

Bug Fixes:
- Fix 1
- Fix 2
"

# Push tag
git push origin v0.3.0
```

**4. Build and Publish**

```bash
# GoReleaser will automatically:
# - Build binaries for all platforms
# - Create GitHub release
# - Upload artifacts
# - Generate checksums

# Manual verification
gh release view v0.3.0
```

**5. Post-Release**

- [ ] Verify release artifacts
- [ ] Test installation from release
- [ ] Update documentation site
- [ ] Announce on GitHub Discussions
- [ ] Update package managers (if applicable)

### Release Configuration

**GoReleaser** (`.goreleaser.yaml`):

```yaml
project_name: portscan

builds:
  - main: ./cmd
    binary: portscan
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - format: tar.gz
    name_template: '{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}'
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'
```

### Hotfix Process

For critical bugs in production:

```bash
# Create hotfix branch from tag
git checkout -b hotfix/v0.3.1 v0.3.0

# Fix the issue
# ... make changes ...

# Commit and tag
git commit -m "fix: critical bug in scanner"
git tag v0.3.1
git push origin hotfix/v0.3.1
git push origin v0.3.1

# Merge back to main
git checkout main
git merge hotfix/v0.3.1
git push origin main
```

---

## 4. Version Compatibility

### Go Version Support

**Current Policy:**
- Support latest 2 major Go versions
- Current: Go 1.24 and 1.23
- Drop support for Go 1.23 when Go 1.25 is released

### Compatibility Matrix

| portscan Version | Go Version | OS Support | Architectures |
|------------------|------------|------------|---------------|
| v0.3.x           | 1.24+      | Linux, macOS, Windows | amd64, arm64 |
| v0.2.x           | 1.23+      | Linux, macOS, Windows | amd64, arm64 |
| v0.1.x           | 1.22+      | Linux, macOS | amd64 |

### Backward Compatibility

**Configuration Files:**
```bash
# v0.1.x config
rate: 5000
ports: "1-1024"
timeout_ms: 200

# v0.2.x config (adds new fields, maintains compatibility)
rate: 7500
ports: "1-1024"
timeout_ms: 200
protocol: tcp        # New field with default
udp_worker_ratio: 0.5  # New field with default

# v0.3.x config (deprecates old fields)
rate: 7500
ports: "1-1024"
timeout_ms: 200
protocol: tcp
max_retries: 2       # New field
# timeout_seconds deprecated, use timeout_ms
```

**API Compatibility:**

```go
// v0.1.x
scanner := core.NewScanner(&core.Config{
    Workers: 100,
    Timeout: 200 * time.Millisecond,
})

// v0.2.x (maintains v0.1.x compatibility)
scanner := core.NewScanner(&core.Config{
    Workers:   100,
    Timeout:   200 * time.Millisecond,
    RateLimit: 7500,  // New optional field
})

// v0.3.x (breaking change - requires migration)
scanner := core.NewScanner(&core.ScannerConfig{  // Renamed
    Workers:    100,
    Timeout:    200 * time.Millisecond,
    RateLimit:  7500,
    MaxRetries: 2,  // New required field
})
```

### Migration Guides

For breaking changes, provide migration guides:

**docs/MIGRATION_v0.2_to_v0.3.md:**

```markdown
# Migration Guide: v0.2 to v0.3

## Configuration Changes

### Config struct renamed
**Before (v0.2):**
```go
cfg := &core.Config{Workers: 100}
```

**After (v0.3):**
```go
cfg := &core.ScannerConfig{Workers: 100}
```

### New required fields
- `MaxRetries`: Default value = 2

## CLI Changes

### Renamed flags
- `--timeout` → `--timeout-ms` (more explicit)
- `--workers` → `--concurrency` (clearer naming)

### Deprecated flags
- `--fast` (removed, use `--rate 10000` instead)

## Code Changes

Run automated migration:
```bash
go run scripts/migrate_v0.2_to_v0.3.go
```
```

---

## 5. Monitoring and Health Checks

### Application Metrics

**Key Metrics to Monitor:**

```go
// Scan performance
- Scan rate (packets/second)
- Average latency per port
- Worker utilization
- Memory usage
- Goroutine count

// Error rates
- Connection timeouts
- Connection refused
- DNS resolution failures

// Resource usage
- CPU utilization
- Memory heap size
- File descriptor count
- Network bandwidth
```

### Health Check Script

```bash
#!/bin/bash
# scripts/health_check.sh

echo "=== Port Scanner Health Check ==="

# Check binary exists
if [ ! -f "./bin/portscan" ]; then
    echo "ERROR: Binary not found"
    exit 1
fi

# Check version
VERSION=$(./bin/portscan version --short)
echo "Version: $VERSION"

# Basic functionality test
echo "Testing basic scan..."
OUTPUT=$(./bin/portscan scan 127.0.0.1 --ports 22,80,443 --timeout 500 2>&1)
if [ $? -eq 0 ]; then
    echo "✓ Basic scan works"
else
    echo "✗ Basic scan failed"
    echo "$OUTPUT"
    exit 1
fi

# Check resource usage
echo "Testing resource usage..."
/usr/bin/time -v ./bin/portscan scan 127.0.0.1 --ports 1-1000 2>&1 | grep -E "(Maximum resident|User time|System time)"

echo "=== Health Check Complete ==="
```

### Performance Baselines

Establish performance baselines for regression detection:

```bash
# Run benchmarks and save results
go test -bench=. ./internal/core > benchmarks/baseline_v0.3.0.txt

# Compare against baseline
go test -bench=. ./internal/core > benchmarks/current.txt
benchstat benchmarks/baseline_v0.3.0.txt benchmarks/current.txt
```

---

## 6. Backup and Recovery

### Configuration Backup

```bash
# Backup user configuration
cp ~/.portscan.yaml ~/.portscan.yaml.backup.$(date +%Y%m%d)

# Restore from backup
cp ~/.portscan.yaml.backup.20250101 ~/.portscan.yaml
```

### Source Code Backup

```bash
# Ensure all branches are pushed to remote
git push --all origin
git push --tags origin

# Create local archive
git archive --format=tar.gz --output=portscan-backup-$(date +%Y%m%d).tar.gz HEAD
```

### Disaster Recovery

**Scenario: Corrupted local repository**

```bash
# Clone fresh copy
git clone https://github.com/lucchesi-sec/portscan.git portscan-recovery
cd portscan-recovery

# Verify integrity
git fsck
make test
```

**Scenario: Lost release artifacts**

```bash
# Rebuild from tag
git checkout v0.3.0
make build-all

# Or use GoReleaser
goreleaser release --skip-publish --snapshot --clean
```

---

## 7. Performance Tuning

### System-Level Tuning

**Linux - Increase File Descriptor Limits:**

```bash
# Check current limits
ulimit -n

# Increase for current session
ulimit -n 65535

# Permanent change (add to /etc/security/limits.conf)
*  soft  nofile  65535
*  hard  nofile  65535
```

**Linux - Optimize Network Stack:**

```bash
# Increase ephemeral port range
sudo sysctl -w net.ipv4.ip_local_port_range="1024 65535"

# Increase socket buffer sizes
sudo sysctl -w net.core.rmem_max=16777216
sudo sysctl -w net.core.wmem_max=16777216

# Enable TCP Fast Open
sudo sysctl -w net.ipv4.tcp_fastopen=3
```

**macOS - Increase File Descriptor Limits:**

```bash
# Check current limits
launchctl limit maxfiles

# Increase (requires restart)
sudo launchctl limit maxfiles 65536 200000
```

### Application Tuning

**Optimal Worker Count:**

```bash
# Formula: workers = 2 × CPU cores
WORKERS=$(( $(nproc) * 2 ))

# For CPU-bound tasks (banner parsing)
WORKERS=$(( $(nproc) ))

# For I/O-bound tasks (network scanning)
WORKERS=$(( $(nproc) * 4 ))

# Use in scan
portscan scan target --workers $WORKERS
```

**Rate Limiting:**

```bash
# Conservative (stealth)
portscan scan target --rate 100

# Balanced (default)
portscan scan target --rate 7500

# Aggressive (LAN only)
portscan scan target --rate 50000
```

### Profiling Production Issues

```bash
# Enable profiling in production
export PORTSCAN_PPROF=:6060

# In another terminal, collect profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Analyze
(pprof) top
(pprof) web
```

---

## 8. Troubleshooting

### Common Issues

#### Issue: High Memory Usage

**Symptoms:**
- Memory usage exceeds 500MB
- OOM killer terminating process

**Diagnosis:**
```bash
# Check memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Check result buffer size
grep result_buffer_size ~/.portscan.yaml
```

**Solution:**
```bash
# Reduce result buffer size
portscan scan target --ui.result-buffer-size 1000

# Or in config
echo "ui:
  result_buffer_size: 1000" >> ~/.portscan.yaml
```

#### Issue: Slow Scan Rate

**Symptoms:**
- Scan rate below expected (< 1000 pps)
- Long scan times

**Diagnosis:**
```bash
# Check system limits
ulimit -n  # Should be > 1024

# Check network latency
ping -c 10 target

# Monitor goroutines
curl http://localhost:6060/debug/pprof/goroutine?debug=1 | wc -l
```

**Solution:**
```bash
# Increase workers
portscan scan target --workers 200

# Increase rate limit
portscan scan target --rate 10000

# Decrease timeout
portscan scan target --timeout 100
```

#### Issue: "Too Many Open Files"

**Symptoms:**
```
Error: dial tcp: socket: too many open files
```

**Diagnosis:**
```bash
# Check file descriptor limit
ulimit -n

# Check current FD usage
lsof -p $(pgrep portscan) | wc -l
```

**Solution:**
```bash
# Increase FD limit (temporary)
ulimit -n 65535

# Increase FD limit (permanent)
# Edit /etc/security/limits.conf (Linux)
# Or /Library/LaunchDaemons/limit.maxfiles.plist (macOS)

# Reduce concurrent workers
portscan scan target --workers 50
```

### Debug Mode

Enable verbose logging:

```bash
# Set environment variable
export PORTSCAN_DEBUG=1

# Run scan with verbose output
portscan scan target --verbose

# Save logs
portscan scan target --verbose 2>&1 | tee scan.log
```

### Getting Help

**Internal Issues:**
```bash
# Check internal documentation
make docs

# Generate diagnostic report
./scripts/diagnostic_report.sh > diagnostics.txt
```

**Community Support:**
- GitHub Issues: https://github.com/lucchesi-sec/portscan/issues
- Discussions: https://github.com/lucchesi-sec/portscan/discussions
- Discord: [Link if available]

---

## Maintenance Schedule

| Task | Frequency | Owner | Notes |
|------|-----------|-------|-------|
| Dependency updates | Weekly | Dependabot | Auto-merge if tests pass |
| Vulnerability scan | Daily | CI/CD | Alert on high/critical |
| Performance benchmarks | Per PR | CI/CD | Flag > 10% regression |
| Documentation review | Monthly | Team | Keep current with code |
| Major dependency updates | Quarterly | Maintainers | Test thoroughly |
| Security audit | Annually | External | Professional review |

---

## Emergency Contacts

**Critical Issues:**
- Security vulnerabilities: security@lucchesi-sec.com
- Production outages: [Internal Slack channel]
- Maintainers: See MAINTAINERS.md

**Response Times:**
- **Critical (P0)**: 2 hours
- **High (P1)**: 1 business day
- **Medium (P2)**: 1 week
- **Low (P3)**: Best effort

---

**Last Updated:** 2025-10-25  
**Next Review:** 2025-11-25  
**Maintainer:** Enzo Lucchesi (@lucchesi-sec)
