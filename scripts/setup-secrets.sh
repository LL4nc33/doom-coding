#!/usr/bin/env bash
# Doom Coding - Secrets Management with SOPS/age
set -euo pipefail

# ===========================================
# COLORS
# ===========================================
readonly GREEN='\033[38;2;46;82;29m'
readonly RED='\033[0;31m'
readonly YELLOW='\033[0;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# ===========================================
# CONFIGURATION
# ===========================================
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
readonly SECRETS_DIR="$PROJECT_DIR/secrets"
readonly AGE_KEY_DIR="$HOME/.config/sops/age"
readonly AGE_KEY_FILE="$AGE_KEY_DIR/keys.txt"

# ===========================================
# LOGGING
# ===========================================
log_info() { echo -e "${BLUE}ℹ${NC}  $*"; }
log_success() { echo -e "${GREEN}✅${NC} $*"; }
log_warning() { echo -e "${YELLOW}⚠${NC}  $*"; }
log_error() { echo -e "${RED}❌${NC} $*" >&2; }

# ===========================================
# DEPENDENCY INSTALLATION
# ===========================================
install_sops() {
    if command -v sops &>/dev/null; then
        log_info "SOPS already installed: $(sops --version)"
        return 0
    fi

    log_info "Installing SOPS..."

    local arch
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) log_error "Unsupported architecture"; exit 1 ;;
    esac

    local version="3.8.1"
    local url="https://github.com/getsops/sops/releases/download/v${version}/sops-v${version}.linux.${arch}"

    sudo curl -Lo /usr/local/bin/sops "$url"
    sudo chmod +x /usr/local/bin/sops

    log_success "SOPS installed"
}

install_age() {
    if command -v age &>/dev/null; then
        log_info "age already installed: $(age --version)"
        return 0
    fi

    log_info "Installing age..."

    if command -v apt-get &>/dev/null; then
        sudo apt-get update
        sudo apt-get install -y age
    elif command -v pacman &>/dev/null; then
        sudo pacman -S --noconfirm age
    else
        # Manual installation
        local arch
        case "$(uname -m)" in
            x86_64|amd64) arch="amd64" ;;
            aarch64|arm64) arch="arm64" ;;
        esac

        local version="1.1.1"
        local url="https://github.com/FiloSottile/age/releases/download/v${version}/age-v${version}-linux-${arch}.tar.gz"

        curl -sL "$url" | sudo tar -xz -C /usr/local/bin --strip-components=1 age/age age/age-keygen
    fi

    log_success "age installed"
}

# ===========================================
# KEY MANAGEMENT
# ===========================================
generate_key() {
    log_info "Generating age encryption key..."

    mkdir -p "$AGE_KEY_DIR"
    chmod 700 "$AGE_KEY_DIR"

    if [[ -f "$AGE_KEY_FILE" ]]; then
        log_warning "Key file already exists: $AGE_KEY_FILE"
        log_info "To generate a new key, first remove the existing one"
        return 0
    fi

    age-keygen -o "$AGE_KEY_FILE"
    chmod 600 "$AGE_KEY_FILE"

    local public_key
    public_key=$(grep "public key:" "$AGE_KEY_FILE" | awk '{print $4}')

    log_success "Key generated successfully!"
    echo ""
    echo "Public key: $public_key"
    echo ""
    echo "Add this public key to your .sops.yaml:"
    echo "  age: $public_key"
    echo ""
    log_warning "IMPORTANT: Backup $AGE_KEY_FILE securely!"
}

show_public_key() {
    if [[ ! -f "$AGE_KEY_FILE" ]]; then
        log_error "No key file found. Run: $0 generate-key"
        exit 1
    fi

    local public_key
    public_key=$(grep "public key:" "$AGE_KEY_FILE" | awk '{print $4}')

    echo "$public_key"
}

# ===========================================
# SECRET OPERATIONS
# ===========================================
create_template() {
    log_info "Creating secrets template..."

    mkdir -p "$SECRETS_DIR"

    if [[ ! -f "$SECRETS_DIR/secrets.yaml" ]]; then
        cat > "$SECRETS_DIR/secrets.yaml" << 'EOF'
# Doom Coding Secrets Template
# Encrypt this file with: ./scripts/setup-secrets.sh encrypt secrets/secrets.yaml

# Tailscale
tailscale:
    auth_key: "tskey-auth-XXXXXXXX"

# code-server
code_server:
    password: "your-secure-password"
    sudo_password: "your-sudo-password"

# Claude API
anthropic:
    api_key: "sk-ant-XXXXXXXX"

# Additional secrets
custom:
    key1: "value1"
    key2: "value2"
EOF
        log_success "Template created: $SECRETS_DIR/secrets.yaml"
        log_warning "Edit this file with your actual secrets, then encrypt it!"
    else
        log_warning "Template already exists: $SECRETS_DIR/secrets.yaml"
    fi
}

encrypt_file() {
    local file="$1"

    if [[ ! -f "$file" ]]; then
        log_error "File not found: $file"
        exit 1
    fi

    if [[ ! -f "$AGE_KEY_FILE" ]]; then
        log_error "No encryption key found. Run: $0 generate-key"
        exit 1
    fi

    local public_key
    public_key=$(grep "public key:" "$AGE_KEY_FILE" | awk '{print $4}')

    local encrypted_file="${file%.yaml}.enc.yaml"

    sops --encrypt --age "$public_key" "$file" > "$encrypted_file"

    log_success "Encrypted: $encrypted_file"
    log_info "You can now safely delete the unencrypted file: $file"
}

decrypt_file() {
    local file="$1"

    if [[ ! -f "$file" ]]; then
        log_error "File not found: $file"
        exit 1
    fi

    export SOPS_AGE_KEY_FILE="$AGE_KEY_FILE"

    sops --decrypt "$file"
}

edit_file() {
    local file="$1"

    if [[ ! -f "$file" ]]; then
        log_error "File not found: $file"
        exit 1
    fi

    export SOPS_AGE_KEY_FILE="$AGE_KEY_FILE"

    sops "$file"
}

# ===========================================
# EXPORT FOR DOCKER
# ===========================================
export_secrets() {
    log_info "Exporting secrets for Docker..."

    local secrets_file="$SECRETS_DIR/secrets.enc.yaml"

    if [[ ! -f "$secrets_file" ]]; then
        log_error "Encrypted secrets file not found: $secrets_file"
        log_info "Create and encrypt your secrets first"
        exit 1
    fi

    export SOPS_AGE_KEY_FILE="$AGE_KEY_FILE"

    # Export individual secrets
    local api_key
    api_key=$(sops --decrypt --extract '["anthropic"]["api_key"]' "$secrets_file" 2>/dev/null || echo "")

    if [[ -n "$api_key" ]]; then
        echo "$api_key" > "$SECRETS_DIR/anthropic_api_key.txt"
        chmod 600 "$SECRETS_DIR/anthropic_api_key.txt"
        log_success "Exported: anthropic_api_key.txt"
    fi

    local ts_key
    ts_key=$(sops --decrypt --extract '["tailscale"]["auth_key"]' "$secrets_file" 2>/dev/null || echo "")

    if [[ -n "$ts_key" ]]; then
        echo "$ts_key" > "$SECRETS_DIR/tailscale_auth_key.txt"
        chmod 600 "$SECRETS_DIR/tailscale_auth_key.txt"
        log_success "Exported: tailscale_auth_key.txt"
    fi

    local cs_pass
    cs_pass=$(sops --decrypt --extract '["code_server"]["password"]' "$secrets_file" 2>/dev/null || echo "")

    if [[ -n "$cs_pass" ]]; then
        echo "$cs_pass" > "$SECRETS_DIR/code_server_password.txt"
        chmod 600 "$SECRETS_DIR/code_server_password.txt"
        log_success "Exported: code_server_password.txt"
    fi

    log_success "Secrets exported for Docker"
}

# ===========================================
# INITIALIZATION
# ===========================================
init() {
    log_info "Initializing secrets management..."

    install_sops
    install_age

    mkdir -p "$SECRETS_DIR"
    chmod 700 "$SECRETS_DIR"

    if [[ ! -f "$AGE_KEY_FILE" ]]; then
        generate_key
    else
        log_info "Using existing key: $AGE_KEY_FILE"
    fi

    # Update .sops.yaml with actual public key
    local public_key
    public_key=$(grep "public key:" "$AGE_KEY_FILE" | awk '{print $4}')

    if [[ -f "$PROJECT_DIR/.sops.yaml" ]]; then
        sed -i "s/age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx/$public_key/g" "$PROJECT_DIR/.sops.yaml"
        log_success "Updated .sops.yaml with your public key"
    fi

    log_success "Secrets management initialized"
}

# ===========================================
# HELP
# ===========================================
show_help() {
    cat << EOF
Doom Coding - Secrets Management

USAGE:
    $0 <command> [options]

COMMANDS:
    init            Initialize secrets management (install tools, generate key)
    generate-key    Generate a new age encryption key
    show-key        Display the public key
    template        Create a secrets template file
    encrypt FILE    Encrypt a secrets file
    decrypt FILE    Decrypt and display a secrets file
    edit FILE       Edit an encrypted file in place
    export          Export secrets for Docker consumption

EXAMPLES:
    $0 init
    $0 template
    $0 encrypt secrets/secrets.yaml
    $0 decrypt secrets/secrets.enc.yaml
    $0 export

EOF
}

# ===========================================
# MAIN
# ===========================================
main() {
    if [[ $# -eq 0 ]]; then
        show_help
        exit 0
    fi

    local command="$1"
    shift

    case "$command" in
        init)
            init
            ;;
        generate-key)
            install_age
            generate_key
            ;;
        show-key)
            show_public_key
            ;;
        template)
            create_template
            ;;
        encrypt)
            [[ $# -eq 0 ]] && { log_error "File required"; exit 1; }
            encrypt_file "$1"
            ;;
        decrypt)
            [[ $# -eq 0 ]] && { log_error "File required"; exit 1; }
            decrypt_file "$1"
            ;;
        edit)
            [[ $# -eq 0 ]] && { log_error "File required"; exit 1; }
            edit_file "$1"
            ;;
        export)
            export_secrets
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
