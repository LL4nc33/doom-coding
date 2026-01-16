# Docker Setup Overview

Architecture and configuration of the Doom Coding Docker stack.

## Architecture

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

The Tailscale container provides networking for all other services.

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

### Starting Services

```bash
# Start all services
docker compose up -d

# Start specific service
docker compose up -d code-server

# Rebuild and start
docker compose up -d --build
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

- [Container Customization](customization.md)
- [Docker Security](security.md)
- [Backup and Restore](../advanced/backup-recovery.md)