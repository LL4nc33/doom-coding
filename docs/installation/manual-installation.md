# Manual Installation Guide

Step-by-step manual installation instructions for Doom Coding when automatic installation isn't suitable.

## Prerequisites

- Linux server (Ubuntu 20.04+, Debian 11+, or Arch)
- Root or sudo access
- Internet connection
- Git installed

## Step 1: Clone Repository

```bash
git clone https://github.com/LL4nc33/doom-coding.git
cd doom-coding
```

## Step 2: Environment Setup

```bash
# Copy environment template
cp .env.example .env

# Edit with your settings
vim .env
```

Required environment variables:
- `TS_AUTHKEY`: Get from https://login.tailscale.com/admin/settings/keys
- `CODE_SERVER_PASSWORD`: Set a strong password
- `ANTHROPIC_API_KEY`: Get from https://console.anthropic.com (optional)

## Step 3: Install Docker

### Ubuntu/Debian
```bash
# Remove old versions
sudo apt-get remove docker docker-engine docker.io containerd runc

# Add Docker GPG key
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Add repository
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Start and enable
sudo systemctl start docker
sudo systemctl enable docker

# Add user to docker group
sudo usermod -aG docker $USER
```

### Arch Linux
```bash
sudo pacman -S docker docker-compose
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER
```

## Step 4: Install Tailscale (Optional)

Skip this step if using `docker-compose.lxc.yml` for local network access.

```bash
curl -fsSL https://tailscale.com/install.sh | sh
```

## Step 5: Set Up Secrets

```bash
# Create secrets directory
mkdir -p secrets

# Add your Anthropic API key
echo "sk-ant-your-api-key-here" > secrets/anthropic_api_key.txt

# Set permissions
chmod 600 secrets/anthropic_api_key.txt
```

## Step 6: SSH Hardening (Optional)

```bash
./scripts/setup-host.sh
```

This will:
- Apply SSH hardening configuration
- Configure fail2ban
- Set up automatic security updates

## Step 7: Terminal Tools (Optional)

```bash
./scripts/setup-terminal.sh
```

Installs and configures:
- zsh with Oh My Zsh
- tmux with custom configuration
- Modern terminal tools (exa, bat, fzf, ripgrep)

## Step 8: Start Services

### Standard Deployment (with Tailscale)
```bash
docker compose up -d
```

### LXC Deployment (local network)
```bash
docker compose -f docker-compose.lxc.yml up -d
```

## Step 9: Connect Tailscale

Only needed for standard deployment:

```bash
sudo tailscale up --authkey=$TS_AUTHKEY
```

## Step 10: Verify Installation

```bash
./scripts/health-check.sh
```

## Access Your Environment

### With Tailscale
1. Get your Tailscale IP: `tailscale status`
2. Open browser: `https://TAILSCALE-IP:8443`

### Local Network (LXC)
1. Get your local IP: `hostname -I`
2. Open browser: `https://LOCAL-IP:8443`

## Troubleshooting

### Docker Permission Denied
```bash
# Log out and back in, or:
newgrp docker
```

### Tailscale Not Connecting
```bash
sudo tailscale status
sudo tailscale up --authkey=YOUR-KEY
```

### code-server Not Accessible
```bash
# Check container logs
docker compose logs code-server

# Verify ports
docker compose ps
```

### SSL Certificate Warnings
Self-signed certificates will show browser warnings. Click "Advanced" → "Proceed to site".

## Advanced Configuration

### Custom SSL Certificates

Replace self-signed certificates:
```bash
# Place your certificates in:
# - ./ssl/server.crt
# - ./ssl/server.key
# Then restart:
docker compose restart
```

### Resource Limits

Edit docker-compose.yml to add resource limits:
```yaml
services:
  code-server:
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '1.0'
```

### Custom code-server Extensions

Mount a custom extensions directory:
```yaml
volumes:
  - ./extensions:/config/extensions
```

## Uninstallation

```bash
# Stop containers
docker compose down

# Remove images
docker rmi $(docker images -q doom-*)

# Remove volumes (optional)
docker volume prune

# Remove Tailscale
sudo tailscale logout
```

## Next Steps

- [Configuration Reference](../configuration/basic-setup.md)
- [Security Guide](../security/hardening.md)
- [Troubleshooting](../troubleshooting/)