#!/bin/bash
#===============================================================================
# Terminal Development Environment - Health Check Script
# Verifies all components are working correctly
#===============================================================================

set -euo pipefail

readonly INSTALL_DIR="/opt/terminal-dev-env"

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# Counters
CHECKS_PASSED=0
CHECKS_FAILED=0
CHECKS_WARNED=0

#===============================================================================
# Output Functions
#===============================================================================

print_header() {
    echo ""
    echo "=========================================="
    echo "  Terminal Development Environment"
    echo "  Health Check"
    echo "=========================================="
    echo ""
}

check_pass() {
    echo -e "  ${GREEN}✓${NC} $*"
    ((CHECKS_PASSED++))
}

check_fail() {
    echo -e "  ${RED}✗${NC} $*"
    ((CHECKS_FAILED++))
}

check_warn() {
    echo -e "  ${YELLOW}!${NC} $*"
    ((CHECKS_WARNED++))
}

section() {
    echo ""
    echo -e "${BLUE}[$1]${NC}"
}

#===============================================================================
# Check Functions
#===============================================================================

check_binaries() {
    section "Binary Checks"

    local binaries=(
        "ttyd:ttyd"
        "tmux:tmux"
        "nvim:neovim"
        "nginx:nginx"
        "zsh:zsh"
        "openssl:openssl"
        "curl:curl"
    )

    for entry in "${binaries[@]}"; do
        local cmd="${entry%%:*}"
        local name="${entry##*:}"

        if command -v "${cmd}" &>/dev/null; then
            local version=$(${cmd} --version 2>&1 | head -1 | grep -oP '[\d.]+' | head -1 || echo "unknown")
            check_pass "${name} (${version})"
        else
            check_fail "${name} not found"
        fi
    done

    # Optional tools
    local optional_tools=(
        "fzf:fzf"
        "rg:ripgrep"
        "fd:fd-find"
        "bat:bat"
        "exa:exa"
        "claude:claude-cli"
    )

    for entry in "${optional_tools[@]}"; do
        local cmd="${entry%%:*}"
        local name="${entry##*:}"

        if command -v "${cmd}" &>/dev/null; then
            check_pass "${name} (optional)"
        else
            check_warn "${name} not installed (optional)"
        fi
    done
}

check_services() {
    section "Service Checks"

    # Check if systemd is available
    if [[ -d /run/systemd/system ]]; then
        # Check ttyd service
        if systemctl is-active --quiet ttyd 2>/dev/null; then
            check_pass "ttyd service running"
        else
            check_fail "ttyd service not running"
        fi

        # Check nginx service
        if systemctl is-active --quiet nginx 2>/dev/null; then
            check_pass "nginx service running"
        else
            check_fail "nginx service not running"
        fi

        # Check if services are enabled
        if systemctl is-enabled --quiet ttyd 2>/dev/null; then
            check_pass "ttyd enabled at boot"
        else
            check_warn "ttyd not enabled at boot"
        fi

        if systemctl is-enabled --quiet nginx 2>/dev/null; then
            check_pass "nginx enabled at boot"
        else
            check_warn "nginx not enabled at boot"
        fi
    else
        # Non-systemd environment (WSL without systemd)
        if pgrep -x ttyd &>/dev/null; then
            check_pass "ttyd process running"
        else
            check_fail "ttyd process not running"
        fi

        if pgrep -x nginx &>/dev/null; then
            check_pass "nginx process running"
        else
            check_fail "nginx process not running"
        fi
    fi
}

check_ports() {
    section "Port Checks"

    # Check ttyd port (7681)
    if ss -tlnp 2>/dev/null | grep -q ":7681 " || netstat -tlnp 2>/dev/null | grep -q ":7681 "; then
        check_pass "ttyd listening on port 7681"
    else
        check_fail "ttyd not listening on port 7681"
    fi

    # Check nginx HTTPS port (443)
    if ss -tlnp 2>/dev/null | grep -q ":443 " || netstat -tlnp 2>/dev/null | grep -q ":443 "; then
        check_pass "nginx listening on port 443 (HTTPS)"
    else
        check_fail "nginx not listening on port 443"
    fi

    # Check nginx HTTP port (80)
    if ss -tlnp 2>/dev/null | grep -q ":80 " || netstat -tlnp 2>/dev/null | grep -q ":80 "; then
        check_pass "nginx listening on port 80 (HTTP redirect)"
    else
        check_warn "nginx not listening on port 80"
    fi
}

check_ssl() {
    section "SSL Certificate Checks"

    local cert_file="${INSTALL_DIR}/ssl/server.crt"
    local key_file="${INSTALL_DIR}/ssl/server.key"

    # Check certificate exists
    if [[ -f "${cert_file}" ]]; then
        check_pass "SSL certificate exists"

        # Check certificate validity
        local expiry=$(openssl x509 -enddate -noout -in "${cert_file}" 2>/dev/null | cut -d= -f2)
        local expiry_epoch=$(date -d "${expiry}" +%s 2>/dev/null || echo 0)
        local now_epoch=$(date +%s)
        local days_left=$(( (expiry_epoch - now_epoch) / 86400 ))

        if [[ ${days_left} -gt 30 ]]; then
            check_pass "Certificate valid for ${days_left} days"
        elif [[ ${days_left} -gt 0 ]]; then
            check_warn "Certificate expires in ${days_left} days"
        else
            check_fail "Certificate expired or invalid"
        fi
    else
        check_fail "SSL certificate not found"
    fi

    # Check key exists
    if [[ -f "${key_file}" ]]; then
        check_pass "SSL private key exists"

        # Check key permissions
        local key_perms=$(stat -c %a "${key_file}" 2>/dev/null || echo "000")
        if [[ "${key_perms}" == "600" ]]; then
            check_pass "Key permissions correct (600)"
        else
            check_warn "Key permissions are ${key_perms} (should be 600)"
        fi
    else
        check_fail "SSL private key not found"
    fi
}

check_nginx_config() {
    section "Nginx Configuration"

    # Test nginx configuration
    if nginx -t &>/dev/null; then
        check_pass "nginx configuration valid"
    else
        check_fail "nginx configuration invalid"
        nginx -t 2>&1 | head -5
    fi

    # Check htpasswd file
    if [[ -f "${INSTALL_DIR}/config/nginx/.htpasswd" ]]; then
        check_pass "Authentication file exists"
    else
        check_fail "Authentication file missing"
    fi
}

check_directories() {
    section "Directory Structure"

    local dirs=(
        "${INSTALL_DIR}"
        "${INSTALL_DIR}/bin"
        "${INSTALL_DIR}/config"
        "${INSTALL_DIR}/ssl"
        "${INSTALL_DIR}/logs"
    )

    for dir in "${dirs[@]}"; do
        if [[ -d "${dir}" ]]; then
            check_pass "${dir}"
        else
            check_fail "${dir} missing"
        fi
    done
}

check_configs() {
    section "Configuration Files"

    local configs=(
        "${INSTALL_DIR}/config/tmux/tmux.conf:tmux config"
        "${INSTALL_DIR}/config/neovim/init.lua:neovim config"
        "${INSTALL_DIR}/config/zsh/.zshrc:zsh config"
        "${INSTALL_DIR}/bin/terminal-session.sh:session script"
    )

    for entry in "${configs[@]}"; do
        local file="${entry%%:*}"
        local name="${entry##*:}"

        if [[ -f "${file}" ]]; then
            check_pass "${name}"
        else
            check_warn "${name} missing"
        fi
    done
}

check_connectivity() {
    section "Connectivity Tests"

    # Test localhost HTTPS
    if curl -sk --connect-timeout 5 "https://127.0.0.1/health" &>/dev/null; then
        check_pass "HTTPS localhost accessible"
    else
        check_fail "HTTPS localhost not accessible"
    fi

    # Test HTTP redirect
    local redirect=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 "http://127.0.0.1/" 2>/dev/null || echo "000")
    if [[ "${redirect}" == "301" ]]; then
        check_pass "HTTP to HTTPS redirect working"
    else
        check_warn "HTTP redirect returned ${redirect}"
    fi

    # Test WebSocket endpoint (basic check)
    if curl -sk --connect-timeout 5 -H "Upgrade: websocket" -H "Connection: Upgrade" "https://127.0.0.1/" &>/dev/null; then
        check_pass "WebSocket endpoint accessible"
    else
        check_warn "WebSocket endpoint check inconclusive"
    fi
}

check_memory() {
    section "Resource Usage"

    # Get memory usage of key processes
    local ttyd_mem=$(ps -C ttyd -o rss= 2>/dev/null | awk '{sum+=$1} END {print sum/1024}' || echo "0")
    local nginx_mem=$(ps -C nginx -o rss= 2>/dev/null | awk '{sum+=$1} END {print sum/1024}' || echo "0")
    local total_mem=$(echo "${ttyd_mem} + ${nginx_mem}" | bc 2>/dev/null || echo "0")

    if [[ $(echo "${total_mem} < 200" | bc 2>/dev/null || echo 1) -eq 1 ]]; then
        check_pass "Memory usage: ~${total_mem}MB (target: <200MB)"
    else
        check_warn "Memory usage: ~${total_mem}MB (target: <200MB)"
    fi

    # Individual process memory
    if [[ "${ttyd_mem}" != "0" ]]; then
        echo "      ttyd: ~${ttyd_mem}MB"
    fi
    if [[ "${nginx_mem}" != "0" ]]; then
        echo "      nginx: ~${nginx_mem}MB"
    fi
}

check_logs() {
    section "Log Files"

    local logs=(
        "${INSTALL_DIR}/logs/nginx-access.log"
        "${INSTALL_DIR}/logs/nginx-error.log"
        "${INSTALL_DIR}/logs/ttyd.log"
    )

    for log in "${logs[@]}"; do
        if [[ -f "${log}" ]]; then
            local size=$(du -h "${log}" 2>/dev/null | cut -f1)
            check_pass "$(basename ${log}) (${size})"

            # Check for recent errors
            if [[ "$(basename ${log})" == *"error"* ]]; then
                local recent_errors=$(tail -10 "${log}" 2>/dev/null | grep -ci "error\|critical\|fatal" || echo 0)
                if [[ ${recent_errors} -gt 0 ]]; then
                    check_warn "  ${recent_errors} recent error(s) in log"
                fi
            fi
        else
            check_warn "$(basename ${log}) not found"
        fi
    done
}

#===============================================================================
# Summary
#===============================================================================

print_summary() {
    echo ""
    echo "=========================================="
    echo "  Summary"
    echo "=========================================="
    echo ""
    echo -e "  ${GREEN}Passed:${NC}  ${CHECKS_PASSED}"
    echo -e "  ${YELLOW}Warnings:${NC} ${CHECKS_WARNED}"
    echo -e "  ${RED}Failed:${NC}  ${CHECKS_FAILED}"
    echo ""

    if [[ ${CHECKS_FAILED} -eq 0 ]]; then
        echo -e "${GREEN}All critical checks passed!${NC}"
        echo ""
        echo "Access terminal at: https://192.168.178.78/"
        return 0
    else
        echo -e "${RED}Some checks failed. Please review the issues above.${NC}"
        return 1
    fi
}

#===============================================================================
# Main
#===============================================================================

main() {
    print_header

    check_binaries
    check_services
    check_ports
    check_ssl
    check_nginx_config
    check_directories
    check_configs
    check_connectivity
    check_memory
    check_logs

    print_summary
}

main "$@"
