# ⚡ Quick Start Guide

Get your Doom Coding environment running in under 5 minutes!

## 🎯 Prerequisites

Before you begin, ensure you have:
- [ ] Linux server (Ubuntu 20.04+, Debian 11+, or Arch)
- [ ] Root or sudo access
- [ ] Internet connection
- [ ] [Tailscale account](https://tailscale.com) (free tier works)

## 🚀 One-Line Install

```bash
curl -fsSL https://raw.githubusercontent.com/LL4nc33/doom-coding/main/scripts/install.sh | bash
```

**Or download and inspect first** (recommended):
```bash
wget https://raw.githubusercontent.com/LL4nc33/doom-coding/main/scripts/install.sh
chmod +x install.sh
./install.sh
```

## 🎮 Interactive Installation

The installer will guide you through:

1. **System Detection**
   ```
   🔍 Detecting system...
   ✅ Ubuntu 22.04 LTS (amd64) detected
   ✅ Docker available
   ```

2. **Component Selection**
   ```
   📦 Select components to install:
   [✓] Docker & Docker Compose
   [✓] Tailscale integration
   [✓] code-server
   [✓] Claude Code
   [✓] Terminal tools (zsh, tmux, etc.)
   [✓] SSH hardening
   ```

3. **Configuration**
   ```
   🔧 Configure your environment:
   Tailscale auth key: [paste your key]
   code-server password: [enter secure password]
   User ID (PUID): 1000 [auto-detected]
   ```

4. **Installation Progress**
   ```
   ⏳ Installing Docker...
   ⏳ Setting up Tailscale...
   ⏳ Configuring code-server...
   ⏳ Installing Claude Code...
   ⏳ Setting up terminal tools...
   ```

## 🎯 Unattended Installation

For automated deployments:

```bash
./install.sh --unattended \
  --tailscale-key="tskey-auth-xxx" \
  --code-password="your-secure-password" \
  --anthropic-key="sk-ant-xxx"
```

### Environment File Method

1. Create configuration:
   ```bash
   cat > .env << EOF
   TS_AUTHKEY=tskey-auth-xxx
   CODE_SERVER_PASSWORD=your-secure-password
   ANTHROPIC_API_KEY=sk-ant-xxx
   PUID=1000
   PGID=1000
   TZ=Europe/Berlin
   EOF
   ```

2. Run unattended install:
   ```bash
   ./install.sh --unattended --env-file=.env
   ```

## 📍 Find Your Services

After installation, get your Tailscale IP:

```bash
tailscale status
# Example output:
# 100.64.1.2   doom-coding         LL4nc33@ linux
```

Access your services:
- **code-server**: `https://100.64.1.2:8443`
- **SSH**: `ssh user@100.64.1.2`

## ✅ Verify Installation

### Basic Health Check
Run the health check to verify core functionality:

```bash
./scripts/health-check.sh
```

Expected output:
```
🏥 Doom Coding Health Check
===========================

✅ Docker: Running (v24.0.7)
✅ Tailscale: Connected (100.64.1.2)
✅ code-server: Accessible (https://100.64.1.2:8443)
✅ Claude Code: Available (v0.8.3)
✅ SSH: Hardened and accessible
✅ Terminal: zsh, tmux, and tools ready

🎉 All systems operational!
```

### Additional Testing
For comprehensive validation, run our testing suite:

```bash
# Quick smoke test (5 minutes)
./scripts/test-runner.sh --smoke-test

# Security validation
./scripts/test-runner.sh --category=security

# Full deployment validation
./scripts/test-runner.sh --deployment=docker-tailscale --iterations=1-10
```

**Testing Documentation**: See [`/docs/testing/`](../testing/) for complete 70-iteration testing framework.

## 🐛 Quick Troubleshooting

### Installation Fails

1. **Check logs**:
   ```bash
   tail -f /var/log/doom-coding-install.log
   ```

2. **Retry specific step**:
   ```bash
   ./install.sh --retry-failed
   ```

### Can't Access Services

1. **Check Tailscale connection**:
   ```bash
   tailscale status
   ```

2. **Verify containers are running**:
   ```bash
   docker compose ps
   ```

3. **Check service logs**:
   ```bash
   docker compose logs code-server
   ```

### Permission Issues

1. **Fix ownership**:
   ```bash
   sudo chown -R $(id -u):$(id -g) /path/to/doom-coding
   ```

2. **Restart with correct PUID/PGID**:
   ```bash
   docker compose down
   docker compose up -d
   ```

## 🎯 Next Steps

Once everything is running:

1. **[Run comprehensive testing](../testing/)** - Validate your installation with our 70-iteration framework
2. **[Configure your environment](../configuration/basic-setup.md)**
3. **[Customize your terminal](../terminal/customization.md)**
4. **[Set up backups](../advanced/backup-recovery.md)**
5. **[Learn advanced features](../advanced/)**

## 🔧 Installation Options

| Flag | Description | Example |
|------|-------------|---------|
| `--unattended` | No interactive prompts | `./install.sh --unattended` |
| `--skip-docker` | Skip Docker installation | `./install.sh --skip-docker` |
| `--skip-terminal` | Skip terminal tools | `./install.sh --skip-terminal` |
| `--skip-hardening` | Skip SSH hardening | `./install.sh --skip-hardening` |
| `--env-file=FILE` | Use environment file | `./install.sh --env-file=.env` |
| `--retry-failed` | Retry failed installation steps | `./install.sh --retry-failed` |

---

**Need more help?** Check the [detailed installation guide](installation-guide.md) or [troubleshooting section](../troubleshooting/).