#===============================================================================
# Zsh Configuration - Mobile-Optimized
# Terminal Development Environment
#===============================================================================

#-------------------------------------------------------------------------------
# Oh My Zsh Configuration
#-------------------------------------------------------------------------------

export ZSH="$HOME/.oh-my-zsh"

# Theme (simple and fast)
ZSH_THEME="robbyrussell"

# Plugins
plugins=(
    git
    docker
    docker-compose
    npm
    node
    python
    rust
    tmux
    fzf
    z
    history
    sudo
    copypath
    copyfile
    web-search
    extract
    colored-man-pages
)

# Load Oh My Zsh
source $ZSH/oh-my-zsh.sh

#-------------------------------------------------------------------------------
# Environment Variables
#-------------------------------------------------------------------------------

export EDITOR="nvim"
export VISUAL="nvim"
export PAGER="less"
export LANG="en_US.UTF-8"
export LC_ALL="en_US.UTF-8"

# Path additions
export PATH="$HOME/.local/bin:$PATH"
export PATH="$HOME/.cargo/bin:$PATH"
export PATH="/usr/local/go/bin:$PATH"
export PATH="$HOME/go/bin:$PATH"

# Node.js / npm
export NPM_CONFIG_PREFIX="$HOME/.npm-global"
export PATH="$HOME/.npm-global/bin:$PATH"

# Terminal Dev Environment
export TERMINAL_DEV_ENV="/opt/terminal-dev-env"

#-------------------------------------------------------------------------------
# Aliases - File Operations
#-------------------------------------------------------------------------------

# Use modern replacements if available
if command -v exa &> /dev/null; then
    alias ls="exa --icons"
    alias ll="exa -la --icons"
    alias la="exa -a --icons"
    alias lt="exa --tree --icons"
else
    alias ls="ls --color=auto"
    alias ll="ls -la"
    alias la="ls -a"
fi

if command -v bat &> /dev/null; then
    alias cat="bat --style=plain"
    alias catn="bat"
fi

if command -v fd &> /dev/null; then
    alias find="fd"
fi

# Common file operations
alias ..="cd .."
alias ...="cd ../.."
alias ....="cd ../../.."
alias mkdir="mkdir -pv"
alias cp="cp -iv"
alias mv="mv -iv"
alias rm="rm -iv"

#-------------------------------------------------------------------------------
# Aliases - Git
#-------------------------------------------------------------------------------

alias g="git"
alias gs="git status"
alias ga="git add"
alias gaa="git add --all"
alias gc="git commit"
alias gcm="git commit -m"
alias gp="git push"
alias gpl="git pull"
alias gco="git checkout"
alias gb="git branch"
alias gd="git diff"
alias gl="git log --oneline -10"
alias glog="git log --graph --oneline --all"

#-------------------------------------------------------------------------------
# Aliases - Docker
#-------------------------------------------------------------------------------

alias d="docker"
alias dc="docker compose"
alias dps="docker ps"
alias dpsa="docker ps -a"
alias di="docker images"
alias dex="docker exec -it"
alias dlogs="docker logs -f"

#-------------------------------------------------------------------------------
# Aliases - Development
#-------------------------------------------------------------------------------

alias v="nvim"
alias vim="nvim"
alias py="python3"
alias pip="pip3"
alias nr="npm run"
alias ni="npm install"

#-------------------------------------------------------------------------------
# Aliases - System
#-------------------------------------------------------------------------------

alias ports="netstat -tulanp"
alias meminfo="free -h"
alias diskinfo="df -h"
alias cpu="htop"

# Service management
alias sctl="sudo systemctl"
alias sstart="sudo systemctl start"
alias sstop="sudo systemctl stop"
alias srestart="sudo systemctl restart"
alias sstatus="sudo systemctl status"

#-------------------------------------------------------------------------------
# Aliases - Terminal Dev Environment
#-------------------------------------------------------------------------------

alias tde-status="sudo systemctl status ttyd nginx"
alias tde-logs="tail -f /opt/terminal-dev-env/logs/*.log"
alias tde-restart="sudo systemctl restart ttyd nginx"

#-------------------------------------------------------------------------------
# Functions
#-------------------------------------------------------------------------------

# Create directory and cd into it
mkcd() {
    mkdir -p "$1" && cd "$1"
}

# Quick file search
ff() {
    find . -type f -name "*$1*"
}

# Quick grep in files
fgrep() {
    grep -rn "$1" .
}

# Extract any archive
extract() {
    if [ -f "$1" ]; then
        case $1 in
            *.tar.bz2)   tar xjf $1    ;;
            *.tar.gz)    tar xzf $1    ;;
            *.bz2)       bunzip2 $1    ;;
            *.rar)       unrar x $1    ;;
            *.gz)        gunzip $1     ;;
            *.tar)       tar xf $1     ;;
            *.tbz2)      tar xjf $1    ;;
            *.tgz)       tar xzf $1    ;;
            *.zip)       unzip $1      ;;
            *.Z)         uncompress $1 ;;
            *.7z)        7z x $1       ;;
            *)           echo "'$1' cannot be extracted via extract()" ;;
        esac
    else
        echo "'$1' is not a valid file"
    fi
}

# Quick server
serve() {
    local port="${1:-8000}"
    python3 -m http.server "$port"
}

# Git clone and cd
gclone() {
    git clone "$1" && cd "$(basename "$1" .git)"
}

# Show PATH entries one per line
showpath() {
    echo $PATH | tr ':' '\n'
}

# System update (Debian/Ubuntu)
update() {
    sudo apt update && sudo apt upgrade -y && sudo apt autoremove -y
}

# Quick notes
note() {
    local notes_dir="$HOME/.notes"
    mkdir -p "$notes_dir"
    if [ -z "$1" ]; then
        nvim "$notes_dir/$(date +%Y-%m-%d).md"
    else
        nvim "$notes_dir/$1.md"
    fi
}

# Weather
weather() {
    curl -s "wttr.in/${1:-}"
}

#-------------------------------------------------------------------------------
# FZF Configuration
#-------------------------------------------------------------------------------

if command -v fzf &> /dev/null; then
    # FZF default options
    export FZF_DEFAULT_OPTS="
        --height 40%
        --layout=reverse
        --border
        --preview-window=right:50%
        --color=dark
        --color=fg:-1,bg:-1,hl:#5fff87,fg+:-1,bg+:-1,hl+:#ffaf5f
        --color=info:#af87ff,prompt:#5fff87,pointer:#ff87d7,marker:#ff87d7,spinner:#ff87d7
    "

    # Use fd for fzf if available
    if command -v fd &> /dev/null; then
        export FZF_DEFAULT_COMMAND="fd --type f --hidden --follow --exclude .git"
        export FZF_CTRL_T_COMMAND="$FZF_DEFAULT_COMMAND"
    fi

    # FZF key bindings and completion
    [ -f ~/.fzf.zsh ] && source ~/.fzf.zsh
fi

#-------------------------------------------------------------------------------
# History Configuration
#-------------------------------------------------------------------------------

HISTSIZE=50000
SAVEHIST=50000
HISTFILE=~/.zsh_history

setopt EXTENDED_HISTORY       # Record timestamp
setopt HIST_EXPIRE_DUPS_FIRST # Expire duplicates first
setopt HIST_IGNORE_DUPS       # Don't record duplicates
setopt HIST_IGNORE_ALL_DUPS   # Delete old duplicates
setopt HIST_FIND_NO_DUPS      # Don't display duplicates
setopt HIST_IGNORE_SPACE      # Don't record lines starting with space
setopt HIST_SAVE_NO_DUPS      # Don't write duplicates
setopt SHARE_HISTORY          # Share history between sessions

#-------------------------------------------------------------------------------
# Zsh Options
#-------------------------------------------------------------------------------

setopt AUTO_CD              # cd by typing directory name
setopt AUTO_PUSHD           # Push directories to stack
setopt PUSHD_IGNORE_DUPS    # Don't push duplicates
setopt PUSHD_SILENT         # Don't print stack after push/pop
setopt CORRECT              # Auto-correct commands
setopt NO_CASE_GLOB         # Case-insensitive globbing
setopt EXTENDED_GLOB        # Extended globbing
setopt INTERACTIVE_COMMENTS # Allow comments in interactive shell

#-------------------------------------------------------------------------------
# Completion
#-------------------------------------------------------------------------------

# Case-insensitive completion
zstyle ':completion:*' matcher-list 'm:{a-zA-Z}={A-Za-z}'

# Colored completion
zstyle ':completion:*' list-colors "${(s.:.)LS_COLORS}"

# Menu completion
zstyle ':completion:*' menu select

# Completion caching
zstyle ':completion:*' use-cache on
zstyle ':completion:*' cache-path ~/.zsh/cache

#-------------------------------------------------------------------------------
# Key Bindings
#-------------------------------------------------------------------------------

# Vi mode
bindkey -v

# But keep some Emacs bindings for convenience
bindkey '^A' beginning-of-line
bindkey '^E' end-of-line
bindkey '^R' history-incremental-search-backward
bindkey '^P' up-line-or-history
bindkey '^N' down-line-or-history

# Better history navigation
bindkey '^[[A' history-substring-search-up 2>/dev/null
bindkey '^[[B' history-substring-search-down 2>/dev/null

#-------------------------------------------------------------------------------
# Prompt Customization (Optional - simple prompt for mobile)
#-------------------------------------------------------------------------------

# Uncomment for a simpler prompt better suited for small screens:
# PROMPT='%F{green}%n@%m%f:%F{blue}%~%f$ '

#-------------------------------------------------------------------------------
# Auto-start tmux (if not already in tmux)
#-------------------------------------------------------------------------------

# Uncomment to auto-attach to tmux session
# if command -v tmux &> /dev/null && [ -n "$PS1" ] && [[ ! "$TERM" =~ screen ]] && [[ ! "$TERM" =~ tmux ]] && [ -z "$TMUX" ]; then
#     tmux attach -t dev 2>/dev/null || tmux new -s dev
# fi

#-------------------------------------------------------------------------------
# Welcome Message
#-------------------------------------------------------------------------------

echo ""
echo "Terminal Development Environment"
echo "================================"
echo "Type 'tde-status' to check services"
echo "Type 'tde-logs' to view logs"
echo ""
