#!/usr/bin/env bash
# Doom Coding - Terminal Tools Setup
# Installs zsh, tmux, and development tools in proper order
set -euo pipefail

# ===========================================
# COLORS
# ===========================================
readonly GREEN='\033[38;2;46;82;29m'
readonly BROWN='\033[38;2;124;94;70m'
readonly RED='\033[0;31m'
readonly YELLOW='\033[0;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# ===========================================
# CONFIGURATION
# ===========================================
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
readonly ZSH_CUSTOM="${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}"
readonly NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
readonly PYENV_ROOT="${PYENV_ROOT:-$HOME/.pyenv}"

# ===========================================
# LOGGING
# ===========================================
log_info() { echo -e "${BLUE}ℹ${NC}  $*"; }
log_success() { echo -e "${GREEN}✅${NC} $*"; }
log_warning() { echo -e "${YELLOW}⚠${NC}  $*"; }
log_error() { echo -e "${RED}❌${NC} $*" >&2; }
log_step() { echo -e "${BROWN}⏳${NC} $*"; }

# ===========================================
# PACKAGE INSTALLATION
# ===========================================
detect_package_manager() {
    if command -v apt-get &>/dev/null; then
        echo "apt"
    elif command -v pacman &>/dev/null; then
        echo "pacman"
    elif command -v dnf &>/dev/null; then
        echo "dnf"
    else
        log_error "No supported package manager found"
        exit 1
    fi
}

install_package() {
    local package="$1"
    local pkg_mgr
    pkg_mgr="$(detect_package_manager)"

    case "$pkg_mgr" in
        apt)
            if ! dpkg-query -W -f='${Status}' "$package" 2>/dev/null | grep -q "install ok"; then
                sudo apt-get install -y "$package"
            fi
            ;;
        pacman)
            if ! pacman -Q "$package" &>/dev/null; then
                sudo pacman -S --noconfirm "$package"
            fi
            ;;
        dnf)
            if ! rpm -q "$package" &>/dev/null; then
                sudo dnf install -y "$package"
            fi
            ;;
    esac
}

# ===========================================
# STEP 1: BASE PACKAGES
# ===========================================
install_base_packages() {
    log_step "Installing base packages..."

    local pkg_mgr
    pkg_mgr="$(detect_package_manager)"

    if [[ "$pkg_mgr" == "apt" ]]; then
        sudo apt-get update
    fi

    local packages=(
        git
        curl
        wget
        jq
        ripgrep
        fzf
        htop
        tree
        unzip
        build-essential
        make
    )

    # Adjust package names for different distros
    if [[ "$pkg_mgr" == "pacman" ]]; then
        packages=("${packages[@]/build-essential/base-devel}")
    fi

    for pkg in "${packages[@]}"; do
        install_package "$pkg" || log_warning "Failed to install $pkg"
    done

    log_success "Base packages installed"
}

# ===========================================
# STEP 2: ZSH + OH MY ZSH
# ===========================================
install_zsh() {
    log_step "Installing Zsh..."

    if ! command -v zsh &>/dev/null; then
        install_package zsh
    fi

    log_success "Zsh installed: $(zsh --version)"
}

install_oh_my_zsh() {
    log_step "Installing Oh My Zsh..."

    if [[ -d "$HOME/.oh-my-zsh" ]]; then
        log_info "Oh My Zsh already installed"
        return 0
    fi

    # Install Oh My Zsh without changing shell automatically
    sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended

    log_success "Oh My Zsh installed"
}

# ===========================================
# STEP 3: ZSH PLUGINS
# ===========================================
install_zsh_plugins() {
    log_step "Installing Zsh plugins..."

    local plugins_dir="${ZSH_CUSTOM}/plugins"
    mkdir -p "$plugins_dir"

    # zsh-autosuggestions
    if [[ ! -d "$plugins_dir/zsh-autosuggestions" ]]; then
        git clone https://github.com/zsh-users/zsh-autosuggestions "$plugins_dir/zsh-autosuggestions"
        log_success "zsh-autosuggestions installed"
    else
        log_info "zsh-autosuggestions already installed"
    fi

    # zsh-syntax-highlighting
    if [[ ! -d "$plugins_dir/zsh-syntax-highlighting" ]]; then
        git clone https://github.com/zsh-users/zsh-syntax-highlighting "$plugins_dir/zsh-syntax-highlighting"
        log_success "zsh-syntax-highlighting installed"
    else
        log_info "zsh-syntax-highlighting already installed"
    fi

    # zsh-completions
    if [[ ! -d "$plugins_dir/zsh-completions" ]]; then
        git clone https://github.com/zsh-users/zsh-completions "$plugins_dir/zsh-completions"
        log_success "zsh-completions installed"
    else
        log_info "zsh-completions already installed"
    fi

    log_success "Zsh plugins installed"
}

install_powerlevel10k() {
    log_step "Installing Powerlevel10k theme..."

    local themes_dir="${ZSH_CUSTOM}/themes"
    mkdir -p "$themes_dir"

    if [[ ! -d "$themes_dir/powerlevel10k" ]]; then
        git clone --depth=1 https://github.com/romkatv/powerlevel10k.git "$themes_dir/powerlevel10k"
        log_success "Powerlevel10k installed"
    else
        log_info "Powerlevel10k already installed"
    fi
}

# ===========================================
# STEP 4: TMUX + TPM
# ===========================================
install_tmux() {
    log_step "Installing Tmux..."

    if ! command -v tmux &>/dev/null; then
        install_package tmux
    fi

    log_success "Tmux installed: $(tmux -V)"
}

install_tpm() {
    log_step "Installing Tmux Plugin Manager..."

    if [[ ! -d "$HOME/.tmux/plugins/tpm" ]]; then
        git clone https://github.com/tmux-plugins/tpm "$HOME/.tmux/plugins/tpm"
        log_success "TPM installed"
    else
        log_info "TPM already installed"
    fi
}

# ===========================================
# STEP 5: NVM + NODE.JS
# ===========================================
install_nvm() {
    log_step "Installing NVM..."

    if [[ -d "$NVM_DIR" ]]; then
        log_info "NVM already installed"
        return 0
    fi

    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash

    # Load NVM
    export NVM_DIR="$HOME/.nvm"
    # shellcheck source=/dev/null
    [[ -s "$NVM_DIR/nvm.sh" ]] && source "$NVM_DIR/nvm.sh"

    log_success "NVM installed"
}

install_nodejs() {
    log_step "Installing Node.js LTS..."

    # Load NVM if available
    export NVM_DIR="$HOME/.nvm"
    # shellcheck source=/dev/null
    [[ -s "$NVM_DIR/nvm.sh" ]] && source "$NVM_DIR/nvm.sh"

    if ! command -v nvm &>/dev/null; then
        log_warning "NVM not available, skipping Node.js installation"
        return 0
    fi

    if nvm ls --no-colors 2>/dev/null | grep -q "lts"; then
        log_info "Node.js LTS already installed"
    else
        nvm install --lts
        nvm use --lts
        nvm alias default 'lts/*'
        log_success "Node.js LTS installed: $(node --version)"
    fi
}

# ===========================================
# STEP 6: PYENV + PYTHON
# ===========================================
install_pyenv_deps() {
    log_step "Installing pyenv dependencies..."

    local pkg_mgr
    pkg_mgr="$(detect_package_manager)"

    case "$pkg_mgr" in
        apt)
            sudo apt-get install -y \
                libssl-dev \
                zlib1g-dev \
                libbz2-dev \
                libreadline-dev \
                libsqlite3-dev \
                libncursesw5-dev \
                xz-utils \
                tk-dev \
                libxml2-dev \
                libxmlsec1-dev \
                libffi-dev \
                liblzma-dev
            ;;
        pacman)
            sudo pacman -S --noconfirm \
                openssl \
                zlib \
                bzip2 \
                readline \
                sqlite \
                ncurses \
                xz \
                tk \
                libxml2 \
                libffi
            ;;
    esac

    log_success "pyenv dependencies installed"
}

install_pyenv() {
    log_step "Installing pyenv..."

    if [[ -d "$PYENV_ROOT" ]]; then
        log_info "pyenv already installed"
        return 0
    fi

    curl https://pyenv.run | bash

    log_success "pyenv installed"
}

install_python() {
    log_step "Installing Python 3.12..."

    # Load pyenv
    export PYENV_ROOT="$HOME/.pyenv"
    export PATH="$PYENV_ROOT/bin:$PATH"

    if ! command -v pyenv &>/dev/null; then
        log_warning "pyenv not available, skipping Python installation"
        return 0
    fi

    eval "$(pyenv init -)"

    if pyenv versions --bare 2>/dev/null | grep -q "3.12"; then
        log_info "Python 3.12 already installed"
    else
        pyenv install 3.12
        pyenv global 3.12
        log_success "Python 3.12 installed"
    fi
}

# ===========================================
# CONFIGURATION COPY
# ===========================================
copy_configs() {
    log_step "Copying configuration files..."

    # Copy .zshrc if it exists in project
    if [[ -f "$PROJECT_DIR/config/zsh/.zshrc" ]]; then
        cp "$PROJECT_DIR/config/zsh/.zshrc" "$HOME/.zshrc"
        log_success "Copied .zshrc"
    fi

    # Copy .tmux.conf if it exists
    if [[ -f "$PROJECT_DIR/config/tmux/tmux.conf" ]]; then
        cp "$PROJECT_DIR/config/tmux/tmux.conf" "$HOME/.tmux.conf"
        log_success "Copied .tmux.conf"
    fi
}

# ===========================================
# SET DEFAULT SHELL
# ===========================================
set_default_shell() {
    log_step "Setting Zsh as default shell..."

    local zsh_path
    zsh_path="$(which zsh)"

    if [[ "$SHELL" == "$zsh_path" ]]; then
        log_info "Zsh is already the default shell"
        return 0
    fi

    # Add zsh to /etc/shells if not present
    if ! grep -q "$zsh_path" /etc/shells; then
        echo "$zsh_path" | sudo tee -a /etc/shells
    fi

    # Change shell
    chsh -s "$zsh_path" || log_warning "Could not change shell. Run: chsh -s $zsh_path"

    log_success "Default shell set to Zsh"
}

# ===========================================
# MAIN
# ===========================================
main() {
    echo -e "${GREEN}Doom Coding - Terminal Setup${NC}"
    echo "=============================="
    echo ""
    echo "Installation order:"
    echo "  1. Base packages"
    echo "  2. Zsh + Oh My Zsh"
    echo "  3. Zsh plugins"
    echo "  4. Tmux + TPM"
    echo "  5. NVM + Node.js LTS"
    echo "  6. pyenv + Python 3.12"
    echo ""

    # Step 1
    install_base_packages

    # Step 2
    install_zsh
    install_oh_my_zsh

    # Step 3
    install_zsh_plugins
    install_powerlevel10k

    # Step 4
    install_tmux
    install_tpm

    # Step 5
    install_nvm
    install_nodejs

    # Step 6
    install_pyenv_deps
    install_pyenv
    install_python

    # Copy configs
    copy_configs

    # Set default shell
    set_default_shell

    echo ""
    log_success "Terminal setup completed!"
    echo ""
    echo "To apply changes, either:"
    echo "  - Log out and back in"
    echo "  - Or run: exec zsh"
}

main "$@"
