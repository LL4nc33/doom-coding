# Docker Setup Overview

Architecture and configuration of the Doom Coding Docker stack.

## Deployment Modes

Doom Coding supports multiple deployment modes to fit different environments:

| Mode | Description | Use Case | TUN Device Required |
|------|-------------|----------|---------------------|
| **Standard** | Containerized Tailscale (default) | General use, full isolation | Yes |
| **Native Tailscale** | Uses host Tailscale installation | LXC containers, performance | No |

### Mode Selection

- **Standard Mode** (this document): Best for most deployments with full container isolation
- **[Native Tailscale Mode](native-tailscale.md)**: Perfect for LXC containers and when TUN device is not available

### Quick Comparison

| Feature | Standard Mode | Native Tailscale Mode |
|---------|---------------|------------------------|
| **Complexity** | Medium | Low |
| **Performance** | Good | Better |
| **Isolation** | High | Medium |
| **LXC Support** | Requires TUN device | Full support |
| **Setup Time** | Longer | Shorter |
| **Maintenance** | Container updates needed | Uses host Tailscale |

**Choose Standard Mode if:**
- You want complete service isolation
- TUN device is available
- You prefer containerized networking

**Choose Native Tailscale Mode if:**
- Running in LXC containers
- TUN device is not available
- You want maximum performance
- Tailscale is already running on host

## Architecture (Standard Mode)

This architecture shows the standard containerized Tailscale mode. For native Tailscale mode architecture, see [native-tailscale.md](native-tailscale.md).

```
┌─────────────────────────────────────────────────────────────────┐
│                        Docker Host                              │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                Docker Compose Stack                      │   │
│  │                                                          │   │
│  │  ┌──────────────────────────────────────────────────┐   │   │
│  │  │              Tailscale Container                 │   │   │
│  │  │  - VPN mesh networking                           │   │   │
│  │  │  - Provides network for other containers         │   │   │
│  │  │  - Exposes services on Tailscale IP              │   │   │
│  │  └──────────────────────────────────────────────────┘   │   │
│  │              │                    │                      │   │
│  │              ▼                    ▼                      │   │
│  │  ┌────────────────────┐  ┌────────────────────┐        │   │
│  │  │  code-server       │  │  Claude Code       │        │   │
│  │  │  (network_mode:    │  │  (network_mode:    │        │   │
│  │  │   service:tailscale│  │   service:tailscale│        │   │
│  │  │  Port 8443         │  │  Interactive CLI   │        │   │
│  │  └────────────────────┘  └────────────────────┘        │   │
│  │                                                          │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  Volumes:                                                       │
│  - doom-tailscale-state     (Tailscale persistent state)       │
│  - doom-code-server-config  (code-server settings/extensions)  │
│  - doom-claude-config       (Claude Code configuration)        │
│  - ./workspace              (Your projects - bind mount)       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Services

### Tailscale (Sidecar)

The Tailscale container provides networking for all other services in standard mode.

> **Note**: In [native-tailscale mode](native-tailscale.md), this container is not needed as the host's Tailscale installation is used instead.

**Key Configuration**:
```yaml
tailscale:
  image: tailscale/tailscale:stable
  cap_add:
    - NET_ADMIN
    - SYS_MODULE
  volumes:
    - tailscale-state:/var/lib/tailscale
    - /dev/net/tun:/dev/net/tun
  environment:
    - TS_AUTHKEY=${TS_AUTHKEY}
    - TS_ACCEPT_DNS=false
```

**Why Sidecar Pattern?**
- All services share the same Tailscale IP
- No need to expose ports on host
- Simplified firewall configuration
- Services only accessible via Tailscale

### code-server

Web-based VS Code accessible via browser.

**Key Configuration**:
```yaml
code-server:
  image: lscr.io/linuxserver/code-server:latest
  network_mode: service:tailscale
  depends_on:
    tailscale:
      condition: service_healthy
  environment:
    - PASSWORD=${CODE_SERVER_PASSWORD}
    - PUID=${PUID}
    - PGID=${PGID}
```

**Network Mode**: Uses `network_mode: service:tailscale` to share
Tailscale's network namespace.

### Claude Code

AI-powered development assistant container.

**Key Configuration**:
```yaml
claude:
  build:
    context: .
    dockerfile: Dockerfile.claude
  network_mode: service:tailscale
  secrets:
    - anthropic_api_key
```

**Native Installation**: Uses the official installer, not npm.

## Volumes

> **Note**: In [native-tailscale mode](native-tailscale.md), the `doom-tailscale-state` volume is not needed since the host manages Tailscale state.

### Named Volumes

| Volume | Purpose | Backup Priority |
|--------|---------|-----------------|
| `doom-tailscale-state` | Tailscale authentication | Medium |
| `doom-code-server-config` | VS Code settings, extensions | High |
| `doom-claude-config` | Claude Code settings | Medium |

### Bind Mounts

| Mount | Container Path | Purpose |
|-------|----------------|---------|
| `./workspace` | `/workspace` | Project files |
| `./config/zsh/.zshrc` | `~/.zshrc` | Shell configuration |
| `./config/tmux/tmux.conf` | `~/.tmux.conf` | Tmux configuration |

## Health Checks

Each service has a health check:

```yaml
healthcheck:
  test: ["CMD", "tailscale", "status", "--json"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 30s
```

Use `condition: service_healthy` to ensure proper startup order.

## Docker Secrets

Sensitive data is passed via Docker secrets:

```yaml
secrets:
  anthropic_api_key:
    file: ./secrets/anthropic_api_key.txt
```

Access in container: `/run/secrets/anthropic_api_key`

## Commands

> **Note**: For [native-tailscale mode](native-tailscale.md), use `docker-compose.native-tailscale.yml` instead of the default compose file.

### Starting Services

```bash
# Start all services (standard mode)
docker compose up -d

# Start specific service
docker compose up -d code-server

# Rebuild and start
docker compose up -d --build

# For native-tailscale mode
docker compose -f docker-compose.native-tailscale.yml up -d
```

### Viewing Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f code-server

# Last 100 lines
docker compose logs --tail=100
```

### Managing Containers

```bash
# Stop all
docker compose down

# Stop and remove volumes
docker compose down -v

# Restart service
docker compose restart code-server

# Execute command in container
docker compose exec code-server bash
```

### Updating

```bash
# Pull latest images
docker compose pull

# Recreate containers
docker compose up -d --force-recreate
```

### Testing Your Deployment

Validate your Docker deployment with our comprehensive testing suite:

```bash
# Basic health verification
./scripts/health-check.sh

# Test Docker-specific functionality
./scripts/test-runner.sh --deployment=docker-tailscale --category=foundation

# Security validation for containers
./scripts/test-runner.sh --category=security --focus=containers

# Full deployment validation
./scripts/test-runner.sh --deployment=docker-tailscale --iterations=1-20
```

**Complete Testing Documentation**: [`../testing/`](../testing/)

## Customization

### Adding Services

Add new services to `docker-compose.yml`:

```yaml
services:
  my-service:
    image: my-image:tag
    network_mode: service:tailscale
    depends_on:
      tailscale:
        condition: service_healthy
```

### Custom Dockerfile

Override with build context:

```yaml
services:
  code-server:
    build:
      context: .
      dockerfile: Dockerfile.code-server
```

### Resource Limits

```yaml
services:
  code-server:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          memory: 1G
```

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker compose logs tailscale

# Verify configuration
docker compose config
```

### Network Issues

```bash
# Check Tailscale status
docker compose exec tailscale tailscale status

# Test connectivity
docker compose exec tailscale ping google.com
```

### Permission Issues

```bash
# Fix ownership
sudo chown -R $(id -u):$(id -g) ./workspace

# Verify PUID/PGID
id -u && id -g
```

## Next Steps

### Testing and Validation

- **[Testing Framework](../testing/)** - Comprehensive 70-iteration testing strategy
- **[Quality Assurance](../testing/test-plan-70.md)** - Detailed testing procedures

### Alternative Deployment Modes

- **[Native Tailscale Mode](native-tailscale.md)** - Use host Tailscale for better performance and LXC compatibility

### Configuration and Management

- [Container Customization](customization.md)
- [Docker Security](security.md)
- [Backup and Restore](../advanced/backup-recovery.md)