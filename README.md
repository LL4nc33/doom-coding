# 🌲 Doom Coding

> A remote development environment with Tailscale networking, code-server, and Claude Code integration.
>
> **Version 0.0.6a** - Now with QR code generation for instant mobile access!

<p align="center">
  <img src="logo/favicon.png" width="128" height="128" alt="Doom Coding Logo">
</p>

<p align="center">
  <strong>Secure • Portable • AI-Powered</strong>
</p>

---

## ⚡ Quick Start

Get your development environment running in under 5 minutes:

```bash
curl -fsSL https://raw.githubusercontent.com/LL4nc33/doom-coding/main/scripts/install.sh | bash
```

**What you get:**
- 🔒 **Secure Access**: Tailscale mesh VPN with zero-config networking
- 💻 **Web IDE**: Full VS Code experience in your browser
- 📱 **Mobile Ready**: QR code generation for instant smartphone access
- 🤖 **AI Integration**: Claude Code for intelligent assistance
- 🛠️ **Complete Toolchain**: zsh, tmux, Node.js, Python, and more
- 🔐 **Hardened Security**: SSH hardening, encrypted secrets, container isolation

## 🎯 Features

| Feature | Description |
|---------|-------------|
| **One-Click Install** | Automated setup for Ubuntu, Debian, and Arch Linux |
| **QR Code Integration** | Generate QR codes for instant mobile access |
| **Tailscale Integration** | Secure mesh networking without port forwarding |
| **code-server** | Full VS Code experience accessible from anywhere |
| **Mobile-First Design** | Optimized for smartphone and tablet development |
| **Claude Code** | AI-powered development assistance |
| **Smart Port Detection** | Automatic port conflict detection and resolution |
| **Terminal Tools** | Pre-configured zsh, tmux, and development tools |
| **Terminal Environment** | Lightweight ttyd-based alternative (~200MB RAM) |
| **Secrets Management** | SOPS/age encryption for sensitive configuration |
| **SSH Hardening** | Modern security configurations and best practices |
| **Health Monitoring** | Automated health checks and service monitoring |
| **LXC Support** | Run in Proxmox LXC containers without TUN device |
| **Flexible Networking** | Choose between Tailscale VPN or local network access |

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Docker Host                              │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                  Docker Network                          │   │
│  │  ┌──────────┐  ┌─────────────┐  ┌─────────────┐       │   │
│  │  │          │  │ code-server │  │ Claude Code │       │   │
│  │  │ Tailscale│◄─┤ network_mode│◄─┤ network_mode│       │   │
│  │  │ (sidecar)│  │  :service   │  │  :service   │       │   │
│  │  │          │  │             │  │             │       │   │
│  │  └─────┬────┘  └─────────────┘  └─────────────┘       │   │
│  │        │                                               │   │
│  │        ▼                                               │   │
│  │   Tailscale Network (100.x.x.x)                       │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
                  ┌─────────────────┐
                  │  Your Devices   │
                  │ (via Tailscale) │
                  └─────────────────┘
```

## 🎯 Deployment Options

### Option 1: Native Tailscale Userspace (LXC) ⭐ EMPFOHLEN
Tailscale wird **direkt auf dem LXC-Host** installiert (nicht in Docker).
- **Best for**: Proxmox LXC Container - niedrigster Ressourcenverbrauch!
- **Requirements**: Docker only (kein TUN-Device erforderlich!)
- **Access**: Via Tailscale IP (https://100.x.x.x/)
- **Compose File**: `docker-compose.native-userspace.yml`
- **CLI**: `--native-userspace`

### Option 2: Docker Tailscale Userspace (LXC)
Tailscale in LXC-Containern **ohne TUN-Device** - verwendet Docker Container.
- **Best for**: Proxmox LXC Container mit vollständiger Container-Isolation
- **Requirements**: Docker only (kein TUN-Device erforderlich!)
- **Access**: Via Tailscale IP (100.x.x.x:8443)
- **Compose File**: `docker-compose.lxc-tailscale.yml`

### Option 3: Docker + Tailscale (Standard)
Full-featured deployment with VS Code in browser and secure Tailscale networking.
- **Best for**: Bare-metal Server, VMs mit TUN-Device
- **Requirements**: Docker, TUN device for Tailscale
- **Access**: Via Tailscale IP (100.x.x.x)

### Option 4: Host Tailscale verwenden
Nutzt bereits installiertes Tailscale auf dem Host.
- **Best for**: Systeme mit vorkonfiguriertem Tailscale
- **Requirements**: Tailscale muss bereits auf dem Host installiert sein
- **Access**: Via Host Tailscale IP (100.x.x.x:8443)
- **CLI**: `--native-tailscale`

### Option 5: Docker + Local Network (LXC)
Docker deployment ohne Tailscale - Zugriff via lokales Netzwerk.
- **Best for**: LXC containers, home lab setups ohne VPN
- **Requirements**: Docker only
- **Access**: Via local IP (192.168.x.x:8443)
- **CLI**: `--skip-tailscale` oder `--local-network`

### Option 6: Terminal Environment (Lightweight)
Bare-metal ttyd + tmux + neovim setup without Docker.
- **Best for**: Resource-constrained systems, mobile access
- **Requirements**: systemd, ~200MB RAM
- **Documentation**: [Terminal Dev Environment](terminal-dev-env/docs/README.md)

## 🚀 Getting Started

### Prerequisites
- Linux server (Ubuntu 20.04+, Debian 11+, or Arch)
- Root or sudo access
- Internet connection
- [Tailscale account](https://tailscale.com) (free)

### Installation Options

**One-Line Install (Recommended):**
```bash
curl -fsSL https://raw.githubusercontent.com/LL4nc33/doom-coding/main/scripts/install.sh | bash
```

**Interactive TUI Wizard (Recommended):**
```bash
git clone https://github.com/LL4nc33/doom-coding.git
cd doom-coding
make build-tui && ./bin/doom-tui
```

The TUI provides a visual, guided installation experience with:
- System detection and compatibility checks
- Deployment mode selection (Tailscale/Local/Terminal-only)
- Component selection with checkboxes
- Configuration input with validation
- Real-time installation progress
- Health check verification

See [TUI Documentation](docs/TUI.md) for details.

**CLI Installation:**
```bash
git clone https://github.com/LL4nc33/doom-coding.git
cd doom-coding
./scripts/install.sh
```

**LXC / Local Network Installation (ohne Tailscale):**
```bash
./scripts/install.sh --skip-tailscale
# oder
./scripts/install.sh --local-network
```

**Terminal Environment (Lightweight):**
```bash
git clone https://github.com/LL4nc33/doom-coding.git
cd doom-coding/terminal-dev-env
sudo bash bin/install.sh
```

**Unattended Installation:**
```bash
./scripts/install.sh --unattended \
  --tailscale-key="tskey-auth-xxx" \
  --code-password="your-secure-password" \
  --anthropic-key="sk-ant-xxx"
```

### CLI Options

| Option | Beschreibung |
|--------|--------------|
| `--unattended` | Vollautomatische Installation |
| `--native-userspace` | Native Tailscale Userspace auf LXC Host (empfohlen) |
| `--native-tailscale` | Vorhandenes Host-Tailscale verwenden |
| `--skip-tailscale` | Ohne Tailscale (lokales Netzwerk) |
| `--local-network` | Alias für --skip-tailscale |
| `--skip-docker` | Docker-Installation überspringen |
| `--skip-terminal` | Terminal-Tools überspringen |
| `--skip-hardening` | SSH-Hardening überspringen |
| `--env-file=FILE` | Eigene Umgebungsdatei verwenden |
| `--dry-run` | Nur anzeigen, nichts ausführen |
| `--force` | Neuinstallation erzwingen |

### Docker Compose Variants

| Datei | Verwendung |
|-------|------------|
| `docker-compose.native-userspace.yml` | **Native LXC Tailscale (EMPFOHLEN)** - Tailscale auf Host |
| `docker-compose.lxc-tailscale.yml` | Docker Tailscale Userspace (kein TUN!) |
| `docker-compose.native-tailscale.yml` | Host-Tailscale verwenden (vorkonfiguriert) |
| `docker-compose.yml` | Standard mit Tailscale (TUN erforderlich) |
| `docker-compose.lxc.yml` | LXC ohne Tailscale (nur lokales Netzwerk) |

```bash
# Native Tailscale Userspace (EMPFOHLEN für LXC)
./scripts/install.sh --native-userspace
# oder manuell:
docker compose -f docker-compose.native-userspace.yml up -d

# Docker Tailscale Userspace Mode
docker compose -f docker-compose.lxc-tailscale.yml up -d

# Standard (mit Tailscale, TUN erforderlich)
docker compose up -d

# LXC ohne Tailscale (nur lokales Netzwerk)
docker compose -f docker-compose.lxc.yml up -d
```

## 📱 Mobile Access & QR Codes

Get instant access to your development environment from any smartphone or tablet:

**Quick Mobile Setup:**
```bash
# Generate QR code for immediate access
./scripts/health-check.sh --qr
```

**What you get:**
- 📱 **Instant Access**: Scan QR code with your phone's camera
- 🔗 **Direct Links**: code-server, documentation, and service setup
- 📚 **Mobile Guide**: Complete smartphone setup documentation
- 🛠️ **Cross-Platform**: Works on Android and iOS devices

**Popular Mobile Apps:**
- **Android**: Termux, JuiceSSH for terminal access
- **iOS**: Blink Shell, Termius for SSH connections

**Complete mobile setup guide**: [`docs/mobile/smartphone-setup.md`](/config/repos/doom-coding/docs/mobile/smartphone-setup.md)

**Access Your Environment:**

*With Tailscale:*
1. Get your Tailscale IP: `tailscale status`
2. Open code-server: `https://YOUR-TAILSCALE-IP:8443`
3. SSH access: `ssh user@YOUR-TAILSCALE-IP`

*Local Network (LXC):*
1. Find your local IP: `hostname -I`
2. Open code-server: `https://YOUR-LOCAL-IP:8443`
3. Terminal Environment: `https://YOUR-LOCAL-IP/` (if using ttyd)

## 📖 Documentation

Complete documentation is available in the [`docs/`](docs/) directory:

- **[Quick Start Guide](docs/installation/quick-start.md)** - Get running in 5 minutes
- **[Mobile Setup Guide](docs/mobile/smartphone-setup.md)** - Complete smartphone and tablet setup
- **[Installation Guide](docs/installation/)** - Detailed setup procedures
- **[Testing Framework](docs/testing/)** - Comprehensive 70-iteration testing strategy
- **[Configuration Reference](docs/configuration/)** - All configuration options
- **[Security Guide](docs/security/)** - Security features and best practices
- **[Troubleshooting](docs/troubleshooting/)** - Common issues and solutions
- **[Advanced Topics](docs/advanced/)** - Power user features and customizations

## 🧪 Testing & Quality Assurance

**Comprehensive 70-Iteration Testing Framework**

We've implemented a rigorous testing strategy covering 5 deployment scenarios across 70 structured test iterations:

### Quick Testing Commands
```bash
# Basic health check with QR code
./scripts/health-check.sh --qr

# Test QR code integration
./scripts/test-qr-integration.sh

# Run security tests
./scripts/test-runner.sh --category=security

# Full deployment validation
./scripts/test-runner.sh --deployment=docker-tailscale
```

### Testing Coverage
| Phase | Focus | Iterations | Status |
|-------|-------|------------|---------|
| **Foundation** | Core functionality, deployment | 1-20 | ✅ |
| **Security** | Hardening, vulnerability assessment | 21-35 | 🔄 |
| **CI/CD** | Automation, deployment pipelines | 36-50 | 📋 |
| **Integration** | Cross-platform, edge cases | 51-60 | 📋 |
| **UX/Docs** | User experience, documentation | 61-70 | 📋 |

**Complete Documentation**: [`docs/testing/`](docs/testing/)

## 🔧 Configuration

### Basic Setup
```bash
# Copy environment template
cp .env.example .env

# Edit configuration
vim .env
```

### Key Configuration Options
- `TS_AUTHKEY`: Tailscale authentication key
- `CODE_SERVER_PASSWORD`: Web IDE password
- `ANTHROPIC_API_KEY`: Claude API key
- `PUID`/`PGID`: User permissions

### Secrets Management
```bash
# Setup encryption
./scripts/setup-secrets.sh generate-key

# Manage secrets
./scripts/setup-secrets.sh encrypt secrets.yaml
```

## 🏥 Health Monitoring

Monitor your environment health and generate access QR codes:

```bash
# Standard health check
./scripts/health-check.sh

# Health check with QR code for mobile access
./scripts/health-check.sh --qr
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

## 🛡️ Security

- **Network Security**: Tailscale mesh VPN with WireGuard encryption
- **SSH Hardening**: Modern ciphers, key-only auth, fail2ban ready
- **Container Security**: Non-root users, minimal privileges
- **Secrets Management**: SOPS/age encryption for sensitive data
- **Regular Updates**: Automated security updates available

## 🤝 Contributing

We welcome contributions! Please see our [contributing guidelines](docs/contributing/) for details.

1. Fork the repository
2. Create your feature branch
3. Add tests and documentation
4. Submit a pull request

## 📋 Support

- **Documentation**: Check the [`docs/`](docs/) directory
- **Issues**: Report bugs via GitHub Issues
- **Discussions**: Join GitHub Discussions for questions
- **Security**: Report security issues privately via email

## 📄 License

Choose your license:
- **[AGPLv3](LICENSE-AGPLv3)** - Copyleft license ensuring derivative works remain open
- **[Apache 2.0](LICENSE-Apache2.0)** - Permissive license allowing commercial use

## 🙏 Acknowledgments

**Inspired by:**
- [rberg27/doom-coding](https://github.com/rberg27/doom-coding) - Original concept and inspiration

**Built with these excellent open-source projects:**
- [Tailscale](https://tailscale.com) - Secure networking
- [code-server](https://github.com/coder/code-server) - VS Code in the browser
- [Claude Code](https://claude.ai/claude-code) - AI development assistance
- [go-qrcode](https://github.com/skip2/go-qrcode) - QR code generation library
- [LinuxServer.io](https://www.linuxserver.io/) - Quality container images

## 🤖 AI Development

This project was developed entirely by AI agents using a multi-agent orchestration approach.

**Development Stack:**
- **Claude Opus 4.5** - Strategic orchestration, architecture decisions, and project coordination
- **Claude Sonnet 4** - Implementation, code generation, testing, and comprehensive documentation

**Methodology:**
- Multi-agent workflow with specialized roles (architect, researcher, security, devops, implementer)
- Iterative development with continuous refinement and validation
- Comprehensive testing across multiple platforms and deployment scenarios
- Human oversight for direction, requirements, and final approval

This represents a demonstration of AI-assisted software development at scale, where AI agents handled all aspects of design, implementation, security auditing, testing, and documentation while maintaining professional software engineering standards and best practices.

---

<p align="center">
  <strong>Happy Coding!</strong><br>
  <em>Built with Forest Green (#2E521D) and determination</em>
</p>