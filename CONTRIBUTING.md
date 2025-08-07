# Contributing to Port Scanner

We love your input! We want to make contributing to Port Scanner as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## 📋 Table of Contents

- [Development Process](#development-process)
- [Getting Started](#getting-started)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Community Guidelines](#community-guidelines)

## 🚀 Development Process

We use GitHub to host code, track issues and feature requests, and accept pull requests.

### Quick Start for New Contributors

1. **Read our [Development Guide](DEV_GUIDE.md)** - Essential for beginners
2. **Fork the repository** on GitHub
3. **Create a feature branch** from `main`
4. **Make your changes** following our coding standards
5. **Test your changes** thoroughly
6. **Submit a pull request**

## 🛠️ Getting Started

### Prerequisites

- Go 1.24 or higher
- Git
- Make (for build automation)

### Setup Development Environment

```bash
# Clone your fork
git clone https://github.com/YOUR-USERNAME/portscan.git
cd portscan

# Set up development tools
make dev-setup

# Install dependencies
make deps

# Verify setup
make test
```

## 📝 Pull Request Process

### Before You Start

1. **Check existing issues** to avoid duplicate work
2. **Create an issue** for new features or significant changes
3. **Discuss your approach** in the issue before implementing

### Creating a Pull Request

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our [coding standards](#coding-standards)

3. **Write or update tests** for your changes

4. **Run the full test suite**
   ```bash
   make ci
   ```

5. **Update documentation** if needed

6. **Commit with conventional messages**
   ```bash
   git commit -m "feat: add UDP scanning support"
   ```

7. **Push your branch and create a PR**
   ```bash
   git push origin feature/your-feature-name
   ```

### Pull Request Guidelines

- **Title**: Use conventional commit format (`feat:`, `fix:`, `docs:`, etc.)
- **Description**: Clearly explain what and why
- **Link related issues**: Use "Closes #123" or "Fixes #123"
- **Keep PRs focused**: One feature/fix per PR
- **Update tests**: Ensure test coverage doesn't decrease
- **Update docs**: Keep documentation current

### PR Review Process

1. **Automated checks** must pass (CI, tests, linting)
2. **Code review** by maintainers
3. **Address feedback** promptly
4. **Maintainer approval** required before merge
5. **Squash and merge** for clean history

## 🐛 Issue Guidelines

### Reporting Bugs

Use our bug report template and include:

- **Go version**: `go version`
- **OS and architecture**
- **Steps to reproduce**
- **Expected vs actual behavior**
- **Error messages or logs**
- **Configuration details**

### Feature Requests

Use our feature request template and include:

- **Problem description**: What are you trying to solve?
- **Proposed solution**: How should it work?
- **Alternatives considered**: Other approaches?
- **Use cases**: Real-world examples

### Security Issues

**DO NOT** create public issues for security vulnerabilities.

Email: security@lucchesi-sec.com

## 📊 Coding Standards

### Go Style Guide

We follow standard Go conventions:

- **gofmt/gofumpt** for formatting
- **golangci-lint** for linting
- **Effective Go** guidelines
- **Go Code Review Comments**

### Code Organization

```go
// Package documentation
package scanner

import (
    // Standard library first
    "context"
    "fmt"
    "time"
    
    // Third-party packages
    "github.com/spf13/cobra"
    
    // Local packages
    "github.com/lucchesi-sec/portscan/internal/core"
)
```

### Naming Conventions

- **Packages**: Short, lowercase, no underscores
- **Variables**: camelCase, descriptive names
- **Functions**: camelCase, verb-noun pattern
- **Constants**: CamelCase or UPPER_CASE for exported
- **Interfaces**: -er suffix (Scanner, Parser)

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to scan port %d: %w", port, err)
}

// Good: Return meaningful errors
func validatePort(port int) error {
    if port < 1 || port > 65535 {
        return ErrInvalidPort
    }
    return nil
}
```

### Comments

- **Exported items**: Must have doc comments
- **Complex logic**: Explain the why, not the what
- **TODOs**: Include issue number or explanation

```go
// Scanner performs network port scanning with configurable rate limiting.
// It supports TCP scanning with banner grabbing and real-time progress updates.
type Scanner struct {
    // rateLimiter controls scan rate to prevent network congestion
    rateLimiter *rate.Limiter
}
```

## 🧪 Testing

### Test Requirements

- **Unit tests** for all new functions
- **Integration tests** for new features
- **Benchmark tests** for performance-critical code
- **Minimum 80% coverage** for new code

### Test Structure

```go
func TestPortParser(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected []int
        wantErr  bool
    }{
        {
            name:     "single port",
            input:    "80",
            expected: []int{80},
            wantErr:  false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := ParsePorts(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParsePorts() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            // Verify result...
        })
    }
}
```

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Race detection
make test-race

# Benchmarks
make benchmark

# Specific package
go test ./internal/core

# Specific test
go test ./internal/core -run TestPortParser
```

## 📚 Documentation

### Code Documentation

- **Godoc comments** for all exported items
- **README updates** for new features
- **Architecture docs** for significant changes

### User Documentation

- **CLI help text** for new commands/flags
- **Configuration examples**
- **Usage examples** in README

### Documentation Style

- **Clear and concise** language
- **Practical examples**
- **Up-to-date** with code changes
- **Accessible** to beginners

## 🤝 Community Guidelines

### Code of Conduct

We follow the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).

### Communication

- **Be respectful** and constructive
- **Ask questions** when unclear
- **Help others** learn and grow
- **Share knowledge** and experiences

### Issue Discussions

- **Stay on topic**
- **Provide constructive feedback**
- **Be patient** with response times
- **Search existing issues** before creating new ones

## 🏷️ Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):

- **Major** (v1.0.0): Breaking changes
- **Minor** (v0.1.0): New features, backward compatible
- **Patch** (v0.0.1): Bug fixes, backward compatible

### Release Workflow

1. **Create release branch**: `git checkout -b release/v0.3.0`
2. **Update version**: Update relevant files
3. **Update CHANGELOG**: Add release notes
4. **Create PR**: For release branch
5. **Tag release**: After merge to main
6. **Publish**: Automated via GitHub Actions

## ❓ Getting Help

### Where to Get Help

- **GitHub Discussions**: General questions and discussions
- **GitHub Issues**: Bug reports and feature requests
- **Development Guide**: [DEV_GUIDE.md](DEV_GUIDE.md)
- **Code Review**: Pull request comments

### Before Asking for Help

1. **Search existing issues** and discussions
2. **Read the documentation**
3. **Check the Development Guide**
4. **Prepare a minimal example**

## 🎯 Good First Issues

New contributors should look for issues labeled:

- `good first issue`: Perfect for beginners
- `help wanted`: Need community assistance
- `documentation`: Improve docs
- `bug`: Fix reported issues
- `enhancement`: Small improvements

## 📈 Contribution Recognition

We recognize contributions through:

- **Contributors list** in README
- **Release notes** mentions
- **GitHub achievements**
- **Maintainer opportunities** for regular contributors

## 🔄 Development Workflow Summary

```bash
# 1. Set up
git clone https://github.com/YOUR-USERNAME/portscan.git
cd portscan
make dev-setup

# 2. Create branch
git checkout -b feature/my-feature

# 3. Develop
# Make changes...
make quick  # format, lint, test

# 4. Test
make test
make test-coverage

# 5. Commit
git add .
git commit -m "feat: implement my feature"

# 6. Submit
git push origin feature/my-feature
# Create PR on GitHub
```

---

Thank you for contributing to Port Scanner! 🎉

Your contributions help make this tool better for the entire security community.
