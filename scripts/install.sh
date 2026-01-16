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
SKIP_TERMINAL=false
SKIP_HARDENING=false
SKIP_SECRETS=false
DRY_RUN=false
VERBOSE=false
FORCE=false
ENV_FILE=""

# Detected system info
OS_ID=""
OS_VERSION=""
ARCH=""
PKG_MANAGER=""

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
    --skip-terminal     Skip terminal tools setup
    --skip-hardening    Skip SSH hardening
    --skip-secrets      Skip SOPS/age setup
    --env-file=FILE     Use specific environment file
    --dry-run           Show what would be done without executing
    --force             Force reinstallation of all components
    --verbose           Enable verbose output
    --retry-failed      Retry previously failed installation steps
    --help, -h          Show this help message
    --version, -v       Show version information

EXAMPLES:
    $0                              Interactive installation
    $0 --unattended                 Fully automated installation
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

    if [[ "$UNATTENDED" == "true" ]]; then
        eval "$var_name=\"$default\""
        return
    fi

    local value
    if [[ "$secret" == "true" ]]; then
        read -rsp "$prompt [$default]: " value
        echo ""
    else
        read -rp "$prompt [$default]: " value
    fi

    eval "$var_name=\"${value:-$default}\""
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

install_tailscale() {
    log_step "Installing Tailscale..."

    if command -v tailscale &>/dev/null && [[ "$FORCE" != "true" ]]; then
        log_success "Tailscale already installed: $(tailscale version | head -1)"
        return 0
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would install Tailscale"
        return 0
    fi

    curl -fsSL https://tailscale.com/install.sh | sh

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
        echo "your-api-key-here" > secrets/anthropic_api_key.txt
        log_warning "Please update secrets/anthropic_api_key.txt with your actual API key"
    fi

    log_success "Environment configured"
}

start_services() {
    log_step "Starting Docker services..."

    cd "$PROJECT_DIR"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would start Docker services"
        return 0
    fi

    # Validate compose file
    docker compose config > /dev/null

    # Build and start
    docker compose build
    docker compose up -d

    log_success "Services started"
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

    # System detection
    detect_os
    detect_arch
    detect_package_manager

    # Prerequisites
    check_root
    check_internet

    # Installation steps
    install_base_packages
    install_docker
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
