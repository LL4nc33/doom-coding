#!/usr/bin/env bash
# =============================================================================
# Doom Coding - Common Library Functions
# =============================================================================
# Shared functions used across all installation and management scripts.
# Source this file at the beginning of other scripts.
#
# Usage:
#   source "$(dirname "${BASH_SOURCE[0]}")/lib/common.sh"
# =============================================================================

# Prevent double-sourcing
[[ -n "${_DOOM_COMMON_SOURCED:-}" ]] && return
readonly _DOOM_COMMON_SOURCED=1

# =============================================================================
# COLORS
# =============================================================================
readonly DC_GREEN='\033[38;2;46;82;29m'
readonly DC_BROWN='\033[38;2;124;94;70m'
readonly DC_RED='\033[0;31m'
readonly DC_YELLOW='\033[0;33m'
readonly DC_BLUE='\033[0;34m'
readonly DC_NC='\033[0m'

# =============================================================================
# LOGGING
# =============================================================================
log_info() {
    echo -e "${DC_BLUE}i${DC_NC}  $*"
}

log_success() {
    echo -e "${DC_GREEN}[OK]${DC_NC} $*"
}

log_warning() {
    echo -e "${DC_YELLOW}[WARN]${DC_NC}  $*" >&2
}

log_error() {
    echo -e "${DC_RED}[ERROR]${DC_NC} $*" >&2
}

log_step() {
    echo -e "\n${DC_GREEN}>${DC_NC}  ${DC_BROWN}$*${DC_NC}"
}

log_pass() {
    echo -e "${DC_GREEN}[PASS]${DC_NC} $*"
}

log_fail() {
    echo -e "${DC_RED}[FAIL]${DC_NC} $*" >&2
}

# =============================================================================
# SYSTEM DETECTION
# =============================================================================
detect_package_manager() {
    if command -v apt-get &>/dev/null; then
        echo "apt"
    elif command -v pacman &>/dev/null; then
        echo "pacman"
    elif command -v dnf &>/dev/null; then
        echo "dnf"
    elif command -v yum &>/dev/null; then
        echo "yum"
    else
        echo "unknown"
    fi
}

detect_os() {
    if [[ -f /etc/os-release ]]; then
        # shellcheck source=/dev/null
        source /etc/os-release
        echo "${ID:-unknown}"
    else
        echo "unknown"
    fi
}

detect_arch() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        armv7l) echo "armhf" ;;
        *) echo "$arch" ;;
    esac
}

detect_container_type() {
    if [[ -f /.dockerenv ]]; then
        echo "docker"
    elif grep -q "lxc" /proc/1/cgroup 2>/dev/null; then
        echo "lxc"
    elif grep -q "microsoft" /proc/version 2>/dev/null; then
        echo "wsl"
    else
        echo ""
    fi
}

is_root() {
    [[ $EUID -eq 0 ]]
}

has_sudo() {
    command -v sudo &>/dev/null && sudo -n true 2>/dev/null
}

# =============================================================================
# PACKAGE MANAGEMENT
# =============================================================================
install_package() {
    local package="$1"
    local pkg_manager
    pkg_manager=$(detect_package_manager)

    log_info "Installing $package..."

    case "$pkg_manager" in
        apt)
            sudo apt-get install -y "$package"
            ;;
        pacman)
            sudo pacman -S --noconfirm "$package"
            ;;
        dnf)
            sudo dnf install -y "$package"
            ;;
        yum)
            sudo yum install -y "$package"
            ;;
        *)
            log_error "Unknown package manager"
            return 1
            ;;
    esac
}

update_package_lists() {
    local pkg_manager
    pkg_manager=$(detect_package_manager)

    case "$pkg_manager" in
        apt)
            sudo apt-get update
            ;;
        pacman)
            sudo pacman -Sy
            ;;
        dnf|yum)
            # dnf/yum refresh automatically
            :
            ;;
    esac
}

# =============================================================================
# UTILITIES
# =============================================================================
confirm() {
    local prompt="${1:-Continue?}"
    local default="${2:-n}"

    if [[ "$default" == "y" ]]; then
        read -rp "$prompt [Y/n] " response
        [[ -z "$response" || "$response" =~ ^[Yy] ]]
    else
        read -rp "$prompt [y/N] " response
        [[ "$response" =~ ^[Yy] ]]
    fi
}

command_exists() {
    command -v "$1" &>/dev/null
}

require_command() {
    local cmd="$1"
    local install_hint="${2:-}"

    if ! command_exists "$cmd"; then
        log_error "Required command '$cmd' not found"
        [[ -n "$install_hint" ]] && log_info "Install with: $install_hint"
        return 1
    fi
}

# =============================================================================
# FILE OPERATIONS
# =============================================================================
backup_file() {
    local file="$1"
    if [[ -f "$file" ]]; then
        local backup="${file}.backup.$(date +%Y%m%d_%H%M%S)"
        cp "$file" "$backup"
        log_info "Backed up $file to $backup"
    fi
}

ensure_directory() {
    local dir="$1"
    if [[ ! -d "$dir" ]]; then
        mkdir -p "$dir"
        log_info "Created directory: $dir"
    fi
}
