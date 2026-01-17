#!/usr/bin/env bash
# Doom Coding - One-Click Installer
# Entry point for complete environment setup
set -eo pipefail

# ===========================================
# BRAND COLORS (ANSI 256)
# ===========================================
readonly GREEN='\033[38;2;46;82;29m'      # Forest Green #2E521D
readonly BROWN='\033[38;2;124;94;70m'     # Tan Brown #7C5E46
readonly LIGHT_BROWN='\033[38;2;164;125;91m' # Light Brown #A47D5B
readonly RED='\033[0;31m'
readonly YELLOW='\033[0;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# ===========================================
# CONFIGURATION
# ===========================================
# Handle curl | bash case where BASH_SOURCE is empty
if [[ -n "${BASH_SOURCE[0]:-}" ]]; then
    readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    readonly PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
else
    # Running via curl | bash - clone to temp directory
    readonly SCRIPT_DIR="/tmp/doom-coding-install"
    readonly PROJECT_DIR="/tmp/doom-coding"
    if [[ ! -d "$PROJECT_DIR" ]]; then
        echo -e "${BLUE}ℹ${NC}  Cloning repository for installation..."
        git clone --depth 1 https://github.com/LL4nc33/doom-coding.git "$PROJECT_DIR" 2>/dev/null || {
            echo -e "${RED}❌${NC} Failed to clone repository"
            exit 1
        }
    fi
fi
readonly LOG_FILE="${LOG_FILE:-/var/log/doom-coding-install.log}"
readonly INSTALLER_VERSION="0.0.6a"

# Default options
UNATTENDED=false
SKIP_DOCKER=false
SKIP_TAILSCALE=false
SKIP_TERMINAL=false
SKIP_HARDENING=false
SKIP_SECRETS=false
DRY_RUN=false
VERBOSE=false
FORCE=false
ENV_FILE=""
USE_TAILSCALE=true
NATIVE_TAILSCALE=false
NATIVE_USERSPACE=false

# Unattended installation values
TAILSCALE_KEY=""
CODE_PASSWORD=""
ANTHROPIC_KEY=""

# Detected system info
OS_ID=""
OS_VERSION=""
ARCH=""
PKG_MANAGER=""

# ===========================================
# TRAP HANDLER FOR CLEANUP
# ===========================================
cleanup_on_error() {
    local exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        log_warning "Installation interrupted or failed (exit code: $exit_code)"
        log_warning "Check log file for details: $LOG_FILE"
        # Add any cleanup needed here
    fi
}
trap cleanup_on_error ERR INT TERM

# ===========================================
# SOURCE SERVICE MANAGEMENT LIBRARY
# ===========================================
# Load the service management functions if available
if [[ -f "${SCRIPT_DIR}/lib/service-manager.sh" ]]; then
    source "${SCRIPT_DIR}/lib/service-manager.sh"
    SERVICE_MANAGER_LOADED=true
else
    SERVICE_MANAGER_LOADED=false
fi

# ===========================================
# LOGGING FUNCTIONS
# ===========================================
setup_logging() {
    if [[ ! -f "$LOG_FILE" ]]; then
        sudo touch "$LOG_FILE" 2>/dev/null || touch "$LOG_FILE" 2>/dev/null || true
        sudo chmod 666 "$LOG_FILE" 2>/dev/null || chmod 666 "$LOG_FILE" 2>/dev/null || true
    fi
}

log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp
    timestamp="$(date '+%Y-%m-%d %H:%M:%S')"
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE" 2>/dev/null || true
}

log_info() {
    log "INFO" "$*"
    echo -e "${BLUE}ℹ${NC}  $*"
}

log_success() {
    log "SUCCESS" "$*"
    echo -e "${GREEN}✅${NC} $*"
}

log_warning() {
    log "WARNING" "$*"
    echo -e "${YELLOW}⚠${NC}  $*"
}

log_error() {
    log "ERROR" "$*"
    echo -e "${RED}❌${NC} $*" >&2
}

log_step() {
    log "STEP" "$*"
    echo -e "${BROWN}⏳${NC} $*"
}

# ===========================================
# QR CODE HELPER FUNCTIONS
# ===========================================
# Generate QR code for terminal display (requires qrencode)
generate_qr() {
    local url="$1"
    local label="${2:-Scan to open}"

    if command -v qrencode &>/dev/null; then
        echo ""
        qrencode -t ansiutf8 -m 2 "$url"
        echo "    ${label} ↑"
        echo ""
    else
        log_info "QR code display requires qrencode (install with: apt install qrencode)"
        echo "    URL: $url"
        echo ""
    fi
}

# Show access QR code after successful installation
show_access_qr() {
    local ip="$1"
    local port="${2:-8443}"
    local protocol="${3:-https}"
    local url="${protocol}://${ip}:${port}"

    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  Access your code-server on any device:${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    generate_qr "$url" "Scan to open code-server"
    echo "    Desktop: $url"
    echo ""
}

# Show QR code linking to external service for credentials
show_service_qr() {
    local service="$1"
    local url=""
    local label=""

    case "$service" in
        tailscale)
            url="https://login.tailscale.com/admin/settings/keys"
            label="Scan to create Tailscale auth key"
            ;;
        anthropic)
            url="https://console.anthropic.com/account/keys"
            label="Scan to get Anthropic API key"
            ;;
        github)
            url="https://github.com/doom-coding/doom-coding"
            label="Scan to view project on GitHub"
            ;;
        termux)
            url="https://play.google.com/store/apps/details?id=com.termux"
            label="Scan to install Termux (Android)"
            ;;
        blink)
            url="https://apps.apple.com/app/blink-shell-mosh-ssh-client/id1594898306"
            label="Scan to install Blink Shell (iOS)"
            ;;
        *)
            log_warning "Unknown service: $service"
            return 1
            ;;
    esac

    generate_qr "$url" "$label"
}

# Show troubleshooting QR code for specific error
show_troubleshoot_qr() {
    local error_code="$1"
    local url="https://doom-coding.dev/troubleshoot/${error_code}"

    echo ""
    echo -e "${YELLOW}Need help? Scan for troubleshooting guide:${NC}"
    generate_qr "$url" "Scan for detailed help"
}

# Handle port conflict with QR-enhanced error message
handle_port_conflict_with_qr() {
    local port="$1"
    local process="${2:-unknown}"

    echo ""
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    log_error "Port $port is already in use"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "This usually means another web server is running."
    echo "Process using port: $process"
    echo ""
    echo "Options:"
    echo "  1) Stop the conflicting service manually"
    echo "  2) Use --force to automatically stop doom-coding containers"
    echo "  3) Choose a different port"
    echo ""
    show_troubleshoot_qr "port-${port}"
}

# Show mobile setup guide QR
show_mobile_setup_qr() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  Mobile Setup Guide${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "Recommended mobile apps:"
    echo ""
    echo "Android:"
    show_service_qr "termux"
    echo ""
    echo "iOS:"
    show_service_qr "blink"
}

# ===========================================
# SECURE DOWNLOAD FUNCTIONS
# ===========================================

# Known checksums for external scripts (update as needed)
# These should be verified from official sources before deployment
# To get checksum: curl -fsSL <URL> | sha256sum
declare -A KNOWN_CHECKSUMS=(
    # Tailscale install script - empty means warn but allow (script changes frequently)
    # Verify at: https://tailscale.com/install.sh
    ["tailscale"]=""
)

# Secure download function that verifies checksums before execution
# Usage: verified_download_and_run <url> [expected_sha256] [script_name]
verified_download_and_run() {
    local url="$1"
    local expected_sha256="${2:-}"
    local script_name="${3:-install.sh}"

    local temp_file
    temp_file=$(mktemp)
    # Ensure cleanup on function exit
    trap 'rm -f "$temp_file"' RETURN

    log_step "Downloading $script_name from $url..."

    if ! curl -fsSL "$url" -o "$temp_file"; then
        log_error "Failed to download from $url"
        return 1
    fi

    # Verify checksum if provided
    if [[ -n "$expected_sha256" ]]; then
        local actual_sha256
        actual_sha256=$(sha256sum "$temp_file" | cut -d' ' -f1)
        if [[ "$actual_sha256" != "$expected_sha256" ]]; then
            log_error "Checksum verification failed for $script_name"
            log_error "Expected: $expected_sha256"
            log_error "Got: $actual_sha256"
            return 1
        fi
        log_success "Checksum verified for $script_name"
    else
        log_warning "No checksum provided for $script_name - skipping verification"
        log_warning "Downloaded script hash: $(sha256sum "$temp_file" | cut -d' ' -f1)"
    fi

    # Make executable and run
    chmod +x "$temp_file"
    bash "$temp_file"
    local exit_code=$?

    return $exit_code
}

# ===========================================
# UTILITY FUNCTIONS
# ===========================================
print_banner() {
    echo -e "${GREEN}"
    cat << 'EOF'
    ____                           ______          ___
   / __ \____  ____  ____ ___     / ____/___  ____/ (_)___  ____ _
  / / / / __ \/ __ \/ __ `__ \   / /   / __ \/ __  / / __ \/ __ `/
 / /_/ / /_/ / /_/ / / / / / /  / /___/ /_/ / /_/ / / / / / /_/ /
/_____/\____/\____/_/ /_/ /_/   \____/\____/\__,_/_/_/ /_/\__, /
                                                         /____/
EOF
    echo -e "${NC}"
    echo -e "${BROWN}Remote Development Environment v${INSTALLER_VERSION}${NC}"
    echo ""
}

print_help() {
    cat << EOF
Usage: $0 [OPTIONS]

Doom Coding - One-Click Remote Development Environment Installer

OPTIONS:
    --unattended        Run without interactive prompts
    --skip-docker       Skip Docker installation
    --skip-tailscale    Skip Tailscale, use local network (for LXC)
    --local-network     Alias for --skip-tailscale
    --native-tailscale  Use existing Tailscale on host (no container)
    --native-userspace  Install Tailscale in userspace mode directly on LXC host
    --skip-terminal     Skip terminal tools setup
    --skip-hardening    Skip SSH hardening
    --skip-secrets      Skip SOPS/age setup
    --env-file=FILE     Use specific environment file
    --tailscale-key=KEY Tailscale auth key for unattended setup
    --code-password=PWD code-server password for unattended setup
    --anthropic-key=KEY Anthropic API key for Claude Code
    --dry-run           Show what would be done without executing
    --force             Force reinstallation, remove conflicting containers
    --verbose           Enable verbose output
    --retry-failed      Retry previously failed installation steps
    --help, -h          Show this help message
    --version, -v       Show version information

CONFLICT RESOLUTION:
    The installer automatically detects port conflicts (8443, 7681) and
    existing doom-coding installations. In interactive mode, you can choose
    to stop/remove conflicting containers. Use --force for automatic cleanup.

EXAMPLES:
    $0                              Interactive installation
    $0 --unattended                 Fully automated installation
    $0 --force                      Remove conflicts automatically
    $0 --skip-tailscale             LXC without TUN device
    $0 --native-tailscale           Use host's existing Tailscale
    $0 --native-userspace           Install native Tailscale in LXC (userspace)
    $0 --unattended --force \\
      --tailscale-key="tskey-auth-xxx" \\
      --code-password="secure-password" \\
      --anthropic-key="sk-ant-xxx"      Fully automated with conflict cleanup
    $0 --skip-docker --skip-terminal  Minimal installation
    $0 --env-file=production.env    Use custom environment file

EOF
}

print_version() {
    echo "Doom Coding Installer v${INSTALLER_VERSION}"
}

confirm() {
    local prompt="$1"
    local default="${2:-y}"

    if [[ "$UNATTENDED" == "true" ]]; then
        return 0
    fi

    local yn
    if [[ "$default" == "y" ]]; then
        read -rp "$prompt [Y/n]: " yn < /dev/tty
        yn="${yn:-y}"
    else
        read -rp "$prompt [y/N]: " yn < /dev/tty
        yn="${yn:-n}"
    fi

    [[ "${yn,,}" == "y" || "${yn,,}" == "yes" ]]
}

prompt_value() {
    local prompt="$1"
    local default="$2"
    local var_name="$3"
    local secret="${4:-false}"

    # Validate variable name to prevent injection (alphanumeric and underscore only)
    if [[ ! "$var_name" =~ ^[a-zA-Z_][a-zA-Z0-9_]*$ ]]; then
        log_error "Invalid variable name: $var_name"
        return 1
    fi

    if [[ "$UNATTENDED" == "true" ]]; then
        # Use declare -g for safer global variable assignment
        declare -g "$var_name=$default"
        return
    fi

    local value
    if [[ "$secret" == "true" ]]; then
        read -rsp "$prompt [$default]: " value
        echo ""
    else
        read -rp "$prompt [$default]: " value < /dev/tty
    fi

    # Use declare -g for safer global variable assignment
    declare -g "$var_name=${value:-$default}"
}

check_root() {
    if [[ $EUID -eq 0 ]]; then
        log_warning "Running as root. Some operations will use sudo for consistency."
    fi
}

check_internet() {
    log_step "Checking internet connectivity..."
    if ! curl -sf --max-time 5 https://google.com > /dev/null 2>&1; then
        log_error "No internet connection detected"
        return 1
    fi
    log_success "Internet connection available"
}

# ===========================================
# PRE-FLIGHT CHECKS
# ===========================================
preflight_check() {
    log_step "Running pre-flight checks..."
    local failed=0

    # Check sudo/root access
    if [[ $EUID -ne 0 ]] && ! sudo -n true 2>/dev/null; then
        if ! sudo -v; then
            log_error "Sudo access required"
            failed=1
        fi
    fi
    log_info "Sudo access: OK"

    # Check internet connectivity
    if ! curl -sf --max-time 10 https://google.com &>/dev/null; then
        log_error "Internet connectivity check failed"
        failed=1
    else
        log_info "Internet connectivity: OK"
    fi

    # Check disk space (need at least 10GB)
    local available_gb
    available_gb=$(df -BG . 2>/dev/null | awk 'NR==2 {print $4}' | tr -d 'G')
    if [[ -n "$available_gb" ]] && [[ $available_gb -lt 10 ]]; then
        log_error "Insufficient disk space: ${available_gb}GB available (need 10GB)"
        failed=1
    else
        log_info "Disk space: ${available_gb:-unknown}GB available"
    fi

    # Check memory (recommend at least 2GB)
    local mem_gb
    mem_gb=$(free -g 2>/dev/null | awk '/^Mem:/{print $2}')
    if [[ -n "$mem_gb" ]] && [[ $mem_gb -lt 2 ]]; then
        log_warning "Low memory: ${mem_gb}GB (recommended 2GB+)"
    else
        log_info "Memory: ${mem_gb:-unknown}GB available"
    fi

    # Check required commands
    for cmd in curl git; do
        if command -v "$cmd" &>/dev/null; then
            log_info "Required command '$cmd': found"
        else
            log_error "Required command '$cmd': not found"
            failed=1
        fi
    done

    if [[ $failed -eq 1 ]]; then
        log_error "Pre-flight checks failed. Please fix the issues above."
        return 1
    fi

    log_success "All pre-flight checks passed"
    return 0
}

# Check for port conflicts and existing containers
check_port_conflicts() {
    log_step "Checking for port conflicts and existing installations..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would check for port conflicts"
        return 0
    fi

    # Skip port checks if Docker is not installed yet
    if ! command -v docker &>/dev/null; then
        log_info "Docker not installed yet, skipping port conflict check"
        return 0
    fi

    local conflicts_found=0
    local ports_to_check=()

    # Determine which ports will be used based on compose file
    # All non-sidecar compose files expose ports 8443 and 7681
    if [[ "$COMPOSE_FILE" == "docker-compose.lxc.yml" ]] || \
       [[ "$COMPOSE_FILE" == "docker-compose.native-tailscale.yml" ]] || \
       [[ "$COMPOSE_FILE" == "docker-compose.lxc-tailscale.yml" ]]; then
        ports_to_check=(8443 7681)
    else
        # docker-compose.yml uses sidecar pattern, no exposed ports
        log_info "Using sidecar network mode, no port exposure needed"
    fi

    # Check for port conflicts
    for port in "${ports_to_check[@]}"; do
        if docker ps --format '{{.Names}}\t{{.Ports}}' | grep -q ":${port}->"; then
            local conflicting_container
            conflicting_container=$(docker ps --format '{{.Names}}\t{{.Ports}}' | grep ":${port}->" | awk '{print $1}')
            log_warning "Port ${port} already in use by container: ${conflicting_container}"
            conflicts_found=1
        fi
    done

    # Check for existing doom-coding containers
    local existing_doom_containers
    existing_doom_containers=$(docker ps -a --filter "name=doom-" --format '{{.Names}}\t{{.Status}}' 2>/dev/null || true)

    if [[ -n "$existing_doom_containers" ]]; then
        log_warning "Found existing doom-coding containers:"
        echo "$existing_doom_containers" | while read -r line; do
            log_info "  $line"
        done
        conflicts_found=1
    fi

    # If conflicts found, offer resolution options
    if [[ $conflicts_found -eq 1 ]]; then
        echo ""
        log_warning "Installation conflicts detected!"
        echo ""

        if [[ "$FORCE" == "true" ]]; then
            log_info "Force mode enabled, will attempt cleanup..."
            cleanup_existing_installation
            return 0
        fi

        if [[ "$UNATTENDED" == "true" ]]; then
            log_error "Cannot proceed in unattended mode with conflicts"
            log_error "Use --force to automatically remove conflicting containers"
            return 1
        fi

        echo -e "${YELLOW}Options:${NC}"
        echo "  1) Stop and remove conflicting doom-coding containers"
        echo "  2) Stop and remove ALL containers using ports 8443/7681"
        echo "  3) Show detailed conflict information"
        echo "  4) Abort installation"
        echo ""

        local choice
        read -rp "Select option [1/2/3/4] (default: 1): " choice < /dev/tty
        choice="${choice:-1}"

        case "$choice" in
            1)
                cleanup_existing_installation
                return 0
                ;;
            2)
                cleanup_port_conflicts "${ports_to_check[@]}"
                return 0
                ;;
            3)
                show_conflict_details "${ports_to_check[@]}"
                # Ask again after showing details
                if confirm "Proceed with cleanup of doom-coding containers?" "y"; then
                    cleanup_existing_installation
                    return 0
                else
                    log_error "Installation aborted by user"
                    exit 0
                fi
                ;;
            4)
                log_info "Installation aborted by user"
                exit 0
                ;;
            *)
                log_error "Invalid selection"
                exit 1
                ;;
        esac
    fi

    log_success "No port conflicts detected"
    return 0
}

# Cleanup existing doom-coding installation
cleanup_existing_installation() {
    log_step "Cleaning up existing doom-coding containers..."

    # Find all doom-* containers
    local doom_containers
    doom_containers=$(docker ps -aq --filter "name=doom-" 2>/dev/null || true)

    if [[ -n "$doom_containers" ]]; then
        log_info "Stopping doom-coding containers..."
        docker stop $doom_containers 2>/dev/null || true

        log_info "Removing doom-coding containers..."
        docker rm $doom_containers 2>/dev/null || true

        log_success "Existing doom-coding containers removed"
    else
        log_info "No doom-coding containers to remove"
    fi
}

# Cleanup all containers using specific ports
cleanup_port_conflicts() {
    local ports=("$@")
    log_step "Cleaning up containers using ports: ${ports[*]}"

    for port in "${ports[@]}"; do
        local containers
        containers=$(docker ps --format '{{.Names}}' | while read -r name; do
            if docker ps --format '{{.Names}}\t{{.Ports}}' | grep "^${name}" | grep -q ":${port}->"; then
                echo "$name"
            fi
        done)

        if [[ -n "$containers" ]]; then
            log_info "Stopping containers using port ${port}..."
            echo "$containers" | xargs docker stop 2>/dev/null || true

            log_info "Removing containers that used port ${port}..."
            echo "$containers" | xargs docker rm 2>/dev/null || true
        fi
    done

    log_success "Port conflicts resolved"
}

# Show detailed conflict information
show_conflict_details() {
    local ports=("$@")
    echo ""
    log_info "=== Detailed Conflict Information ==="
    echo ""

    # Show all doom-* containers
    echo -e "${BLUE}Doom-coding containers:${NC}"
    docker ps -a --filter "name=doom-" --format "table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "  None found"
    echo ""

    # Show port usage
    for port in "${ports[@]}"; do
        echo -e "${BLUE}Port ${port} usage:${NC}"
        docker ps --format '{{.Names}}\t{{.Image}}\t{{.Ports}}' | grep ":${port}->" || echo "  Not in use"
        echo ""
    done

    # Show process listening on ports
    echo -e "${BLUE}System port listeners:${NC}"
    for port in "${ports[@]}"; do
        if command -v ss &>/dev/null; then
            ss -tlnp 2>/dev/null | grep ":${port}" || echo "  Port ${port}: not in use"
        elif command -v netstat &>/dev/null; then
            netstat -tlnp 2>/dev/null | grep ":${port}" || echo "  Port ${port}: not in use"
        fi
    done
    echo ""
}

# ===========================================
# INPUT VALIDATION FUNCTIONS
# ===========================================
validate_password() {
    local password="$1"
    local min_length="${2:-8}"

    if [[ ${#password} -lt $min_length ]]; then
        log_error "Password must be at least $min_length characters"
        return 1
    fi
    return 0
}

validate_tailscale_key() {
    local key="$1"
    if [[ -n "$key" ]] && [[ ! "$key" =~ ^tskey- ]]; then
        log_warning "Tailscale key format may be invalid (should start with 'tskey-')"
    fi
    return 0
}

# ===========================================
# SYSTEM DETECTION
# ===========================================
detect_os() {
    log_step "Detecting operating system..."

    if [[ -f /etc/os-release ]]; then
        # Read os-release without sourcing (avoids variable conflicts)
        OS_ID="$(grep -E '^ID=' /etc/os-release | cut -d= -f2 | tr -d '"')"
        OS_VERSION="$(grep -E '^VERSION_ID=' /etc/os-release | cut -d= -f2 | tr -d '"' || echo 'unknown')"
    elif [[ -f /etc/debian_version ]]; then
        OS_ID="debian"
        OS_VERSION="$(cat /etc/debian_version)"
    elif [[ -f /etc/arch-release ]]; then
        OS_ID="arch"
        OS_VERSION="rolling"
    else
        log_error "Unsupported operating system"
        return 1
    fi

    log_success "Detected: ${OS_ID} ${OS_VERSION}"
}

detect_arch() {
    log_step "Detecting architecture..."

    case "$(uname -m)" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="armhf"
            ;;
        *)
            log_error "Unsupported architecture: $(uname -m)"
            return 1
            ;;
    esac

    log_success "Architecture: ${ARCH}"
}

detect_package_manager() {
    if command -v apt-get &>/dev/null; then
        PKG_MANAGER="apt"
    elif command -v pacman &>/dev/null; then
        PKG_MANAGER="pacman"
    elif command -v dnf &>/dev/null; then
        PKG_MANAGER="dnf"
    else
        log_error "No supported package manager found"
        return 1
    fi

    log_info "Package manager: ${PKG_MANAGER}"
}

# ===========================================
# INSTALLATION FUNCTIONS
# ===========================================
install_package() {
    local package="$1"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would install: $package"
        return 0
    fi

    case "$PKG_MANAGER" in
        apt)
            if ! dpkg-query -W -f='${Status}' "$package" 2>/dev/null | grep -q "install ok"; then
                sudo apt-get install -y "$package"
            fi
            ;;
        pacman)
            if ! pacman -Q "$package" &>/dev/null; then
                sudo pacman -S --noconfirm "$package"
            fi
            ;;
        dnf)
            if ! rpm -q "$package" &>/dev/null; then
                sudo dnf install -y "$package"
            fi
            ;;
    esac
}

install_base_packages() {
    log_step "Installing base packages..."

    local packages=(
        curl
        wget
        git
        jq
        unzip
        ca-certificates
        gnupg
        lsb-release
        qrencode
    )

    if [[ "$PKG_MANAGER" == "apt" ]]; then
        sudo apt-get update
    fi

    for pkg in "${packages[@]}"; do
        install_package "$pkg"
    done

    log_success "Base packages installed"
}

install_docker() {
    if [[ "$SKIP_DOCKER" == "true" ]]; then
        log_info "Skipping Docker installation (--skip-docker)"
        return 0
    fi

    log_step "Installing Docker..."

    if command -v docker &>/dev/null && [[ "$FORCE" != "true" ]]; then
        log_success "Docker already installed: $(docker --version)"
        return 0
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would install Docker"
        return 0
    fi

    case "$OS_ID" in
        ubuntu|debian)
            # Remove old versions
            sudo apt-get remove -y docker docker-engine docker.io containerd runc 2>/dev/null || true

            # Add Docker GPG key
            sudo install -m 0755 -d /etc/apt/keyrings
            curl -fsSL "https://download.docker.com/linux/${OS_ID}/gpg" | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
            sudo chmod a+r /etc/apt/keyrings/docker.gpg

            # Add repository
            echo \
                "deb [arch=${ARCH} signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/${OS_ID} \
                $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
                sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

            # Install Docker
            sudo apt-get update
            sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
            ;;
        arch)
            sudo pacman -S --noconfirm docker docker-compose
            ;;
        *)
            log_error "Docker installation not supported for ${OS_ID}"
            return 1
            ;;
    esac

    # Start and enable Docker
    sudo systemctl start docker
    sudo systemctl enable docker

    # Add current user to docker group
    if [[ -n "${SUDO_USER:-}" ]]; then
        sudo usermod -aG docker "$SUDO_USER"
    else
        sudo usermod -aG docker "$USER"
    fi

    log_success "Docker installed successfully"
    log_warning "You may need to log out and back in for docker group membership to take effect"
}

# ===========================================
# TAILSCALE SETUP
# ===========================================

# Global variable for compose file selection
COMPOSE_FILE="docker-compose.yml"

check_tun_device() {
    # Check if TUN device is available (required for Tailscale)
    if [[ -c /dev/net/tun ]]; then
        return 0
    elif [[ -e /dev/net/tun ]]; then
        # Exists but not a character device
        return 1
    else
        return 1
    fi
}

detect_container_type() {
    # Detect if running in LXC, Docker, or bare-metal
    if [[ -f /proc/1/environ ]] && grep -q "container=lxc" /proc/1/environ 2>/dev/null; then
        echo "lxc"
    elif [[ -f /.dockerenv ]]; then
        echo "docker"
    elif grep -q "docker\|lxc" /proc/1/cgroup 2>/dev/null; then
        echo "container"
    else
        echo "bare-metal"
    fi
}

setup_tailscale_choice() {
    # Skip if already decided via command line
    if [[ "$SKIP_TAILSCALE" == "true" ]]; then
        log_info "Tailscale uebersprungen (--skip-tailscale)"
        USE_TAILSCALE=false
        COMPOSE_FILE="docker-compose.lxc.yml"
        return 0
    fi

    local container_type
    container_type=$(detect_container_type)
    local tun_available=false

    log_step "Checking Tailscale compatibility..."

    # Check TUN device
    if check_tun_device; then
        tun_available=true
        log_success "TUN device available (/dev/net/tun)"
    else
        log_warning "TUN device not available"
    fi

    log_info "Environment: ${container_type}"

    # If in LXC without TUN, warn user
    if [[ "$container_type" == "lxc" ]] && [[ "$tun_available" == "false" ]]; then
        echo ""
        log_warning "Du bist in einem LXC-Container ohne TUN-Device."
        echo ""
        echo -e "${YELLOW}Optionen:${NC}"
        echo "  1) ${GREEN}Native Tailscale Userspace (EMPFOHLEN fuer LXC)${NC}"
        echo "     - Installiert Tailscale direkt auf dem LXC Host"
        echo "     - Kein Docker-Container fuer Tailscale noetig"
        echo "     - Geringster Ressourcenverbrauch"
        echo ""
        echo "  2) Docker Tailscale Userspace"
        echo "     - Tailscale laeuft in Docker Container"
        echo "     - SOCKS5 Proxy fuer Services"
        echo ""
        echo "  3) Ohne Tailscale (nur lokales Netzwerk)"
        echo "     - Zugriff nur via LAN IP (z.B. 192.168.x.x)"
        echo ""
        echo "  4) Host Tailscale verwenden (vorkonfiguriert)"
        echo "     - Nutzt bereits installiertes Host-Tailscale"
        echo ""
        echo "  5) TUN in LXC aktivieren (Proxmox-Host Konfiguration)"
        echo "     - Erfordert Aenderungen auf dem Proxmox Host"
        echo ""
        echo "  6) Installation abbrechen"
        echo ""

        if [[ "$UNATTENDED" == "true" ]]; then
            log_info "Unattended mode: Verwende Native Tailscale Userspace Mode"
            NATIVE_USERSPACE=true
            USE_TAILSCALE=true
            COMPOSE_FILE="docker-compose.native-userspace.yml"
            return 0
        fi

        local choice
        read -rp "Auswahl [1/2/3/4/5/6] (Standard: 1): " choice < /dev/tty
        choice="${choice:-1}"
        case "$choice" in
            1)
                NATIVE_USERSPACE=true
                USE_TAILSCALE=true
                COMPOSE_FILE="docker-compose.native-userspace.yml"
                log_success "Verwende Native Tailscale Userspace Mode"
                log_info "Tailscale wird direkt auf dem LXC Host installiert"
                ;;
            2)
                USE_TAILSCALE=true
                TS_USERSPACE=true
                COMPOSE_FILE="docker-compose.lxc-tailscale.yml"
                log_success "Verwende Docker Tailscale Userspace Mode (docker-compose.lxc-tailscale.yml)"
                log_info "Kein TUN-Device erforderlich!"
                ;;
            3)
                USE_TAILSCALE=false
                COMPOSE_FILE="docker-compose.lxc.yml"
                log_info "Verwende lokales Netzwerk (docker-compose.lxc.yml)"
                ;;
            4)
                NATIVE_TAILSCALE=true
                USE_TAILSCALE=false
                SKIP_TAILSCALE=true
                COMPOSE_FILE="docker-compose.native-tailscale.yml"
                log_info "Verwende vorhandenes Host-Tailscale (docker-compose.native-tailscale.yml)"
                ;;
            5)
                echo ""
                echo -e "${BLUE}Auf dem Proxmox-Host ausfuehren:${NC}"
                echo ""
                echo "  nano /etc/pve/lxc/<CONTAINER_ID>.conf"
                echo ""
                echo "  # Diese Zeilen hinzufuegen:"
                echo "  lxc.cgroup2.devices.allow: c 10:200 rwm"
                echo "  lxc.mount.entry: /dev/net/tun dev/net/tun none bind,create=file"
                echo ""
                echo "  # Dann Container neustarten:"
                echo "  pct restart <CONTAINER_ID>"
                echo ""
                log_error "Bitte TUN aktivieren und Installer erneut ausfuehren."
                exit 0
                ;;
            6)
                log_info "Installation abgebrochen."
                exit 0
                ;;
            *)
                log_error "Ungueltige Auswahl"
                exit 1
                ;;
        esac
    else
        # TUN available or bare-metal - ask if user wants Tailscale
        if [[ "$UNATTENDED" == "true" ]]; then
            USE_TAILSCALE=true
            return 0
        fi

        echo ""
        echo -e "${BLUE}Netzwerk-Konfiguration:${NC}"
        echo ""
        echo "  1) Mit Tailscale (empfohlen fuer Remote-Zugriff via VPN)"
        echo "  2) Lokales Netzwerk (direkter Zugriff via IP, z.B. 192.168.178.78)"
        echo ""

        local choice
        read -rp "Auswahl [1/2] (Standard: 1): " choice < /dev/tty
        choice="${choice:-1}"

        case "$choice" in
            1)
                USE_TAILSCALE=true
                COMPOSE_FILE="docker-compose.yml"
                log_info "Verwende Tailscale (docker-compose.yml)"
                ;;
            2)
                USE_TAILSCALE=false
                COMPOSE_FILE="docker-compose.lxc.yml"
                log_info "Verwende lokales Netzwerk (docker-compose.lxc.yml)"
                ;;
            *)
                USE_TAILSCALE=true
                COMPOSE_FILE="docker-compose.yml"
                ;;
        esac
    fi
}

install_tailscale() {
    if [[ "${USE_TAILSCALE:-true}" != "true" ]]; then
        log_info "Tailscale uebersprungen (lokales Netzwerk gewaehlt)"
        return 0
    fi

    log_step "Installing Tailscale..."

    if command -v tailscale &>/dev/null && [[ "$FORCE" != "true" ]]; then
        log_success "Tailscale already installed: $(tailscale version | head -1)"
        return 0
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would install Tailscale"
        return 0
    fi

    # Use verified download instead of curl|sh for security
    # Tailscale script changes frequently, so checksum verification is optional
    # The function will warn if no checksum is provided and display the actual hash
    verified_download_and_run \
        "https://tailscale.com/install.sh" \
        "${KNOWN_CHECKSUMS[tailscale]:-}" \
        "Tailscale installer"

    log_success "Tailscale installed"
}

setup_native_tailscale() {
    log_step "Checking host Tailscale for native-tailscale mode..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would verify host Tailscale configuration"
        return 0
    fi

    # Check if Tailscale is installed
    if ! command -v tailscale &>/dev/null; then
        log_error "Tailscale is not installed on the host"
        log_error "Native-tailscale mode requires Tailscale to be installed on the host"
        log_info "Install Tailscale with: curl -fsSL https://tailscale.com/install.sh | sh"
        return 1
    fi

    # Check if Tailscale is running
    if ! tailscale status &>/dev/null; then
        log_warning "Tailscale is installed but not running"
        log_info "Start Tailscale with: sudo tailscale up"

        if [[ "$UNATTENDED" != "true" ]]; then
            if confirm "Would you like to start Tailscale now?" "y"; then
                sudo tailscale up
            fi
        fi
    fi

    # Get Tailscale IP
    local ts_ip
    ts_ip=$(tailscale ip -4 2>/dev/null || echo "")

    if [[ -n "$ts_ip" ]]; then
        log_success "Host Tailscale is running"
        log_info "Tailscale IP: $ts_ip"
        echo ""
        log_info "Services will be accessible via:"
        echo "  code-server: https://${ts_ip}:8443"
        echo "  Claude ttyd: http://${ts_ip}:7681"
        echo ""

        # Optional: Ask about tailscale serve
        if [[ "$UNATTENDED" != "true" ]]; then
            echo -e "${BLUE}Optional: Tailscale Serve${NC}"
            echo "You can use 'tailscale serve' to expose services with HTTPS"
            echo "and automatic certificates. This is optional."
            echo ""
            echo "Example:"
            echo "  tailscale serve --bg 8443"
            echo ""
        fi
    else
        log_warning "Could not determine Tailscale IP"
        log_info "Make sure Tailscale is connected with: tailscale up"
    fi

    log_success "Native Tailscale mode configured"
}

install_native_tailscale_userspace() {
    log_step "Installing Tailscale in native userspace mode..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would install Tailscale in userspace mode"
        return 0
    fi

    # Install Tailscale if not already present
    if ! command -v tailscale &>/dev/null; then
        log_step "Installing Tailscale..."
        verified_download_and_run \
            "https://tailscale.com/install.sh" \
            "${KNOWN_CHECKSUMS[tailscale]:-}" \
            "Tailscale installer"
    else
        log_success "Tailscale already installed: $(tailscale version | head -1)"
    fi

    # Stop existing tailscaled if running
    sudo systemctl stop tailscaled 2>/dev/null || true

    # Create state directory
    sudo mkdir -p /var/lib/tailscale

    # Create systemd service for userspace mode
    log_step "Creating Tailscale userspace systemd service..."
    sudo tee /etc/systemd/system/tailscaled-userspace.service > /dev/null << 'SYSTEMD_EOF'
[Unit]
Description=Tailscale node agent (userspace networking mode)
Documentation=https://tailscale.com/kb/
Wants=network-pre.target
After=network-pre.target NetworkManager.service systemd-resolved.service

[Service]
ExecStart=/usr/sbin/tailscaled --state=/var/lib/tailscale/tailscaled.state --socket=/var/run/tailscale/tailscaled.sock --tun=userspace-networking
ExecStopPost=/usr/sbin/tailscaled --cleanup
Restart=on-failure
RuntimeDirectory=tailscale
RuntimeDirectoryMode=0755
StateDirectory=tailscale
StateDirectoryMode=0700
CacheDirectory=tailscale
CacheDirectoryMode=0750
Type=notify
Environment=TS_DEBUG_FIREWALL_MODE=auto

[Install]
WantedBy=multi-user.target
SYSTEMD_EOF

    # Disable standard tailscaled service if exists
    sudo systemctl disable tailscaled 2>/dev/null || true

    # Enable and start userspace service
    sudo systemctl daemon-reload
    sudo systemctl enable tailscaled-userspace
    sudo systemctl start tailscaled-userspace

    # Wait for service to be ready
    log_step "Waiting for Tailscale daemon..."
    local attempts=0
    while [[ $attempts -lt 30 ]]; do
        if sudo tailscale status &>/dev/null 2>&1; then
            break
        fi
        ((attempts++))
        sleep 1
    done

    # Authenticate with Tailscale
    if [[ -n "$TAILSCALE_KEY" ]]; then
        log_step "Authenticating with Tailscale using auth key..."
        sudo tailscale up --authkey="$TAILSCALE_KEY" --accept-routes=false
    else
        log_step "Starting Tailscale authentication..."
        echo ""
        log_info "Please authenticate Tailscale in your browser"
        echo ""

        if [[ "$UNATTENDED" == "true" ]]; then
            log_warning "No auth key provided in unattended mode"
            log_info "Run 'sudo tailscale up' after installation to authenticate"
        else
            sudo tailscale up --accept-routes=false
        fi
    fi

    # Get and display Tailscale IP
    local ts_ip
    ts_ip=$(tailscale ip -4 2>/dev/null || echo "")

    if [[ -n "$ts_ip" ]]; then
        log_success "Tailscale connected: $ts_ip"
    else
        log_warning "Could not get Tailscale IP - run 'tailscale ip' after authentication"
    fi

    log_success "Native Tailscale userspace mode installed"
}

setup_tailscale_serve() {
    log_step "Setting up Tailscale Serve..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would configure Tailscale Serve"
        return 0
    fi

    # Check if Tailscale is running
    if ! tailscale status &>/dev/null; then
        log_warning "Tailscale not running, skipping serve setup"
        return 0
    fi

    # Wait for services to be available
    log_step "Waiting for Docker services to start..."
    local attempts=0
    while [[ $attempts -lt 60 ]]; do
        if curl -sf -o /dev/null "http://127.0.0.1:8443" -k 2>/dev/null; then
            break
        fi
        ((attempts++))
        sleep 2
    done

    # Run the setup script if available
    if [[ -x "${SCRIPT_DIR}/setup-tailscale-serve.sh" ]]; then
        "${SCRIPT_DIR}/setup-tailscale-serve.sh" setup
    else
        # Fallback: configure serve manually
        log_step "Configuring Tailscale Serve for code-server..."
        tailscale serve --bg --https=443 http://127.0.0.1:8443 2>/dev/null || \
            tailscale serve --bg 443 http://127.0.0.1:8443 2>/dev/null || \
            log_warning "Could not configure Tailscale Serve"
    fi

    # Display access info
    local ts_ip
    ts_ip=$(tailscale ip -4 2>/dev/null || echo "")
    if [[ -n "$ts_ip" ]]; then
        echo ""
        log_success "Services erreichbar ueber Tailscale:"
        echo "  code-server: https://${ts_ip}/"
        echo ""
    fi
}

setup_environment() {
    log_step "Setting up environment..."

    cd "$PROJECT_DIR"

    if [[ ! -f .env ]]; then
        if [[ -n "$ENV_FILE" && -f "$ENV_FILE" ]]; then
            cp "$ENV_FILE" .env
            log_info "Copied environment from $ENV_FILE"
        else
            cp .env.example .env
            log_info "Created .env from template"

            if [[ "$UNATTENDED" != "true" ]]; then
                log_warning "Please edit .env with your configuration before starting services"
            fi
        fi
    else
        log_info ".env already exists"
    fi

    # Create workspace directory
    mkdir -p workspace

    # Create secrets directory with placeholder
    mkdir -p secrets
    if [[ ! -f secrets/anthropic_api_key.txt ]]; then
        if [[ -n "$ANTHROPIC_KEY" ]]; then
            echo "$ANTHROPIC_KEY" > secrets/anthropic_api_key.txt
            log_info "Anthropic API key configured from CLI parameter"
        else
            echo "your-api-key-here" > secrets/anthropic_api_key.txt
            log_warning "Please update secrets/anthropic_api_key.txt with your actual API key"
        fi
    fi

    # Update .env with CLI parameters if provided
    if [[ "$UNATTENDED" == "true" ]]; then
        if [[ -n "$TAILSCALE_KEY" ]]; then
            sed -i "s/^TS_AUTHKEY=.*/TS_AUTHKEY=${TAILSCALE_KEY}/" .env
            log_info "Tailscale key configured"
        fi

        if [[ -n "$CODE_PASSWORD" ]]; then
            sed -i "s/^CODE_SERVER_PASSWORD=.*/CODE_SERVER_PASSWORD=${CODE_PASSWORD}/" .env
            log_info "code-server password configured"
        fi

        if [[ -n "$ANTHROPIC_KEY" ]]; then
            sed -i "s|^ANTHROPIC_API_KEY=.*|ANTHROPIC_API_KEY=${ANTHROPIC_KEY}|" .env
            log_info "Anthropic API key configured in .env"
        fi
    fi

    log_success "Environment configured"
}

start_services() {
    log_step "Starting Docker services..."

    cd "$PROJECT_DIR"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would start Docker services with ${COMPOSE_FILE}"
        return 0
    fi

    # Check if compose file exists
    if [[ ! -f "$COMPOSE_FILE" ]]; then
        log_error "Compose file not found: ${COMPOSE_FILE}"
        return 1
    fi

    log_info "Using compose file: ${COMPOSE_FILE}"

    # Validate compose file
    docker compose -f "$COMPOSE_FILE" config > /dev/null

    # Service detection and conflict resolution
    if [[ "$SERVICE_MANAGER_LOADED" == "true" ]]; then
        log_step "Detecting existing services..."

        # Show service summary
        show_service_summary

        # Check for existing doom installation
        if has_existing_installation; then
            log_info "Existing doom-coding installation detected"

            if [[ "$FORCE" == "true" ]]; then
                log_info "Force mode: stopping and removing existing installation..."
                backup_existing_config || log_warning "Backup failed, continuing anyway"
                stop_doom_services 30 "$COMPOSE_FILE" 2>/dev/null || true
                docker rm -f doom-tailscale doom-code-server doom-claude 2>/dev/null || true
            else
                log_info "Upgrading existing installation..."
                backup_existing_config || log_warning "Backup failed, continuing anyway"
                stop_doom_services 30 "$COMPOSE_FILE" 2>/dev/null || true
            fi
        fi

        # Check for port conflicts (non-doom services)
        local conflicts
        conflicts=$(check_port_conflicts 2>/dev/null || echo "[]")
        if [[ "$conflicts" != "[]" ]] && [[ -n "$conflicts" ]]; then
            log_warning "Port conflicts detected"

            # Extract conflicting ports and handle them
            local ports
            ports=$(echo "$conflicts" | jq -r '.[].port' 2>/dev/null || true)

            for port in $ports; do
                [[ -z "$port" ]] && continue

                # Skip if conflict is from our own containers (being stopped)
                local info
                info=$(get_port_info "$port" 2>/dev/null || echo '{}')
                local container
                container=$(echo "$info" | jq -r '.container // empty' 2>/dev/null || true)

                if [[ "$container" =~ ^doom- ]]; then
                    log_info "Port $port will be freed by stopping existing doom container"
                    continue
                fi

                local process_name
                process_name=$(echo "$info" | jq -r '.process // "unknown"' 2>/dev/null || echo "unknown")
                log_warning "Port $port is in use by: $process_name"

                if [[ "$FORCE" == "true" ]]; then
                    # In force mode, try to stop if it's a container
                    if [[ -n "$container" ]]; then
                        log_info "Force mode: stopping conflicting container $container"
                        docker stop "$container" 2>/dev/null || true
                    else
                        log_warning "Cannot automatically stop process using port $port"
                    fi
                elif [[ "$UNATTENDED" != "true" ]]; then
                    local alt_port
                    alt_port=$(find_available_port $((port + 1)))
                    echo ""
                    echo "Options:"
                    echo "  1) Continue anyway (may fail if port is still in use)"
                    echo "  2) Abort installation"
                    echo ""
                    echo "  Note: Alternative port $alt_port is available"
                    echo ""

                    local choice
                    read -rp "Choice [1/2] (default: 1): " choice < /dev/tty
                    choice="${choice:-1}"

                    if [[ "$choice" == "2" ]]; then
                        log_info "Installation aborted by user"
                        return 1
                    fi
                fi
            done
        fi
    else
        # Fallback port check without service manager
        log_step "Checking port availability..."
        for port in 8443 7681; do
            if command -v nc &>/dev/null && nc -z localhost "$port" 2>/dev/null; then
                log_warning "Port $port appears to be in use"
                if [[ "$FORCE" == "true" ]]; then
                    # Try to stop doom containers that might be using it
                    docker stop doom-code-server doom-claude 2>/dev/null || true
                fi
            elif command -v ss &>/dev/null && ss -tln 2>/dev/null | grep -q ":${port} "; then
                log_warning "Port $port appears to be in use"
            fi
        done
    fi

    # Pull images with filtered output
    log_step "Pulling container images..."
    if [[ "$SERVICE_MANAGER_LOADED" == "true" ]] && [[ "$VERBOSE" != "true" ]]; then
        docker compose -f "$COMPOSE_FILE" pull 2>&1 | filter_docker_output
    else
        docker compose -f "$COMPOSE_FILE" pull
    fi

    # Build and start with filtered output
    log_step "Building and starting containers..."
    if [[ "$SERVICE_MANAGER_LOADED" == "true" ]] && [[ "$VERBOSE" != "true" ]]; then
        docker compose -f "$COMPOSE_FILE" build 2>&1 | filter_docker_output
        docker compose -f "$COMPOSE_FILE" up -d 2>&1 | filter_docker_output
    else
        docker compose -f "$COMPOSE_FILE" build
        docker compose -f "$COMPOSE_FILE" up -d
    fi

    # Check if services actually started
    local failed_containers
    failed_containers=$(docker ps -a --filter "name=doom-" --filter "status=created" --format '{{.Names}}' 2>/dev/null || true)

    if [[ -n "$failed_containers" ]]; then
        log_error "Some containers failed to start:"
        echo "$failed_containers" | while read -r container; do
            log_error "  ${container}: $(docker inspect ${container} --format '{{.State.Error}}' 2>/dev/null)"
        done
        log_warning "Cleaning up failed containers..."
        echo "$failed_containers" | xargs docker rm 2>/dev/null || true
        return 1
    fi

    # Wait for health and show access info
    if [[ "$SERVICE_MANAGER_LOADED" == "true" ]]; then
        wait_for_services 120 || log_warning "Some services are not yet healthy"
        show_access_info "$COMPOSE_FILE"
    else
        log_success "Services started"
        # Fallback access info
        echo ""
        if [[ "$NATIVE_TAILSCALE" == "true" ]]; then
            local ts_ip
            ts_ip=$(tailscale ip -4 2>/dev/null || echo "<TAILSCALE-IP>")
            log_info "Access via Host-Tailscale:"
            echo "  code-server: https://${ts_ip}:8443"
            echo "  Tip: 'tailscale ip' shows the Tailscale IP"
        elif [[ "${USE_TAILSCALE:-true}" == "true" ]]; then
            log_info "Access via Tailscale IP (after 'tailscale up'):"
            echo "  code-server: https://<TAILSCALE-IP>:8443"
        else
            local host_ip
            host_ip=$(hostname -I | awk '{print $1}')
            log_info "Access via local network:"
            echo "  code-server: https://${host_ip}:8443"
        fi
    fi
}

# ===========================================
# MAIN EXECUTION
# ===========================================
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --unattended)
                UNATTENDED=true
                shift
                ;;
            --skip-docker)
                SKIP_DOCKER=true
                shift
                ;;
            --skip-tailscale|--local-network|--no-tailscale)
                SKIP_TAILSCALE=true
                USE_TAILSCALE=false
                COMPOSE_FILE="docker-compose.lxc.yml"
                shift
                ;;
            --native-tailscale)
                NATIVE_TAILSCALE=true
                USE_TAILSCALE=false
                SKIP_TAILSCALE=true
                COMPOSE_FILE="docker-compose.native-tailscale.yml"
                shift
                ;;
            --native-userspace)
                NATIVE_USERSPACE=true
                USE_TAILSCALE=true
                SKIP_TAILSCALE=false
                COMPOSE_FILE="docker-compose.native-userspace.yml"
                shift
                ;;
            --skip-terminal)
                SKIP_TERMINAL=true
                shift
                ;;
            --skip-hardening)
                SKIP_HARDENING=true
                shift
                ;;
            --skip-secrets)
                SKIP_SECRETS=true
                shift
                ;;
            --env-file=*)
                ENV_FILE="${1#*=}"
                shift
                ;;
            --tailscale-key=*)
                TAILSCALE_KEY="${1#*=}"
                shift
                ;;
            --code-password=*)
                CODE_PASSWORD="${1#*=}"
                shift
                ;;
            --anthropic-key=*)
                ANTHROPIC_KEY="${1#*=}"
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --force)
                FORCE=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                set -x
                shift
                ;;
            --retry-failed)
                # TODO: Implement retry logic
                shift
                ;;
            --help|-h)
                print_help
                exit 0
                ;;
            --version|-v)
                print_version
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                print_help
                exit 1
                ;;
        esac
    done
}

main() {
    parse_arguments "$@"

    print_banner
    setup_logging

    log_info "Starting Doom Coding installation..."
    log_info "Log file: $LOG_FILE"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_warning "DRY RUN MODE - No changes will be made"
    fi

    # Validate input parameters if provided
    if [[ -n "$CODE_PASSWORD" ]]; then
        validate_password "$CODE_PASSWORD" || exit 1
    fi
    if [[ -n "$TAILSCALE_KEY" ]]; then
        validate_tailscale_key "$TAILSCALE_KEY"
    fi

    # Run pre-flight checks (sudo, internet, disk, memory, required commands)
    preflight_check || exit 1

    # System detection
    detect_os
    detect_arch
    detect_package_manager

    # Additional prerequisite checks
    check_root

    # Installation steps
    install_base_packages
    install_docker

    # Tailscale setup (interactive choice or native mode)
    if [[ "$NATIVE_TAILSCALE" == "true" ]]; then
        setup_native_tailscale
    elif [[ "$NATIVE_USERSPACE" == "true" ]]; then
        install_native_tailscale_userspace
    else
        setup_tailscale_choice
        # After choice, check if native userspace was selected
        if [[ "$NATIVE_USERSPACE" == "true" ]]; then
            install_native_tailscale_userspace
        else
            install_tailscale
        fi
    fi

    # Terminal tools
    if [[ "$SKIP_TERMINAL" != "true" ]]; then
        if [[ -x "$SCRIPT_DIR/setup-terminal.sh" ]]; then
            log_step "Running terminal setup..."
            "$SCRIPT_DIR/setup-terminal.sh"
        fi
    fi

    # SSH Hardening
    if [[ "$SKIP_HARDENING" != "true" ]]; then
        if [[ -x "$SCRIPT_DIR/setup-host.sh" ]]; then
            log_step "Running host setup..."
            "$SCRIPT_DIR/setup-host.sh"
        fi
    fi

    # Secrets setup
    if [[ "$SKIP_SECRETS" != "true" ]]; then
        if [[ -x "$SCRIPT_DIR/setup-secrets.sh" ]]; then
            log_step "Running secrets setup..."
            "$SCRIPT_DIR/setup-secrets.sh" init
        fi
    fi

    # Environment and services
    setup_environment

    if confirm "Start Docker services now?"; then
        start_services

        # Setup Tailscale Serve for native userspace mode
        if [[ "$NATIVE_USERSPACE" == "true" ]]; then
            setup_tailscale_serve
        fi
    fi

    # Health check
    if [[ -x "$SCRIPT_DIR/health-check.sh" ]]; then
        log_step "Running health check..."
        "$SCRIPT_DIR/health-check.sh" || true
    fi

    echo ""
    log_success "Installation completed!"
    echo ""

    # Show access QR code based on deployment mode
    if [[ "$NATIVE_TAILSCALE" == "true" ]]; then
        local ts_ip
        ts_ip=$(tailscale ip -4 2>/dev/null || echo "")
        if [[ -n "$ts_ip" ]]; then
            show_access_qr "$ts_ip" "8443" "https"
        fi
    elif [[ "$NATIVE_USERSPACE" == "true" ]]; then
        local ts_ip
        ts_ip=$(tailscale ip -4 2>/dev/null || echo "")
        if [[ -n "$ts_ip" ]]; then
            echo -e "${BLUE}Native Tailscale Userspace Mode:${NC}"
            echo "  code-server: https://${ts_ip}/"
            echo ""
            echo "  (Port 443 via Tailscale Serve)"
            show_access_qr "$ts_ip" "" "https"
        else
            echo -e "${BLUE}After Tailscale connects:${NC}"
            echo "  Run 'tailscale ip' to get your Tailscale IP"
            echo "  code-server: https://<TAILSCALE-IP>/"
            echo ""
        fi
    elif [[ "${USE_TAILSCALE:-true}" == "true" ]]; then
        echo -e "${BLUE}After Tailscale connects, access via:${NC}"
        echo "  https://<TAILSCALE-IP>:8443"
        echo ""
        echo "  Run 'tailscale ip' to get your Tailscale IP"
        echo ""
    else
        local host_ip
        host_ip=$(hostname -I | awk '{print $1}')
        if [[ -n "$host_ip" ]]; then
            show_access_qr "$host_ip" "8443" "https"
        fi
    fi

    echo -e "${GREEN}Next steps:${NC}"
    echo "  1. Edit .env with your configuration"
    echo "  2. Update secrets/anthropic_api_key.txt"
    echo "  3. Run: docker compose up -d"
    echo "  4. Check status: ./scripts/health-check.sh"
    echo ""

    # Offer mobile setup guide
    if [[ "$UNATTENDED" != "true" ]]; then
        if confirm "Show mobile setup guide with QR codes?" "n"; then
            show_mobile_setup_qr
        fi
    fi
}

main "$@"
