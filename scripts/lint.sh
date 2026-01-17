#!/bin/bash

# Linting and Pre-commit Setup Script
# Run this script to install and configure linting tools

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}🔧 URL Shortener - Linting Setup${NC}"
echo -e "${CYAN}================================${NC}"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install tools
install_tools() {
    echo -e "\n${YELLOW}📦 Installing Required Tools${NC}"

    # Check if Go is installed
    if ! command_exists go; then
        echo -e "${RED}Go is required but not installed. Please install Go first.${NC}"
        exit 1
    fi

    # Install golangci-lint
    echo "Installing golangci-lint..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command_exists brew; then
            brew install golangci-lint
        else
            curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
        fi
    else
        # Linux
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
    fi

    # Install pre-commit
    echo "Installing pre-commit..."
    if command_exists pip3; then
        pip3 install pre-commit
    elif command_exists pip; then
        pip install pre-commit
    elif command_exists brew; then
        brew install pre-commit
    elif command_exists apt-get; then
        sudo apt-get update && sudo apt-get install -y python3-pip
        pip3 install pre-commit
    elif command_exists yum; then
        sudo yum install -y python3-pip
        pip3 install pre-commit
    else
        echo -e "${YELLOW}Please install pre-commit manually: https://pre-commit.com/${NC}"
    fi

    # Install pre-commit hooks
    if command_exists pre-commit; then
        echo "Installing pre-commit hooks..."
        pre-commit install
        pre-commit install --hook-type commit-msg
    fi

    # Install additional tools
    echo "Installing additional Go tools..."
    go install golang.org/x/tools/cmd/goimports@latest

    echo -e "${GREEN}✅ Installation complete!${NC}"
}

# Function to run linters
run_linters() {
    echo -e "\n${YELLOW}🔍 Running Linters${NC}"

    # Run golangci-lint
    if command_exists golangci-lint; then
        echo "Running golangci-lint..."
        if golangci-lint run; then
            echo -e "${GREEN}✅ golangci-lint passed${NC}"
        else
            echo -e "${YELLOW}⚠️  golangci-lint found issues${NC}"
        fi
    else
        echo -e "${YELLOW}golangci-lint not found, skipping...${NC}"
    fi

    # Run pre-commit on all files
    if command_exists pre-commit; then
        echo "Running pre-commit hooks..."
        if pre-commit run --all-files; then
            echo -e "${GREEN}✅ pre-commit passed${NC}"
        else
            echo -e "${YELLOW}⚠️  pre-commit found issues${NC}"
        fi
    else
        echo -e "${YELLOW}pre-commit not found, skipping...${NC}"
    fi
}

# Function to fix issues
fix_issues() {
    echo -e "\n${YELLOW}🔧 Fixing Issues${NC}"

    # Run golangci-lint with fix
    if command_exists golangci-lint; then
        echo "Running golangci-lint with auto-fix..."
        golangci-lint run --fix
    fi

    # Run go fmt
    if command_exists go; then
        echo "Running go fmt..."
        go fmt ./...
    fi

    # Run goimports if available
    if command_exists goimports; then
        echo "Running goimports..."
        find . -name "*.go" -not -path "./vendor/*" -exec goimports -w {} \;
    fi

    echo -e "${GREEN}✅ Auto-fix complete!${NC}"
}

# Function to show usage
show_usage() {
    cat << 'EOF'
Usage: ./scripts/lint.sh [OPTIONS]

OPTIONS:
    install     Install linting tools and pre-commit hooks
    run         Run all linters and checks
    fix         Attempt to automatically fix issues
    help        Show this help message

Examples:
    ./scripts/lint.sh install          # Install tools
    ./scripts/lint.sh run              # Run linters
    ./scripts/lint.sh fix              # Fix issues
    ./scripts/lint.sh install run      # Install and run

Configuration Files:
    .golangci.yml           - golangci-lint configuration
    .pre-commit-config.yaml - pre-commit hooks configuration
    .hadolint.yaml          - Dockerfile linting configuration
    .secrets.baseline       - Secret detection baseline
EOF
}

# Parse arguments
INSTALL=false
RUN=false
FIX=false
HELP=false

while [[ $# -gt 0 ]]; do
    case $1 in
        install)
            INSTALL=true
            shift
            ;;
        run)
            RUN=true
            shift
            ;;
        fix)
            FIX=true
            shift
            ;;
        help)
            HELP=true
            shift
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
if [[ "$HELP" == true ]]; then
    show_usage
    exit 0
fi

if [[ "$INSTALL" == true ]]; then
    install_tools
fi

if [[ "$FIX" == true ]]; then
    fix_issues
fi

if [[ "$RUN" == true ]]; then
    run_linters
fi

if [[ "$INSTALL" == false && "$RUN" == false && "$FIX" == false && "$HELP" == false ]]; then
    show_usage
fi

echo -e "\n${GREEN}🎉 Linting setup complete!${NC}"
echo -e "${CYAN}Run 'pre-commit run --all-files' to validate all files${NC}"
