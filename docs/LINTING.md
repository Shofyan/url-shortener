# Code Quality & Linting Guide

This document describes the code quality tools and linting setup for the URL Shortener project.

## Overview

The project uses a comprehensive linting and code quality setup that includes:

- **golangci-lint** with gocyclo and revive linters
- **Pre-commit hooks** for automated quality checks
- **Security scanning** with gosec
- **Dockerfile linting** with hadolint
- **Secret detection** with detect-secrets

## Quick Start

### Install Tools
```bash
# Using make
make lint-install

# Or using scripts
./scripts/lint.sh install     # Linux/macOS
.\scripts\lint.ps1 -Install   # Windows PowerShell
```

### Run Linting
```bash
# Run all linters
make lint

# Run with auto-fix
make lint-fix

# Run pre-commit hooks
make pre-commit

# Run all quality checks
make quality
```

## Linting Configuration

### golangci-lint (.golangci.yml)

The project is configured with strict linting rules:

**Required Linters (from ERD):**
- `gocyclo` - Cyclomatic complexity checker (max 10 per function)
- `revive` - Fast, configurable Go linter

**Additional Production Linters:**
- `errcheck` - Check for unchecked errors
- `gosec` - Security vulnerability checker
- `staticcheck` - Go static analysis
- `govet` - Standard Go vet
- `ineffassign` - Detect ineffectual assignments
- `unused` - Find unused code
- `misspell` - Spell checker

### Key Rules

1. **Cyclomatic Complexity**: Maximum 10 per function
2. **Security**: All security issues must be addressed
3. **Error Handling**: All errors must be checked
4. **Code Style**: Consistent formatting and naming
5. **Documentation**: Exported functions must be documented

## Pre-commit Hooks (.pre-commit-config.yaml)

Automated checks that run before each commit:

### Go-specific Hooks
- `golangci-lint` - Run full linting suite
- `go-fmt` - Code formatting
- `go-imports` - Import organization
- `go-vet` - Standard Go analysis
- `go-mod-tidy` - Clean up go.mod

### Security & Quality
- `detect-secrets` - Find secrets in code
- `hadolint` - Dockerfile linting
- `trailing-whitespace` - Remove trailing spaces
- `end-of-file-fixer` - Ensure files end with newline
- `check-yaml` - YAML syntax validation
- `check-json` - JSON syntax validation

## Usage Examples

### Command Line

```bash
# Install everything
make lint-install

# Run basic linting
make lint

# Fix auto-fixable issues
make lint-fix

# Check cyclomatic complexity
make complexity-check

# Run security scan
make security-scan

# Format code
make fmt

# Run all quality checks
make quality

# Fix all auto-fixable quality issues
make quality-fix
```

### Pre-commit Integration

```bash
# Install pre-commit hooks
pre-commit install

# Run hooks on all files
pre-commit run --all-files

# Run specific hook
pre-commit run golangci-lint

# Update hooks to latest versions
pre-commit autoupdate
```

## CI/CD Integration

The GitHub Actions workflow (`.github/workflows/ci-cd.yml`) includes:

### Linting Stage
- golangci-lint with full configuration
- Cyclomatic complexity check (max 10)
- Code formatting verification
- Go vet analysis

### Security Stage
- gosec security scanning
- SARIF report upload to GitHub Security tab
- Dependency vulnerability scanning

### Quality Gates
All checks must pass before:
- Merging pull requests
- Deploying to production
- Creating releases

## Error Resolution

### Common Issues

**High Cyclomatic Complexity**
```bash
# Check complexity
gocyclo -over 10 .

# Refactor complex functions by:
# 1. Breaking into smaller functions
# 2. Using early returns
# 3. Extracting conditional logic
```

**Linting Failures**
```bash
# See detailed issues
golangci-lint run -v

# Auto-fix when possible
golangci-lint run --fix

# Disable specific rules (use sparingly)
//nolint:rulename // Justification required
```

**Security Issues**
```bash
# Run security scan
gosec ./...

# Review and fix security vulnerabilities
# Never ignore security issues without proper justification
```

### IDE Integration

**VS Code**
```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
```

**GoLand/IntelliJ**
- Install golangci-lint plugin
- Configure to use project .golangci.yml
- Enable format on save

## Configuration Files

| File | Purpose |
|------|---------|
| `.golangci.yml` | golangci-lint configuration |
| `.pre-commit-config.yaml` | Pre-commit hooks setup |
| `.hadolint.yaml` | Dockerfile linting rules |
| `.secrets.baseline` | Secret detection baseline |
| `scripts/lint.sh` | Linux/macOS lint setup script |
| `scripts/lint.ps1` | Windows PowerShell lint script |

## Compliance Checklist

### ERD Requirements ✅
- [x] golangci-lint enabled
- [x] gocyclo linter enabled (max 10 complexity)
- [x] revive linter enabled
- [x] GitHub Actions CI/CD pipeline
- [x] Automated quality checks

### Production Ready ✅
- [x] Security scanning (gosec)
- [x] Dockerfile linting (hadolint)
- [x] Secret detection (detect-secrets)
- [x] Pre-commit hooks
- [x] Comprehensive test coverage
- [x] Performance benchmarking

## Troubleshooting

### Installation Issues
```bash
# Verify Go installation
go version

# Check PATH for Go binaries
echo $GOPATH/bin

# Install tools manually
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
```

### Performance Issues
```bash
# Run with timeout
golangci-lint run --timeout=10m

# Skip vendor directory
golangci-lint run --skip-dirs=vendor

# Run specific linters only
golangci-lint run --enable-only=gocyclo,revive
```

For additional help, see:
- [golangci-lint documentation](https://golangci-lint.run/)
- [pre-commit documentation](https://pre-commit.com/)
- [Project contributing guidelines](CONTRIBUTING.md)
