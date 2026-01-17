#!/usr/bin/env bash
# Doom Coding - Tailscale Serve Setup Script
# Configures Tailscale Serve to expose services over Tailscale network
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

# Default ports
CODE_SERVER_PORT="${CODE_SERVER_PORT:-8443}"
CLAUDE_PORT="${CLAUDE_PORT:-7681}"

# ===========================================
# LOGGING
# ===========================================
log_info() {
    echo -e "${BLUE}ℹ${NC}  $*"
}

log_success() {
    echo -e "${GREEN}✅${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC}  $*"
}

log_error() {
    echo -e "${RED}❌${NC} $*" >&2
}

log_step() {
    echo -e "${BROWN}⏳${NC} $*"
}

# ===========================================
# HELPER FUNCTIONS
# ===========================================
check_tailscale() {
    if ! command -v tailscale &>/dev/null; then
        log_error "Tailscale is not installed"
        return 1
    fi

    if ! tailscale status &>/dev/null; then
        log_error "Tailscale is not connected"
        log_info "Run 'tailscale up' to connect"
        return 1
    fi

    return 0
}

get_tailscale_ip() {
    tailscale ip -4 2>/dev/null || echo ""
}

wait_for_service() {
    local port="$1"
    local max_attempts="${2:-30}"
    local attempt=0

    while [[ $attempt -lt $max_attempts ]]; do
        if curl -sf -o /dev/null "http://127.0.0.1:${port}" 2>/dev/null || \
           curl -sf -o /dev/null "https://127.0.0.1:${port}" -k 2>/dev/null; then
            return 0
        fi
        ((attempt++))
        sleep 1
    done

    return 1
}

# ===========================================
# TAILSCALE SERVE FUNCTIONS
# ===========================================
setup_code_server_serve() {
    log_step "Configuring Tailscale Serve for code-server (port ${CODE_SERVER_PORT})..."

    # Check if service is running
    if ! wait_for_service "$CODE_SERVER_PORT" 5; then
        log_warning "code-server not responding on port ${CODE_SERVER_PORT}"
        log_info "Make sure Docker services are running first"
    fi

    # Configure Tailscale Serve
    # Using port 443 for HTTPS access (standard)
    if tailscale serve --bg --https=443 "http://127.0.0.1:${CODE_SERVER_PORT}" 2>/dev/null; then
        log_success "code-server exposed via Tailscale Serve on port 443"
    else
        # Try alternative syntax for older versions
        tailscale serve --bg 443 "http://127.0.0.1:${CODE_SERVER_PORT}" 2>/dev/null || {
            log_error "Failed to configure Tailscale Serve for code-server"
            return 1
        }
        log_success "code-server exposed via Tailscale Serve on port 443"
    fi
}

setup_claude_serve() {
    log_step "Configuring Tailscale Serve for Claude ttyd (port ${CLAUDE_PORT})..."

    # Check if service is running
    if ! wait_for_service "$CLAUDE_PORT" 5; then
        log_warning "Claude ttyd not responding on port ${CLAUDE_PORT}"
        log_info "This is optional - ttyd may not be enabled"
        return 0
    fi

    # Configure Tailscale Serve on a different port
    if tailscale serve --bg --https=7681 "http://127.0.0.1:${CLAUDE_PORT}" 2>/dev/null; then
        log_success "Claude ttyd exposed via Tailscale Serve on port 7681"
    else
        # Try alternative syntax for older versions
        tailscale serve --bg 7681 "http://127.0.0.1:${CLAUDE_PORT}" 2>/dev/null || {
            log_warning "Could not configure Tailscale Serve for Claude ttyd"
            return 0
        }
        log_success "Claude ttyd exposed via Tailscale Serve on port 7681"
    fi
}

show_serve_status() {
    local ts_ip
    ts_ip=$(get_tailscale_ip)

    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  Tailscale Serve Konfiguration${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    if [[ -n "$ts_ip" ]]; then
        echo "Tailscale IP: $ts_ip"
        echo ""
        echo "Services erreichbar unter:"
        echo "  code-server: https://${ts_ip}/"
        echo "  Claude ttyd: https://${ts_ip}:7681/"
        echo ""
        echo "Oder via MagicDNS (falls aktiviert):"
        local hostname
        hostname=$(tailscale status --json | jq -r '.Self.DNSName' 2>/dev/null | sed 's/\.$//')
        if [[ -n "$hostname" && "$hostname" != "null" ]]; then
            echo "  code-server: https://${hostname}/"
            echo "  Claude ttyd: https://${hostname}:7681/"
        fi
    else
        log_warning "Tailscale IP konnte nicht ermittelt werden"
    fi

    echo ""
    echo "Aktuelle Tailscale Serve Konfiguration:"
    tailscale serve status 2>/dev/null || echo "  Keine aktive Konfiguration"
    echo ""
}

reset_serve() {
    log_step "Resetting Tailscale Serve configuration..."

    tailscale serve reset 2>/dev/null || {
        # Fallback: remove individual serves
        tailscale serve off 443 2>/dev/null || true
        tailscale serve off 7681 2>/dev/null || true
    }

    log_success "Tailscale Serve configuration reset"
}

# ===========================================
# MAIN
# ===========================================
print_help() {
    cat << EOF
Usage: $0 [COMMAND] [OPTIONS]

Doom Coding - Tailscale Serve Setup

COMMANDS:
    setup       Configure Tailscale Serve for all services (default)
    status      Show current Tailscale Serve status
    reset       Remove all Tailscale Serve configurations

OPTIONS:
    --code-port=PORT    code-server port (default: 8443)
    --claude-port=PORT  Claude ttyd port (default: 7681)
    --help, -h          Show this help

EXAMPLES:
    $0                  Setup Tailscale Serve with defaults
    $0 setup            Same as above
    $0 status           Show current serve status
    $0 reset            Remove all serve configurations

EOF
}

main() {
    local command="setup"

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            setup|status|reset)
                command="$1"
                shift
                ;;
            --code-port=*)
                CODE_SERVER_PORT="${1#*=}"
                shift
                ;;
            --claude-port=*)
                CLAUDE_PORT="${1#*=}"
                shift
                ;;
            --help|-h)
                print_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                print_help
                exit 1
                ;;
        esac
    done

    # Check prerequisites
    if ! check_tailscale; then
        exit 1
    fi

    # Execute command
    case "$command" in
        setup)
            log_info "Setting up Tailscale Serve..."
            setup_code_server_serve
            setup_claude_serve
            show_serve_status
            ;;
        status)
            show_serve_status
            ;;
        reset)
            reset_serve
            ;;
    esac
}

main "$@"
