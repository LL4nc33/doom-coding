# Native Tailscale Mode

> Use your existing host Tailscale installation instead of a containerized one for optimal performance and simplified networking.

## Overview

Native Tailscale mode leverages your host's existing Tailscale installation to provide VPN access to your Docker containers. Instead of running Tailscale in a container, your services bind directly to the host's ports and are accessible via the host's Tailscale IP address.

**Key Benefits:**
- No TUN device requirements (perfect for LXC containers)
- Better performance (no container networking overhead)
- Simplified architecture (fewer containers to manage)
- Uses existing host Tailscale configuration

**How it differs from standard mode:**

| Aspect | Standard Mode | Native Tailscale Mode |
|--------|---------------|----------------------|
| **Tailscale** | Runs in container | Uses host installation |
| **TUN Device** | Required (`/dev/net/tun`) | Not needed |
| **Port Binding** | Internal container network | Host ports (8443, 7681) |
| **Access Method** | Container's Tailscale IP | Host's Tailscale IP |
| **LXC Support** | Requires TUN device | Works without TUN device |

## When to Use

Choose native-tailscale mode when:

✅ **Tailscale is already running on your host**
- You have an existing Tailscale setup
- Want to reuse existing network configuration
- Multiple services share the same Tailscale node

✅ **Running in LXC containers without TUN device**
- Proxmox LXC containers with unprivileged mode
- Hosting providers that don't support TUN devices
- VPS environments with restricted capabilities

✅ **Performance is critical**
- Eliminating container networking overhead
- Direct host port binding for maximum speed
- Reduced resource usage (no Tailscale container)

❌ **Don't use native-tailscale mode when:**
- Host doesn't have Tailscale installed/running
- You want isolated Tailscale networks per service
- Each container needs its own Tailscale identity

## Requirements

### Host Tailscale Installation

1. **Tailscale must be installed on the host:**
   ```bash
   curl -fsSL https://tailscale.com/install.sh | sh
   ```

2. **Tailscale must be running and authenticated:**
   ```bash
   sudo tailscale up
   ```

3. **Verify Tailscale status:**
   ```bash
   tailscale status
   # Should show: Connected to Tailscale

   tailscale ip -4
   # Should return your Tailscale IP (100.x.x.x)
   ```

### System Requirements

- Docker and Docker Compose installed
- Host ports 8443 and 7681 available
- Internet connectivity for container images

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Docker Host                              │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                Host Tailscale                            │   │
│  │  - Running on host (not in container)                   │   │
│  │  - IP: 100.x.x.x                                        │   │
│  │  - Provides secure mesh networking                       │   │
│  └─────────────────────────────────────────────────────────┘   │
│                             │                                   │
│                             ▼                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Docker Compose Stack                        │   │
│  │                                                          │   │
│  │  ┌────────────────────┐  ┌────────────────────┐        │   │
│  │  │  code-server       │  │  Claude Code       │        │   │
│  │  │  Port: 8443        │  │  Port: 7681        │        │   │
│  │  │  (bind to host)    │  │  (bind to host)    │        │   │
│  │  └────────────────────┘  └────────────────────┘        │   │
│  │                                                          │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  Access via: https://100.x.x.x:8443 (code-server)             │
│              http://100.x.x.x:7681  (Claude ttyd)             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Network Flow

1. **Client Connection**: Client connects to host's Tailscale IP
2. **Port Binding**: Docker containers bind directly to host ports
3. **Service Access**: Services accessible via `100.x.x.x:port`
4. **No Container Networking**: No shared network namespaces needed

## Installation

### Interactive Installation

The installer automatically detects when host Tailscale is running and recommends native mode:

```bash
curl -fsSL https://raw.githubusercontent.com/LL4nc33/doom-coding/main/scripts/install.sh | bash
```

**Installation flow:**
1. System detection identifies running Tailscale
2. Installer recommends native-tailscale mode
3. Select "Docker + Host Tailscale" option
4. Installer verifies host Tailscale configuration
5. Deploys containers with port binding

### Command Line Installation

For automated/unattended installation:

```bash
# Download installer
wget https://raw.githubusercontent.com/LL4nc33/doom-coding/main/scripts/install.sh

# Run with native-tailscale flag
./install.sh --native-tailscale --unattended \
  --code-password="your-secure-password" \
  --anthropic-key="sk-ant-your-key"
```

**CLI flags for native-tailscale mode:**
```bash
./install.sh --native-tailscale          # Enable native mode
./install.sh --native-tailscale \        # With customization
  --code-password="secure123" \
  --anthropic-key="sk-ant-xxx" \
  --workspace="/opt/projects"
```

### Manual Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/LL4nc33/doom-coding.git
   cd doom-coding
   ```

2. **Verify host Tailscale:**
   ```bash
   tailscale status
   tailscale ip -4
   ```

3. **Create environment file:**
   ```bash
   cp .env.example .env.native-tailscale
   ```

4. **Configure secrets:**
   ```bash
   mkdir -p secrets
   echo "sk-ant-your-key" > secrets/anthropic_api_key.txt
   ```

5. **Deploy with native compose file:**
   ```bash
   docker compose -f docker-compose.native-tailscale.yml up -d
   ```

## Configuration

### Docker Compose File

The native-tailscale mode uses `docker-compose.native-tailscale.yml`:

```yaml
services:
  # code-server - VS Code in browser
  code-server:
    image: lscr.io/linuxserver/code-server:latest
    container_name: doom-code-server
    restart: unless-stopped
    ports:
      - "8443:8443"    # Bind to host port
    environment:
      - PUID=${PUID:-1000}
      - PGID=${PGID:-1000}
      - TZ=${TZ:-Europe/Berlin}
      - PASSWORD=${CODE_SERVER_PASSWORD}
      - SUDO_PASSWORD=${SUDO_PASSWORD}
    volumes:
      - code-server-config:/config
      - ${WORKSPACE_PATH:-./workspace}:/workspace

  # Claude Code Container
  claude:
    build:
      context: .
      dockerfile: Dockerfile.claude
    container_name: doom-claude
    restart: unless-stopped
    ports:
      - "7681:7681"    # ttyd web terminal
    secrets:
      - anthropic_api_key
    volumes:
      - claude-config:/home/claude/.claude
      - ${WORKSPACE_PATH:-./workspace}:/workspace
```

### Environment Variables

Create `.env` file with these settings:

```bash
# User Configuration
PUID=1000
PGID=1000
TZ=Europe/Berlin

# Service Passwords
CODE_SERVER_PASSWORD=your-secure-password
SUDO_PASSWORD=your-sudo-password

# Paths
WORKSPACE_PATH=/opt/workspace

# Optional: Resource Limits
CODE_SERVER_MEMORY=2G
CLAUDE_MEMORY=1G
```

### Required Secrets

1. **Anthropic API Key:**
   ```bash
   mkdir -p secrets
   echo "sk-ant-your-api-key" > secrets/anthropic_api_key.txt
   chmod 600 secrets/anthropic_api_key.txt
   ```

2. **Verify secrets mount:**
   ```bash
   docker compose exec claude cat /run/secrets/anthropic_api_key
   ```

### Port Configuration

| Service | Port | Protocol | Description |
|---------|------|----------|-------------|
| code-server | 8443 | HTTPS | VS Code web interface |
| Claude ttyd | 7681 | HTTP | Web-based terminal |

**Note:** These ports are bound directly to the host, accessible via your Tailscale IP.

## Access

### Getting Your Tailscale IP

```bash
# Get your Tailscale IPv4 address
tailscale ip -4
# Output: 100.64.1.2

# Get full Tailscale status
tailscale status
```

### Service URLs

Once deployed, access services via your Tailscale IP:

```bash
# Get your Tailscale IP
TS_IP=$(tailscale ip -4)

echo "🌐 Access URLs:"
echo "  code-server: https://$TS_IP:8443"
echo "  Claude ttyd: http://$TS_IP:7681"
```

**Example URLs:**
- **code-server**: `https://100.64.1.2:8443`
- **Claude terminal**: `http://100.64.1.2:7681`

### First Time Access

1. **code-server login:**
   - Open `https://YOUR_TAILSCALE_IP:8443`
   - Enter your CODE_SERVER_PASSWORD
   - Accept self-signed certificate warning

2. **Claude Code access:**
   - Open `http://YOUR_TAILSCALE_IP:7681`
   - This opens a web terminal with Claude Code available
   - Run `claude` to start the AI assistant

### Mobile Access

Access from mobile devices using Tailscale app:

1. Install Tailscale on your mobile device
2. Login with same Tailscale account
3. Access via browser: `https://100.x.x.x:8443`

## Troubleshooting

### Host Tailscale Issues

**Problem:** Tailscale not running on host
```bash
# Check Tailscale status
tailscale status
# Error: Tailscale is stopped

# Solution: Start Tailscale
sudo tailscale up
```

**Problem:** No Tailscale IP assigned
```bash
# Check if authenticated
tailscale status
# Shows: Logged out

# Solution: Authenticate
sudo tailscale up
# Follow authentication link
```

**Problem:** Tailscale installed but not in PATH
```bash
# Check installation
which tailscale
# Returns: tailscale not found

# Solution: Add to PATH or use full path
export PATH=$PATH:/usr/bin
```

### Container Access Issues

**Problem:** Can't access code-server on 8443
```bash
# Check if port is bound
netstat -tulpn | grep 8443
# Should show Docker binding

# Check container status
docker compose ps
# code-server should be "Up"

# Check logs
docker compose logs code-server
```

**Problem:** Port already in use
```bash
# Check what's using the port
sudo lsof -i :8443
# Kill conflicting process or change port

# Alternative: Use different port
# Edit docker-compose.native-tailscale.yml
ports:
  - "8444:8443"  # Use 8444 instead
```

### Network Connectivity

**Problem:** Can access locally but not via Tailscale IP
```bash
# Test local access
curl -k https://localhost:8443/healthz

# Test Tailscale access
TS_IP=$(tailscale ip -4)
curl -k https://$TS_IP:8443/healthz

# Check Tailscale firewall
tailscale status --json | jq .TailscaleIPs
```

**Problem:** Services not accessible from other devices
```bash
# Check ACLs in Tailscale admin console
# Ensure your device can access the host

# Check if ports are actually bound
docker port doom-code-server
# Should show: 8443/tcp -> 0.0.0.0:8443
```

### Docker Issues

**Problem:** Containers not starting
```bash
# Check Docker daemon
sudo systemctl status docker

# Check compose file syntax
docker compose -f docker-compose.native-tailscale.yml config

# Check logs
docker compose logs --follow
```

**Problem:** Volume permissions
```bash
# Fix workspace permissions
sudo chown -R $(id -u):$(id -g) ./workspace

# Check PUID/PGID in .env
id -u && id -g
```

### Health Checks

**Verify complete setup:**
```bash
#!/bin/bash
echo "🔍 Health Check for Native Tailscale Mode"
echo

# Host Tailscale
echo "📡 Host Tailscale:"
if tailscale status &>/dev/null; then
    echo "  ✅ Tailscale running"
    echo "  📍 IP: $(tailscale ip -4)"
else
    echo "  ❌ Tailscale not running"
fi
echo

# Docker containers
echo "🐳 Docker Services:"
docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
echo

# Port accessibility
echo "🔌 Port Tests:"
TS_IP=$(tailscale ip -4)
if curl -k --connect-timeout 5 https://$TS_IP:8443/healthz &>/dev/null; then
    echo "  ✅ code-server accessible"
else
    echo "  ❌ code-server not accessible"
fi

if curl --connect-timeout 5 http://$TS_IP:7681 &>/dev/null; then
    echo "  ✅ Claude ttyd accessible"
else
    echo "  ❌ Claude ttyd not accessible"
fi
```

## Security Considerations

### Network Security

**Tailscale Security Model:**
- End-to-end encryption between devices
- Zero Trust networking (no open internet ports)
- Device authentication via your identity provider
- Fine-grained ACL controls available

**Port Binding:**
```bash
# Ports are bound to all interfaces (0.0.0.0)
# But only accessible via Tailscale due to firewall
netstat -tulpn | grep -E "(8443|7681)"
```

### Container Security

**Resource Limits:**
```yaml
# Prevent resource exhaustion
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 2G
    reservations:
      memory: 512M
```

**Read-Only Configurations:**
```yaml
# Mount configs as read-only
volumes:
  - ./config/zsh/.zshrc:/config/.zshrc:ro
  - ./config/tmux/tmux.conf:/config/.tmux.conf:ro
```

### Access Control

**Tailscale ACLs:**
```json
{
  "acls": [
    {
      "action": "accept",
      "src": ["user@example.com"],
      "dst": ["tag:doom-coding:8443,7681"]
    }
  ]
}
```

**code-server Authentication:**
- Always use strong passwords
- Consider additional authentication layers
- Monitor access logs

### Secrets Management

**Best Practices:**
```bash
# Use Docker secrets (not environment variables)
secrets:
  anthropic_api_key:
    file: ./secrets/anthropic_api_key.txt

# Secure file permissions
chmod 600 secrets/*.txt
```

## Examples

### Basic Deployment

```bash
# 1. Verify host Tailscale
tailscale status

# 2. Clone and configure
git clone https://github.com/LL4nc33/doom-coding.git
cd doom-coding
cp .env.example .env

# 3. Set credentials
echo "your-password" > .env
echo "sk-ant-key" > secrets/anthropic_api_key.txt

# 4. Deploy
docker compose -f docker-compose.native-tailscale.yml up -d

# 5. Access
echo "Visit: https://$(tailscale ip -4):8443"
```

### Custom Workspace

```bash
# Deploy with custom workspace location
export WORKSPACE_PATH="/home/projects"
mkdir -p $WORKSPACE_PATH
chmod 755 $WORKSPACE_PATH

docker compose -f docker-compose.native-tailscale.yml up -d
```

### Multi-User Setup

```bash
# Deploy multiple instances for different users
# User 1
export COMPOSE_PROJECT_NAME=doom-user1
export CODE_SERVER_PORT=8443
docker compose up -d

# User 2
export COMPOSE_PROJECT_NAME=doom-user2
export CODE_SERVER_PORT=8444
docker compose up -d

# Access:
# User 1: https://TAILSCALE_IP:8443
# User 2: https://TAILSCALE_IP:8444
```

### Development Workflow

```bash
# Start development environment
docker compose -f docker-compose.native-tailscale.yml up -d

# Access code-server in browser
open https://$(tailscale ip -4):8443

# Use Claude Code in terminal
docker compose exec claude claude

# Monitor logs
docker compose logs -f code-server claude
```

### Backup and Restore

```bash
# Backup volumes
docker run --rm -v doom-code-server-config:/data \
  -v $(pwd):/backup alpine \
  tar czf /backup/code-server-config.tar.gz -C /data .

# Restore volumes
docker run --rm -v doom-code-server-config:/data \
  -v $(pwd):/backup alpine \
  tar xzf /backup/code-server-config.tar.gz -C /data
```

### Performance Monitoring

```bash
# Monitor resource usage
docker stats doom-code-server doom-claude

# Check Tailscale performance
tailscale ping YOUR_DEVICE_IP

# Monitor network traffic
tailscale netcheck
```

## Next Steps

**Explore Advanced Features:**
- [Container Customization](../advanced/customization.md)
- [Backup and Recovery](../advanced/backup-recovery.md)
- [Monitoring Setup](../advanced/monitoring.md)

**Alternative Modes:**
- [Standard Docker + Tailscale](overview.md) - Full container mode
- [Local Network Mode](../installation/manual-installation.md#local-mode) - No VPN
- [Terminal-Only Mode](../terminal/customization.md) - Lightweight setup

**Security Hardening:**
- [SSH Configuration](../security/ssh-hardening.md)
- [Secrets Management](../security/secrets-management.md)
- [Tailscale ACLs](../security/tailscale-security.md)

---

> 💡 **Tip**: Native Tailscale mode is perfect for LXC containers and high-performance setups. It leverages your existing Tailscale configuration while providing full containerized development capabilities.