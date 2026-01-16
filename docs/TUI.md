# Doom Coding TUI - Interactive Setup Wizard

## Overview

The Doom Coding TUI (Text User Interface) provides an interactive, visual way to configure and install the Doom Coding development environment. Built with Go and the Bubble Tea framework, it offers a modern terminal-based experience while preserving full compatibility with the existing bash installation scripts.

## Features

- **Visual guided setup** - Step-by-step wizard with clear navigation
- **System detection** - Automatically detects OS, architecture, container type, and TUN availability
- **Deployment modes** - Choose between Docker+Tailscale, Docker+Local Network, or Terminal-only
- **Component selection** - Pick which components to install with checkbox interface
- **Input validation** - Real-time validation of credentials and configuration
- **Progress tracking** - Live installation progress with real-time output
- **Health checks** - Post-installation verification of all services
- **CLI automation** - Full CLI mode for scripted deployments

## Installation

### Building from Source

```bash
# Navigate to the TUI directory
cd cmd/doom-tui

# Download dependencies
go mod download

# Build the binary
go build -o doom-tui .

# Optional: Install globally
sudo mv doom-tui /usr/local/bin/
```

### Quick Build (with Make)

```bash
make build-tui
```

## Usage

### Interactive Mode (Default)

```bash
# Run the interactive TUI wizard
./doom-tui

# Or if installed globally
doom-tui
```

### CLI Mode (Automation)

```bash
# Fully automated installation
./doom-tui cli --unattended \
  --tailscale-key="tskey-auth-xxxxx" \
  --code-password="secure-password" \
  --anthropic-key="sk-ant-api03-xxxxx"

# Skip specific components
./doom-tui cli --unattended \
  --skip-tailscale \
  --skip-hardening

# Dry run (show what would be executed)
./doom-tui cli --dry-run --show-commands

# Local network mode (for LXC without TUN)
./doom-tui cli --unattended --skip-tailscale \
  --code-password="password"
```

### Status Check

```bash
# Check installation health
./doom-tui status
```

## CLI Flags

| Flag | Description |
|------|-------------|
| `--unattended` | Run without prompts (requires credentials) |
| `--tailscale-key=KEY` | Tailscale auth key |
| `--code-password=PWD` | code-server password |
| `--anthropic-key=KEY` | Anthropic API key |
| `--sudo-password=PWD` | Container sudo password |
| `--config=FILE` | Load config from JSON file |
| `--dry-run` | Show commands without executing |
| `--show-commands` | Display equivalent bash command |
| `--skip-docker` | Skip Docker installation |
| `--skip-tailscale` | Skip Tailscale (use local network) |
| `--skip-terminal` | Skip terminal tools setup |
| `--skip-hardening` | Skip SSH hardening |
| `--skip-secrets` | Skip secrets management |
| `--verbose` | Enable verbose output |

## Screen Flow

```
┌─────────────────┐
│    Welcome      │
│     Screen      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    System       │
│   Detection     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Deployment    │
│  Mode Selection │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Component     │
│   Selection     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Configuration  │
│     Input       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Preview &    │
│   Confirmation  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Installation   │
│    Progress     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Results &    │
│   Health Check  │
└─────────────────┘
```

## Keyboard Navigation

### Global Keys
| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit application |
| `Esc` | Go back to previous screen |

### Welcome Screen
| Key | Action |
|-----|--------|
| `Enter` | Continue to system detection |
| `h` | Show help |

### Detection Screen
| Key | Action |
|-----|--------|
| `Enter` | Continue with detected settings |
| `r` | Re-run detection |

### Selection Screens
| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `Space` | Toggle selection |
| `Enter` | Confirm and continue |
| `a` | Select all (checkboxes) |
| `n` | Select none (checkboxes) |

### Configuration Screen
| Key | Action |
|-----|--------|
| `Tab` / `↓` | Next field |
| `Shift+Tab` / `↑` | Previous field |
| `Enter` | Submit form / Next field |
| `Ctrl+V` | Toggle password visibility |

### Preview Screen
| Key | Action |
|-----|--------|
| `Enter` / `i` | Start installation |
| `e` | Export configuration |
| `s` | Show bash command |

### Results Screen
| Key | Action |
|-----|--------|
| `Enter` / `q` | Exit |
| `r` | Re-run health check |
| `l` | View logs |

## Deployment Modes

### 1. Docker + Tailscale (Recommended)
- Full deployment with secure VPN access
- Requires TUN device (not available in all LXC containers)
- Access code-server from anywhere via Tailscale

### 2. Docker + Local Network
- Containers accessible on local network only
- Best for LXC containers without TUN device
- Port 8443 exposed directly on host

### 3. Terminal Tools Only
- Minimal installation (~200MB RAM)
- Installs zsh, tmux, Oh My Zsh, nvm, pyenv
- No Docker containers

## Components

| Component | Description |
|-----------|-------------|
| **Docker** | Container runtime for services |
| **Tailscale** | Secure VPN mesh network |
| **Terminal Tools** | zsh, tmux, Oh My Zsh, nvm, pyenv |
| **SSH Hardening** | Security configuration, fail2ban |
| **Secrets Management** | SOPS/age encryption |

## Configuration

### Configuration File Format (JSON)

```json
{
  "deployment_mode": "tailscale",
  "components": {
    "docker": true,
    "tailscale": true,
    "terminal_tools": true,
    "ssh_hardening": true,
    "secrets_manager": true
  },
  "credentials": {
    "tailscale_key": "tskey-auth-xxxxx",
    "code_password": "secure-password",
    "sudo_password": "sudo-password",
    "anthropic_key": "sk-ant-api03-xxxxx"
  },
  "environment": {
    "puid": "1000",
    "pgid": "1000",
    "timezone": "Europe/Berlin",
    "workspace_path": "./workspace"
  },
  "advanced": {
    "code_server_port": 8443,
    "ts_accept_dns": false
  }
}
```

### Load Configuration from File

```bash
./doom-tui cli --config=my-config.json --unattended
```

## Integration with Existing Scripts

The TUI acts as a visual frontend that generates configuration and calls the existing bash scripts:

1. **install.sh** - Main installation orchestration
2. **setup-terminal.sh** - Terminal tools setup
3. **setup-host.sh** - SSH hardening
4. **setup-secrets.sh** - SOPS/age configuration
5. **health-check.sh** - Service verification

The TUI:
- Generates the `.env` file based on user input
- Writes secrets to `secrets/anthropic_api_key.txt`
- Calls `install.sh --unattended` with appropriate flags
- Parses output for progress display
- Runs health checks after completion

## Architecture

```
doom-tui/
├── cmd/doom-tui/
│   ├── main.go          # Entry point, CLI parsing
│   ├── model.go         # Main TUI model and state
│   └── view.go          # Screen rendering
├── internal/
│   ├── config/          # Configuration management
│   │   └── config.go    # Config struct, .env generation
│   ├── system/          # System detection
│   │   └── detect.go    # OS, arch, container detection
│   └── executor/        # Script execution
│       └── executor.go  # Script orchestration, health checks
└── tui/
    ├── components/      # Reusable UI components
    │   ├── checkbox.go  # Checkbox group
    │   ├── radio.go     # Radio button group
    │   ├── form.go      # Form with validation
    │   └── progress.go  # Progress indicator
    ├── screens/         # Individual screens
    │   ├── welcome.go   # Welcome screen
    │   ├── detection.go # System detection
    │   ├── deployment.go# Mode selection
    │   ├── components.go# Component selection
    │   ├── config.go    # Configuration input
    │   ├── preview.go   # Preview & confirmation
    │   └── results.go   # Results & health check
    └── styles/          # Styling definitions
        └── styles.go    # Colors, typography
```

## Troubleshooting

### TUI doesn't start
- Ensure terminal supports 256 colors
- Try: `TERM=xterm-256color ./doom-tui`

### Detection shows wrong container type
- Check `/proc/1/cgroup` for container hints
- Use `--skip-tailscale` if TUN detection fails

### Installation fails
- Check logs: `/var/log/doom-coding-install.log`
- Run with `--verbose` for detailed output
- Ensure network connectivity

### Tailscale mode unavailable
- TUN device not detected
- For LXC: Enable TUN in container settings
- Or use "Docker + Local Network" mode

## Development

### Prerequisites
- Go 1.22+
- Terminal with 256 color support

### Running in Development
```bash
cd cmd/doom-tui
go run .
```

### Building
```bash
go build -o doom-tui .
```

### Testing
```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## License

MIT License - See LICENSE file for details.
