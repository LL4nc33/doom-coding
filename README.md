# 🌲 Doom Coding

> A remote development environment with Tailscale networking, code-server, and Claude Code integration.

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
- 🤖 **AI Integration**: Claude Code for intelligent assistance
- 🛠️ **Complete Toolchain**: zsh, tmux, Node.js, Python, and more
- 🔐 **Hardened Security**: SSH hardening, encrypted secrets, container isolation

## 🎯 Features

| Feature | Description |
|---------|-------------|
| **One-Click Install** | Automated setup for Ubuntu, Debian, and Arch Linux |
| **Tailscale Integration** | Secure mesh networking without port forwarding |
| **code-server** | Full VS Code experience accessible from anywhere |
| **Claude Code** | AI-powered development assistance |
| **Terminal Tools** | Pre-configured zsh, tmux, and development tools |
| **Secrets Management** | SOPS/age encryption for sensitive configuration |
| **SSH Hardening** | Modern security configurations and best practices |
| **Health Monitoring** | Automated health checks and service monitoring |

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

## 🚀 Getting Started

### Prerequisites
- Linux server (Ubuntu 20.04+, Debian 11+, or Arch)
- Root or sudo access
- Internet connection
- [Tailscale account](https://tailscale.com) (free)

### Installation Options

**Interactive Installation:**
```bash
git clone https://github.com/LL4nc33/doom-coding.git
# Or via SSH:
git clone git@github.com:LL4nc33/doom-coding.git
cd doom-coding
./scripts/install.sh
```

**Unattended Installation:**
```bash
./scripts/install.sh --unattended \
  --tailscale-key="tskey-auth-xxx" \
  --code-password="your-secure-password" \
  --anthropic-key="sk-ant-xxx"
```

**Access Your Environment:**
1. Get your Tailscale IP: `tailscale status`
2. Open code-server: `https://YOUR-TAILSCALE-IP:8443`
3. SSH access: `ssh user@YOUR-TAILSCALE-IP`

## 📖 Documentation

Complete documentation is available in the [`docs/`](docs/) directory:

- **[Quick Start Guide](docs/installation/quick-start.md)** - Get running in 5 minutes
- **[Installation Guide](docs/installation/)** - Detailed setup procedures
- **[Configuration Reference](docs/configuration/)** - All configuration options
- **[Security Guide](docs/security/)** - Security features and best practices
- **[Troubleshooting](docs/troubleshooting/)** - Common issues and solutions
- **[Advanced Topics](docs/advanced/)** - Power user features and customizations

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

Monitor your environment health:

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

Built with these excellent open-source projects:
- [Tailscale](https://tailscale.com) - Secure networking
- [code-server](https://github.com/coder/code-server) - VS Code in the browser
- [Claude Code](https://claude.ai/claude-code) - AI development assistance
- [LinuxServer.io](https://www.linuxserver.io/) - Quality container images

---

<p align="center">
  <strong>Happy Coding!</strong><br>
  <em>Built with Forest Green (#2E521D) and determination</em>
</p>