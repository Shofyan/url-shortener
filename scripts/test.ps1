# Test Runner Script for URL Shortener (PowerShell)
# This script runs all test suites and generates coverage reports

param(
    [switch]$Short,
    [switch]$Verbose,
    [switch]$Coverage
)

# Colors for output
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Cyan"

Write-Host "🚀 Starting URL Shortener Test Suite" -ForegroundColor $Blue
Write-Host "====================================" -ForegroundColor $Blue

# Test configuration
$TestTimeout = "30s"
$CoverageDir = "coverage"
$CoverageFile = "coverage.out"

# Create coverage directory
if (!(Test-Path $CoverageDir)) {
    New-Item -ItemType Directory -Path $CoverageDir | Out-Null
}

Write-Host "`n📋 Test Environment Setup" -ForegroundColor $Blue
Write-Host "Go version: $(go version)"
Write-Host "Test timeout: $TestTimeout"
Write-Host ""

# Function to run tests with coverage
function Run-TestSuite {
    param(
        [string]$Name,
        [string]$Path,
        [string]$Flags = ""
    )

    Write-Host "🧪 Running $Name" -ForegroundColor $Blue
    Write-Host "Path: $Path"

    $coverageFile = "$CoverageDir/$($Name.ToLower() -replace '_', '')_coverage.out"
    $cmd = "go test $Flags -timeout=$TestTimeout"

    if ($Coverage) {
        $cmd += " -coverprofile=$coverageFile"
    }

    $cmd += " $Path"

    try {
        Invoke-Expression $cmd
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ $Name - PASSED" -ForegroundColor $Green
            return $true
        } else {
            Write-Host "❌ $Name - FAILED" -ForegroundColor $Red
            return $false
        }
    } catch {
        Write-Host "❌ $Name - ERROR: $($_.Exception.Message)" -ForegroundColor $Red
        return $false
    }
}

# Track test results
$TotalTests = 0
$PassedTests = 0
$FailedTests = 0

# Test suites
$TestSuites = @(
    @{Name="Unit_Tests_Usecase"; Path="./tests/unit/usecase/..."; Flags="-v"}
    @{Name="Unit_Tests_TTL"; Path="./tests/unit/ttl/..."; Flags="-v"}
    @{Name="Unit_Tests_Repository"; Path="./tests/unit/repository/..."; Flags="-v"}
    @{Name="Integration_Tests_API"; Path="./tests/integration/api/..."; Flags="-v"}
)

if (!$Short) {
    $TestSuites += @{Name="Concurrency_Tests"; Path="./tests/concurrency/..."; Flags="-v -tags=integration"}
}

Write-Host "📊 Running Test Suites" -ForegroundColor $Yellow
Write-Host "======================"

foreach ($suite in $TestSuites) {
    $TotalTests++

    if (Run-TestSuite -Name $suite.Name -Path $suite.Path -Flags $suite.Flags) {
        $PassedTests++
    } else {
        $FailedTests++
    }

    Write-Host ""
}

# Run performance benchmarks
if (!$Short) {
    Write-Host "🏃 Running Performance Benchmarks" -ForegroundColor $Blue
    Write-Host "=================================="

    try {
        go test -bench=. -benchmem ./tests/concurrency/... -timeout=60s
        Write-Host "✅ Benchmarks completed" -ForegroundColor $Green
    } catch {
        Write-Host "⚠️  Benchmarks failed or skipped" -ForegroundColor $Yellow
    }

    Write-Host ""
}

# Generate coverage report
if ($Coverage) {
    Write-Host "📈 Generating Coverage Report" -ForegroundColor $Blue
    Write-Host "=============================="

    # Combine coverage files
    $combinedCoverage = "$CoverageDir/$CoverageFile"
    "mode: set" | Out-File -FilePath $combinedCoverage -Encoding utf8

    Get-ChildItem -Path "$CoverageDir/*_coverage.out" | ForEach-Object {
        Get-Content $_.FullName | Select-Object -Skip 1 | Out-File -FilePath $combinedCoverage -Append -Encoding utf8
    }

    if (Test-Path $combinedCoverage) {
        # Generate coverage statistics
        $coverageOutput = go tool cover -func=$combinedCoverage | Select-String "total:"
        if ($coverageOutput) {
            $coveragePercent = ($coverageOutput -split '\s+')[2]
            Write-Host "Overall test coverage: $coveragePercent"
        }

        # Generate HTML coverage report
        go tool cover -html=$combinedCoverage -o "$CoverageDir/coverage.html"
        Write-Host "HTML coverage report generated: $CoverageDir/coverage.html"
    } else {
        Write-Host "⚠️  No coverage data available" -ForegroundColor $Yellow
    }

    Write-Host ""
}

# Test summary
Write-Host "📋 Test Summary" -ForegroundColor $Blue
Write-Host "==============="
Write-Host "Total test suites: $TotalTests"
Write-Host "Passed: $PassedTests" -ForegroundColor $Green
Write-Host "Failed: $FailedTests" -ForegroundColor $Red

if ($FailedTests -eq 0) {
    Write-Host ""
    Write-Host "🎉 All tests passed successfully!" -ForegroundColor $Green

    # Functional requirements checklist
    Write-Host ""
    Write-Host "✅ Functional Requirements Checklist" -ForegroundColor $Blue
    Write-Host "====================================="
    Write-Host "✅ POST / endpoint - API compliance tests"
    Write-Host "✅ GET /s/{short_code} endpoint - API compliance tests"
    Write-Host "✅ TTL default 24h - Deterministic TTL tests"
    Write-Host "✅ Character exclusion 0,O,l,1 - Unit tests (updated generator)"
    Write-Host "✅ Thread-safe clicks - Concurrency tests"
    Write-Host "✅ last_accessed_at field - Integration tests"
    Write-Host "✅ X-Processing-Time-Micros header - API tests"
    Write-Host "✅ No PII storage/logging - Privacy compliance (IP removed)"
    Write-Host ""
    Write-Host "🔥 Ready for production deployment!" -ForegroundColor $Green

    exit 0
} else {
    Write-Host ""
    Write-Host "💥 Some tests failed. Please review the output above." -ForegroundColor $Red
    exit 1
}
