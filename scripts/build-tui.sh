#!/bin/bash
# =============================================================================
# Doom Coding TUI Build Script
# =============================================================================
# Builds the doom-tui binary from source
#
# Usage:
#   ./scripts/build-tui.sh              # Build for current platform
#   ./scripts/build-tui.sh --all        # Build for all platforms
#   ./scripts/build-tui.sh --install    # Build and install to /usr/local/bin
#
# Requirements:
#   - Go 1.22 or later
# =============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TUI_DIR="$PROJECT_ROOT/cmd/doom-tui"
BIN_DIR="$PROJECT_ROOT/bin"
BINARY_NAME="doom-tui"

# Version info
VERSION=$(git -C "$PROJECT_ROOT" describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME"

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed!"
        echo ""
        echo "Please install Go 1.22 or later:"
        echo ""
        echo "  Ubuntu/Debian:"
        echo "    sudo apt update && sudo apt install golang-go"
        echo ""
        echo "  Or download from: https://go.dev/dl/"
        echo ""
        exit 1
    fi

    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    GO_MAJOR=$(echo "$GO_VERSION" | cut -d. -f1)
    GO_MINOR=$(echo "$GO_VERSION" | cut -d. -f2)

    if [[ "$GO_MAJOR" -lt 1 ]] || [[ "$GO_MAJOR" -eq 1 && "$GO_MINOR" -lt 22 ]]; then
        log_warn "Go version $GO_VERSION detected. Go 1.22+ is recommended."
    else
        log_info "Go version $GO_VERSION detected"
    fi
}

# Download dependencies
download_deps() {
    log_info "Downloading dependencies..."
    cd "$TUI_DIR"
    go mod download
    go mod tidy
    log_success "Dependencies downloaded"
}

# Build for current platform
build_current() {
    log_info "Building doom-tui for current platform..."
    mkdir -p "$BIN_DIR"
    cd "$TUI_DIR"
    go build -ldflags "$LDFLAGS" -o "$BIN_DIR/$BINARY_NAME" .
    log_success "Built: $BIN_DIR/$BINARY_NAME"
}

# Build for all platforms
build_all() {
    log_info "Building doom-tui for all platforms..."
    mkdir -p "$BIN_DIR"
    cd "$TUI_DIR"

    platforms=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
    )

    for platform in "${platforms[@]}"; do
        GOOS="${platform%/*}"
        GOARCH="${platform#*/}"
        OUTPUT="$BIN_DIR/$BINARY_NAME-$GOOS-$GOARCH"

        log_info "Building for $GOOS/$GOARCH..."
        GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "$LDFLAGS" -o "$OUTPUT" .
        log_success "Built: $OUTPUT"
    done

    log_success "All platforms built in $BIN_DIR/"
}

# Install to system
install_binary() {
    build_current

    log_info "Installing doom-tui to /usr/local/bin..."
    if [[ -w /usr/local/bin ]]; then
        cp "$BIN_DIR/$BINARY_NAME" /usr/local/bin/
        chmod +x /usr/local/bin/$BINARY_NAME
    else
        sudo cp "$BIN_DIR/$BINARY_NAME" /usr/local/bin/
        sudo chmod +x /usr/local/bin/$BINARY_NAME
    fi
    log_success "Installed: /usr/local/bin/$BINARY_NAME"
}

# Show help
show_help() {
    echo "Doom Coding TUI Build Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  (none)      Build for current platform"
    echo "  --all       Build for all supported platforms"
    echo "  --install   Build and install to /usr/local/bin"
    echo "  --deps      Only download dependencies"
    echo "  --clean     Remove build artifacts"
    echo "  --help      Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                    # Build for current platform"
    echo "  $0 --all              # Build for Linux and macOS"
    echo "  $0 --install          # Build and install globally"
    echo ""
}

# Clean build artifacts
clean() {
    log_info "Cleaning build artifacts..."
    rm -rf "$BIN_DIR"
    log_success "Clean complete"
}

# Main
main() {
    case "${1:-}" in
        --help|-h)
            show_help
            ;;
        --all)
            check_go
            download_deps
            build_all
            ;;
        --install)
            check_go
            download_deps
            install_binary
            ;;
        --deps)
            check_go
            download_deps
            ;;
        --clean)
            clean
            ;;
        *)
            check_go
            download_deps
            build_current
            echo ""
            echo "Run the TUI with:"
            echo "  $BIN_DIR/$BINARY_NAME"
            ;;
    esac
}

main "$@"
