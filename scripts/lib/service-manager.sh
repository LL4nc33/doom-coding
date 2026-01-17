#!/usr/bin/env bash
# Service Management Library for Doom Coding
# Provides functions for service detection, conflict resolution, and lifecycle management
#
# Usage: source this file from install.sh or other scripts

# Ensure we have colors defined
if [[ -z "${NC:-}" ]]; then
    readonly SM_GREEN='\033[38;2;46;82;29m'
    readonly SM_BROWN='\033[38;2;124;94;70m'
    readonly SM_RED='\033[0;31m'
    readonly SM_YELLOW='\033[0;33m'
    readonly SM_BLUE='\033[0;34m'
    readonly SM_GRAY='\033[0;90m'
    readonly SM_NC='\033[0m'
else
    SM_GREEN="${GREEN}"
    SM_BROWN="${BROWN}"
    SM_RED="${RED}"
    SM_YELLOW="${YELLOW}"
    SM_BLUE="${BLUE}"
    SM_NC="${NC}"
    SM_GRAY='\033[0;90m'
fi

# ===========================================
# SERVICE DETECTION
# ===========================================

# Detect all doom-coding related services
# Returns JSON-formatted list of services
detect_doom_services() {
    local services=()

    # Check Docker containers
    if command -v docker &>/dev/null && docker info &>/dev/null; then
        while IFS= read -r line; do
            [[ -z "$line" ]] && continue
            local name state
            name=$(echo "$line" | jq -r '.Names // empty')
            state=$(echo "$line" | jq -r '.State // "unknown"')

            # Check if doom-managed
            if [[ "$name" =~ ^doom- ]] || echo "$line" | grep -q "com.doom-coding"; then
                services+=("{\"name\":\"$name\",\"type\":\"doom\",\"state\":\"$state\",\"container\":true}")
            fi
        done < <(docker ps -a --format '{{json .}}' 2>/dev/null)
    fi

    # Output as JSON array
    printf '['
    local first=true
    for svc in "${services[@]}"; do
        if [[ "$first" == "true" ]]; then
            first=false
        else
            printf ','
        fi
        printf '%s' "$svc"
    done
    printf ']'
}

# Check if a port is in use
# Args: $1 = port number
# Returns: 0 if in use, 1 if free
is_port_in_use() {
    local port="$1"

    # Try netcat first (most reliable)
    if command -v nc &>/dev/null; then
        nc -z localhost "$port" 2>/dev/null
        return $?
    fi

    # Try ss
    if command -v ss &>/dev/null; then
        ss -tln 2>/dev/null | grep -q ":${port} " && return 0
    fi

    # Try lsof
    if command -v lsof &>/dev/null; then
        lsof -i ":${port}" &>/dev/null && return 0
    fi

    # Try netstat as last resort
    if command -v netstat &>/dev/null; then
        netstat -tln 2>/dev/null | grep -q ":${port} " && return 0
    fi

    return 1
}

# Get information about what's using a port
# Args: $1 = port number
# Returns: JSON object with process info
get_port_info() {
    local port="$1"
    local pid=""
    local process_name=""
    local container_name=""

    # Check if it's a Docker container
    if command -v docker &>/dev/null; then
        container_name=$(docker ps --format '{{.Names}}' --filter "publish=${port}" 2>/dev/null | head -1)
    fi

    # Get PID using lsof
    if command -v lsof &>/dev/null; then
        pid=$(lsof -ti ":${port}" 2>/dev/null | head -1)
        if [[ -n "$pid" ]]; then
            process_name=$(ps -p "$pid" -o comm= 2>/dev/null)
        fi
    fi

    # Get PID using ss as fallback
    if [[ -z "$pid" ]] && command -v ss &>/dev/null; then
        local ss_output
        ss_output=$(ss -tlnp "sport = :${port}" 2>/dev/null)
        if [[ -n "$ss_output" ]]; then
            pid=$(echo "$ss_output" | grep -oP 'pid=\K\d+' | head -1)
            if [[ -n "$pid" ]]; then
                process_name=$(ps -p "$pid" -o comm= 2>/dev/null)
            fi
        fi
    fi

    printf '{"port":%d,"pid":%s,"process":"%s","container":"%s"}' \
        "$port" \
        "${pid:-null}" \
        "${process_name:-unknown}" \
        "${container_name:-}"
}

# Check for port conflicts with doom-coding default ports
# Returns: JSON array of conflicts
check_port_conflicts() {
    local conflicts=()
    local doom_ports=(8443 7681)

    for port in "${doom_ports[@]}"; do
        if is_port_in_use "$port"; then
            local info
            info=$(get_port_info "$port")
            conflicts+=("$info")
        fi
    done

    # Output as JSON array
    printf '['
    local first=true
    for conflict in "${conflicts[@]}"; do
        if [[ "$first" == "true" ]]; then
            first=false
        else
            printf ','
        fi
        printf '%s' "$conflict"
    done
    printf ']'
}

# Find an available port starting from a preferred port
# Args: $1 = preferred port, $2 = max attempts (default 100)
# Returns: available port number or 0 if none found
find_available_port() {
    local preferred="$1"
    local max_attempts="${2:-100}"
    local port="$preferred"
    local attempts=0

    while [[ $attempts -lt $max_attempts ]]; do
        if ! is_port_in_use "$port"; then
            echo "$port"
            return 0
        fi
        ((port++))
        ((attempts++))
    done

    echo "0"
    return 1
}

# ===========================================
# CONFLICT RESOLUTION
# ===========================================

# Display port conflict and offer resolution
# Args: $1 = port, $2 = service name requesting port
handle_port_conflict() {
    local port="$1"
    local service="$2"
    local info
    info=$(get_port_info "$port")

    local container
    container=$(echo "$info" | jq -r '.container // empty')
    local process
    process=$(echo "$info" | jq -r '.process // "unknown"')
    local pid
    pid=$(echo "$info" | jq -r '.pid // empty')

    echo ""
    echo -e "${SM_YELLOW}Port Conflict Detected${SM_NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "Port ${SM_BROWN}$port${SM_NC} is required by ${SM_GREEN}$service${SM_NC}"
    echo ""

    if [[ -n "$container" ]]; then
        echo -e "Currently used by container: ${SM_BLUE}$container${SM_NC}"

        # Check if it's our own container
        if [[ "$container" =~ ^doom- ]]; then
            echo -e "  ${SM_GRAY}(Previous doom-coding installation)${SM_NC}"
            return 0  # Will be handled by upgrade
        fi
    else
        echo -e "Currently used by process: ${SM_BLUE}$process${SM_NC} (PID: $pid)"
    fi

    echo ""
    local alt_port
    alt_port=$(find_available_port $((port + 1)))

    echo "Options:"
    echo "  1) Use alternative port: $alt_port"
    if [[ -n "$container" ]]; then
        echo "  2) Stop the conflicting container"
    elif [[ -n "$pid" ]]; then
        echo "  2) Stop the conflicting process (requires manual action)"
    fi
    echo "  3) Skip this service"
    echo "  4) Abort installation"
    echo ""

    # Return suggested alternative port
    echo "$alt_port"
}

# ===========================================
# SERVICE LIFECYCLE
# ===========================================

# Stop doom-coding services gracefully
# Args: $1 = timeout in seconds (default 30)
stop_doom_services() {
    local timeout="${1:-30}"
    local compose_file="${2:-docker-compose.yml}"

    echo -e "${SM_BLUE}Stopping doom-coding services...${SM_NC}"

    # Use docker-compose if available
    if [[ -f "$compose_file" ]]; then
        docker compose -f "$compose_file" down --timeout "$timeout" 2>/dev/null || true
    fi

    # Stop any remaining doom containers
    local containers
    containers=$(docker ps -q --filter "name=doom-" 2>/dev/null)
    if [[ -n "$containers" ]]; then
        echo -e "${SM_GRAY}Stopping remaining containers...${SM_NC}"
        echo "$containers" | xargs -r docker stop -t "$timeout" 2>/dev/null || true
    fi

    echo -e "${SM_GREEN}Services stopped${SM_NC}"
}

# Start doom-coding services
# Args: $1 = compose file
start_doom_services() {
    local compose_file="${1:-docker-compose.yml}"

    if [[ ! -f "$compose_file" ]]; then
        echo -e "${SM_RED}Compose file not found: $compose_file${SM_NC}" >&2
        return 1
    fi

    echo -e "${SM_BLUE}Starting doom-coding services...${SM_NC}"

    # Pull images with progress filtering
    echo -e "${SM_GRAY}Pulling images (this may take a while on first run)...${SM_NC}"
    docker compose -f "$compose_file" pull 2>&1 | filter_docker_output

    # Start services
    docker compose -f "$compose_file" up -d 2>&1 | filter_docker_output

    echo -e "${SM_GREEN}Services started${SM_NC}"
}

# ===========================================
# OUTPUT FILTERING
# ===========================================

# Filter verbose Docker output for cleaner user experience
filter_docker_output() {
    local line
    local pull_count=0
    local showing_pull=false

    while IFS= read -r line; do
        # Skip empty lines
        [[ -z "$line" ]] && continue

        # Skip layer-level pull progress (e.g., "abc123: Downloading...")
        if [[ "$line" =~ ^[a-f0-9]+:\ (Pulling|Waiting|Downloading|Extracting|Pull\ complete|Already\ exists|Verifying) ]]; then
            if [[ "$showing_pull" == false ]]; then
                echo -e "${SM_GRAY}Pulling image layers...${SM_NC}"
                showing_pull=true
            fi
            ((pull_count++))
            # Show progress every 10 layers
            if ((pull_count % 10 == 0)); then
                echo -ne "\r${SM_GRAY}  Processed $pull_count layers...${SM_NC}"
            fi
            continue
        fi

        # Complete pull progress line
        if [[ "$showing_pull" == true ]] && [[ ! "$line" =~ ^[a-f0-9]+: ]]; then
            echo ""
            showing_pull=false
            pull_count=0
        fi

        # Skip digest lines
        [[ "$line" =~ ^Digest:\ sha256: ]] && continue

        # Skip status lines (but show the final one)
        if [[ "$line" =~ ^Status: ]]; then
            echo -e "${SM_GRAY}$line${SM_NC}"
            continue
        fi

        # Show container events
        if [[ "$line" =~ Container.*Started ]] || [[ "$line" =~ Container.*Created ]]; then
            echo -e "${SM_GREEN}$line${SM_NC}"
            continue
        fi

        if [[ "$line" =~ Container.*Stopped ]]; then
            echo -e "${SM_YELLOW}$line${SM_NC}"
            continue
        fi

        # Show errors
        if [[ "$line" =~ [Ee]rror ]] || [[ "$line" =~ [Ff]ailed ]]; then
            echo -e "${SM_RED}$line${SM_NC}"
            continue
        fi

        # Show warnings
        if [[ "$line" =~ [Ww]arning ]]; then
            echo -e "${SM_YELLOW}$line${SM_NC}"
            continue
        fi

        # Everything else goes to debug (hidden unless verbose)
        if [[ "${VERBOSE:-false}" == "true" ]]; then
            echo -e "${SM_GRAY}$line${SM_NC}"
        fi
    done
}

# ===========================================
# MIGRATION HELPERS
# ===========================================

# Check if there's an existing doom-coding installation
has_existing_installation() {
    # Check for existing containers
    if docker ps -a --format '{{.Names}}' 2>/dev/null | grep -q "^doom-"; then
        return 0
    fi

    # Check for existing volumes
    if docker volume ls --format '{{.Name}}' 2>/dev/null | grep -q "^doom-"; then
        return 0
    fi

    return 1
}

# Detect existing code-server installations (non-doom)
detect_external_code_server() {
    local found=()

    # Check Docker containers
    while IFS= read -r container; do
        [[ -z "$container" ]] && continue
        # Skip doom containers
        [[ "$container" =~ ^doom- ]] && continue

        if [[ "$container" =~ code-server ]] || docker inspect "$container" 2>/dev/null | grep -q "code-server"; then
            found+=("$container")
        fi
    done < <(docker ps -a --format '{{.Names}}' 2>/dev/null)

    # Check for host installation
    if command -v code-server &>/dev/null; then
        found+=("host:code-server")
    fi

    # Check common config locations
    local config_paths=(
        "$HOME/.local/share/code-server"
        "/config/.local/share/code-server"
        "/home/coder/.local/share/code-server"
    )

    for path in "${config_paths[@]}"; do
        if [[ -d "$path" ]]; then
            found+=("config:$path")
        fi
    done

    printf '%s\n' "${found[@]}"
}

# Create a backup of existing configuration
backup_existing_config() {
    local backup_dir="${1:-/tmp/doom-coding-backup-$(date +%Y%m%d-%H%M%S)}"

    mkdir -p "$backup_dir"

    echo -e "${SM_BLUE}Creating backup at: $backup_dir${SM_NC}"

    # Backup .env if exists
    if [[ -f .env ]]; then
        cp .env "$backup_dir/"
    fi

    # Backup secrets if they exist
    if [[ -d secrets ]]; then
        cp -r secrets "$backup_dir/"
    fi

    # Export Docker volumes
    for volume in doom-code-server-config doom-claude-config doom-tailscale-state; do
        if docker volume inspect "$volume" &>/dev/null; then
            echo -e "${SM_GRAY}Backing up volume: $volume${SM_NC}"
            docker run --rm \
                -v "${volume}:/data" \
                -v "${backup_dir}:/backup" \
                alpine tar czf "/backup/${volume}.tar.gz" -C /data . 2>/dev/null || true
        fi
    done

    echo -e "${SM_GREEN}Backup complete: $backup_dir${SM_NC}"
    echo "$backup_dir"
}

# ===========================================
# HEALTH CHECKS
# ===========================================

# Wait for a container to be healthy
# Args: $1 = container name, $2 = timeout (default 60)
wait_for_healthy() {
    local container="$1"
    local timeout="${2:-60}"
    local elapsed=0
    local interval=2

    while [[ $elapsed -lt $timeout ]]; do
        local health
        health=$(docker inspect --format '{{.State.Health.Status}}' "$container" 2>/dev/null || echo "none")

        case "$health" in
            healthy)
                return 0
                ;;
            none)
                # No healthcheck, check if running
                local state
                state=$(docker inspect --format '{{.State.Status}}' "$container" 2>/dev/null || echo "unknown")
                if [[ "$state" == "running" ]]; then
                    return 0
                fi
                ;;
        esac

        sleep $interval
        ((elapsed += interval))
    done

    return 1
}

# Wait for all doom services to be healthy
wait_for_services() {
    local timeout="${1:-120}"
    local all_healthy=true

    echo -e "${SM_BLUE}Waiting for services to be healthy...${SM_NC}"

    for container in doom-tailscale doom-code-server doom-claude; do
        echo -ne "  ${SM_GRAY}$container: checking...${SM_NC}"
        if wait_for_healthy "$container" "$timeout"; then
            echo -e "\r  ${SM_GREEN}$container: healthy${SM_NC}     "
        else
            echo -e "\r  ${SM_YELLOW}$container: not ready${SM_NC}  "
            all_healthy=false
        fi
    done

    if [[ "$all_healthy" == true ]]; then
        return 0
    else
        return 1
    fi
}

# ===========================================
# USER INTERACTION
# ===========================================

# Display a summary of detected services
show_service_summary() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${SM_GREEN}Service Detection Summary${SM_NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    # Existing doom installation
    if has_existing_installation; then
        echo -e "  ${SM_BLUE}Existing doom-coding installation detected${SM_NC}"
        echo -e "  ${SM_GRAY}Will upgrade and preserve your settings${SM_NC}"
    fi

    # Port conflicts
    local conflicts
    conflicts=$(check_port_conflicts)
    if [[ "$conflicts" != "[]" ]]; then
        echo ""
        echo -e "  ${SM_YELLOW}Port conflicts detected:${SM_NC}"
        echo "$conflicts" | jq -r '.[] | "    Port \(.port): \(.process // .container // "unknown")"'
    fi

    # External code-server
    local external
    external=$(detect_external_code_server)
    if [[ -n "$external" ]]; then
        echo ""
        echo -e "  ${SM_BROWN}External code-server detected:${SM_NC}"
        while IFS= read -r item; do
            echo "    $item"
        done <<< "$external"
    fi

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Display access URLs after successful start
show_access_info() {
    local compose_file="${1:-docker-compose.yml}"

    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${SM_GREEN}Services Started Successfully${SM_NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # Determine access method
    local access_ip=""

    # Try Tailscale IP first (host)
    if command -v tailscale &>/dev/null; then
        access_ip=$(tailscale ip -4 2>/dev/null || true)
    fi

    # Try container Tailscale
    if [[ -z "$access_ip" ]]; then
        access_ip=$(docker exec doom-tailscale tailscale ip -4 2>/dev/null || true)
    fi

    # Fallback to local IP
    if [[ -z "$access_ip" ]]; then
        access_ip=$(hostname -I 2>/dev/null | awk '{print $1}')
    fi

    # Final fallback
    if [[ -z "$access_ip" ]]; then
        access_ip="localhost"
    fi

    echo -e "  ${SM_BROWN}code-server:${SM_NC}"
    echo -e "    https://${access_ip}:8443"
    echo ""
    echo -e "  ${SM_BROWN}Claude Code (ttyd):${SM_NC}"
    echo -e "    http://${access_ip}:7681"
    echo ""

    if [[ "$access_ip" =~ ^100\. ]]; then
        echo -e "  ${SM_GRAY}Access via Tailscale VPN${SM_NC}"
    elif [[ "$access_ip" =~ ^192\.168\. ]] || [[ "$access_ip" =~ ^10\. ]]; then
        echo -e "  ${SM_GRAY}Access via local network${SM_NC}"
    fi

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}
