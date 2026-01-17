# Linting and Pre-commit Setup Script
# Run this script to install and configure linting tools

param(
    [switch]$Install,
    [switch]$Run,
    [switch]$Fix
)

$ErrorActionPreference = "Stop"

Write-Host "🔧 URL Shortener - Linting Setup" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan

function Test-CommandExists {
    param($Command)
    $null = Get-Command $Command -ErrorAction SilentlyContinue
    return $?
}

function Install-Tools {
    Write-Host "`n📦 Installing Required Tools" -ForegroundColor Yellow

    # Check if Go is installed
    if (-not (Test-CommandExists "go")) {
        Write-Error "Go is required but not installed. Please install Go first."
        exit 1
    }

    # Install golangci-lint
    Write-Host "Installing golangci-lint..."
    if ($IsLinux -or $IsMacOS) {
        # Linux/macOS installation
        Invoke-WebRequest -Uri "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" -OutFile "install.sh"
        bash install.sh -b $env:GOPATH/bin v1.55.2
        Remove-Item "install.sh"
    } else {
        # Windows installation
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
    }

    # Install pre-commit
    Write-Host "Installing pre-commit..."
    if (Test-CommandExists "pip") {
        pip install pre-commit
    } elseif (Test-CommandExists "brew") {
        brew install pre-commit
    } elseif (Test-CommandExists "conda") {
        conda install -c conda-forge pre-commit
    } else {
        Write-Warning "Please install pre-commit manually: https://pre-commit.com/"
    }

    # Install pre-commit hooks
    if (Test-CommandExists "pre-commit") {
        Write-Host "Installing pre-commit hooks..."
        pre-commit install
        pre-commit install --hook-type commit-msg
    }

    Write-Host "✅ Installation complete!" -ForegroundColor Green
}

function Run-Linters {
    Write-Host "`n🔍 Running Linters" -ForegroundColor Yellow

    # Run golangci-lint
    if (Test-CommandExists "golangci-lint") {
        Write-Host "Running golangci-lint..."
        golangci-lint run

        if ($LASTEXITCODE -ne 0) {
            Write-Warning "golangci-lint found issues"
        }
    } else {
        Write-Warning "golangci-lint not found, skipping..."
    }

    # Run pre-commit on all files
    if (Test-CommandExists "pre-commit") {
        Write-Host "Running pre-commit hooks..."
        pre-commit run --all-files

        if ($LASTEXITCODE -ne 0) {
            Write-Warning "pre-commit found issues"
        }
    } else {
        Write-Warning "pre-commit not found, skipping..."
    }
}

function Fix-Issues {
    Write-Host "`n🔧 Fixing Issues" -ForegroundColor Yellow

    # Run golangci-lint with fix
    if (Test-CommandExists "golangci-lint") {
        Write-Host "Running golangci-lint with auto-fix..."
        golangci-lint run --fix
    }

    # Run go fmt
    if (Test-CommandExists "go") {
        Write-Host "Running go fmt..."
        go fmt ./...
    }

    # Run goimports if available
    if (Test-CommandExists "goimports") {
        Write-Host "Running goimports..."
        goimports -w .
    }

    Write-Host "✅ Auto-fix complete!" -ForegroundColor Green
}

function Show-Usage {
    Write-Host @"
Usage: .\scripts\lint.ps1 [OPTIONS]

OPTIONS:
    -Install    Install linting tools and pre-commit hooks
    -Run        Run all linters and checks
    -Fix        Attempt to automatically fix issues

Examples:
    .\scripts\lint.ps1 -Install          # Install tools
    .\scripts\lint.ps1 -Run              # Run linters
    .\scripts\lint.ps1 -Fix              # Fix issues
    .\scripts\lint.ps1 -Install -Run     # Install and run

Configuration Files:
    .golangci.yml           - golangci-lint configuration
    .pre-commit-config.yaml - pre-commit hooks configuration
    .hadolint.yaml          - Dockerfile linting configuration
    .secrets.baseline       - Secret detection baseline
"@
}

# Main execution
if ($Install) {
    Install-Tools
}

if ($Fix) {
    Fix-Issues
}

if ($Run) {
    Run-Linters
}

if (-not ($Install -or $Run -or $Fix)) {
    Show-Usage
}

Write-Host "`n🎉 Linting setup complete!" -ForegroundColor Green
Write-Host "Run 'pre-commit run --all-files' to validate all files" -ForegroundColor Cyan
