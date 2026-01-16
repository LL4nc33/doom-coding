# Installation Guide

Complete installation guide for Doom Coding remote development environment.

## System Requirements

### Minimum Requirements
- **CPU**: 2 cores
- **RAM**: 4 GB
- **Disk**: 20 GB free space
- **Network**: Internet connection

### Recommended Requirements
- **CPU**: 4+ cores
- **RAM**: 8+ GB
- **Disk**: 50+ GB SSD
- **Network**: Stable broadband connection

### Supported Operating Systems
- Ubuntu 20.04 LTS or newer
- Debian 11 (Bullseye) or newer
- Arch Linux (rolling release)

### Supported Architectures
- x86_64 (amd64)
- ARM64 (aarch64)

## Prerequisites

### Required
1. **Tailscale Account**: Sign up at https://tailscale.com
2. **Git**: For cloning the repository
3. **Root/sudo access**: For system-level installations

### Optional (for Claude Code)
- Anthropic API key: Get from https://console.anthropic.com

## Installation Methods

### Method 1: One-Line Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/LL4nc33/doom-coding/main/scripts/install.sh | bash
```

### Method 2: Clone and Install

```bash
# Clone repository
git clone https://github.com/LL4nc33/doom-coding.git
# Or via SSH:
git clone git@github.com:LL4nc33/doom-coding.git

cd doom-coding

# Run installer
./scripts/install.sh
```

### Method 3: Manual Installation

See [Manual Installation](manual-installation.md) for step-by-step instructions.

## Installation Options

### Interactive Mode (Default)

```bash
./scripts/install.sh
```

The installer will prompt you for:
- Tailscale authentication key
- code-server password
- User ID (PUID/PGID)
- Component selection

### Unattended Mode

For automated deployments:

```bash
./scripts/install.sh --unattended \
  --tailscale-key="tskey-auth-xxx" \
  --code-password="your-password" \
  --anthropic-key="sk-ant-xxx"
```

### Selective Installation

Skip specific components:

```bash
# Skip Docker (if already installed)
./scripts/install.sh --skip-docker

# Skip terminal tools
./scripts/install.sh --skip-terminal

# Skip SSH hardening
./scripts/install.sh --skip-hardening

# Skip secrets management
./scripts/install.sh --skip-secrets

# Combine flags
./scripts/install.sh --skip-docker --skip-terminal
```

## Post-Installation Steps

### 1. Configure Environment

Edit the `.env` file with your settings:

```bash
cp .env.example .env
vim .env
```

### 2. Set Up Secrets

Create and encrypt your secrets:

```bash
# Initialize secrets management
./scripts/setup-secrets.sh init

# Create secrets template
./scripts/setup-secrets.sh template

# Edit secrets
vim secrets/secrets.yaml

# Encrypt secrets
./scripts/setup-secrets.sh encrypt secrets/secrets.yaml

# Export for Docker
./scripts/setup-secrets.sh export
```

### 3. Start Services

```bash
docker compose up -d
```

### 4. Verify Installation

```bash
./scripts/health-check.sh
```

### 5. Access Your Environment

1. Get your Tailscale IP:
   ```bash
   tailscale status
   ```

2. Access code-server:
   ```
   https://YOUR-TAILSCALE-IP:8443
   ```

3. SSH access:
   ```bash
   ssh user@YOUR-TAILSCALE-IP
   ```

## Troubleshooting Installation

### Docker Permission Denied

```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Log out and back in, or run:
newgrp docker
```

### Tailscale Not Connecting

```bash
# Check Tailscale status
tailscale status

# Re-authenticate
sudo tailscale up --auth-key=YOUR_KEY
```

### code-server Not Accessible

```bash
# Check container logs
docker compose logs code-server

# Verify port is open
docker compose ps
```

### Installation Log

All installation steps are logged:

```bash
tail -f /var/log/doom-coding-install.log
```

## Uninstallation

To completely remove Doom Coding:

```bash
# Stop and remove containers
docker compose down -v

# Remove Docker images
docker rmi $(docker images -q doom-*)

# Remove configuration
rm -rf ~/.doom-coding
rm -rf /etc/ssh/sshd_config.d/99-doom-hardening.conf

# Restore SSH config
sudo systemctl restart sshd
```

## Next Steps

- [Configuration Guide](../configuration/basic-setup.md)
- [Security Setup](../security/hardening.md)
- [Terminal Customization](../terminal/customization.md)