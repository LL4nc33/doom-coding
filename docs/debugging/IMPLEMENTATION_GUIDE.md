# Debugging Infrastructure - Implementation Guide

This guide provides ready-to-implement code for improving the doom-coding project's debugging and error handling infrastructure.

---

## 1. Enhanced Install Script with Error Recovery

### File: `/config/repos/doom-coding/scripts/install-enhanced.sh`

Add these functions to the install.sh script:

```bash
#!/usr/bin/env bash
# Enhanced Installation Script with Error Recovery
set -euo pipefail

# ===========================================
# TRAP HANDLER FOR CLEANUP
# ===========================================

INSTALLED_COMPONENTS=()
TEMP_FILES=()

cleanup_on_error() {
    local exit_code=$?

    if [[ $exit_code -ne 0 ]]; then
        log_error "Installation failed with exit code: $exit_code"
        log_warning "Rolling back changes..."

        # Rollback in reverse order
        for ((i=${#INSTALLED_COMPONENTS[@]}-1; i>=0; i--)); do
            rollback_component "${INSTALLED_COMPONENTS[i]}"
        done

        # Clean temp files
        for temp_file in "${TEMP_FILES[@]}"; do
            rm -f "$temp_file" 2>/dev/null || true
        done

        log_info "Rollback complete. System restored to pre-installation state."
        log_info "Check logs for details: $LOG_FILE"
    fi
}

trap cleanup_on_error EXIT ERR INT TERM

# ===========================================
# COMPONENT TRACKING
# ===========================================

track_component() {
    local component="$1"
    INSTALLED_COMPONENTS+=("$component")
    log_info "Installed: $component"
}

rollback_component() {
    local component="$1"
    log_step "Rolling back: $component"

    case "$component" in
        docker_containers)
            docker compose down -v 2>/dev/null || true
            ;;

        docker_service)
            sudo systemctl stop docker 2>/dev/null || true
            sudo systemctl disable docker 2>/dev/null || true
            ;;

        docker_packages)
            if [[ "$PKG_MANAGER" == "apt" ]]; then
                sudo apt-get remove -y docker-ce docker-ce-cli containerd.io \
                    docker-buildx-plugin docker-compose-plugin 2>/dev/null || true
            fi
            ;;

        tailscale)
            sudo tailscale down 2>/dev/null || true
            ;;

        env_file)
            rm -f "$PROJECT_DIR/.env" 2>/dev/null || true
            ;;

        ssh_hardening)
            sudo rm -f /etc/ssh/sshd_config.d/99-doom-hardening.conf
            sudo systemctl reload sshd 2>/dev/null || true
            ;;

        terminal_tools)
            # Reverting terminal setup is complex, just log
            log_info "Terminal tools installed - manual cleanup needed if desired"
            ;;

        *)
            log_warning "Unknown component: $component"
            ;;
    esac
}

# ===========================================
# PRE-FLIGHT CHECKS
# ===========================================

preflight_checks() {
    log_info "Running pre-flight checks..."
    local failed=0

    # Check 1: Sudo access
    log_step "Checking sudo access..."
    if [[ $EUID -ne 0 ]] && ! sudo -n true 2>/dev/null; then
        log_fail "Sudo access required"
        log_info "Please run with sudo or configure passwordless sudo"
        failed=1
    else
        log_pass "Sudo access: OK"
    fi

    # Check 2: Internet connectivity
    log_step "Checking internet connectivity..."
    local test_urls=(
        "https://google.com"
        "https://github.com"
        "https://download.docker.com"
    )

    local connected=false
    for url in "${test_urls[@]}"; do
        if curl -sf --max-time 5 "$url" >/dev/null 2>&1; then
            connected=true
            break
        fi
    done

    if [[ "$connected" == "false" ]]; then
        log_fail "Internet connectivity: FAILED"
        log_info "Cannot reach test URLs. Check your network connection."
        failed=1
    else
        log_pass "Internet connectivity: OK"
    fi

    # Check 3: Disk space
    log_step "Checking disk space..."
    local available_gb
    available_gb=$(df -BG . | awk 'NR==2 {print $4}' | tr -d 'G')

    if [[ $available_gb -lt 10 ]]; then
        log_fail "Disk space: ${available_gb}GB (minimum 10GB required)"
        log_info "Free up disk space: docker system prune -af"
        failed=1
    elif [[ $available_gb -lt 20 ]]; then
        log_warn "Disk space: ${available_gb}GB (recommended 20GB+)"
    else
        log_pass "Disk space: ${available_gb}GB available"
    fi

    # Check 4: Memory
    log_step "Checking system memory..."
    local mem_gb
    mem_gb=$(free -g | awk '/^Mem:/{print $2}')

    if [[ $mem_gb -lt 2 ]]; then
        log_warn "Memory: ${mem_gb}GB (recommended 2GB+)"
        log_info "Installation may be slow with limited memory"
    else
        log_pass "Memory: ${mem_gb}GB available"
    fi

    # Check 5: CPU cores
    log_step "Checking CPU cores..."
    local cores
    cores=$(nproc)

    if [[ $cores -lt 2 ]]; then
        log_warn "CPU cores: $cores (recommended 2+)"
    else
        log_pass "CPU cores: $cores"
    fi

    # Check 6: Required commands
    log_step "Checking required commands..."
    local required_commands=(curl wget git)

    for cmd in "${required_commands[@]}"; do
        if command -v "$cmd" &>/dev/null; then
            log_pass "Command '$cmd': found"
        else
            log_fail "Command '$cmd': not found"
            failed=1
        fi
    done

    # Check 7: Port availability (if not skipping Docker)
    if [[ "$SKIP_DOCKER" != "true" ]]; then
        log_step "Checking port availability..."
        local ports=(8443)

        for port in "${ports[@]}"; do
            if lsof -Pi ":${port}" -sTCP:LISTEN -t >/dev/null 2>&1; then
                local pid
                pid=$(lsof -Pi ":${port}" -sTCP:LISTEN -t | head -1)
                local process
                process=$(ps -p "$pid" -o comm= 2>/dev/null || echo "unknown")

                log_fail "Port $port already in use by $process (PID: $pid)"
                log_info "Fix: Kill process with 'kill $pid' or change port in config"
                failed=1
            else
                log_pass "Port $port: available"
            fi
        done
    fi

    # Check 8: Container environment detection
    log_step "Detecting container environment..."
    local container_type
    container_type=$(detect_container_type)

    if [[ "$container_type" != "bare-metal" ]]; then
        log_info "Running in: $container_type"

        if [[ "$container_type" == "lxc" ]]; then
            # Check LXC-specific requirements
            if ! grep -q "1" /sys/fs/cgroup/nesting 2>/dev/null && [[ "$SKIP_DOCKER" != "true" ]]; then
                log_warn "Docker in LXC requires nesting enabled"
                log_info "Enable on Proxmox host: pct set <CTID> -features nesting=1"

                if [[ "$UNATTENDED" != "true" ]]; then
                    if ! confirm "Continue anyway? (may fail)"; then
                        exit 1
                    fi
                fi
            fi
        fi
    else
        log_pass "Environment: bare-metal"
    fi

    # Summary
    echo ""
    if [[ $failed -eq 1 ]]; then
        log_error "Pre-flight checks failed. Please fix the issues above."
        return 1
    else
        log_success "All pre-flight checks passed!"
        return 0
    fi
}

# ===========================================
# NETWORK OPERATIONS WITH RETRY
# ===========================================

curl_with_retry() {
    local url="$1"
    local output="${2:--}"  # Default to stdout
    local max_attempts="${3:-3}"
    local timeout="${4:-30}"

    local attempt=1
    local delay=2

    while [[ $attempt -le $max_attempts ]]; do
        log_step "Downloading from $url (attempt $attempt/$max_attempts)..."

        if [[ "$output" == "-" ]]; then
            if curl --max-time "$timeout" --retry 2 -fsSL "$url"; then
                return 0
            fi
        else
            if curl --max-time "$timeout" --retry 2 -fsSL "$url" -o "$output"; then
                return 0
            fi
        fi

        if [[ $attempt -lt $max_attempts ]]; then
            log_warning "Download failed, retrying in ${delay}s..."
            sleep "$delay"
            delay=$((delay * 2))
        fi

        attempt=$((attempt + 1))
    done

    log_error "Download failed after $max_attempts attempts: $url"
    return 1
}

git_clone_with_retry() {
    local repo="$1"
    local dest="$2"
    local max_attempts="${3:-3}"

    local attempt=1
    local delay=2

    while [[ $attempt -le $max_attempts ]]; do
        log_step "Cloning repository (attempt $attempt/$max_attempts)..."

        if timeout 120 git clone --depth 1 "$repo" "$dest" 2>&1; then
            return 0
        fi

        # Clean up partial clone
        rm -rf "$dest"

        if [[ $attempt -lt $max_attempts ]]; then
            log_warning "Clone failed, retrying in ${delay}s..."
            sleep "$delay"
            delay=$((delay * 2))
        fi

        attempt=$((attempt + 1))
    done

    log_error "Git clone failed after $max_attempts attempts: $repo"
    return 1
}

# ===========================================
# VALIDATION FUNCTIONS
# ===========================================

validate_api_key() {
    local key="$1"
    local key_type="$2"

    case "$key_type" in
        anthropic)
            if [[ ! "$key" =~ ^sk-ant- ]]; then
                log_error "Invalid Anthropic API key format"
                log_info "Should start with: sk-ant-"
                return 1
            fi

            if [[ ${#key} -lt 50 ]]; then
                log_warning "API key seems short (${#key} chars)"
            fi
            ;;

        tailscale)
            if [[ ! "$key" =~ ^tskey- ]]; then
                log_error "Invalid Tailscale key format"
                log_info "Should start with: tskey-auth- or tskey-client-"
                return 1
            fi

            if [[ ${#key} -lt 30 ]]; then
                log_error "Tailscale key too short (${#key} chars)"
                return 1
            fi
            ;;

        *)
            log_warning "Unknown key type: $key_type"
            ;;
    esac

    return 0
}

validate_password() {
    local password="$1"
    local min_length="${2:-8}"

    if [[ ${#password} -lt $min_length ]]; then
        log_error "Password too short (${#password} chars, need $min_length)"
        return 1
    fi

    # Check for whitespace
    if [[ "$password" =~ [[:space:]] ]]; then
        log_warning "Password contains whitespace"
    fi

    # Check strength
    local strength=0
    [[ "$password" =~ [a-z] ]] && ((strength++))
    [[ "$password" =~ [A-Z] ]] && ((strength++))
    [[ "$password" =~ [0-9] ]] && ((strength++))
    [[ "$password" =~ [^a-zA-Z0-9] ]] && ((strength++))

    if [[ $strength -lt 3 ]]; then
        log_warning "Weak password - use a mix of upper, lower, numbers, and symbols"
    fi

    return 0
}

validate_compose_config() {
    local compose_file="$1"

    log_step "Validating Docker Compose configuration..."

    if [[ ! -f "$compose_file" ]]; then
        log_error "Compose file not found: $compose_file"
        return 1
    fi

    # Validate syntax
    if ! docker compose -f "$compose_file" config >/dev/null 2>&1; then
        log_error "Invalid Docker Compose syntax"
        docker compose -f "$compose_file" config 2>&1 | tail -10
        return 1
    fi

    # Check required environment variables
    local required_vars=(
        CODE_SERVER_PASSWORD
    )

    for var in "${required_vars[@]}"; do
        if [[ -z "${!var:-}" ]] && ! grep -q "^${var}=" .env 2>/dev/null; then
            log_error "Required variable not set: $var"
            return 1
        fi
    done

    # Check secrets files
    if grep -q "secrets:" "$compose_file"; then
        if [[ ! -f "secrets/anthropic_api_key.txt" ]]; then
            log_error "Missing required file: secrets/anthropic_api_key.txt"
            return 1
        fi
    fi

    log_pass "Compose configuration valid"
    return 0
}

# ===========================================
# ENHANCED DOCKER INSTALLATION
# ===========================================

install_docker_enhanced() {
    if [[ "$SKIP_DOCKER" == "true" ]]; then
        log_info "Skipping Docker installation (--skip-docker)"
        return 0
    fi

    log_step "Installing Docker..."

    # Check if already installed
    if command -v docker &>/dev/null && [[ "$FORCE" != "true" ]]; then
        log_success "Docker already installed: $(docker --version)"

        # Verify it's running
        if ! docker info &>/dev/null; then
            log_warning "Docker installed but daemon not running"
            start_docker_daemon || return 1
        fi

        return 0
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would install Docker"
        return 0
    fi

    case "$OS_ID" in
        ubuntu|debian)
            install_docker_debian || return 1
            ;;
        arch)
            install_docker_arch || return 1
            ;;
        *)
            log_error "Docker installation not supported for ${OS_ID}"
            return 1
            ;;
    esac

    track_component "docker_packages"

    # Start and verify daemon
    start_docker_daemon || return 1
    track_component "docker_service"

    # Add user to docker group
    add_user_to_docker_group

    log_success "Docker installed successfully"
}

install_docker_debian() {
    log_step "Installing Docker on Debian/Ubuntu..."

    # Remove old versions
    sudo apt-get remove -y docker docker-engine docker.io containerd runc 2>/dev/null || true

    # Install prerequisites
    sudo apt-get update
    sudo apt-get install -y ca-certificates curl gnupg lsb-release

    # Add GPG key with retry
    sudo install -m 0755 -d /etc/apt/keyrings

    if ! curl_with_retry "https://download.docker.com/linux/${OS_ID}/gpg" \
         /tmp/docker.gpg 3 30; then
        return 1
    fi

    sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg < /tmp/docker.gpg
    sudo chmod a+r /etc/apt/keyrings/docker.gpg
    rm /tmp/docker.gpg

    # Add repository
    echo \
        "deb [arch=${ARCH} signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/${OS_ID} \
        $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
        sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

    # Install Docker
    sudo apt-get update
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io \
        docker-buildx-plugin docker-compose-plugin

    return 0
}

install_docker_arch() {
    log_step "Installing Docker on Arch Linux..."
    sudo pacman -S --noconfirm docker docker-compose
    return 0
}

start_docker_daemon() {
    log_step "Starting Docker daemon..."

    sudo systemctl start docker
    sudo systemctl enable docker

    # Wait for daemon to be ready
    local max_wait=30
    local waited=0

    while [[ $waited -lt $max_wait ]]; do
        if docker info &>/dev/null; then
            log_success "Docker daemon is running"
            return 0
        fi
        sleep 1
        waited=$((waited + 1))
    done

    log_error "Docker daemon failed to start within ${max_wait}s"

    # Show diagnostics
    log_info "Diagnostics:"
    sudo systemctl status docker --no-pager || true
    sudo journalctl -u docker -n 20 --no-pager || true

    return 1
}

add_user_to_docker_group() {
    local target_user="${SUDO_USER:-$USER}"

    if groups "$target_user" | grep -q docker; then
        log_info "User '$target_user' already in docker group"
        return 0
    fi

    log_step "Adding user to docker group..."
    sudo usermod -aG docker "$target_user"

    log_success "User added to docker group"
    log_warning "Note: You may need to log out and back in for group membership to take effect"
}

# ===========================================
# ENHANCED SERVICE STARTUP
# ===========================================

start_services_enhanced() {
    log_step "Starting Docker services..."

    cd "$PROJECT_DIR"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would start Docker services with ${COMPOSE_FILE}"
        return 0
    fi

    # Validate configuration first
    if ! validate_compose_config "$COMPOSE_FILE"; then
        log_error "Compose configuration validation failed"
        return 1
    fi

    log_info "Using compose file: ${COMPOSE_FILE}"

    # Pull images first (with progress)
    log_step "Pulling Docker images..."
    if ! docker compose -f "$COMPOSE_FILE" pull; then
        log_warning "Some images failed to pull, will try building..."
    fi

    # Build custom images
    log_step "Building custom images..."
    if ! docker compose -f "$COMPOSE_FILE" build; then
        log_error "Docker build failed"
        return 1
    fi

    # Start services
    log_step "Starting containers..."
    if ! docker compose -f "$COMPOSE_FILE" up -d; then
        log_error "Failed to start containers"
        docker compose -f "$COMPOSE_FILE" logs --tail=50
        return 1
    fi

    track_component "docker_containers"

    # Wait for services to be healthy
    wait_for_healthy_services || return 1

    log_success "Services started successfully"

    # Show access info
    show_access_info
}

wait_for_healthy_services() {
    log_step "Waiting for services to be healthy..."

    local services=()

    if [[ "${USE_TAILSCALE:-true}" == "true" ]]; then
        services+=("doom-tailscale")
    fi

    services+=("doom-code-server")

    local max_wait=120
    local waited=0

    while [[ $waited -lt $max_wait ]]; do
        local all_healthy=true

        for service in "${services[@]}"; do
            local health
            health=$(docker inspect "$service" --format='{{.State.Health.Status}}' 2>/dev/null || echo "unknown")

            if [[ "$health" != "healthy" ]] && [[ "$health" != "none" ]]; then
                all_healthy=false
                break
            fi
        done

        if [[ "$all_healthy" == "true" ]]; then
            log_success "All services are healthy"
            return 0
        fi

        sleep 5
        waited=$((waited + 5))

        log_info "Waiting for services... (${waited}s/${max_wait}s)"
    done

    log_error "Services did not become healthy within ${max_wait}s"

    # Show logs for debugging
    for service in "${services[@]}"; do
        log_info "Logs for $service:"
        docker logs "$service" --tail 20 2>&1 | sed 's/^/  /'
    done

    return 1
}

show_access_info() {
    echo ""
    log_success "Installation completed successfully!"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    if [[ "${USE_TAILSCALE:-true}" == "true" ]]; then
        log_info "Access via Tailscale:"
        echo "  1. Run: tailscale up"
        echo "  2. Get IP: tailscale status"
        echo "  3. Open: https://<TAILSCALE-IP>:8443"
    else
        local host_ip
        host_ip=$(hostname -I | awk '{print $1}')
        log_info "Access via local network:"
        echo "  Open: https://${host_ip}:8443"
    fi

    echo ""
    log_info "Next steps:"
    echo "  1. Run health check: ./scripts/health-check.sh"
    echo "  2. Check logs: docker compose logs -f"
    echo "  3. View docs: https://github.com/LL4nc33/doom-coding"
    echo ""
}

# ===========================================
# MAIN EXECUTION WITH ENHANCEMENTS
# ===========================================

main_enhanced() {
    parse_arguments "$@"
    print_banner
    setup_logging

    log_info "Starting Doom Coding installation (Enhanced)..."
    log_info "Log file: $LOG_FILE"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_warning "DRY RUN MODE - No changes will be made"
    fi

    # PRE-FLIGHT CHECKS (NEW)
    preflight_checks || exit 1

    echo ""
    log_info "Pre-flight checks passed. Starting installation..."
    echo ""

    # System detection
    detect_os
    detect_arch
    detect_package_manager

    # Prerequisites
    check_root

    # Installation steps with tracking
    install_base_packages
    install_docker_enhanced
    setup_tailscale_choice
    install_tailscale

    # Terminal tools
    if [[ "$SKIP_TERMINAL" != "true" ]]; then
        if [[ -x "$SCRIPT_DIR/setup-terminal.sh" ]]; then
            log_step "Running terminal setup..."
            "$SCRIPT_DIR/setup-terminal.sh"
            track_component "terminal_tools"
        fi
    fi

    # SSH Hardening
    if [[ "$SKIP_HARDENING" != "true" ]]; then
        if [[ -x "$SCRIPT_DIR/setup-host.sh" ]]; then
            log_step "Running host setup..."
            "$SCRIPT_DIR/setup-host.sh"
            track_component "ssh_hardening"
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
    track_component "env_file"

    if confirm "Start Docker services now?"; then
        start_services_enhanced
    fi

    # Health check
    if [[ -x "$SCRIPT_DIR/health-check.sh" ]]; then
        echo ""
        log_step "Running health check..."
        "$SCRIPT_DIR/health-check.sh" || true
    fi

    # Success! Disable error trap
    trap - EXIT ERR INT TERM

    echo ""
    log_success "Installation completed successfully!"
}
```

---

## 2. Diagnostic Collection Script

### File: `/config/repos/doom-coding/scripts/collect-diagnostics.sh`

```bash
#!/usr/bin/env bash
# Doom Coding - Diagnostic Collection Script
set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
readonly OUTPUT_DIR="doom-diagnostics-$(date +%Y%m%d-%H%M%S)"

# Colors
readonly GREEN='\033[38;2;46;82;29m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

log_info() { echo -e "${BLUE}ℹ${NC}  $*"; }
log_success() { echo -e "${GREEN}✅${NC} $*"; }

collect_system_info() {
    log_info "Collecting system information..."

    {
        echo "=== System Information ==="
        echo "Hostname: $(hostname)"
        echo "Kernel: $(uname -r)"
        echo "OS: $(cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d= -f2)"
        echo ""

        echo "=== Hardware ==="
        echo "CPU: $(nproc) cores"
        echo "Memory: $(free -h | awk '/^Mem:/{print $2}')"
        echo "Disk: $(df -h . | awk 'NR==2 {print $4}') free"
        echo ""

        echo "=== Container Detection ==="
        if [[ -f /.dockerenv ]]; then
            echo "Environment: Docker container"
        elif grep -q "lxc" /proc/1/cgroup 2>/dev/null; then
            echo "Environment: LXC container"
        else
            echo "Environment: Bare metal / VM"
        fi

        if [[ -c /dev/net/tun ]]; then
            echo "TUN device: Available"
        else
            echo "TUN device: Not available"
        fi
        echo ""

    } > "$OUTPUT_DIR/system-info.txt"
}

collect_docker_info() {
    log_info "Collecting Docker information..."

    {
        echo "=== Docker Version ==="
        docker version 2>&1 || echo "Docker not available"
        echo ""

        echo "=== Docker Info ==="
        docker info 2>&1 || echo "Docker daemon not running"
        echo ""

        echo "=== Running Containers ==="
        docker ps -a 2>&1 || echo "Cannot list containers"
        echo ""

        echo "=== Docker Networks ==="
        docker network ls 2>&1 || echo "Cannot list networks"
        echo ""

        echo "=== Docker Volumes ==="
        docker volume ls 2>&1 || echo "Cannot list volumes"
        echo ""

    } > "$OUTPUT_DIR/docker-info.txt"
}

collect_container_logs() {
    log_info "Collecting container logs..."

    local containers=(doom-tailscale doom-code-server doom-claude)

    for container in "${containers[@]}"; do
        if docker ps -a --format '{{.Names}}' 2>/dev/null | grep -q "^${container}$"; then
            docker logs "$container" --tail 200 &> "$OUTPUT_DIR/${container}.log"
            docker inspect "$container" > "$OUTPUT_DIR/${container}-inspect.json" 2>&1
        fi
    done
}

collect_configuration() {
    log_info "Collecting configuration (sanitized)..."

    # Sanitize .env file
    if [[ -f "$PROJECT_DIR/.env" ]]; then
        sed -E 's/(PASSWORD|KEY|SECRET|TOKEN)=.*/\1=<REDACTED>/g' \
            "$PROJECT_DIR/.env" > "$OUTPUT_DIR/env-sanitized.txt"
    fi

    # Resolved docker-compose
    cd "$PROJECT_DIR"
    docker compose config > "$OUTPUT_DIR/compose-resolved.yml" 2>&1 || true

    # Copy compose files
    cp docker-compose.yml "$OUTPUT_DIR/" 2>/dev/null || true
    cp docker-compose.lxc.yml "$OUTPUT_DIR/" 2>/dev/null || true
}

collect_network_info() {
    log_info "Collecting network information..."

    {
        echo "=== Network Interfaces ==="
        ip addr show 2>&1 || ifconfig 2>&1 || echo "Cannot get network info"
        echo ""

        echo "=== Routing Table ==="
        ip route show 2>&1 || route -n 2>&1 || echo "Cannot get routes"
        echo ""

        echo "=== Listening Ports ==="
        ss -tlnp 2>&1 || netstat -tlnp 2>&1 || echo "Cannot get listening ports"
        echo ""

        echo "=== DNS Configuration ==="
        cat /etc/resolv.conf 2>&1 || echo "Cannot read resolv.conf"
        echo ""

        echo "=== Tailscale Status ==="
        if command -v tailscale &>/dev/null; then
            tailscale status 2>&1 || echo "Tailscale not running"
        else
            echo "Tailscale not installed"
        fi
        echo ""

    } > "$OUTPUT_DIR/network-info.txt"
}

collect_logs() {
    log_info "Collecting application logs..."

    # Installation log
    if [[ -f /var/log/doom-coding-install.log ]]; then
        cp /var/log/doom-coding-install.log "$OUTPUT_DIR/install.log"
    fi

    # System journal
    if command -v journalctl &>/dev/null; then
        journalctl -u docker -n 100 --no-pager > "$OUTPUT_DIR/docker-journal.log" 2>&1 || true
        journalctl -u tailscaled -n 100 --no-pager > "$OUTPUT_DIR/tailscale-journal.log" 2>&1 || true
    fi
}

run_health_check() {
    log_info "Running health check..."

    if [[ -x "$SCRIPT_DIR/health-check.sh" ]]; then
        "$SCRIPT_DIR/health-check.sh" > "$OUTPUT_DIR/health-check.txt" 2>&1 || true
        "$SCRIPT_DIR/health-check.sh" --json > "$OUTPUT_DIR/health-check.json" 2>&1 || true
    fi
}

create_archive() {
    log_info "Creating archive..."

    tar czf "${OUTPUT_DIR}.tar.gz" "$OUTPUT_DIR"
    rm -rf "$OUTPUT_DIR"

    log_success "Diagnostics collected: ${OUTPUT_DIR}.tar.gz"

    local size
    size=$(du -h "${OUTPUT_DIR}.tar.gz" | cut -f1)

    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Archive: ${OUTPUT_DIR}.tar.gz"
    echo "Size: $size"
    echo ""
    echo "Please attach this file when reporting issues on GitHub:"
    echo "  https://github.com/LL4nc33/doom-coding/issues"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

main() {
    echo "Doom Coding - Diagnostic Collection"
    echo "===================================="
    echo ""

    # Create output directory
    mkdir -p "$OUTPUT_DIR"

    # Collect information
    collect_system_info
    collect_docker_info
    collect_container_logs
    collect_configuration
    collect_network_info
    collect_logs
    run_health_check

    # Create archive
    create_archive
}

main "$@"
```

Make the script executable:
```bash
chmod +x scripts/collect-diagnostics.sh
```

---

## 3. Interactive Troubleshooter

### File: `/config/repos/doom-coding/scripts/troubleshoot.sh`

```bash
#!/usr/bin/env bash
# Doom Coding - Interactive Troubleshooter
set -euo pipefail

# Colors
readonly GREEN='\033[38;2;46;82;29m'
readonly RED='\033[0;31m'
readonly YELLOW='\033[0;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

log_info() { echo -e "${BLUE}ℹ${NC}  $*"; }
log_success() { echo -e "${GREEN}✅${NC} $*"; }
log_warning() { echo -e "${YELLOW}⚠${NC}  $*"; }
log_error() { echo -e "${RED}❌${NC} $*"; }

show_menu() {
    clear
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Doom Coding - Interactive Troubleshooter"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "What issue are you experiencing?"
    echo ""
    echo "  1) Installation fails or hangs"
    echo "  2) Docker containers won't start"
    echo "  3) Cannot access web UI"
    echo "  4) Tailscale not connecting"
    echo "  5) Performance issues (slow/laggy)"
    echo "  6) Permission errors"
    echo "  7) Run diagnostic collection"
    echo "  8) View logs"
    echo "  9) Exit"
    echo ""
    read -rp "Select (1-9): " choice
    echo ""

    case "$choice" in
        1) troubleshoot_installation ;;
        2) troubleshoot_containers ;;
        3) troubleshoot_web_access ;;
        4) troubleshoot_tailscale ;;
        5) troubleshoot_performance ;;
        6) troubleshoot_permissions ;;
        7) collect_diagnostics ;;
        8) view_logs ;;
        9) exit 0 ;;
        *) log_error "Invalid choice"; sleep 2; show_menu ;;
    esac
}

troubleshoot_installation() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Installation Troubleshooter"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # Check installation log
    if [[ -f /var/log/doom-coding-install.log ]]; then
        log_info "Checking installation log..."
        local errors
        errors=$(grep -iE "error|failed|fatal" /var/log/doom-coding-install.log | tail -10)

        if [[ -n "$errors" ]]; then
            log_warning "Recent errors found:"
            echo "$errors"
            echo ""
        else
            log_success "No recent errors in installation log"
        fi
    else
        log_warning "Installation log not found"
    fi

    # Check Docker
    echo ""
    log_info "Checking Docker..."
    if ! command -v docker &>/dev/null; then
        log_error "Docker not installed"
        echo "Fix: Run the installer to install Docker"
    elif ! docker info &>/dev/null; then
        log_error "Docker daemon not running"
        echo "Fix: sudo systemctl start docker"
    else
        log_success "Docker is running"
    fi

    # Check disk space
    echo ""
    log_info "Checking disk space..."
    local available_gb
    available_gb=$(df -BG . | awk 'NR==2 {print $4}' | tr -d 'G')
    if [[ $available_gb -lt 10 ]]; then
        log_error "Low disk space: ${available_gb}GB"
        echo "Fix: docker system prune -af"
    else
        log_success "Disk space OK: ${available_gb}GB"
    fi

    # Check internet
    echo ""
    log_info "Checking internet connectivity..."
    if curl -sf --max-time 5 https://google.com >/dev/null 2>&1; then
        log_success "Internet connection OK"
    else
        log_error "No internet connection"
        echo "Fix: Check network settings"
    fi

    pause_and_return
}

troubleshoot_containers() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Container Troubleshooter"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    local containers=(doom-tailscale doom-code-server doom-claude)

    for container in "${containers[@]}"; do
        echo "Checking $container..."

        if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${container}$"; then
            local status
            status=$(docker inspect "$container" --format '{{.State.Status}}')
            local health
            health=$(docker inspect "$container" --format '{{.State.Health.Status}}' 2>/dev/null || echo "none")

            if [[ "$status" == "running" ]]; then
                if [[ "$health" == "healthy" ]] || [[ "$health" == "none" ]]; then
                    log_success "$container: running and healthy"
                else
                    log_warning "$container: running but $health"
                    echo "  View logs: docker logs $container"
                fi
            else
                log_error "$container: $status"
                echo "  View logs: docker logs $container"
            fi
        else
            log_error "$container: not found"
            echo "  Start: docker compose up -d"
        fi
        echo ""
    done

    # Check common issues
    log_info "Checking for common issues..."

    # Port conflicts
    if lsof -Pi :8443 -sTCP:LISTEN -t >/dev/null 2>&1; then
        local pid
        pid=$(lsof -Pi :8443 -sTCP:LISTEN -t | head -1)
        local process
        process=$(ps -p "$pid" -o comm= 2>/dev/null || echo "unknown")

        if [[ "$process" != "docker-proxy" ]]; then
            log_warning "Port 8443 in use by: $process (PID: $pid)"
            echo "  Fix: kill $pid"
            echo ""
        fi
    fi

    # Volume permissions
    if [[ -d ./workspace ]]; then
        if [[ ! -w ./workspace ]]; then
            log_warning "Workspace not writable"
            echo "  Fix: sudo chown -R \$(id -u):\$(id -g) ./workspace"
            echo ""
        fi
    fi

    pause_and_return
}

troubleshoot_web_access() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Web Access Troubleshooter"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # Check code-server container
    log_info "Checking code-server container..."
    if docker ps --format '{{.Names}}' | grep -q "doom-code-server"; then
        log_success "code-server container is running"

        # Check port
        if docker port doom-code-server 2>/dev/null | grep -q "8443"; then
            log_success "Port 8443 is exposed"
        else
            log_warning "Port 8443 not exposed - check docker-compose config"
        fi
    else
        log_error "code-server container not running"
        echo "Fix: docker compose up -d"
        pause_and_return
        return
    fi

    # Determine access URL
    echo ""
    log_info "Determining access URL..."

    local access_url=""

    # Check if using Tailscale
    if docker ps --format '{{.Names}}' | grep -q "doom-tailscale"; then
        if command -v tailscale &>/dev/null; then
            local ts_ip
            ts_ip=$(tailscale ip -4 2>/dev/null)

            if [[ -n "$ts_ip" ]]; then
                access_url="https://${ts_ip}:8443"
                log_success "Tailscale IP: $ts_ip"
            else
                log_warning "Tailscale not connected"
                echo "Fix: tailscale up"
            fi
        fi
    else
        # Local network access
        local host_ip
        host_ip=$(hostname -I | awk '{print $1}')
        access_url="https://${host_ip}:8443"
        log_info "Local network access"
    fi

    if [[ -n "$access_url" ]]; then
        echo ""
        echo "Access URL: $access_url"
        echo ""

        # Test connectivity
        log_info "Testing connectivity..."
        if curl -k -sf --max-time 5 "$access_url/healthz" >/dev/null 2>&1; then
            log_success "code-server is responding"
        else
            log_warning "Cannot reach code-server"
            echo "  Check firewall rules"
            echo "  View logs: docker logs doom-code-server"
        fi
    fi

    pause_and_return
}

troubleshoot_tailscale() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Tailscale Troubleshooter"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # Check TUN device
    log_info "Checking TUN device..."
    if [[ -c /dev/net/tun ]]; then
        log_success "TUN device available"
    else
        log_error "TUN device not available"
        echo "Tailscale requires /dev/net/tun"
        echo ""

        # Check environment
        if grep -q "lxc" /proc/1/cgroup 2>/dev/null; then
            echo "Running in LXC container."
            echo ""
            echo "To enable TUN in Proxmox LXC:"
            echo "  1. On Proxmox host, edit: /etc/pve/lxc/<CTID>.conf"
            echo "  2. Add lines:"
            echo "     lxc.cgroup2.devices.allow: c 10:200 rwm"
            echo "     lxc.mount.entry: /dev/net/tun dev/net/tun none bind,create=file"
            echo "  3. Restart: pct restart <CTID>"
            echo ""
        fi

        pause_and_return
        return
    fi

    # Check Tailscale installation
    echo ""
    log_info "Checking Tailscale installation..."
    if command -v tailscale &>/dev/null; then
        log_success "Tailscale installed"

        # Check status
        local status
        status=$(tailscale status --json 2>/dev/null || echo "{}")
        local backend_state
        backend_state=$(echo "$status" | jq -r '.BackendState // "Unknown"' 2>/dev/null || echo "Unknown")

        if [[ "$backend_state" == "Running" ]]; then
            log_success "Tailscale connected"

            local ip
            ip=$(tailscale ip -4 2>/dev/null)
            echo "  IP: $ip"
        else
            log_warning "Tailscale not connected: $backend_state"

            if [[ "$backend_state" == "NeedsLogin" ]]; then
                echo "  Fix: tailscale up --authkey=<YOUR_KEY>"
            elif [[ "$backend_state" == "Stopped" ]]; then
                echo "  Fix: sudo systemctl start tailscaled"
            fi
        fi
    else
        log_warning "Tailscale not installed on host"
        echo "  Check container: docker exec doom-tailscale tailscale status"
    fi

    pause_and_return
}

troubleshoot_performance() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Performance Troubleshooter"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # System resources
    log_info "Checking system resources..."
    echo ""

    echo "CPU:"
    top -bn1 | head -3 | tail -1

    echo ""
    echo "Memory:"
    free -h

    echo ""
    echo "Disk:"
    df -h .

    echo ""
    echo ""

    # Container resources
    log_info "Container resource usage:"
    echo ""
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}" 2>/dev/null || echo "Cannot get stats"

    echo ""
    echo ""
    log_info "Recommendations:"
    echo "  - Close unused applications"
    echo "  - Prune Docker: docker system prune -af"
    echo "  - Check for disk I/O: iostat -x 1"
    echo "  - Monitor network: iftop"

    pause_and_return
}

troubleshoot_permissions() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Permission Troubleshooter"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # Check Docker group membership
    log_info "Checking Docker permissions..."
    if groups | grep -q docker; then
        log_success "User is in docker group"
    else
        log_warning "User not in docker group"
        echo "  Fix: sudo usermod -aG docker \$USER"
        echo "  Then: Log out and back in"
        echo ""
    fi

    # Check workspace permissions
    echo ""
    log_info "Checking workspace permissions..."
    if [[ -d ./workspace ]]; then
        local owner
        owner=$(stat -c '%U' ./workspace 2>/dev/null || stat -f '%Su' ./workspace 2>/dev/null)

        if [[ "$owner" == "$USER" ]]; then
            log_success "Workspace owned by current user"
        else
            log_warning "Workspace owned by: $owner"
            echo "  Fix: sudo chown -R \$USER:\$USER ./workspace"
            echo ""
        fi

        if [[ -w ./workspace ]]; then
            log_success "Workspace is writable"
        else
            log_error "Workspace not writable"
            echo "  Fix: chmod -R u+w ./workspace"
            echo ""
        fi
    fi

    # Check secrets permissions
    echo ""
    log_info "Checking secrets permissions..."
    if [[ -d ./secrets ]]; then
        local perms
        perms=$(stat -c '%a' ./secrets 2>/dev/null || stat -f '%Lp' ./secrets 2>/dev/null)

        if [[ "$perms" == "700" ]] || [[ "$perms" == "600" ]]; then
            log_success "Secrets directory has correct permissions"
        else
            log_warning "Secrets directory permissions: $perms"
            echo "  Fix: chmod 700 ./secrets"
            echo ""
        fi
    fi

    pause_and_return
}

collect_diagnostics() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Diagnostic Collection"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    if [[ -x ./scripts/collect-diagnostics.sh ]]; then
        log_info "Running diagnostic collection..."
        ./scripts/collect-diagnostics.sh
    else
        log_error "Diagnostic script not found"
        echo "  Ensure you're in the doom-coding directory"
    fi

    pause_and_return
}

view_logs() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  View Logs"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Which logs would you like to view?"
    echo ""
    echo "  1) Installation log"
    echo "  2) Docker container logs"
    echo "  3) System journal (Docker)"
    echo "  4) System journal (Tailscale)"
    echo "  5) Back to main menu"
    echo ""
    read -rp "Select (1-5): " log_choice
    echo ""

    case "$log_choice" in
        1)
            if [[ -f /var/log/doom-coding-install.log ]]; then
                less +G /var/log/doom-coding-install.log
            else
                log_error "Installation log not found"
            fi
            ;;
        2)
            echo "Select container:"
            echo "  1) doom-tailscale"
            echo "  2) doom-code-server"
            echo "  3) doom-claude"
            read -rp "Select: " container_choice

            case "$container_choice" in
                1) docker logs doom-tailscale --tail 100 -f ;;
                2) docker logs doom-code-server --tail 100 -f ;;
                3) docker logs doom-claude --tail 100 -f ;;
            esac
            ;;
        3)
            if command -v journalctl &>/dev/null; then
                journalctl -u docker -f
            else
                log_error "journalctl not available"
            fi
            ;;
        4)
            if command -v journalctl &>/dev/null; then
                journalctl -u tailscaled -f
            else
                log_error "journalctl not available"
            fi
            ;;
    esac

    pause_and_return
}

pause_and_return() {
    echo ""
    read -rp "Press Enter to return to main menu..."
    show_menu
}

main() {
    show_menu
}

main "$@"
```

Make executable:
```bash
chmod +x scripts/troubleshoot.sh
```

---

## Summary

These implementation files provide:

1. **Enhanced install.sh** with:
   - Automatic rollback on failure
   - Pre-flight validation
   - Network retry logic
   - Input validation
   - Better error messages

2. **Diagnostic collection** that gathers:
   - System information
   - Docker state
   - Container logs
   - Network configuration
   - Health check results

3. **Interactive troubleshooter** for:
   - Guided problem diagnosis
   - Common issue detection
   - Quick fix suggestions
   - Log viewing

To integrate these improvements into the main codebase, you can either:

1. Replace the existing `install.sh` with the enhanced version
2. Create a new `install-enhanced.sh` as an alternative installer
3. Gradually merge the improvements into existing scripts

The modular design allows you to cherry-pick specific improvements based on priority.
