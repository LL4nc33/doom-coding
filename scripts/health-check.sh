#!/usr/bin/env bash
# Doom Coding - Health Check Script
# Verifies all services are running correctly
set -euo pipefail

# ===========================================
# COLORS
# ===========================================
readonly GREEN='\033[38;2;46;82;29m'
readonly BROWN='\033[38;2;124;94;70m'
readonly RED='\033[0;31m'
readonly YELLOW='\033[0;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# ===========================================
# CONFIGURATION
# ===========================================
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Counters
PASSED=0
FAILED=0
WARNINGS=0

# Output format
OUTPUT_FORMAT="human"  # human or json
SHOW_QR=false

# ===========================================
# LOGGING
# ===========================================
log_pass() {
    ((PASSED++))
    if [[ "$OUTPUT_FORMAT" == "human" ]]; then
        echo -e "${GREEN}✅${NC} $*"
    fi
}

log_fail() {
    ((FAILED++))
    if [[ "$OUTPUT_FORMAT" == "human" ]]; then
        echo -e "${RED}❌${NC} $*"
    fi
}

log_warn() {
    ((WARNINGS++))
    if [[ "$OUTPUT_FORMAT" == "human" ]]; then
        echo -e "${YELLOW}⚠${NC}  $*"
    fi
}

log_info() {
    if [[ "$OUTPUT_FORMAT" == "human" ]]; then
        echo -e "${BLUE}ℹ${NC}  $*"
    fi
}

# ===========================================
# QR CODE FUNCTIONS
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
        echo ""
        echo "    QR: Install 'qrencode' to display QR code"
        echo "    URL: $url"
        echo ""
    fi
}

# Detect if native userspace mode is active
is_native_userspace_mode() {
    systemctl is-active tailscaled-userspace &>/dev/null
}

# Get access URL based on current configuration
get_access_url() {
    local ip=""
    local port="8443"
    local protocol="https"

    # Check for native userspace mode (uses port 443 via Tailscale Serve)
    if is_native_userspace_mode; then
        ip=$(tailscale ip -4 2>/dev/null || echo "")
        if [[ -n "$ip" ]]; then
            # Native userspace uses Tailscale Serve on port 443
            echo "${protocol}://${ip}/"
            return
        fi
    fi

    # Try to get Tailscale IP first
    if command -v tailscale &>/dev/null; then
        ip=$(tailscale ip -4 2>/dev/null || echo "")
    fi

    # Check for container Tailscale
    if [[ -z "$ip" ]] && docker ps --format '{{.Names}}' 2>/dev/null | grep -q "doom-tailscale"; then
        ip=$(docker exec doom-tailscale tailscale ip -4 2>/dev/null || echo "")
    fi

    # Fallback to local IP
    if [[ -z "$ip" ]]; then
        ip=$(hostname -I 2>/dev/null | awk '{print $1}')
    fi

    if [[ -n "$ip" ]]; then
        echo "${protocol}://${ip}:${port}"
    else
        echo ""
    fi
}

# Show access QR code
show_access_qr() {
    local url
    url=$(get_access_url)

    if [[ -n "$url" ]]; then
        echo ""
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}  Access your code-server on any device:${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        generate_qr "$url" "Scan to open code-server"
        echo "    Desktop: $url"
        echo ""
    else
        echo ""
        echo -e "${YELLOW}Could not determine access URL${NC}"
        echo "Run 'tailscale ip' or check your local network IP"
        echo ""
    fi
}

# ===========================================
# CHECKS
# ===========================================
check_docker() {
    if command -v docker &>/dev/null; then
        if docker info &>/dev/null; then
            local version
            version=$(docker --version | awk '{print $3}' | tr -d ',')
            log_pass "Docker: Running (v$version)"
            return 0
        else
            log_fail "Docker: Not running or no permissions"
            return 1
        fi
    else
        log_fail "Docker: Not installed"
        return 1
    fi
}

check_docker_compose() {
    if docker compose version &>/dev/null; then
        local version
        version=$(docker compose version --short 2>/dev/null || echo "unknown")
        log_pass "Docker Compose: Available (v$version)"
        return 0
    else
        log_fail "Docker Compose: Not available"
        return 1
    fi
}

check_containers() {
    cd "$PROJECT_DIR" 2>/dev/null || return 1

    if [[ ! -f "docker-compose.yml" ]]; then
        log_warn "docker-compose.yml not found"
        return 0
    fi

    local containers
    containers=$(docker compose ps --format json 2>/dev/null || echo "")

    if [[ -z "$containers" ]]; then
        log_warn "No containers running"
        return 0
    fi

    # Check each container
    local running=0
    local total=0

    while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        ((total++))

        local name state health
        name=$(echo "$line" | jq -r '.Name // .name // "unknown"' 2>/dev/null || echo "unknown")
        state=$(echo "$line" | jq -r '.State // .state // "unknown"' 2>/dev/null || echo "unknown")
        health=$(echo "$line" | jq -r '.Health // .health // "N/A"' 2>/dev/null || echo "N/A")

        if [[ "$state" == "running" ]]; then
            ((running++))
            if [[ "$health" == "healthy" || "$health" == "N/A" ]]; then
                log_pass "Container $name: Running ($health)"
            else
                log_warn "Container $name: Running but $health"
            fi
        else
            log_fail "Container $name: $state"
        fi
    done <<< "$containers"

    if [[ $running -eq $total && $total -gt 0 ]]; then
        log_pass "All containers running ($running/$total)"
    elif [[ $total -eq 0 ]]; then
        log_warn "No containers defined"
    else
        log_warn "Some containers not running ($running/$total)"
    fi
}

check_tailscale() {
    # First check for native userspace mode (systemd service)
    if systemctl is-active tailscaled-userspace &>/dev/null; then
        local ip
        ip=$(tailscale ip -4 2>/dev/null || echo "N/A")
        log_pass "Tailscale (native userspace): Running ($ip)"
        return 0
    fi

    if command -v tailscale &>/dev/null; then
        local status
        status=$(tailscale status --json 2>/dev/null || echo "{}")

        local backend_state
        backend_state=$(echo "$status" | jq -r '.BackendState // "Unknown"' 2>/dev/null || echo "Unknown")

        if [[ "$backend_state" == "Running" ]]; then
            local ip
            ip=$(echo "$status" | jq -r '.Self.TailscaleIPs[0] // "N/A"' 2>/dev/null || echo "N/A")
            log_pass "Tailscale: Connected ($ip)"
            return 0
        else
            log_fail "Tailscale: Not connected ($backend_state)"
            return 1
        fi
    else
        # Check if running in container
        if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "tailscale"; then
            local container_status
            container_status=$(docker inspect doom-tailscale --format '{{.State.Health.Status}}' 2>/dev/null || echo "unknown")
            if [[ "$container_status" == "healthy" ]]; then
                log_pass "Tailscale (container): Healthy"
                return 0
            else
                log_warn "Tailscale (container): $container_status"
                return 0
            fi
        fi
        log_warn "Tailscale: Not installed on host"
        return 0
    fi
}

check_tailscale_serve() {
    # Only check if native userspace mode is active
    if ! is_native_userspace_mode; then
        return 0
    fi

    # Check if tailscale serve is configured
    local serve_status
    serve_status=$(tailscale serve status 2>/dev/null || echo "")

    if [[ -n "$serve_status" ]] && [[ "$serve_status" != *"No serve config"* ]]; then
        log_pass "Tailscale Serve: Configured"
        return 0
    else
        log_warn "Tailscale Serve: Not configured"
        log_info "  Run: ./scripts/setup-tailscale-serve.sh setup"
        return 0
    fi
}

check_code_server() {
    # Check if container is running
    if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "code-server"; then
        local health
        health=$(docker inspect doom-code-server --format '{{.State.Health.Status}}' 2>/dev/null || echo "unknown")

        if [[ "$health" == "healthy" ]]; then
            log_pass "code-server: Healthy"
            return 0
        else
            log_warn "code-server: Container running but $health"
            return 0
        fi
    else
        log_warn "code-server: Container not running"
        return 0
    fi
}

check_claude_code() {
    # Check in container first
    if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "doom-claude"; then
        local version
        version=$(docker exec doom-claude claude --version 2>/dev/null || echo "")
        if [[ -n "$version" ]]; then
            log_pass "Claude Code (container): $version"
            return 0
        else
            log_warn "Claude Code: Container running but claude command failed"
            return 0
        fi
    fi

    # Check on host
    if command -v claude &>/dev/null; then
        local version
        version=$(claude --version 2>/dev/null || echo "unknown")
        log_pass "Claude Code (host): $version"
        return 0
    fi

    log_warn "Claude Code: Not found"
    return 0
}

check_ssh_hardening() {
    local hardening_file="/etc/ssh/sshd_config.d/99-doom-hardening.conf"

    if [[ -f "$hardening_file" ]]; then
        # Check key settings
        local root_login pass_auth

        root_login=$(grep -i "^PermitRootLogin" "$hardening_file" 2>/dev/null | awk '{print $2}' || echo "")
        pass_auth=$(grep -i "^PasswordAuthentication" "$hardening_file" 2>/dev/null | awk '{print $2}' || echo "")

        if [[ "$root_login" == "no" && "$pass_auth" == "no" ]]; then
            log_pass "SSH Hardening: Active (root login disabled, key-only auth)"
        else
            log_warn "SSH Hardening: Partial configuration"
        fi
    else
        log_warn "SSH Hardening: Config not found"
    fi
}

check_terminal_tools() {
    local tools_found=0
    local tools_missing=0

    # Check zsh
    if command -v zsh &>/dev/null; then
        ((tools_found++))
    else
        ((tools_missing++))
    fi

    # Check tmux
    if command -v tmux &>/dev/null; then
        ((tools_found++))
    else
        ((tools_missing++))
    fi

    # Check Oh My Zsh
    if [[ -d "$HOME/.oh-my-zsh" ]]; then
        ((tools_found++))
    else
        ((tools_missing++))
    fi

    # Check NVM
    if [[ -d "${NVM_DIR:-$HOME/.nvm}" ]]; then
        ((tools_found++))
    else
        ((tools_missing++))
    fi

    # Check pyenv
    if [[ -d "${PYENV_ROOT:-$HOME/.pyenv}" ]]; then
        ((tools_found++))
    else
        ((tools_missing++))
    fi

    if [[ $tools_missing -eq 0 ]]; then
        log_pass "Terminal Tools: All installed ($tools_found/5)"
    elif [[ $tools_found -gt 0 ]]; then
        log_warn "Terminal Tools: Partial ($tools_found/5)"
    else
        log_warn "Terminal Tools: Not configured"
    fi
}

check_secrets() {
    # Check SOPS
    if command -v sops &>/dev/null; then
        log_pass "SOPS: Installed"
    else
        log_warn "SOPS: Not installed"
    fi

    # Check age
    if command -v age &>/dev/null; then
        log_pass "age: Installed"
    else
        log_warn "age: Not installed"
    fi

    # Check key file
    local key_file="$HOME/.config/sops/age/keys.txt"
    if [[ -f "$key_file" ]]; then
        log_pass "Encryption key: Present"
    else
        log_warn "Encryption key: Not found"
    fi
}

check_disk_space() {
    local available
    available=$(df -BG "$PROJECT_DIR" 2>/dev/null | awk 'NR==2 {print $4}' | tr -d 'G')

    if [[ -n "$available" ]]; then
        if [[ "$available" -gt 10 ]]; then
            log_pass "Disk Space: ${available}GB available"
        elif [[ "$available" -gt 5 ]]; then
            log_warn "Disk Space: ${available}GB available (low)"
        else
            log_fail "Disk Space: ${available}GB available (critical)"
        fi
    fi
}

# ===========================================
# OUTPUT
# ===========================================
print_summary_human() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "Summary: ${GREEN}$PASSED passed${NC}, ${RED}$FAILED failed${NC}, ${YELLOW}$WARNINGS warnings${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    if [[ $FAILED -eq 0 ]]; then
        echo ""
        echo -e "${GREEN}🎉 All systems operational!${NC}"

        # Show access QR code if requested or by default on success
        if [[ "$SHOW_QR" == "true" ]]; then
            show_access_qr
        else
            # Show compact access info
            local url
            url=$(get_access_url)
            if [[ -n "$url" ]]; then
                echo ""
                echo -e "${BLUE}Access:${NC} $url"
                echo -e "${BLUE}Tip:${NC} Run with --qr to show QR code for mobile access"
            fi
        fi
    else
        echo ""
        echo -e "${RED}⚠ Some checks failed. Review the output above.${NC}"
    fi
}

print_summary_json() {
    cat << EOF
{
    "passed": $PASSED,
    "failed": $FAILED,
    "warnings": $WARNINGS,
    "healthy": $([ $FAILED -eq 0 ] && echo "true" || echo "false")
}
EOF
}

# ===========================================
# MAIN
# ===========================================
main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --json)
                OUTPUT_FORMAT="json"
                shift
                ;;
            --qr)
                SHOW_QR=true
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [--json] [--qr]"
                echo ""
                echo "Options:"
                echo "  --json    Output in JSON format"
                echo "  --qr      Show access QR code after health check"
                echo "  --help    Show this help"
                exit 0
                ;;
            *)
                shift
                ;;
        esac
    done

    if [[ "$OUTPUT_FORMAT" == "human" ]]; then
        echo -e "${GREEN}🏥 Doom Coding Health Check${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
    fi

    # Run all checks
    check_docker
    check_docker_compose
    check_containers
    check_tailscale
    check_tailscale_serve
    check_code_server
    check_claude_code
    check_ssh_hardening
    check_terminal_tools
    check_secrets
    check_disk_space

    # Print summary
    if [[ "$OUTPUT_FORMAT" == "human" ]]; then
        print_summary_human
    else
        print_summary_json
    fi

    # Exit code based on failures
    [[ $FAILED -eq 0 ]]
}

main "$@"
