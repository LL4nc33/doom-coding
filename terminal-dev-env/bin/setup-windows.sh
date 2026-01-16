#!/bin/bash
#===============================================================================
# Terminal Development Environment - WSL2 Setup Script
# Handles Windows/WSL2-specific installation and configuration
#===============================================================================

set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly INSTALL_DIR="/opt/terminal-dev-env"
readonly LOG_FILE="${INSTALL_DIR}/logs/setup-windows.log"

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

log() { echo -e "${BLUE}[INFO]${NC} $*" | tee -a "${LOG_FILE}" 2>/dev/null || echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC} $*" | tee -a "${LOG_FILE}" 2>/dev/null || echo -e "${GREEN}[OK]${NC} $*"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*" | tee -a "${LOG_FILE}" 2>/dev/null || echo -e "${YELLOW}[WARN]${NC} $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*" | tee -a "${LOG_FILE}" 2>/dev/null || echo -e "${RED}[ERROR]${NC} $*"; }
die() { error "$*"; exit 1; }

#===============================================================================
# WSL2 Detection and Verification
#===============================================================================

is_wsl() {
    if grep -qi microsoft /proc/version 2>/dev/null; then
        return 0
    fi
    return 1
}

is_wsl2() {
    if is_wsl; then
        # WSL2 has a different kernel than WSL1
        if grep -qi "microsoft.*WSL2" /proc/version 2>/dev/null || \
           [[ -d /run/WSL ]]; then
            return 0
        fi
    fi
    return 1
}

check_systemd() {
    # Check if systemd is available (WSL 0.67.6+)
    if [[ -d /run/systemd/system ]]; then
        return 0
    fi
    return 1
}

get_windows_username() {
    # Get Windows username from WSL
    local win_user=""
    win_user=$(cmd.exe /c "echo %USERNAME%" 2>/dev/null | tr -d '\r\n') || true
    echo "${win_user}"
}

get_wsl_ip() {
    # Get WSL2 IP address
    ip addr show eth0 2>/dev/null | grep 'inet ' | awk '{print $2}' | cut -d/ -f1
}

get_windows_host_ip() {
    # Get Windows host IP from WSL2 perspective
    cat /etc/resolv.conf | grep nameserver | awk '{print $2}'
}

#===============================================================================
# WSL2 Configuration
#===============================================================================

enable_systemd_in_wsl() {
    log "Enabling systemd in WSL2..."

    local wsl_conf="/etc/wsl.conf"

    # Check if systemd is already enabled
    if grep -q "systemd=true" "${wsl_conf}" 2>/dev/null; then
        success "systemd already enabled in WSL2"
        return 0
    fi

    # Create or update wsl.conf
    cat > "${wsl_conf}" << 'EOF'
[boot]
systemd=true

[network]
generateResolvConf=true

[interop]
enabled=true
appendWindowsPath=true
EOF

    warn "systemd enabled - WSL2 restart required!"
    warn "Run: wsl --shutdown (in PowerShell), then restart WSL"

    success "WSL2 configuration updated"
}

configure_wsl_networking() {
    log "Configuring WSL2 networking..."

    # Get current WSL IP
    local wsl_ip=$(get_wsl_ip)
    local host_ip=$(get_windows_host_ip)

    log "WSL2 IP: ${wsl_ip}"
    log "Windows Host IP: ${host_ip}"

    # Create port forwarding script for Windows
    create_windows_portforward_script

    success "WSL2 networking configured"
}

create_windows_portforward_script() {
    log "Creating Windows port forwarding script..."

    local win_script_dir="/mnt/c/terminal-dev-env"
    mkdir -p "${win_script_dir}"

    # PowerShell script for port forwarding
    cat > "${win_script_dir}/setup-portforward.ps1" << 'POWERSHELL'
#Requires -RunAsAdministrator
#===============================================================================
# Terminal Development Environment - Windows Port Forwarding Setup
# Run this script as Administrator to enable WSL2 port forwarding
#===============================================================================

Write-Host "Terminal Development Environment - Port Forwarding Setup" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# Get WSL2 IP address
$wslIP = (wsl hostname -I).Trim().Split(" ")[0]

if ([string]::IsNullOrEmpty($wslIP)) {
    Write-Host "ERROR: Could not get WSL2 IP address. Is WSL2 running?" -ForegroundColor Red
    exit 1
}

Write-Host "WSL2 IP Address: $wslIP" -ForegroundColor Green

# Ports to forward
$ports = @(80, 443)

# Remove existing port proxy rules
Write-Host "`nRemoving existing port proxy rules..." -ForegroundColor Yellow
foreach ($port in $ports) {
    netsh interface portproxy delete v4tov4 listenport=$port listenaddress=0.0.0.0 2>$null
}

# Add new port proxy rules
Write-Host "Adding port proxy rules..." -ForegroundColor Yellow
foreach ($port in $ports) {
    $result = netsh interface portproxy add v4tov4 listenport=$port listenaddress=0.0.0.0 connectport=$port connectaddress=$wslIP
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  Port $port -> WSL2:$port [OK]" -ForegroundColor Green
    } else {
        Write-Host "  Port $port -> WSL2:$port [FAILED]" -ForegroundColor Red
    }
}

# Configure Windows Firewall
Write-Host "`nConfiguring Windows Firewall..." -ForegroundColor Yellow

# Remove existing rules
Remove-NetFirewallRule -DisplayName "Terminal Dev Environment*" -ErrorAction SilentlyContinue

# Add firewall rules
New-NetFirewallRule -DisplayName "Terminal Dev Environment - HTTP" `
    -Direction Inbound -Action Allow -Protocol TCP -LocalPort 80 `
    -Profile Private,Domain -ErrorAction SilentlyContinue | Out-Null

New-NetFirewallRule -DisplayName "Terminal Dev Environment - HTTPS" `
    -Direction Inbound -Action Allow -Protocol TCP -LocalPort 443 `
    -Profile Private,Domain -ErrorAction SilentlyContinue | Out-Null

Write-Host "Firewall rules configured" -ForegroundColor Green

# Show current port proxy configuration
Write-Host "`nCurrent Port Proxy Configuration:" -ForegroundColor Cyan
netsh interface portproxy show v4tov4

# Get local IP addresses
Write-Host "`nAccess URLs:" -ForegroundColor Cyan
$localIPs = Get-NetIPAddress -AddressFamily IPv4 | Where-Object { $_.IPAddress -notlike "127.*" -and $_.IPAddress -notlike "169.*" }
foreach ($ip in $localIPs) {
    Write-Host "  https://$($ip.IPAddress)/" -ForegroundColor Green
}

Write-Host "`nSetup complete!" -ForegroundColor Green
Write-Host "Note: Run this script again if WSL2 IP changes (after reboot)" -ForegroundColor Yellow

# Keep window open
Read-Host "`nPress Enter to exit"
POWERSHELL

    # Create batch file launcher
    cat > "${win_script_dir}/setup-portforward.bat" << 'BATCH'
@echo off
PowerShell -ExecutionPolicy Bypass -File "%~dp0setup-portforward.ps1"
pause
BATCH

    # Create scheduled task setup script
    cat > "${win_script_dir}/create-scheduled-task.ps1" << 'POWERSHELL'
#Requires -RunAsAdministrator
# Create scheduled task to run port forwarding on login

$taskName = "Terminal Dev Environment - Port Forward"
$scriptPath = "C:\terminal-dev-env\setup-portforward.ps1"

# Remove existing task
Unregister-ScheduledTask -TaskName $taskName -Confirm:$false -ErrorAction SilentlyContinue

# Create new task
$action = New-ScheduledTaskAction -Execute "PowerShell.exe" -Argument "-ExecutionPolicy Bypass -WindowStyle Hidden -File `"$scriptPath`""
$trigger = New-ScheduledTaskTrigger -AtLogon
$principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries

Register-ScheduledTask -TaskName $taskName -Action $action -Trigger $trigger -Principal $principal -Settings $settings

Write-Host "Scheduled task created: $taskName" -ForegroundColor Green
POWERSHELL

    chmod +x "${win_script_dir}"/*.ps1 2>/dev/null || true

    success "Windows scripts created in C:\\terminal-dev-env\\"
    log "Run 'setup-portforward.bat' as Administrator on Windows to enable port forwarding"
}

#===============================================================================
# WSL2-Specific Service Management
#===============================================================================

setup_wsl_services() {
    log "Setting up services for WSL2..."

    if check_systemd; then
        log "systemd detected - using standard service management"
        # Use standard systemd setup
        source "${SCRIPT_DIR}/setup-linux.sh"
        configure_systemd
    else
        log "systemd not available - using alternative service management"
        setup_wsl_services_nosystemd
    fi
}

setup_wsl_services_nosystemd() {
    log "Setting up services without systemd..."

    # Create startup script
    cat > "${INSTALL_DIR}/bin/start-services.sh" << 'EOF'
#!/bin/bash
# Start Terminal Development Environment services (non-systemd)

INSTALL_DIR="/opt/terminal-dev-env"
PIDDIR="${INSTALL_DIR}/run"
mkdir -p "${PIDDIR}"

start_ttyd() {
    if [[ -f "${PIDDIR}/ttyd.pid" ]] && kill -0 $(cat "${PIDDIR}/ttyd.pid") 2>/dev/null; then
        echo "ttyd already running"
        return
    fi

    ttyd --port 7681 \
         --interface 127.0.0.1 \
         --max-clients 5 \
         --ping-interval 30 \
         "${INSTALL_DIR}/bin/terminal-session.sh" &

    echo $! > "${PIDDIR}/ttyd.pid"
    echo "ttyd started (PID: $(cat ${PIDDIR}/ttyd.pid))"
}

start_nginx() {
    if [[ -f "${PIDDIR}/nginx.pid" ]] && kill -0 $(cat "${PIDDIR}/nginx.pid") 2>/dev/null; then
        echo "nginx already running"
        return
    fi

    nginx
    echo "nginx started"
}

start_ttyd
start_nginx
echo "Services started"
EOF
    chmod +x "${INSTALL_DIR}/bin/start-services.sh"

    # Create stop script
    cat > "${INSTALL_DIR}/bin/stop-services.sh" << 'EOF'
#!/bin/bash
# Stop Terminal Development Environment services

INSTALL_DIR="/opt/terminal-dev-env"
PIDDIR="${INSTALL_DIR}/run"

if [[ -f "${PIDDIR}/ttyd.pid" ]]; then
    kill $(cat "${PIDDIR}/ttyd.pid") 2>/dev/null
    rm -f "${PIDDIR}/ttyd.pid"
    echo "ttyd stopped"
fi

nginx -s stop 2>/dev/null
echo "nginx stopped"
EOF
    chmod +x "${INSTALL_DIR}/bin/stop-services.sh"

    # Create Windows startup script
    local win_script_dir="/mnt/c/terminal-dev-env"
    cat > "${win_script_dir}/start-wsl-services.bat" << 'BATCH'
@echo off
echo Starting Terminal Development Environment in WSL2...
wsl -u root /opt/terminal-dev-env/bin/start-services.sh
echo Services started. Access at https://localhost/
pause
BATCH

    success "WSL2 service scripts created"
}

#===============================================================================
# Main Setup Flow
#===============================================================================

verify_wsl_environment() {
    log "Verifying WSL environment..."

    if ! is_wsl; then
        die "This script must be run inside WSL"
    fi

    if ! is_wsl2; then
        warn "Running in WSL1 - some features may not work correctly"
        warn "Consider upgrading to WSL2 for better performance"
    else
        success "WSL2 environment detected"
    fi

    if ! check_systemd; then
        warn "systemd not enabled in WSL2"
        enable_systemd_in_wsl
    else
        success "systemd available"
    fi
}

run_linux_setup() {
    log "Running Linux setup in WSL2..."

    if [[ -f "${SCRIPT_DIR}/setup-linux.sh" ]]; then
        source "${SCRIPT_DIR}/setup-linux.sh"

        verify_system_requirements
        install_core_dependencies
        install_terminal_tools
        install_ttyd
        install_nginx

        generate_ssl_certificates "localhost"
        setup_authentication "admin"

        configure_nginx
        setup_user_configs
    else
        die "setup-linux.sh not found"
    fi
}

#===============================================================================
# Main
#===============================================================================

main() {
    echo ""
    echo "=========================================="
    echo "  Terminal Development Environment"
    echo "  WSL2 Setup Script"
    echo "=========================================="
    echo ""

    mkdir -p "${INSTALL_DIR}/logs"

    verify_wsl_environment
    run_linux_setup
    configure_wsl_networking
    setup_wsl_services

    if check_systemd; then
        # Start services with systemd
        systemctl enable ttyd nginx 2>/dev/null || true
        systemctl start ttyd nginx 2>/dev/null || true
    else
        # Start services manually
        "${INSTALL_DIR}/bin/start-services.sh"
    fi

    echo ""
    echo "=========================================="
    echo "  WSL2 Setup Complete!"
    echo "=========================================="
    echo ""
    echo "Next steps for Windows:"
    echo ""
    echo "1. Open PowerShell as Administrator"
    echo "2. Run: C:\\terminal-dev-env\\setup-portforward.bat"
    echo ""
    echo "Access terminal at: https://192.168.178.78/"
    echo "(Or https://localhost/ from the same machine)"
    echo ""
    echo "Credentials in: ${INSTALL_DIR}/config/nginx/.htpasswd.plain"
    echo ""
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
