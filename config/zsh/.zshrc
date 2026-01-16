# Doom Coding Zsh Configuration
# Brand Colors: Forest Green #2E521D, Tan Brown #7C5E46

# ============================================
# OH MY ZSH CONFIGURATION
# ============================================

# Path to Oh My Zsh installation
export ZSH="$HOME/.oh-my-zsh"

# Theme: Powerlevel10k (if installed) or agnoster fallback
if [[ -d "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/themes/powerlevel10k" ]]; then
    ZSH_THEME="powerlevel10k/powerlevel10k"
else
    ZSH_THEME="agnoster"
fi

# Plugins
plugins=(
    git
    docker
    docker-compose
    kubectl
    npm
    node
    python
    pip
    sudo
    history
    z
    zsh-autosuggestions
    zsh-syntax-highlighting
    zsh-completions
)

# Load completions
autoload -U compinit && compinit

# Source Oh My Zsh
source $ZSH/oh-my-zsh.sh

# ============================================
# ENVIRONMENT VARIABLES
# ============================================

# Editor
export EDITOR='vim'
export VISUAL='vim'

# Language
export LANG='en_US.UTF-8'
export LC_ALL='en_US.UTF-8'

# History
export HISTSIZE=50000
export SAVEHIST=50000
export HISTFILE="$HOME/.zsh_history"

# ============================================
# NVM (Node Version Manager)
# ============================================

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"

# ============================================
# PYENV (Python Version Manager)
# ============================================

export PYENV_ROOT="$HOME/.pyenv"
[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"
command -v pyenv &>/dev/null && eval "$(pyenv init -)"

# ============================================
# CLAUDE CODE
# ============================================

export PATH="$HOME/.claude/bin:$PATH"

# Load API key from Docker secret if available
if [[ -f /run/secrets/anthropic_api_key ]]; then
    export ANTHROPIC_API_KEY=$(cat /run/secrets/anthropic_api_key)
fi

# ============================================
# PATH ADDITIONS
# ============================================

# Local binaries
export PATH="$HOME/.local/bin:$PATH"
export PATH="$HOME/bin:$PATH"

# ============================================
# ALIASES - GENERAL
# ============================================

# Directory navigation
alias ..='cd ..'
alias ...='cd ../..'
alias ....='cd ../../..'
alias ll='ls -alF'
alias la='ls -A'
alias l='ls -CF'

# Safety
alias rm='rm -i'
alias cp='cp -i'
alias mv='mv -i'

# Convenience
alias h='history'
alias c='clear'
alias reload='source ~/.zshrc'

# ============================================
# ALIASES - GIT
# ============================================

alias gs='git status'
alias ga='git add'
alias gc='git commit'
alias gp='git push'
alias gl='git pull'
alias gd='git diff'
alias gco='git checkout'
alias gb='git branch'
alias glog='git log --oneline --graph --decorate'
alias gstash='git stash'
alias gpop='git stash pop'

# ============================================
# ALIASES - DOCKER
# ============================================

alias d='docker'
alias dc='docker compose'
alias dps='docker ps'
alias dpsa='docker ps -a'
alias di='docker images'
alias dex='docker exec -it'
alias dlogs='docker logs -f'
alias dprune='docker system prune -af'
alias dstop='docker stop $(docker ps -q)'

# Docker Compose shortcuts
alias dcup='docker compose up -d'
alias dcdown='docker compose down'
alias dcrestart='docker compose restart'
alias dclogs='docker compose logs -f'
alias dcps='docker compose ps'
alias dcbuild='docker compose build'

# ============================================
# ALIASES - CLAUDE CODE
# ============================================

alias cc='claude'
alias ccc='claude --dangerously-skip-permissions'
alias cchelp='claude --help'

# ============================================
# ALIASES - SYSTEM
# ============================================

alias ports='netstat -tulanp'
alias myip='curl -s ifconfig.me'
alias df='df -h'
alias du='du -h'
alias free='free -h'

# ============================================
# FUNCTIONS
# ============================================

# Create directory and cd into it
mkcd() {
    mkdir -p "$1" && cd "$1"
}

# Extract various archive formats
extract() {
    if [[ -f "$1" ]]; then
        case "$1" in
            *.tar.bz2)   tar xjf "$1"     ;;
            *.tar.gz)    tar xzf "$1"     ;;
            *.bz2)       bunzip2 "$1"     ;;
            *.rar)       unrar x "$1"     ;;
            *.gz)        gunzip "$1"      ;;
            *.tar)       tar xf "$1"      ;;
            *.tbz2)      tar xjf "$1"     ;;
            *.tgz)       tar xzf "$1"     ;;
            *.zip)       unzip "$1"       ;;
            *.Z)         uncompress "$1"  ;;
            *.7z)        7z x "$1"        ;;
            *)           echo "'$1' cannot be extracted" ;;
        esac
    else
        echo "'$1' is not a valid file"
    fi
}

# Quick find file
ff() {
    find . -type f -name "*$1*"
}

# Quick find directory
fd() {
    find . -type d -name "*$1*"
}

# Docker shell into container
dsh() {
    docker exec -it "$1" /bin/bash || docker exec -it "$1" /bin/sh
}

# ============================================
# DOOM CODING PROMPT CUSTOMIZATION
# ============================================

# Custom prompt colors (if not using Powerlevel10k)
if [[ "$ZSH_THEME" == "agnoster" ]]; then
    # Override segment colors for Doom Coding theme
    PROMPT_SEGMENT_SEPARATOR=''

    # Customize colors
    export DEFAULT_USER=$(whoami)
fi

# ============================================
# FZF CONFIGURATION
# ============================================

if command -v fzf &>/dev/null; then
    # Use ripgrep for fzf if available
    if command -v rg &>/dev/null; then
        export FZF_DEFAULT_COMMAND='rg --files --hidden --follow --glob "!.git/*"'
    fi

    # FZF options
    export FZF_DEFAULT_OPTS='--height 40% --layout=reverse --border'

    # Key bindings
    [ -f ~/.fzf.zsh ] && source ~/.fzf.zsh
fi

# ============================================
# LOCAL CUSTOMIZATIONS
# ============================================

# Source local config if exists (not tracked in git)
[[ -f ~/.zshrc.local ]] && source ~/.zshrc.local

# ============================================
# STARTUP MESSAGE
# ============================================

# Display Doom Coding banner on first terminal
if [[ -z "$DOOM_CODING_BANNER_SHOWN" ]]; then
    echo -e "\033[38;2;46;82;29m"
    echo "  ____                        ____          _ _             "
    echo " |  _ \\  ___   ___  _ __ ___ / ___|___   __| (_)_ __   __ _ "
    echo " | | | |/ _ \\ / _ \\| '_ \` _ \\ |   / _ \\ / _\` | | '_ \\ / _\` |"
    echo " | |_| | (_) | (_) | | | | | | |__| (_) | (_| | | | | | (_| |"
    echo " |____/ \\___/ \\___/|_| |_| |_|\\____\\___/ \\__,_|_|_| |_|\\__, |"
    echo "                                                       |___/ "
    echo -e "\033[0m"
    echo -e "\033[38;2;124;94;70mRemote Development Environment\033[0m"
    echo ""
    export DOOM_CODING_BANNER_SHOWN=1
fi
