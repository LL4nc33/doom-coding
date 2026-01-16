# Debugging Quick Reference

Quick reference for debugging doom-coding installation and runtime issues.

---

## Diagnostic Commands

### Installation Diagnostics

```bash
# View installation log
tail -f /var/log/doom-coding-install.log

# Run health check
./scripts/health-check.sh

# Collect full diagnostics
./scripts/collect-diagnostics.sh

# Interactive troubleshooter
./scripts/troubleshoot.sh
```

### Docker Diagnostics

```bash
# Container status
docker compose ps

# View all logs
docker compose logs -f

# Specific container logs
docker logs doom-tailscale --tail 50
docker logs doom-code-server --tail 50
docker logs doom-claude --tail 50

# Container resource usage
docker stats

# Inspect container
docker inspect doom-code-server

# Network status
docker network inspect doom-coding-network

# Volume list
docker volume ls
```

### System Diagnostics

```bash
# System resources
htop                           # CPU/Memory
df -h                          # Disk space
free -h                        # Memory details
iotop                          # Disk I/O

# Network
ip addr                        # Network interfaces
ss -tlnp                       # Listening ports
netstat -tlnp                  # Alternative

# Process info
ps aux | grep docker           # Docker processes
systemctl status docker        # Docker service status
journalctl -u docker -f        # Docker journal

# Container environment
docker exec doom-code-server env  # Environment variables
docker exec doom-code-server pwd  # Working directory
```

### Tailscale Diagnostics

```bash
# Tailscale status
tailscale status
tailscale status --json

# Get IP
tailscale ip -4

# Check TUN device
ls -l /dev/net/tun
cat /proc/modules | grep tun

# Service logs
journalctl -u tailscaled -f

# Inside container
docker exec doom-tailscale tailscale status
```

---

## Common Error Patterns

### "Cannot connect to Docker daemon"

**Cause**: Docker not running

**Fix**:
```bash
sudo systemctl start docker
sudo systemctl enable docker
```

**Verify**:
```bash
docker info
```

---

### "Port already allocated"

**Cause**: Port 8443 in use

**Diagnose**:
```bash
sudo lsof -i :8443
```

**Fix**:
```bash
# Kill process
sudo kill $(lsof -t -i:8443)

# Or change port in docker-compose.yml
```

---

### "Permission denied" on Docker

**Cause**: User not in docker group

**Fix**:
```bash
sudo usermod -aG docker $USER
newgrp docker
# Or log out and back in
```

**Verify**:
```bash
groups | grep docker
docker ps
```

---

### "No space left on device"

**Cause**: Disk full

**Diagnose**:
```bash
df -h
docker system df
```

**Fix**:
```bash
docker system prune -af
docker volume prune -f
```

---

### "TUN device not available"

**Cause**: /dev/net/tun missing

**For LXC (Proxmox)**:
```bash
# On Proxmox host:
nano /etc/pve/lxc/<CTID>.conf

# Add:
lxc.cgroup2.devices.allow: c 10:200 rwm
lxc.mount.entry: /dev/net/tun dev/net/tun none bind,create=file

# Restart
pct restart <CTID>
```

**Verify**:
```bash
ls -l /dev/net/tun
# Should show: crw-rw-rw- 1 root root 10, 200
```

---

### "Container exits immediately"

**Diagnose**:
```bash
# Check exit code
docker inspect doom-code-server --format='{{.State.ExitCode}}'

# View logs
docker logs doom-code-server

# Check for missing files
docker exec doom-code-server ls -la /run/secrets/
```

**Common causes**:
1. Missing secrets file
2. Invalid environment variables
3. Command syntax error
4. Dependency not met

---

### "Cannot access web UI"

**Diagnose network mode**:
```bash
# Check container network
docker inspect doom-code-server --format='{{.HostConfig.NetworkMode}}'

# Should show "service:tailscale" or "default"
```

**For Tailscale mode**:
```bash
# Get Tailscale IP
tailscale status
# Access: https://<TAILSCALE-IP>:8443
```

**For local network mode**:
```bash
# Get host IP
hostname -I
# Access: https://<LOCAL-IP>:8443
```

**Test connectivity**:
```bash
curl -k https://localhost:8443/healthz
```

---

## Validation Checks

### Pre-Installation Checklist

```bash
# Internet connectivity
curl -sf https://google.com

# Disk space (need 10GB+)
df -BG . | awk 'NR==2 {print $4}'

# Memory (recommended 2GB+)
free -g | awk '/^Mem:/{print $2}'

# Required commands
command -v docker curl git wget

# Port availability
lsof -i :8443
```

### Post-Installation Checklist

```bash
# Docker running
docker info

# Containers up
docker compose ps

# Services healthy
docker inspect doom-code-server --format='{{.State.Health.Status}}'

# Network accessible
curl -k https://localhost:8443/healthz

# Volumes mounted
docker inspect doom-code-server --format='{{range .Mounts}}{{.Source}} -> {{.Destination}}{{"\n"}}{{end}}'
```

---

## Performance Debugging

### Slow Container Startup

```bash
# Check image pull time
time docker compose pull

# Check build time
time docker compose build

# Check system resources
htop
docker stats
```

### High Memory Usage

```bash
# Container memory
docker stats --no-stream

# System memory
free -h

# Memory by process
ps aux --sort=-%mem | head

# Limit container memory
# Edit docker-compose.yml:
services:
  code-server:
    deploy:
      resources:
        limits:
          memory: 2G
```

### High CPU Usage

```bash
# Container CPU
docker stats --no-stream

# Process CPU
top -bn1 | head -20

# Identify bottleneck
strace -c -p <PID>
```

### Slow Network

```bash
# Test inside container
docker exec doom-code-server ping -c 5 8.8.8.8

# DNS issues
docker exec doom-code-server cat /etc/resolv.conf

# Network mode
docker network inspect doom-coding-network
```

---

## Log Analysis Patterns

### Search for Errors

```bash
# Installation log
grep -i "error\|fail" /var/log/doom-coding-install.log | tail -20

# Docker logs
docker compose logs | grep -i "error\|exception\|fatal"

# System journal
journalctl -u docker | grep -i "error\|fail"
```

### Timeline Analysis

```bash
# Last 100 lines with timestamps
docker logs doom-code-server --tail 100 --timestamps

# Logs since specific time
docker logs doom-code-server --since 2024-01-01T10:00:00

# Logs for specific duration
journalctl -u docker --since "1 hour ago"
```

### Filter by Component

```bash
# Only warnings and errors
docker logs doom-code-server 2>&1 | grep -E "WARN|ERROR"

# Specific component
docker logs doom-code-server | grep "health"
```

---

## Recovery Procedures

### Clean Docker State

```bash
# Stop all containers
docker compose down

# Remove volumes (DESTRUCTIVE)
docker compose down -v

# Clean system
docker system prune -af

# Remove all doom-coding resources
docker ps -a | grep doom | awk '{print $1}' | xargs docker rm -f
docker volume ls | grep doom | awk '{print $2}' | xargs docker volume rm
docker network ls | grep doom | awk '{print $1}' | xargs docker network rm
```

### Reset Installation

```bash
# Remove environment
rm -f .env

# Clean secrets (backup first!)
rm -rf secrets/

# Re-run installation
./scripts/install.sh
```

### Rollback Docker Installation

```bash
# Stop service
sudo systemctl stop docker

# Remove packages (Debian/Ubuntu)
sudo apt-get remove -y docker-ce docker-ce-cli containerd.io

# Remove data
sudo rm -rf /var/lib/docker
sudo rm -rf /etc/docker

# Remove GPG key
sudo rm -f /etc/apt/keyrings/docker.gpg
sudo rm -f /etc/apt/sources.list.d/docker.list
```

---

## Debug Flags

### Bash Scripts

```bash
# Verbose mode
./scripts/install.sh --verbose

# Dry run (show commands without executing)
./scripts/install.sh --dry-run

# Enable bash debug output
bash -x ./scripts/install.sh
```

### Docker Compose

```bash
# Verbose output
docker compose --verbose up

# Show config
docker compose config

# No cache build
docker compose build --no-cache
```

### Go TUI

```bash
# Verbose mode
./bin/doom-tui --verbose

# Show commands that would run
./bin/doom-tui --show-commands

# Debug build
go build -gcflags="all=-N -l" -o doom-tui-debug cmd/doom-tui/main.go
dlv exec doom-tui-debug
```

---

## Environment Variables for Debugging

```bash
# Enable Docker debug mode
export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=plain

# Verbose Tailscale
export TS_DEBUG_FIREWALL_MODE=auto

# Go debug
export GODEBUG=gctrace=1

# Bash debug
export PS4='+(${BASH_SOURCE}:${LINENO}): ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'
set -x
```

---

## Emergency Contacts / Resources

### Documentation
- Main docs: `/config/repos/doom-coding/docs/`
- Troubleshooting: `/config/repos/doom-coding/docs/troubleshooting/`
- API docs: `/config/repos/doom-coding/docs/api/`

### External Resources
- Docker docs: https://docs.docker.com/
- Tailscale docs: https://tailscale.com/kb/
- code-server docs: https://coder.com/docs/code-server/

### Community
- GitHub Issues: https://github.com/LL4nc33/doom-coding/issues
- GitHub Discussions: https://github.com/LL4nc33/doom-coding/discussions

---

## Development Debugging

### Build Issues

```bash
# Clean build
make clean
make build

# Verbose build
go build -v ./...

# Check dependencies
go mod tidy
go mod verify
```

### TUI Debugging

```bash
# Run with stdout/stderr to file
./bin/doom-tui 2> debug.log

# Use delve debugger
dlv debug cmd/doom-tui/main.go
```

### Test Failures

```bash
# Run tests with verbose
go test -v ./...

# Run specific test
go test -v -run TestSystemDetect ./internal/system

# With race detector
go test -race ./...
```

---

## Quick Tips

1. **Always check logs first**: `docker compose logs -f`
2. **Run health check**: `./scripts/health-check.sh`
3. **Collect diagnostics before reporting**: `./scripts/collect-diagnostics.sh`
4. **Use dry-run for testing**: `./scripts/install.sh --dry-run`
5. **Keep backups**: Before major changes, backup `.env` and `secrets/`
6. **Check permissions**: Many issues are permission-related
7. **Verify network**: Ensure containers can reach internet
8. **Monitor resources**: Watch disk space and memory
9. **Read error messages**: They often contain the fix
10. **When in doubt**: Use the interactive troubleshooter

---

**Last Updated**: 2026-01-16
**Version**: 0.0.4
