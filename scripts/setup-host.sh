#!/usr/bin/env bash
# Doom Coding - Host Setup Script
# Configures SSH hardening and system settings
set -euo pipefail

# ===========================================
# COLORS
# ===========================================
readonly GREEN='\033[38;2;46;82;29m'
readonly RED='\033[0;31m'
readonly YELLOW='\033[0;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# ===========================================
# CONFIGURATION
# ===========================================
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
readonly SSH_CONFIG_DIR="/etc/ssh/sshd_config.d"
readonly SSH_HARDENING_FILE="99-doom-hardening.conf"

# ===========================================
# LOGGING
# ===========================================
log_info() { echo -e "${BLUE}ℹ${NC}  $*"; }
log_success() { echo -e "${GREEN}✅${NC} $*"; }
log_warning() { echo -e "${YELLOW}⚠${NC}  $*"; }
log_error() { echo -e "${RED}❌${NC} $*" >&2; }

# ===========================================
# SSH HARDENING
# ===========================================
setup_ssh_hardening() {
    log_info "Setting up SSH hardening..."

    # Check if sshd_config.d is supported
    if [[ ! -d "$SSH_CONFIG_DIR" ]]; then
        sudo mkdir -p "$SSH_CONFIG_DIR"
    fi

    # Check if Include directive exists
    if ! grep -q "^Include.*sshd_config.d" /etc/ssh/sshd_config 2>/dev/null; then
        log_warning "Adding Include directive to sshd_config"
        echo "Include ${SSH_CONFIG_DIR}/*.conf" | sudo tee -a /etc/ssh/sshd_config > /dev/null
    fi

    # Copy hardening config
    if [[ -f "$PROJECT_DIR/config/ssh/99-hardening.conf" ]]; then
        sudo cp "$PROJECT_DIR/config/ssh/99-hardening.conf" "$SSH_CONFIG_DIR/$SSH_HARDENING_FILE"
        sudo chmod 644 "$SSH_CONFIG_DIR/$SSH_HARDENING_FILE"
        log_success "SSH hardening config installed"
    else
        log_warning "SSH hardening config not found, creating default..."
        create_default_ssh_config
    fi

    # Test SSH configuration
    if sudo sshd -t 2>/dev/null; then
        log_success "SSH configuration is valid"

        # Restart SSH service
        if systemctl is-active --quiet sshd 2>/dev/null; then
            sudo systemctl reload sshd
            log_success "SSH service reloaded"
        elif systemctl is-active --quiet ssh 2>/dev/null; then
            sudo systemctl reload ssh
            log_success "SSH service reloaded"
        fi
    else
        log_error "SSH configuration test failed!"
        log_warning "Removing invalid config..."
        sudo rm -f "$SSH_CONFIG_DIR/$SSH_HARDENING_FILE"
        return 1
    fi
}

create_default_ssh_config() {
    sudo tee "$SSH_CONFIG_DIR/$SSH_HARDENING_FILE" > /dev/null << 'EOF'
# Doom Coding SSH Hardening
# Mozilla Modern SSH Configuration

# Authentication
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
AuthenticationMethods publickey
PermitEmptyPasswords no
MaxAuthTries 3

# Security
X11Forwarding no
AllowTcpForwarding no
AllowAgentForwarding no
PermitTunnel no

# Cryptography (Mozilla Modern)
KexAlgorithms curve25519-sha256@libssh.org,curve25519-sha256
Ciphers chacha20-poly1305@openssh.com,aes256-gcm@openssh.com,aes128-gcm@openssh.com
MACs hmac-sha2-512-etm@openssh.com,hmac-sha2-256-etm@openssh.com
HostKeyAlgorithms ssh-ed25519,rsa-sha2-512,rsa-sha2-256

# Timeouts
ClientAliveInterval 300
ClientAliveCountMax 2
LoginGraceTime 30

# Logging
LogLevel VERBOSE
EOF
    sudo chmod 644 "$SSH_CONFIG_DIR/$SSH_HARDENING_FILE"
    log_success "Default SSH hardening config created"
}

# ===========================================
# FAIL2BAN SETUP
# ===========================================
setup_fail2ban() {
    log_info "Setting up fail2ban..."

    if ! command -v fail2ban-client &>/dev/null; then
        log_info "Installing fail2ban..."
        if command -v apt-get &>/dev/null; then
            sudo apt-get update
            sudo apt-get install -y fail2ban
        elif command -v pacman &>/dev/null; then
            sudo pacman -S --noconfirm fail2ban
        else
            log_warning "Could not install fail2ban - unsupported package manager"
            return 0
        fi
    fi

    # Create jail.local if it doesn't exist
    if [[ ! -f /etc/fail2ban/jail.local ]]; then
        sudo tee /etc/fail2ban/jail.local > /dev/null << 'EOF'
[DEFAULT]
bantime = 1h
findtime = 10m
maxretry = 3
backend = systemd

[sshd]
enabled = true
port = ssh
logpath = %(sshd_log)s
maxretry = 3
bantime = 1h
EOF
        log_success "fail2ban jail.local created"
    fi

    # Enable and start fail2ban
    sudo systemctl enable fail2ban
    sudo systemctl restart fail2ban

    log_success "fail2ban configured and started"
}

# ===========================================
# FIREWALL SETUP (UFW)
# ===========================================
setup_firewall() {
    log_info "Checking firewall..."

    if command -v ufw &>/dev/null; then
        # Check if UFW is active
        if sudo ufw status | grep -q "Status: active"; then
            log_info "UFW is active"

            # Ensure SSH is allowed
            sudo ufw allow ssh
            log_success "SSH allowed through firewall"
        else
            log_warning "UFW is installed but not active"
            log_info "To enable: sudo ufw enable"
        fi
    else
        log_info "UFW not installed - firewall configuration skipped"
    fi
}

# ===========================================
# SYSTEM TWEAKS
# ===========================================
apply_system_tweaks() {
    log_info "Applying system tweaks..."

    # Increase file descriptor limits
    if ! grep -q "doom-coding" /etc/security/limits.conf 2>/dev/null; then
        sudo tee -a /etc/security/limits.conf > /dev/null << 'EOF'

# Doom Coding - Increased limits
* soft nofile 65535
* hard nofile 65535
* soft nproc 65535
* hard nproc 65535
EOF
        log_success "File descriptor limits increased"
    fi

    # Sysctl tweaks for Docker
    if [[ ! -f /etc/sysctl.d/99-doom-coding.conf ]]; then
        sudo tee /etc/sysctl.d/99-doom-coding.conf > /dev/null << 'EOF'
# Doom Coding - Sysctl tweaks
net.ipv4.ip_forward = 1
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOF
        sudo sysctl --system > /dev/null 2>&1 || true
        log_success "Sysctl tweaks applied"
    fi
}

# ===========================================
# MAIN
# ===========================================
main() {
    echo -e "${GREEN}Doom Coding - Host Setup${NC}"
    echo "========================="
    echo ""

    setup_ssh_hardening
    setup_fail2ban
    setup_firewall
    apply_system_tweaks

    echo ""
    log_success "Host setup completed!"
}

main "$@"
