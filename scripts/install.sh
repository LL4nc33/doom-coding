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
readonly INSTALLER_VERSION="1.0.0"

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
    --skip-terminal     Skip terminal tools setup
    --skip-hardening    Skip SSH hardening
    --skip-secrets      Skip SOPS/age setup
    --env-file=FILE     Use specific environment file
    --tailscale-key=KEY Tailscale auth key for unattended setup
    --code-password=PWD code-server password for unattended setup
    --anthropic-key=KEY Anthropic API key for Claude Code
    --dry-run           Show what would be done without executing
    --force             Force reinstallation of all components
    --verbose           Enable verbose output
    --retry-failed      Retry previously failed installation steps
    --help, -h          Show this help message
    --version, -v       Show version information

EXAMPLES:
    $0                              Interactive installation
    $0 --unattended                 Fully automated installation
    $0 --skip-tailscale             LXC without TUN device
    $0 --unattended \\
      --tailscale-key="tskey-auth-xxx" \\
      --code-password="secure-password" \\
      --anthropic-key="sk-ant-xxx"      Fully automated with credentials
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
        read -rp "$prompt [Y/n]: " yn
        yn="${yn:-y}"
    else
        read -rp "$prompt [y/N]: " yn
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
        read -rp "$prompt [$default]: " value
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
        log_warning "Tailscale benoetigt TUN fuer VPN-Funktionalitaet."
        echo ""
        echo -e "${YELLOW}Optionen:${NC}"
        echo "  1) Ohne Tailscale fortfahren (lokales Netzwerk, z.B. 192.168.x.x)"
        echo "  2) TUN in LXC aktivieren (erfordert Proxmox-Host Konfiguration)"
        echo "  3) Installation abbrechen"
        echo ""

        if [[ "$UNATTENDED" == "true" ]]; then
            log_info "Unattended mode: Verwende lokales Netzwerk (ohne Tailscale)"
            USE_TAILSCALE=false
            COMPOSE_FILE="docker-compose.lxc.yml"
            return 0
        fi

        local choice
        read -rp "Auswahl [1/2/3]: " choice
        case "$choice" in
            1)
                USE_TAILSCALE=false
                COMPOSE_FILE="docker-compose.lxc.yml"
                log_info "Verwende lokales Netzwerk (docker-compose.lxc.yml)"
                ;;
            2)
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
            3)
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
        read -rp "Auswahl [1/2] (Standard: 1): " choice
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

    # Build and start
    docker compose -f "$COMPOSE_FILE" build
    docker compose -f "$COMPOSE_FILE" up -d

    log_success "Services started"

    # Show access info
    echo ""
    if [[ "${USE_TAILSCALE:-true}" == "true" ]]; then
        log_info "Zugriff via Tailscale IP (nach 'tailscale up'):"
        echo "  code-server: https://<TAILSCALE-IP>:8443"
    else
        local host_ip
        host_ip=$(hostname -I | awk '{print $1}')
        log_info "Zugriff via lokales Netzwerk:"
        echo "  code-server: https://${host_ip}:8443"
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

    # Tailscale setup (interactive choice)
    setup_tailscale_choice
    install_tailscale

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
    fi

    # Health check
    if [[ -x "$SCRIPT_DIR/health-check.sh" ]]; then
        log_step "Running health check..."
        "$SCRIPT_DIR/health-check.sh" || true
    fi

    echo ""
    log_success "Installation completed!"
    echo ""
    echo -e "${GREEN}Next steps:${NC}"
    echo "  1. Edit .env with your configuration"
    echo "  2. Update secrets/anthropic_api_key.txt"
    echo "  3. Run: docker compose up -d"
    echo "  4. Check status: ./scripts/health-check.sh"
    echo ""
}

main "$@"
