# Security Scanning Guide

This document describes the security scanning and vulnerability management process for the portscan project.

## Overview

The project uses multiple security tools to ensure code quality and identify vulnerabilities:

- **govulncheck**: Official Go vulnerability scanner from the Go Security Team
- **gosec**: Static security analyzer for Go code
- **golangci-lint**: Meta-linter that includes security-focused linters

## Quick Start

### Run Vulnerability Scan

```bash
# Run vulnerability check only
make vulncheck

# Run all security checks (govulncheck + gosec)
make security
```

### Install Security Tools

If you haven't set up your development environment yet:

```bash
make dev-setup
```

This installs:
- `govulncheck` - Vulnerability scanner
- `gosec` - Security analyzer
- `golangci-lint` - Code linter
- Other development tools

## Vulnerability Scanning with govulncheck

### What is govulncheck?

`govulncheck` is the official Go vulnerability scanner maintained by the Go Security Team. It:

- Checks your code and dependencies against the Go Vulnerability Database
- Analyzes which vulnerable functions are actually called (not just present)
- Provides accurate, actionable vulnerability reports
- Integrates seamlessly with the Go toolchain

### Usage

#### Command Line

```bash
# Scan entire project
go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Or use the Makefile target
make vulncheck
```

#### Output Example

When vulnerabilities are found:
```
Scanning your code and 123 packages across 45 dependent modules for known vulnerabilities...

Vulnerability #1: GO-2023-1234
    A vulnerability in package X allows Y

  More info: https://pkg.go.dev/vuln/GO-2023-1234

  Module: github.com/example/package
    Found in: github.com/example/package@v1.2.3
    Fixed in: github.com/example/package@v1.2.4

  Call stacks in your code:
    internal/core/scanner.go:123:45: your.function calls vulnerable.function
```

When no vulnerabilities are found:
```
No vulnerabilities found.
```

### Interpreting Results

govulncheck reports are divided into:

1. **Direct vulnerabilities**: In packages you directly import
2. **Indirect vulnerabilities**: In transitive dependencies
3. **Called vulnerabilities**: Functions your code actually calls
4. **Unreached vulnerabilities**: Present but not called by your code

**Priority**: Focus on fixing called vulnerabilities first, as these pose actual risk to your application.

## Static Security Analysis with gosec

### What is gosec?

`gosec` inspects Go source code for common security problems such as:

- SQL injection
- Command injection
- Weak cryptography
- File path traversal
- Unsafe operations
- And more...

### Usage

```bash
# Run gosec
gosec ./...

# Run with specific rules
gosec -include=G401,G501 ./...

# Generate report
gosec -fmt=json -out=results.json ./...
```

### Common Issues and Fixes

#### G104: Unchecked errors
```go
// Bad
file.Close()

// Good
defer func() {
    if err := file.Close(); err != nil {
        log.Printf("failed to close file: %v", err)
    }
}()
```

#### G401: Weak cryptographic hash
```go
// Bad
hash := md5.Sum(data)

// Good
hash := sha256.Sum256(data)
```

## Continuous Integration

### GitHub Actions

Security scans run automatically on:
- Every push to `main`
- Every pull request
- Scheduled daily scans

See `.github/workflows/ci.yml` for configuration.

### Local Pre-commit Checks

Set up pre-commit hooks to catch issues early:

```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install

# Run manually
pre-commit run --all-files
```

## Dependency Management

### Updating Dependencies

```bash
# Check for available updates
go list -u -m all

# Update all dependencies
go get -u ./...
go mod tidy

# Update specific package
go get -u github.com/example/package@latest

# Run security scan after updates
make vulncheck
make test
```

### Security-Critical Packages

Pay special attention when updating:

- `golang.org/x/crypto` - Cryptographic primitives
- `golang.org/x/net` - Network protocols
- `golang.org/x/sys` - System calls
- `golang.org/x/text` - Text processing
- Authentication/authorization packages
- Database drivers
- HTTP clients and servers

### Best Practices

1. **Regular Updates**: Update dependencies monthly or when vulnerabilities are announced
2. **Test After Updates**: Always run full test suite after updating
3. **Review Changes**: Check release notes for breaking changes
4. **Pin Versions**: Use exact versions in go.mod for reproducible builds
5. **Audit New Dependencies**: Review security history before adding new dependencies

## Vulnerability Response

### When a Vulnerability is Found

1. **Assess Severity**
   - Critical: Exploitable in production, immediate action required
   - High: Potential for exploitation, fix within 7 days
   - Medium: Limited impact, fix within 30 days
   - Low: Minimal risk, fix in next release cycle

2. **Determine Impact**
   - Is the vulnerable code path actually used?
   - Does it affect production deployments?
   - Are there existing mitigations?

3. **Plan Remediation**
   - Update to patched version if available
   - Apply workarounds if patch not available
   - Consider replacing vulnerable dependency
   - Document why vulnerability doesn't apply (if applicable)

4. **Verify Fix**
   ```bash
   # After updating
   make vulncheck
   make test
   make build
   ```

5. **Deploy**
   - Follow normal deployment process
   - Monitor for issues
   - Update documentation

### Reporting Vulnerabilities

If you discover a security vulnerability in this project:

1. **DO NOT** open a public issue
2. Email security@lucchesi-sec.com with:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)
3. Allow 90 days for response and fix

## Automated Scanning Schedule

- **Pre-commit**: gosec on changed files
- **CI Pipeline**: Full scan on every PR and push
- **Daily**: Automated vulnerability scans
- **Monthly**: Dependency update review

## Resources

- [Go Vulnerability Database](https://vuln.go.dev/)
- [govulncheck Documentation](https://go.dev/doc/security/vuln/)
- [gosec Rules](https://github.com/securego/gosec#available-rules)
- [OWASP Go Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Go_Security_Cheat_Sheet.html)
- [Go Security Policy](https://go.dev/security/policy)

## Version History

- **2025-10-25**: Initial security scanning documentation
  - Added govulncheck integration
  - Updated Makefile with security targets
  - Enhanced CI/CD pipeline
  - Documented vulnerability response process
