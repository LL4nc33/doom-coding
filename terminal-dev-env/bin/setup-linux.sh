#!/bin/bash
#===============================================================================
# Terminal Development Environment - Linux Setup Script
# Handles Linux-specific installation and configuration
#===============================================================================

set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly INSTALL_DIR="/opt/terminal-dev-env"
readonly LOG_FILE="${INSTALL_DIR}/logs/setup-linux.log"

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

log() { echo -e "${BLUE}[INFO]${NC} $*" | tee -a "${LOG_FILE}"; }
success() { echo -e "${GREEN}[OK]${NC} $*" | tee -a "${LOG_FILE}"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*" | tee -a "${LOG_FILE}"; }
error() { echo -e "${RED}[ERROR]${NC} $*" | tee -a "${LOG_FILE}"; }
die() { error "$*"; exit 1; }

#===============================================================================
# System Verification
#===============================================================================

verify_system_requirements() {
    log "Verifying system requirements..."

    # Check memory (minimum 512MB)
    local total_mem=$(grep MemTotal /proc/meminfo | awk '{print $2}')
    if [[ ${total_mem} -lt 524288 ]]; then
        warn "System has less than 512MB RAM - performance may be affected"
    fi

    # Check disk space (minimum 1GB)
    local free_space=$(df -k / | awk 'NR==2 {print $4}')
    if [[ ${free_space} -lt 1048576 ]]; then
        die "Insufficient disk space. Need at least 1GB free."
    fi

    # Check kernel version (minimum 4.x for modern features)
    local kernel_version=$(uname -r | cut -d. -f1)
    if [[ ${kernel_version} -lt 4 ]]; then
        warn "Kernel version is older than 4.x - some features may not work"
    fi

    success "System requirements verified"
}

#===============================================================================
# Dependency Installation
#===============================================================================

install_core_dependencies() {
    log "Installing core dependencies..."

    if command -v apt-get &>/dev/null; then
        apt-get update -qq

        # Build tools
        apt-get install -y --no-install-recommends \
            build-essential \
            cmake \
            pkg-config \
            autoconf \
            automake \
            libtool

        # Libraries for ttyd
        apt-get install -y --no-install-recommends \
            libwebsockets-dev \
            libjson-c-dev \
            libssl-dev \
            zlib1g-dev

        # Core tools
        apt-get install -y --no-install-recommends \
            git \
            curl \
            wget \
            ca-certificates \
            gnupg \
            lsb-release

    elif command -v dnf &>/dev/null; then
        dnf install -y \
            gcc gcc-c++ cmake pkgconfig \
            libwebsockets-devel json-c-devel openssl-devel zlib-devel \
            git curl wget ca-certificates

    elif command -v pacman &>/dev/null; then
        pacman -Syu --noconfirm
        pacman -S --noconfirm --needed \
            base-devel cmake pkgconf \
            libwebsockets json-c openssl zlib \
            git curl wget ca-certificates

    else
        die "Unsupported package manager"
    fi

    success "Core dependencies installed"
}

install_terminal_tools() {
    log "Installing terminal tools..."

    if command -v apt-get &>/dev/null; then
        apt-get install -y --no-install-recommends \
            tmux \
            neovim \
            zsh \
            fzf \
            ripgrep \
            jq \
            htop \
            tree \
            ncdu \
            unzip \
            zip

        # fd-find (Debian/Ubuntu names it differently)
        apt-get install -y fd-find 2>/dev/null && \
            ln -sf $(which fdfind) /usr/local/bin/fd 2>/dev/null || true

        # bat
        apt-get install -y bat 2>/dev/null && \
            ln -sf $(which batcat) /usr/local/bin/bat 2>/dev/null || true

        # exa
        apt-get install -y exa 2>/dev/null || install_exa_from_release

    elif command -v dnf &>/dev/null; then
        dnf install -y \
            tmux neovim zsh fzf ripgrep fd-find bat jq htop tree ncdu

    elif command -v pacman &>/dev/null; then
        pacman -S --noconfirm --needed \
            tmux neovim zsh fzf ripgrep fd bat exa jq htop tree ncdu
    fi

    success "Terminal tools installed"
}

install_exa_from_release() {
    log "Installing exa from GitHub release..."

    local version="0.10.1"
    local arch=$(uname -m)

    case "${arch}" in
        x86_64) arch="x86_64" ;;
        aarch64) arch="aarch64" ;;
        *) warn "Unsupported architecture for exa: ${arch}"; return 0 ;;
    esac

    local url="https://github.com/ogham/exa/releases/download/v${version}/exa-linux-${arch}-v${version}.zip"

    curl -sL "${url}" -o /tmp/exa.zip
    unzip -o /tmp/exa.zip -d /tmp/exa
    cp /tmp/exa/bin/exa /usr/local/bin/
    chmod +x /usr/local/bin/exa
    rm -rf /tmp/exa /tmp/exa.zip

    success "exa installed from release"
}

install_ttyd() {
    log "Installing ttyd..."

    # Check if already installed
    if command -v ttyd &>/dev/null; then
        local current_version=$(ttyd --version 2>&1 | head -1 | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
        log "ttyd already installed (version: ${current_version})"
        return 0
    fi

    # Try package manager first
    if command -v apt-get &>/dev/null; then
        if apt-cache show ttyd &>/dev/null; then
            apt-get install -y ttyd
            success "ttyd installed from package manager"
            return 0
        fi
    fi

    # Build from source
    build_ttyd_from_source
}

build_ttyd_from_source() {
    log "Building ttyd from source..."

    local version="1.7.4"
    local build_dir="/tmp/ttyd-build-$$"

    mkdir -p "${build_dir}"
    cd "${build_dir}"

    # Download
    curl -sL "https://github.com/tsl0922/ttyd/archive/refs/tags/${version}.tar.gz" | tar xz
    cd "ttyd-${version}"

    # Build
    mkdir build && cd build
    cmake -DCMAKE_BUILD_TYPE=Release ..
    make -j$(nproc)

    # Install
    make install

    # Cleanup
    cd /
    rm -rf "${build_dir}"

    # Verify
    if command -v ttyd &>/dev/null; then
        success "ttyd built and installed successfully"
    else
        die "ttyd installation failed"
    fi
}

install_nginx() {
    log "Installing nginx..."

    if command -v apt-get &>/dev/null; then
        apt-get install -y nginx

    elif command -v dnf &>/dev/null; then
        dnf install -y nginx

    elif command -v pacman &>/dev/null; then
        pacman -S --noconfirm nginx
    fi

    # Stop nginx if running (will configure and restart later)
    systemctl stop nginx 2>/dev/null || true

    success "nginx installed"
}

#===============================================================================
# Configuration
#===============================================================================

configure_nginx() {
    log "Configuring nginx..."

    local nginx_conf_dir=""
    local nginx_sites_enabled=""

    # Determine nginx config location
    if [[ -d /etc/nginx/sites-available ]]; then
        nginx_conf_dir="/etc/nginx/sites-available"
        nginx_sites_enabled="/etc/nginx/sites-enabled"
    else
        nginx_conf_dir="/etc/nginx/conf.d"
    fi

    # Copy configuration
    if [[ -f "${INSTALL_DIR}/config/nginx/terminal.conf" ]]; then
        cp "${INSTALL_DIR}/config/nginx/terminal.conf" "${nginx_conf_dir}/terminal.conf"
    else
        # Create default nginx config
        create_nginx_config "${nginx_conf_dir}/terminal.conf"
    fi

    # Create symlink for sites-enabled
    if [[ -n "${nginx_sites_enabled}" ]]; then
        ln -sf "${nginx_conf_dir}/terminal.conf" "${nginx_sites_enabled}/terminal.conf"
        rm -f "${nginx_sites_enabled}/default" 2>/dev/null || true
    fi

    # Remove default config that might conflict
    rm -f /etc/nginx/conf.d/default.conf 2>/dev/null || true

    # Test configuration
    nginx -t

    success "nginx configured"
}

create_nginx_config() {
    local config_file="$1"

    cat > "${config_file}" << 'EOF'
# Terminal Development Environment - nginx Configuration

upstream ttyd_backend {
    server 127.0.0.1:7681;
    keepalive 32;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name _;

    ssl_certificate /opt/terminal-dev-env/ssl/server.crt;
    ssl_certificate_key /opt/terminal-dev-env/ssl/server.key;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;

    add_header Strict-Transport-Security "max-age=63072000" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    auth_basic "Terminal Access";
    auth_basic_user_file /opt/terminal-dev-env/config/nginx/.htpasswd;

    access_log /opt/terminal-dev-env/logs/nginx-access.log;
    error_log /opt/terminal-dev-env/logs/nginx-error.log;

    location / {
        proxy_pass http://ttyd_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 7d;
        proxy_send_timeout 7d;
        proxy_read_timeout 7d;
        proxy_buffering off;
    }

    location /health {
        auth_basic off;
        return 200 'OK';
        add_header Content-Type text/plain;
    }
}

server {
    listen 80;
    listen [::]:80;
    server_name _;
    return 301 https://$host$request_uri;
}
EOF
}

configure_systemd() {
    log "Configuring systemd services..."

    # Create ttyd service
    cat > /etc/systemd/system/ttyd.service << 'EOF'
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
    --ping-interval 30 \
    /opt/terminal-dev-env/bin/terminal-session.sh
Restart=always
RestartSec=3
StandardOutput=append:/opt/terminal-dev-env/logs/ttyd.log
StandardError=append:/opt/terminal-dev-env/logs/ttyd-error.log

# Security
ProtectSystem=strict
ReadWritePaths=/opt/terminal-dev-env/logs /tmp /var/tmp /root /home
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

    # Create terminal session script
    cat > "${INSTALL_DIR}/bin/terminal-session.sh" << 'EOF'
#!/bin/bash
# Terminal session wrapper - starts tmux with persistent session

SESSION_NAME="dev"
TMUX_CONF="/opt/terminal-dev-env/config/tmux/tmux.conf"

# Ensure tmux config exists
if [[ ! -f "${TMUX_CONF}" ]]; then
    TMUX_CONF=""
fi

# Check if session exists
if tmux has-session -t "${SESSION_NAME}" 2>/dev/null; then
    exec tmux ${TMUX_CONF:+-f "${TMUX_CONF}"} attach-session -t "${SESSION_NAME}"
else
    exec tmux ${TMUX_CONF:+-f "${TMUX_CONF}"} new-session -s "${SESSION_NAME}"
fi
EOF
    chmod +x "${INSTALL_DIR}/bin/terminal-session.sh"

    # Reload systemd
    systemctl daemon-reload

    success "systemd services configured"
}

configure_firewall() {
    log "Configuring firewall..."

    if command -v ufw &>/dev/null; then
        # UFW (Debian/Ubuntu)
        ufw --force reset
        ufw default deny incoming
        ufw default allow outgoing
        ufw allow 22/tcp comment 'SSH'
        ufw allow 80/tcp comment 'HTTP'
        ufw allow 443/tcp comment 'HTTPS'
        ufw --force enable

        success "UFW firewall configured"

    elif command -v firewall-cmd &>/dev/null; then
        # firewalld (RHEL/Fedora)
        systemctl enable --now firewalld
        firewall-cmd --permanent --add-service=ssh
        firewall-cmd --permanent --add-service=http
        firewall-cmd --permanent --add-service=https
        firewall-cmd --reload

        success "firewalld configured"

    else
        warn "No supported firewall found - please configure manually"
    fi
}

#===============================================================================
# User Setup
#===============================================================================

setup_user_configs() {
    local target_user="${SUDO_USER:-root}"
    local user_home=$(eval echo "~${target_user}")

    log "Setting up user configurations for ${target_user}..."

    # Install Oh My Zsh
    if [[ ! -d "${user_home}/.oh-my-zsh" ]]; then
        log "Installing Oh My Zsh..."
        sudo -u "${target_user}" sh -c 'RUNZSH=no CHSH=no sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"' || true
    fi

    # Install tmux plugin manager
    if [[ ! -d "${user_home}/.tmux/plugins/tpm" ]]; then
        log "Installing tmux plugin manager..."
        sudo -u "${target_user}" git clone https://github.com/tmux-plugins/tpm "${user_home}/.tmux/plugins/tpm" || true
    fi

    # Copy zsh configuration
    if [[ -f "${INSTALL_DIR}/config/zsh/.zshrc" ]]; then
        cp "${INSTALL_DIR}/config/zsh/.zshrc" "${user_home}/.zshrc"
        chown "${target_user}:${target_user}" "${user_home}/.zshrc"
    fi

    # Setup neovim configuration
    local nvim_config="${user_home}/.config/nvim"
    mkdir -p "${nvim_config}"
    if [[ -f "${INSTALL_DIR}/config/neovim/init.lua" ]]; then
        cp "${INSTALL_DIR}/config/neovim/init.lua" "${nvim_config}/init.lua"
        chown -R "${target_user}:${target_user}" "${user_home}/.config"
    fi

    # Set default shell to zsh
    if command -v zsh &>/dev/null; then
        chsh -s "$(which zsh)" "${target_user}" 2>/dev/null || true
    fi

    success "User configurations set up"
}

#===============================================================================
# SSL Setup
#===============================================================================

generate_ssl_certificates() {
    local domain="${1:-localhost}"

    log "Generating SSL certificates..."

    mkdir -p "${INSTALL_DIR}/ssl"

    # Generate private key
    openssl genrsa -out "${INSTALL_DIR}/ssl/server.key" 4096

    # Create config for SAN
    cat > "${INSTALL_DIR}/ssl/openssl.cnf" << EOF
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

    # Generate certificate
    openssl req -new -x509 -sha256 \
        -key "${INSTALL_DIR}/ssl/server.key" \
        -out "${INSTALL_DIR}/ssl/server.crt" \
        -days 365 \
        -config "${INSTALL_DIR}/ssl/openssl.cnf"

    chmod 600 "${INSTALL_DIR}/ssl/server.key"
    chmod 644 "${INSTALL_DIR}/ssl/server.crt"

    success "SSL certificates generated"
}

setup_authentication() {
    local user="${1:-admin}"
    local pass="${2:-}"

    log "Setting up authentication..."

    mkdir -p "${INSTALL_DIR}/config/nginx"

    if [[ -z "${pass}" ]]; then
        pass=$(openssl rand -base64 16)
        log "Generated password: ${pass}"
        echo "${pass}" > "${INSTALL_DIR}/config/nginx/.htpasswd.plain"
        chmod 600 "${INSTALL_DIR}/config/nginx/.htpasswd.plain"
    fi

    echo "${user}:$(openssl passwd -apr1 "${pass}")" > "${INSTALL_DIR}/config/nginx/.htpasswd"
    chmod 600 "${INSTALL_DIR}/config/nginx/.htpasswd"

    success "Authentication configured (user: ${user})"
}

#===============================================================================
# Service Management
#===============================================================================

start_services() {
    log "Starting services..."

    systemctl enable ttyd nginx
    systemctl start ttyd
    systemctl start nginx

    sleep 2

    if systemctl is-active --quiet ttyd && systemctl is-active --quiet nginx; then
        success "Services started successfully"
    else
        error "Some services failed to start"
        systemctl status ttyd --no-pager || true
        systemctl status nginx --no-pager || true
        return 1
    fi
}

#===============================================================================
# Main
#===============================================================================

main() {
    mkdir -p "${INSTALL_DIR}/logs"

    log "Starting Linux setup..."

    verify_system_requirements
    install_core_dependencies
    install_terminal_tools
    install_ttyd
    install_nginx

    generate_ssl_certificates "localhost"
    setup_authentication "admin"

    configure_nginx
    configure_systemd
    configure_firewall
    setup_user_configs

    start_services

    success "Linux setup complete!"

    echo ""
    echo "=========================================="
    echo "  Setup Complete!"
    echo "=========================================="
    echo ""
    echo "Access: https://192.168.178.78/"
    echo ""
    echo "Credentials in: ${INSTALL_DIR}/config/nginx/.htpasswd.plain"
    echo ""
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
