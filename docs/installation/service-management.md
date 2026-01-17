# Service Management Strategy

This document describes doom-coding's approach to service detection, conflict resolution, migration, and lifecycle management.

## Overview

doom-coding uses a sophisticated service management system to:

1. **Detect** existing services and installations
2. **Resolve** port and container name conflicts
3. **Migrate** from existing code-server installations
4. **Manage** service lifecycle (start/stop/restart)
5. **Log** operations with user-friendly output

## Service Detection

### What Gets Detected

The installer automatically detects:

- **doom-coding containers**: Previous installations (doom-tailscale, doom-code-server, doom-claude)
- **External code-server**: Any non-doom code-server instance
- **Port usage**: Services occupying ports 8443 (code-server) and 7681 (ttyd)
- **Host Tailscale**: Native Tailscale installation on the host

### Detection Methods

```
+------------------+     +-------------------+     +------------------+
|  Docker API      | --> |  Service Manager  | <-- |  Port Scanning   |
|  (containers)    |     |                   |     |  (lsof/ss/nc)    |
+------------------+     +-------------------+     +------------------+
                               |
                               v
                         +-----------+
                         |  Analysis |
                         +-----------+
                               |
              +----------------+----------------+
              v                v                v
        +---------+      +-----------+    +---------+
        | Upgrade |      | Migration |    | Parallel|
        | (doom)  |      | (external)|    | Install |
        +---------+      +-----------+    +---------+
```

## Conflict Resolution

### Port Conflicts

When a target port is already in use:

| Scenario | Resolution |
|----------|------------|
| Previous doom-coding | Stop, upgrade, restart |
| External code-server | Offer migration or alternative port |
| Unknown service | Suggest alternative port |

**Port Resolution Priority:**

1. Default port (8443 for code-server)
2. Next available in range 8000-9000
3. System-assigned if range exhausted

### Container Name Conflicts

Container names are fixed (doom-*) to enable:
- Consistent networking
- Easy identification
- Reliable health checks

If names conflict, the existing containers are stopped and replaced during upgrade.

## Migration Strategies

### Fresh Installation

- No existing doom-coding
- No port conflicts
- Standard installation proceeds

### Upgrade Existing

Triggered when previous doom-coding detected:

1. Create backup of current state
2. Stop existing containers
3. Pull new images
4. Start updated containers
5. Verify health

**Preserved Data:**
- Docker volumes (code-server-config, claude-config, tailscale-state)
- .env configuration
- Workspace files

### Migrate from External

Triggered when external code-server detected:

1. Backup external configuration
2. Stop external service (optional)
3. Copy extensions to doom-coding
4. Copy settings to doom-coding
5. Start doom-coding

**Migrated Data:**
- VS Code extensions
- User settings (settings.json)
- Keybindings

### Parallel Installation

When conflicts cannot be resolved automatically:

1. Assign alternative ports
2. Use modified compose file
3. Both systems can run simultaneously

## Logging Strategy

### Log Levels

| Level | User Sees | Log File |
|-------|-----------|----------|
| ERROR | Always | Yes |
| WARNING | Always | Yes |
| INFO | Always | Yes |
| DEBUG | Verbose mode only | Yes |
| PROGRESS | Always (in-place updates) | Yes |

### Docker Output Filtering

The installer filters verbose Docker output:

**Shown to User:**
```
Pulling images...
  doom-code-server: pulling
  doom-code-server: ready
Container doom-code-server Created
Container doom-code-server Started
```

**Hidden (but logged):**
```
abc123def456: Pulling fs layer
abc123def456: Downloading [========>     ] 45%
abc123def456: Download complete
abc123def456: Pull complete
Digest: sha256:...
Status: Downloaded newer image
```

### Log File Location

All operations are logged to: `/var/log/doom-coding-install.log`

```bash
# View recent logs
tail -100 /var/log/doom-coding-install.log

# View only errors
grep '\[ERROR\]' /var/log/doom-coding-install.log

# Follow live installation
tail -f /var/log/doom-coding-install.log
```

## Service Lifecycle

### Starting Services

```bash
# Via installer
./scripts/install.sh

# Direct compose
docker compose -f docker-compose.yml up -d

# Using lifecycle manager (Go)
lifecycle.Start(ctx, plan)
```

### Stopping Services

```bash
# Graceful stop (30s timeout)
docker compose -f docker-compose.yml down

# With longer timeout
docker compose -f docker-compose.yml down --timeout 60

# Force stop
docker compose -f docker-compose.yml down --timeout 0
```

### Health Checks

Each container has health checks:

| Container | Check | Interval | Timeout |
|-----------|-------|----------|---------|
| doom-tailscale | `tailscale status` | 30s | 10s |
| doom-code-server | `curl localhost:8443/healthz` | 30s | 10s |
| doom-claude | Container running | N/A | N/A |

```bash
# Check all health
docker compose ps

# Individual container
docker inspect --format '{{.State.Health.Status}}' doom-code-server
```

## Configuration

### Environment Variables

```bash
# Installation behavior
VERBOSE=true          # Show debug output
DRY_RUN=true          # Preview without changes
FORCE=true            # Overwrite existing

# Port configuration (future)
CODE_SERVER_PORT=8443
TTYD_PORT=7681
```

### Port Ranges

Default port search range: 8000-9000

```go
// Customizing in Go
manager.portRange = PortRange{Start: 9000, End: 10000}
```

## Troubleshooting

### Port Already in Use

```bash
# Find what's using port 8443
lsof -i :8443
# or
ss -tlnp sport = :8443

# Stop the process
kill <PID>
# or for containers
docker stop <container_name>
```

### Migration Failed

```bash
# Check backup location
ls -la /tmp/doom-coding-backup-*

# Restore from backup
cd /tmp/doom-coding-backup-YYYYMMDD-HHMMSS
docker run --rm -v doom-code-server-config:/data -v $(pwd):/backup \
  alpine tar xzf /backup/doom-code-server-config.tar.gz -C /data
```

### Services Not Starting

```bash
# Check logs
docker compose logs

# Check specific service
docker compose logs doom-code-server

# Check container status
docker inspect doom-code-server --format '{{json .State}}'
```

## API Reference

### Go Packages

```go
import "doom-coding/internal/service"

// Service detection
manager := service.NewManager(projectRoot)
services, _ := manager.DetectExistingServices(ctx)
conflicts, _ := manager.CheckPortConflicts(ctx, targetPorts)

// Migration
migrator := service.NewMigrator(manager, projectRoot)
plan, _ := migrator.AnalyzeExisting(ctx, targetPorts)
result, _ := migrator.Execute(ctx, plan)

// Lifecycle
lifecycle := service.NewLifecycleManager(manager, projectRoot, composeFile)
lifecycle.SetLogger(logger)
result, _ := lifecycle.Start(ctx, plan)

// Logging
logger := service.NewLogger(fileWriter, userWriter)
logger.SetVerbose(true)
logger.Info("startup", "Starting services...")
```

### Shell Functions

```bash
source scripts/lib/service-manager.sh

# Detection
detect_doom_services        # JSON array of doom services
check_port_conflicts        # JSON array of port conflicts
has_existing_installation   # Returns 0 if exists

# Resolution
handle_port_conflict 8443 "code-server"
find_available_port 8443

# Lifecycle
stop_doom_services 30
start_doom_services docker-compose.yml
wait_for_services 120

# User feedback
show_service_summary
show_access_info
filter_docker_output
```
