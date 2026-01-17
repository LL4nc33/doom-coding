#!/bin/bash
#===============================================================================
# Terminal Development Environment - Main Installer
# Browser-accessible terminal via ttyd + tmux + neovim + Claude CLI
#===============================================================================

set -euo pipefail

# Configuration
readonly INSTALL_DIR="/opt/terminal-dev-env"
readonly LOG_FILE="${INSTALL_DIR}/logs/install.log"
readonly CONFIG_DIR="${INSTALL_DIR}/config"
readonly SSL_DIR="${INSTALL_DIR}/ssl"
readonly BIN_DIR="${INSTALL_DIR}/bin"

# Default values
DEFAULT_PORT=7681
DEFAULT_DOMAIN="localhost"
TTYD_USER=""
TTYD_PASS=""
SSL_ENABLED=true
FORCE_REINSTALL=false

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

#===============================================================================
# Utility Functions
#===============================================================================

log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${timestamp} [${level}] ${message}" | tee -a "${LOG_FILE}" 2>/dev/null || echo -e "${timestamp} [${level}] ${message}"
}

info()    { echo -e "${BLUE}[INFO]${NC} $*"; log "INFO" "$*"; }
success() { echo -e "${GREEN}[OK]${NC} $*"; log "SUCCESS" "$*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; log "WARN" "$*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*"; log "ERROR" "$*"; }
die()     { error "$*"; exit 1; }

check_root() {
    if [[ $EUID -ne 0 ]]; then
        die "This script must be run as root (use sudo)"
    fi
}

#===============================================================================
# Platform Detection
#===============================================================================

detect_platform() {
    local platform=""
    local distro=""
    local version=""
    local wsl=false

    # Check for WSL
    if grep -qi microsoft /proc/version 2>/dev/null; then
        wsl=true
    fi

    # Detect OS
    if [[ -f /etc/os-release ]]; then
        source /etc/os-release
        distro="${ID:-unknown}"
        version="${VERSION_ID:-unknown}"

        case "${distro}" in
            ubuntu|debian|linuxmint|pop)
                platform="debian"
                ;;
            fedora|centos|rhel|rocky|almalinux)
                platform="redhat"
                ;;
            arch|manjaro|endeavouros)
                platform="arch"
                ;;
            alpine)
                platform="alpine"
                ;;
            *)
                platform="unknown"
                ;;
        esac
    elif [[ "$(uname -s)" == "Darwin" ]]; then
        platform="macos"
        distro="macos"
        version=$(sw_vers -productVersion)
    else
        platform="unknown"
        distro="unknown"
        version="unknown"
    fi

    echo "${platform}|${distro}|${version}|${wsl}"
}

#===============================================================================
# Dependency Installation
#===============================================================================

install_dependencies_debian() {
    info "Installing dependencies for Debian/Ubuntu..."

    apt-get update -qq

    # Core dependencies
    apt-get install -y --no-install-recommends \
        build-essential \
        cmake \
        git \
        curl \
        wget \
        ca-certificates \
        gnupg \
        lsb-release \
        software-properties-common

    # Terminal tools
    apt-get install -y --no-install-recommends \
        tmux \
        neovim \
        zsh \
        fzf \
        ripgrep \
        fd-find \
        bat \
        exa \
        jq \
        htop \
        tree

    # Networking
    apt-get install -y --no-install-recommends \
        nginx \
        openssl \
        ufw

    # ttyd (check if available in repos, otherwise build)
    if apt-cache show ttyd &>/dev/null; then
        apt-get install -y ttyd
    else
        install_ttyd_from_source
    fi

    success "Debian/Ubuntu dependencies installed"
}

install_dependencies_redhat() {
    info "Installing dependencies for RHEL/Fedora..."

    # Enable EPEL if needed
    if command -v dnf &>/dev/null; then
        dnf install -y epel-release 2>/dev/null || true
        dnf install -y \
            gcc gcc-c++ cmake git curl wget \
            tmux neovim zsh fzf ripgrep fd-find bat jq htop tree \
            nginx openssl firewalld
    else
        yum install -y epel-release 2>/dev/null || true
        yum install -y \
            gcc gcc-c++ cmake git curl wget \
            tmux neovim zsh fzf ripgrep jq htop tree \
            nginx openssl firewalld
    fi

    install_ttyd_from_source

    success "RHEL/Fedora dependencies installed"
}

install_dependencies_arch() {
    info "Installing dependencies for Arch Linux..."

    pacman -Syu --noconfirm
    pacman -S --noconfirm --needed \
        base-devel cmake git curl wget \
        tmux neovim zsh fzf ripgrep fd bat exa jq htop tree \
        nginx openssl ufw ttyd

    success "Arch Linux dependencies installed"
}

install_ttyd_from_source() {
    info "Building ttyd from source..."

    local ttyd_version="1.7.4"
    local build_dir="/tmp/ttyd-build"

    # Install build dependencies
    if command -v apt-get &>/dev/null; then
        apt-get install -y --no-install-recommends \
            libwebsockets-dev libjson-c-dev libssl-dev
    elif command -v dnf &>/dev/null; then
        dnf install -y libwebsockets-devel json-c-devel openssl-devel
    fi

    rm -rf "${build_dir}"
    mkdir -p "${build_dir}"
    cd "${build_dir}"

    curl -sL "https://github.com/tsl0922/ttyd/archive/refs/tags/${ttyd_version}.tar.gz" | tar xz
    cd "ttyd-${ttyd_version}"

    mkdir build && cd build
    cmake ..
    make -j$(nproc)
    make install

    cd /
    rm -rf "${build_dir}"

    success "ttyd built and installed"
}

#===============================================================================
# SSL Certificate Generation
#===============================================================================

generate_ssl_certificates() {
    local domain="${1:-localhost}"
    local ssl_dir="${SSL_DIR}"

    info "Generating SSL certificates for ${domain}..."

    mkdir -p "${ssl_dir}"

    # Generate private key
    openssl genrsa -out "${ssl_dir}/server.key" 4096

    # Create certificate signing request config
    cat > "${ssl_dir}/openssl.cnf" << EOF
[req]
default_bits = 4096
prompt = no
default_md = sha256
distinguished_name = dn
x509_extensions = v3_req

[dn]
C = DE
ST = State
L = City
O = Terminal Dev Environment
OU = Development
CN = ${domain}

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = ${domain}
DNS.2 = localhost
IP.1 = 127.0.0.1
IP.2 = 192.168.178.78
EOF

    # Generate self-signed certificate (valid for 365 days)
    openssl req -new -x509 -sha256 \
        -key "${ssl_dir}/server.key" \
        -out "${ssl_dir}/server.crt" \
        -days 365 \
        -config "${ssl_dir}/openssl.cnf"

    # Set permissions
    chmod 600 "${ssl_dir}/server.key"
    chmod 644 "${ssl_dir}/server.crt"

    success "SSL certificates generated"
}

#===============================================================================
# Configuration Deployment
#===============================================================================

deploy_configs() {
    info "Deploying configuration files..."

    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local source_config_dir="$(dirname "${script_dir}")/config"

    # Copy configurations
    cp -r "${source_config_dir}/nginx/"* "${CONFIG_DIR}/nginx/" 2>/dev/null || true
    cp -r "${source_config_dir}/tmux/"* "${CONFIG_DIR}/tmux/" 2>/dev/null || true
    cp -r "${source_config_dir}/neovim/"* "${CONFIG_DIR}/neovim/" 2>/dev/null || true
    cp -r "${source_config_dir}/zsh/"* "${CONFIG_DIR}/zsh/" 2>/dev/null || true

    success "Configurations deployed"
}

setup_nginx() {
    info "Configuring nginx..."

    local nginx_conf="${CONFIG_DIR}/nginx/terminal.conf"

    # Create nginx site configuration
    cat > "${nginx_conf}" << 'NGINX_CONF'
# Terminal Development Environment - nginx Configuration
# Mobile-optimized reverse proxy for ttyd

upstream ttyd_backend {
    server 127.0.0.1:7681;
    keepalive 32;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name _;

    # SSL Configuration
    ssl_certificate /opt/terminal-dev-env/ssl/server.crt;
    ssl_certificate_key /opt/terminal-dev-env/ssl/server.key;

    # Modern SSL settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;

    # HSTS
    add_header Strict-Transport-Security "max-age=63072000" always;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Basic authentication (credentials set during install)
    auth_basic "Terminal Access";
    auth_basic_user_file /opt/terminal-dev-env/config/nginx/.htpasswd;

    # Logging
    access_log /opt/terminal-dev-env/logs/nginx-access.log;
    error_log /opt/terminal-dev-env/logs/nginx-error.log;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=terminal_limit:10m rate=10r/s;

    location / {
        limit_req zone=terminal_limit burst=20 nodelay;

        proxy_pass http://ttyd_backend;
        proxy_http_version 1.1;

        # WebSocket support
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # Headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts for long-running terminal sessions
        proxy_connect_timeout 7d;
        proxy_send_timeout 7d;
        proxy_read_timeout 7d;

        # Buffer settings
        proxy_buffering off;
        proxy_buffer_size 4k;
    }

    # Health check endpoint
    location /health {
        auth_basic off;
        return 200 'OK';
        add_header Content-Type text/plain;
    }
}

# HTTP to HTTPS redirect
server {
    listen 80;
    listen [::]:80;
    server_name _;
    return 301 https://$host$request_uri;
}
NGINX_CONF

    # Link to nginx sites
    ln -sf "${nginx_conf}" /etc/nginx/sites-available/terminal 2>/dev/null || \
    ln -sf "${nginx_conf}" /etc/nginx/conf.d/terminal.conf 2>/dev/null || true

    if [[ -d /etc/nginx/sites-enabled ]]; then
        ln -sf /etc/nginx/sites-available/terminal /etc/nginx/sites-enabled/terminal
        rm -f /etc/nginx/sites-enabled/default 2>/dev/null || true
    fi

    success "nginx configured"
}

setup_htpasswd() {
    local user="${1:-admin}"
    local pass="${2:-}"

    if [[ -z "${pass}" ]]; then
        pass=$(openssl rand -base64 16)
        warn "Generated random password: ${pass}"
        echo "${pass}" > "${CONFIG_DIR}/nginx/.htpasswd.plain"
        chmod 600 "${CONFIG_DIR}/nginx/.htpasswd.plain"
    fi

    # Create htpasswd file
    echo "${user}:$(openssl passwd -apr1 "${pass}")" > "${CONFIG_DIR}/nginx/.htpasswd"
    chmod 600 "${CONFIG_DIR}/nginx/.htpasswd"

    info "Credentials: User=${user}"
}

#===============================================================================
# Service Setup
#===============================================================================

setup_systemd_services() {
    info "Setting up systemd services..."

    # ttyd service
    cat > /etc/systemd/system/ttyd.service << 'TTYD_SERVICE'
[Unit]
Description=ttyd - Terminal sharing over the web
Documentation=https://github.com/tsl0922/ttyd
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/ttyd \
    --port 7681 \
    --interface 127.0.0.1 \
    --max-clients 5 \
    --once \
    --ping-interval 30 \
    /opt/terminal-dev-env/bin/terminal-session.sh
Restart=always
RestartSec=3

# Security hardening
NoNewPrivileges=false
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/opt/terminal-dev-env/logs /tmp /var/tmp
PrivateTmp=true

[Install]
WantedBy=multi-user.target
TTYD_SERVICE

    # Terminal session wrapper script
    cat > "${BIN_DIR}/terminal-session.sh" << 'SESSION_SCRIPT'
#!/bin/bash
# Terminal session wrapper - starts tmux with persistent session

SESSION_NAME="dev"
TMUX_CONF="/opt/terminal-dev-env/config/tmux/tmux.conf"

# Check if session exists
if tmux has-session -t "${SESSION_NAME}" 2>/dev/null; then
    exec tmux -f "${TMUX_CONF}" attach-session -t "${SESSION_NAME}"
else
    exec tmux -f "${TMUX_CONF}" new-session -s "${SESSION_NAME}"
fi
SESSION_SCRIPT
    chmod +x "${BIN_DIR}/terminal-session.sh"

    # Reload systemd
    systemctl daemon-reload

    success "systemd services configured"
}

#===============================================================================
# Firewall Configuration
#===============================================================================

configure_firewall() {
    info "Configuring firewall..."

    if command -v ufw &>/dev/null; then
        ufw --force reset
        ufw default deny incoming
        ufw default allow outgoing

        # Allow SSH
        ufw allow 22/tcp comment 'SSH'

        # Allow HTTPS for terminal access
        ufw allow 443/tcp comment 'HTTPS Terminal'

        # Allow HTTP (redirects to HTTPS)
        ufw allow 80/tcp comment 'HTTP Redirect'

        # Rate limiting for HTTP/HTTPS
        ufw limit 443/tcp

        ufw --force enable
        success "UFW firewall configured"

    elif command -v firewall-cmd &>/dev/null; then
        systemctl enable --now firewalld

        firewall-cmd --permanent --add-service=ssh
        firewall-cmd --permanent --add-service=http
        firewall-cmd --permanent --add-service=https
        firewall-cmd --reload

        success "firewalld configured"
    else
        warn "No supported firewall found"
    fi
}

#===============================================================================
# Shell Configuration
#===============================================================================

setup_zsh() {
    info "Setting up zsh..."

    local target_user="${SUDO_USER:-root}"
    local user_home=$(eval echo "~${target_user}")

    # Install Oh My Zsh if not present
    if [[ ! -d "${user_home}/.oh-my-zsh" ]]; then
        sudo -u "${target_user}" sh -c 'RUNZSH=no CHSH=no sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"' || true
    fi

    # Copy custom zsh configuration
    if [[ -f "${CONFIG_DIR}/zsh/.zshrc" ]]; then
        cp "${CONFIG_DIR}/zsh/.zshrc" "${user_home}/.zshrc"
        chown "${target_user}:${target_user}" "${user_home}/.zshrc"
    fi

    # Set zsh as default shell
    chsh -s "$(which zsh)" "${target_user}" 2>/dev/null || true

    success "zsh configured"
}

setup_neovim() {
    info "Setting up neovim..."

    local target_user="${SUDO_USER:-root}"
    local user_home=$(eval echo "~${target_user}")
    local nvim_config="${user_home}/.config/nvim"

    mkdir -p "${nvim_config}"

    # Copy neovim configuration
    if [[ -f "${CONFIG_DIR}/neovim/init.lua" ]]; then
        cp "${CONFIG_DIR}/neovim/init.lua" "${nvim_config}/init.lua"
        chown -R "${target_user}:${target_user}" "${nvim_config}"
    fi

    success "neovim configured"
}

#===============================================================================
# Claude CLI Installation
#===============================================================================

install_claude_cli() {
    info "Installing Claude CLI..."

    # Check if npm/node is available
    if ! command -v npm &>/dev/null; then
        # Install Node.js
        if command -v apt-get &>/dev/null; then
            curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
            apt-get install -y nodejs
        elif command -v dnf &>/dev/null; then
            dnf install -y nodejs npm
        elif command -v pacman &>/dev/null; then
            pacman -S --noconfirm nodejs npm
        fi
    fi

    # Install Claude CLI globally
    npm install -g @anthropic-ai/claude-code 2>/dev/null || {
        warn "Claude CLI installation failed - may need manual installation"
        return 0
    }

    success "Claude CLI installed"
}

#===============================================================================
# Health Check
#===============================================================================

run_health_check() {
    info "Running health check..."

    local errors=0

    # Check ttyd binary
    if command -v ttyd &>/dev/null; then
        success "ttyd binary found"
    else
        error "ttyd binary not found"
        ((errors++))
    fi

    # Check nginx
    if nginx -t &>/dev/null; then
        success "nginx configuration valid"
    else
        error "nginx configuration invalid"
        ((errors++))
    fi

    # Check SSL certificates
    if [[ -f "${SSL_DIR}/server.crt" ]] && [[ -f "${SSL_DIR}/server.key" ]]; then
        success "SSL certificates present"
    else
        error "SSL certificates missing"
        ((errors++))
    fi

    # Check tmux
    if command -v tmux &>/dev/null; then
        success "tmux available"
    else
        error "tmux not found"
        ((errors++))
    fi

    # Check neovim
    if command -v nvim &>/dev/null; then
        success "neovim available"
    else
        error "neovim not found"
        ((errors++))
    fi

    if [[ ${errors} -eq 0 ]]; then
        success "All health checks passed!"
        return 0
    else
        error "${errors} health check(s) failed"
        return 1
    fi
}

#===============================================================================
# Start Services
#===============================================================================

start_services() {
    info "Starting services..."

    systemctl enable --now ttyd
    systemctl enable --now nginx

    sleep 2

    if systemctl is-active --quiet ttyd && systemctl is-active --quiet nginx; then
        success "Services started successfully"
        info "Access terminal at: https://192.168.178.78/"
    else
        error "Some services failed to start"
        systemctl status ttyd --no-pager || true
        systemctl status nginx --no-pager || true
        return 1
    fi
}

#===============================================================================
# Usage
#===============================================================================

show_usage() {
    cat << EOF
Terminal Development Environment Installer

Usage: $0 [options]

Options:
    -h, --help          Show this help message
    -u, --user USER     Set authentication username (default: admin)
    -p, --pass PASS     Set authentication password (random if not set)
    -d, --domain DOMAIN Set domain name (default: localhost)
    --no-ssl            Disable SSL (not recommended)
    --force             Force reinstallation
    --skip-firewall     Skip firewall configuration
    --health-check      Run health check only

Examples:
    sudo $0
    sudo $0 -u myuser -p mypassword
    sudo $0 --health-check

EOF
}

#===============================================================================
# Main
#===============================================================================

main() {
    local skip_firewall=false
    local health_check_only=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                show_usage
                exit 0
                ;;
            -u|--user)
                TTYD_USER="$2"
                shift 2
                ;;
            -p|--pass)
                TTYD_PASS="$2"
                shift 2
                ;;
            -d|--domain)
                DEFAULT_DOMAIN="$2"
                shift 2
                ;;
            --no-ssl)
                SSL_ENABLED=false
                shift
                ;;
            --force)
                FORCE_REINSTALL=true
                shift
                ;;
            --skip-firewall)
                skip_firewall=true
                shift
                ;;
            --health-check)
                health_check_only=true
                shift
                ;;
            *)
                error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done

    # Health check only mode
    if [[ "${health_check_only}" == true ]]; then
        run_health_check
        exit $?
    fi

    # Start installation
    check_root

    echo ""
    echo "=========================================="
    echo "  Terminal Development Environment"
    echo "  Installer v0.0.6a"
    echo "=========================================="
    echo ""

    # Create installation directory
    mkdir -p "${INSTALL_DIR}"/{bin,config/{nginx,ttyd,tmux,neovim,zsh},ssl,logs,docs}

    # Detect platform
    IFS='|' read -r platform distro version wsl <<< "$(detect_platform)"
    info "Detected: ${distro} ${version} (${platform})"
    [[ "${wsl}" == "true" ]] && info "Running in WSL2"

    # Install dependencies based on platform
    case "${platform}" in
        debian)
            install_dependencies_debian
            ;;
        redhat)
            install_dependencies_redhat
            ;;
        arch)
            install_dependencies_arch
            ;;
        *)
            die "Unsupported platform: ${platform}"
            ;;
    esac

    # Deploy configurations
    deploy_configs

    # Generate SSL certificates
    if [[ "${SSL_ENABLED}" == true ]]; then
        generate_ssl_certificates "${DEFAULT_DOMAIN}"
    fi

    # Setup components
    setup_nginx
    setup_htpasswd "${TTYD_USER:-admin}" "${TTYD_PASS}"
    setup_systemd_services
    setup_zsh
    setup_neovim

    # Install Claude CLI
    install_claude_cli

    # Configure firewall
    if [[ "${skip_firewall}" != true ]]; then
        configure_firewall
    fi

    # Run health check
    run_health_check

    # Start services
    start_services

    echo ""
    echo "=========================================="
    echo "  Installation Complete!"
    echo "=========================================="
    echo ""
    echo "Access your terminal at:"
    echo "  https://192.168.178.78/"
    echo ""
    echo "Credentials stored in:"
    echo "  ${CONFIG_DIR}/nginx/.htpasswd.plain"
    echo ""
    echo "Logs available at:"
    echo "  ${INSTALL_DIR}/logs/"
    echo ""
}

main "$@"
