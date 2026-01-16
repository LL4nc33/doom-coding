# 📜 install.sh Script Reference

The main installer script that orchestrates the complete Doom Coding environment setup.

## 🎯 Overview

`install.sh` is the entry point for setting up your Doom Coding environment. It:
- Detects your operating system and architecture
- Installs dependencies in the correct order
- Handles both interactive and unattended installation
- Provides comprehensive logging and error handling
- Ensures idempotent operation (safe to run multiple times)

## 🚀 Usage

### Basic Usage
```bash
./install.sh                    # Interactive installation
./install.sh --help             # Show help
./install.sh --version          # Show version info
```

### Command Line Options

| Option | Description | Example |
|--------|-------------|---------|
| `--unattended` | Run without prompts | `./install.sh --unattended` |
| `--skip-docker` | Skip Docker installation | `./install.sh --skip-docker` |
| `--skip-terminal` | Skip terminal tools setup | `./install.sh --skip-terminal` |
| `--skip-hardening` | Skip SSH hardening | `./install.sh --skip-hardening` |
| `--skip-secrets` | Skip SOPS/age setup | `./install.sh --skip-secrets` |
| `--env-file=FILE` | Use specific env file | `./install.sh --env-file=prod.env` |
| `--dry-run` | Show what would be done | `./install.sh --dry-run` |
| `--retry-failed` | Retry failed steps | `./install.sh --retry-failed` |
| `--force` | Force reinstallation | `./install.sh --force` |
| `--verbose` | Enable debug output | `./install.sh --verbose` |
| `--log-file=FILE` | Custom log file | `./install.sh --log-file=/tmp/install.log` |

### Unattended Installation Options

```bash
./install.sh --unattended \
  --tailscale-key="tskey-auth-xxx" \
  --code-password="secure-password" \
  --anthropic-key="sk-ant-xxx" \
  --puid=1000 \
  --pgid=1000 \
  --timezone="Europe/Berlin"
```

## 📋 Installation Steps

The installer follows these steps in order:

### 1. System Detection
```bash
# Operating system detection
detect_os() {
    if [[ -f /etc/os-release ]]; then
        source /etc/os-release
        OS_ID="$ID"
        OS_VERSION_ID="$VERSION_ID"
    fi
}

# Architecture detection
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) log_error "Unsupported architecture: $(uname -m)" ;;
    esac
}
```

### 2. Prerequisites Check
- Root/sudo access verification
- Internet connectivity test
- Disk space requirements (minimum 10GB)
- Available ports check (8443, 22)

### 3. Docker Installation
```bash
install_docker() {
    if ! command -v docker &>/dev/null; then
        log_info "Installing Docker..."
        case "$OS_ID" in
            ubuntu|debian)
                install_docker_debian
                ;;
            arch)
                install_docker_arch
                ;;
        esac
    fi
}
```

### 4. Component Installation
1. **Tailscale Setup**
2. **code-server Configuration**
3. **Claude Code Installation** (native, not npm)
4. **Terminal Tools** (optional)
5. **SSH Hardening** (optional)
6. **Secrets Management** (optional)

### 5. Service Startup
- Docker Compose validation
- Container health checks
- Service accessibility verification

## 🔍 Functions Reference

### Core Functions

#### `main()`
Main entry point that orchestrates the installation process.

#### `parse_arguments()`
Parses command line arguments and sets global variables.

#### `setup_logging()`
Configures logging to file and console with appropriate formatting.

#### `check_prerequisites()`
Validates system requirements before installation.

#### `detect_system()`
Detects operating system, architecture, and available package managers.

### Installation Functions

#### `install_docker()`
Installs Docker and Docker Compose based on the detected OS.

```bash
install_docker() {
    local os_id="$1"

    case "$os_id" in
        ubuntu|debian)
            # Add Docker's official GPG key
            curl -fsSL https://download.docker.com/linux/${os_id}/gpg | \
                sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

            # Add Docker repository
            echo "deb [arch=${ARCH} signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] \
                https://download.docker.com/linux/${os_id} $(lsb_release -cs) stable" | \
                sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

            # Install Docker
            sudo apt-get update
            sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
            ;;
        arch)
            sudo pacman -S --noconfirm docker docker-compose
            ;;
    esac
}
```

#### `setup_tailscale()`
Configures Tailscale with OAuth or auth key authentication.

#### `install_claude_code()`
Downloads and installs Claude Code using the official installer.

```bash
install_claude_code() {
    log_info "Installing Claude Code (native)..."

    # Use official installer, NOT npm
    curl -fsSL https://claude.ai/install.sh | bash

    # Verify installation
    if command -v claude &>/dev/null; then
        log_success "Claude Code installed successfully"
        claude --version
    else
        log_error "Claude Code installation failed"
        return 1
    fi
}
```

### Configuration Functions

#### `create_env_file()`
Creates the `.env` file from template with user-provided or default values.

#### `setup_secrets()`
Initializes SOPS/age encryption for secrets management.

#### `configure_docker_compose()`
Validates and customizes docker-compose.yml for the environment.

### Utility Functions

#### `log_info()`, `log_success()`, `log_warning()`, `log_error()`
Logging functions with color-coded output using brand colors.

```bash
log_info() {
    echo -e "${BLUE}ℹ${NC} $*" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}✅${NC} $*" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $*" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}❌${NC} $*" | tee -a "$LOG_FILE"
}
```

#### `prompt_user()`
Interactive user input with validation and default values.

#### `check_service_health()`
Verifies that installed services are running and accessible.

## 🎨 Brand Colors

The installer uses consistent brand colors:

```bash
# Brand color palette
readonly GREEN='\033[38;2;46;82;29m'      # Forest Green #2E521D
readonly BROWN='\033[38;2;124;94;70m'     # Tan Brown #7C5E46
readonly LIGHT_BROWN='\033[38;2;164;125;91m' # Light Brown #A47D5B
readonly DARK_NAVY='\033[38;2;34;32;51m'  # Dark Navy #222033
readonly NC='\033[0m'                      # No Color
```

## 📊 Installation Progress

The installer provides real-time progress feedback:

```
🔍 Detecting system...
✅ Ubuntu 22.04 LTS (amd64) detected

📦 Installing components...
⏳ Installing Docker... (1/6)
✅ Docker installed successfully
⏳ Setting up Tailscale... (2/6)
✅ Tailscale configured
⏳ Installing code-server... (3/6)
✅ code-server ready
⏳ Installing Claude Code... (4/6)
✅ Claude Code installed
⏳ Setting up terminal tools... (5/6)
✅ Terminal tools configured
⏳ Running health checks... (6/6)
✅ All systems operational!

🎉 Installation completed in 3m 42s
```

## 🔧 Error Handling

The installer includes comprehensive error handling:

### Retry Mechanism
Failed steps can be retried individually:
```bash
./install.sh --retry-failed
```

### Rollback on Failure
Critical failures trigger automatic rollback:
```bash
cleanup_on_failure() {
    log_warning "Installation failed, cleaning up..."
    docker compose down 2>/dev/null || true
    sudo systemctl stop tailscaled 2>/dev/null || true
}
```

### Log Analysis
Detailed logs help diagnose issues:
```bash
tail -f /var/log/doom-coding-install.log
```

## 🧪 Testing Mode

Use dry-run mode to see what would be done:
```bash
./install.sh --dry-run

# Output:
# [DRY RUN] Would install Docker
# [DRY RUN] Would configure Tailscale
# [DRY RUN] Would setup code-server
# [DRY RUN] Would install Claude Code
# [DRY RUN] Would configure terminal tools
```

## 🔄 Idempotency

The installer is designed to be run multiple times safely:

```bash
# Safe to run multiple times
./install.sh
./install.sh  # No harmful side effects
./install.sh --force  # Force reinstall everything
```

Each component checks if it's already installed before proceeding.

## 📝 Examples

### Minimal Installation
```bash
./install.sh --skip-terminal --skip-hardening
```

### Developer Setup
```bash
./install.sh --unattended \
  --anthropic-key="$ANTHROPIC_API_KEY" \
  --tailscale-key="$TS_AUTH_KEY" \
  --code-password="dev-password"
```

### Production Setup
```bash
./install.sh --unattended \
  --env-file=production.env \
  --verbose \
  --log-file=/var/log/production-install.log
```

---

**Next:** [setup-terminal.sh Reference](setup-terminal-sh.md)