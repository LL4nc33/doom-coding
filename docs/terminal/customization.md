# Terminal Customization Guide

This guide covers customization options for terminal tools in Doom Coding, including both the Docker-based environment and the lightweight terminal development environment.

## 🎨 Overview

Doom Coding includes several customizable terminal tools:
- **zsh** with Oh My Zsh framework
- **tmux** with custom configuration
- **neovim** with modern Lua configuration
- **Modern CLI tools** (exa, bat, fzf, ripgrep)

## 🐚 Zsh Customization

### Default Configuration
Located in `config/zsh/.zshrc` with these features:
- Oh My Zsh framework
- Useful plugins (git, docker, fzf, z, etc.)
- Modern aliases
- Git shortcuts
- Development environment variables

### Customizing Your Theme

```bash
# Edit zsh configuration
vim ~/.zshrc

# Change theme (default: robbyrussell)
ZSH_THEME="agnoster"  # Or any Oh My Zsh theme

# Popular themes for development:
ZSH_THEME="powerlevel10k/powerlevel10k"  # Requires installation
ZSH_THEME="spaceship"  # Modern and fast
ZSH_THEME="pure"       # Minimal and clean
```

### Installing Powerlevel10k
```bash
# Clone theme
git clone --depth=1 https://github.com/romkatv/powerlevel10k.git ${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/themes/powerlevel10k

# Set theme in .zshrc
ZSH_THEME="powerlevel10k/powerlevel10k"

# Restart shell and configure
exec zsh
p10k configure
```

### Adding Custom Aliases
Add to your `~/.zshrc`:

```bash
# Development shortcuts
alias gcm="git commit -m"
alias gps="git push"
alias gpl="git pull"
alias gst="git status"

# Docker shortcuts
alias dps="docker ps"
alias dpsa="docker ps -a"
alias dex="docker exec -it"
alias dlogs="docker logs -f"

# Navigation
alias ..="cd .."
alias ...="cd ../.."
alias ll="ls -la"
alias la="ls -A"

# Development tools
alias py="python3"
alias serve="python3 -m http.server"
alias ports="netstat -tulanp"
```

### Plugin Configuration
Enable additional Oh My Zsh plugins:

```bash
plugins=(
    git
    docker
    docker-compose
    npm
    node
    python
    rust
    golang
    kubectl
    helm
    terraform
    aws
    gcp
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
    command-not-found
    auto-suggestions      # Requires installation
    syntax-highlighting   # Requires installation
)
```

### Installing Additional Plugins

**Autosuggestions:**
```bash
git clone https://github.com/zsh-users/zsh-autosuggestions ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions
```

**Syntax Highlighting:**
```bash
git clone https://github.com/zsh-users/zsh-syntax-highlighting.git ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting
```

## 📺 Tmux Customization

### Default Configuration
Located in `config/tmux/tmux.conf` with mobile-optimized settings:
- Prefix key: `Ctrl+a` (more touch-friendly than `Ctrl+b`)
- Mouse support enabled
- Session persistence with tmux-resurrect
- Vi-style key bindings

### Key Bindings Reference

| Action | Key Combination |
|--------|----------------|
| Prefix | `Ctrl+a` |
| Vertical split | `Prefix + \|` |
| Horizontal split | `Prefix + -` |
| Switch panes | `Alt + Arrow Keys` (no prefix!) |
| Switch windows | `Shift + Arrow Keys` (no prefix!) |
| Quick windows | `Alt + 1-5` |
| Reload config | `Prefix + r` |
| Save session | `Prefix + Ctrl+s` |
| Restore session | `Prefix + Ctrl+r` |

### Custom Tmux Configuration
Create `~/.tmux.conf.local`:

```bash
# Custom status bar colors
set -g status-bg colour235
set -g status-fg colour136

# Window status colors
setw -g window-status-current-style bg=colour166,fg=colour235

# Pane border colors
set -g pane-border-style fg=colour240
set -g pane-active-border-style fg=colour166

# Custom key bindings
bind-key v split-window -h
bind-key s split-window -v
bind-key h select-pane -L
bind-key j select-pane -D
bind-key k select-pane -U
bind-key l select-pane -R

# Increase history limit
set -g history-limit 50000

# Enable focus events
set -g focus-events on
```

### Installing TPM (Tmux Plugin Manager)
```bash
git clone https://github.com/tmux-plugins/tpm ~/.tmux/plugins/tpm

# Add to your .tmux.conf:
set -g @plugin 'tmux-plugins/tpm'
set -g @plugin 'tmux-plugins/tmux-resurrect'
set -g @plugin 'tmux-plugins/tmux-continuum'

# Initialize TPM
run '~/.tmux/plugins/tpm/tpm'

# Install plugins: Prefix + I
# Update plugins: Prefix + U
# Remove plugins: Prefix + Alt + u
```

### Session Management
```bash
# Create named session
tmux new -s development

# List sessions
tmux ls

# Attach to session
tmux attach -t development

# Rename session
Prefix + $

# Kill session
tmux kill-session -t development
```

## ⚙️ Neovim Customization

### Default Configuration
Located in `config/neovim/init.lua` with:
- Lazy.nvim plugin manager
- LSP support via Mason
- File explorer (nvim-tree)
- Fuzzy finder (Telescope)
- Git integration (Gitsigns)
- Syntax highlighting (Treesitter)

### Custom Configuration
Edit `~/.config/nvim/init.lua`:

```lua
-- Custom leader key (default is Space)
vim.g.mapleader = " "

-- Custom options
vim.opt.number = true
vim.opt.relativenumber = true
vim.opt.tabstop = 4
vim.opt.shiftwidth = 4
vim.opt.expandtab = true
vim.opt.wrap = false

-- Custom keymaps
local map = vim.keymap.set

-- Better navigation
map('n', '<C-h>', '<C-w>h', { desc = 'Go to left window' })
map('n', '<C-j>', '<C-w>j', { desc = 'Go to lower window' })
map('n', '<C-k>', '<C-w>k', { desc = 'Go to upper window' })
map('n', '<C-l>', '<C-w>l', { desc = 'Go to right window' })

-- Quick save
map('n', '<C-s>', ':w<CR>', { desc = 'Save file' })

-- Better indenting
map('v', '<', '<gv', { desc = 'Indent left' })
map('v', '>', '>gv', { desc = 'Indent right' })
```

### Installing Additional Plugins
Add to your lazy.nvim setup:

```lua
{
    "folke/which-key.nvim",
    config = function()
        vim.o.timeout = true
        vim.o.timeoutlen = 300
        require("which-key").setup()
    end
},

{
    "lewis6991/gitsigns.nvim",
    config = function()
        require('gitsigns').setup()
    end
},

{
    "nvim-lualine/lualine.nvim",
    dependencies = { 'nvim-tree/nvim-web-devicons' },
    config = function()
        require('lualine').setup {
            options = {
                theme = 'tokyonight'
            }
        }
    end
}
```

### Language Server Protocol (LSP)
Configure language servers via Mason:

```lua
require("mason").setup()
require("mason-lspconfig").setup({
    ensure_installed = {
        "lua_ls",        -- Lua
        "pyright",       -- Python
        "tsserver",      -- TypeScript
        "rust_analyzer", -- Rust
        "gopls",         -- Go
        "clangd",        -- C/C++
        "bashls",        -- Bash
        "dockerls",      -- Docker
        "yamlls",        -- YAML
    }
})
```

## 🛠️ Modern CLI Tools

### Exa (ls replacement)
```bash
# Aliases already configured in .zshrc
alias ls="exa --icons"
alias ll="exa -la --icons"
alias lt="exa --tree --icons"

# Custom exa configuration
export EXA_COLORS="da=1;34:gm=1;34"
```

### Bat (cat replacement)
```bash
# Set theme
export BAT_THEME="TwoDark"

# Custom aliases
alias cat="bat --style=plain"
alias catn="bat"  # with line numbers
```

### FZF (Fuzzy Finder)
```bash
# Custom FZF options
export FZF_DEFAULT_OPTS="
    --height 40%
    --layout=reverse
    --border
    --preview-window=right:50%
    --color=dark
"

# Use fd for better performance
export FZF_DEFAULT_COMMAND="fd --type f --hidden --follow --exclude .git"

# Custom key bindings
bindkey '^T' fzf-file-widget
bindkey '^R' fzf-history-widget
bindkey '\\C-F' fzf-cd-widget
```

### Ripgrep Configuration
```bash
# Create ~/.ripgreprc
--type-add
--web:*.{html,css,js,ts,jsx,tsx,vue,svelte}
--smart-case
--follow
--hidden
--glob=!.git/*
--glob=!node_modules/*
```

## 🎨 Color Schemes

### Terminal Colors
Most terminals support 256 colors or true color. Popular schemes:
- **Dracula**: Dark theme with purple accents
- **Nord**: Arctic, north-bluish color palette
- **Gruvbox**: Retro groove color scheme
- **Solarized**: Balanced dark/light themes
- **One Dark**: Atom's One Dark theme

### Consistent Theming
For consistent colors across all tools, consider:

```bash
# Add to .zshrc
export BAT_THEME="Nord"
export FZF_DEFAULT_OPTS="--color=bg+:#3B4252,bg:#2E3440,spinner:#81A1C1"

# Tmux colors in .tmux.conf
set -g status-bg '#2E3440'
set -g status-fg '#D8DEE9'
```

## 📱 Mobile Optimizations

### Touch-Friendly Settings
Already configured in default setup:
- tmux mouse support enabled
- Larger click targets
- Simplified key combinations
- Alternative escape sequences (jk/kj in neovim)

### Smartphone-Specific Tips
```bash
# Increase tmux status line size for touch
set -g status-left-length 50
set -g status-right-length 150

# Larger pane indicators
set -g display-panes-time 4000
```

## 🔧 Advanced Customization

### Environment Variables
Add to `~/.zshrc`:

```bash
# Development paths
export GOPATH="$HOME/go"
export PATH="$PATH:$GOPATH/bin"
export PATH="$HOME/.local/bin:$PATH"

# Editor preferences
export EDITOR="nvim"
export VISUAL="nvim"
export PAGER="bat"

# Language settings
export LANG="en_US.UTF-8"
export LC_ALL="en_US.UTF-8"
```

### Custom Functions
```bash
# Quick directory navigation
mkcd() {
    mkdir -p "$1" && cd "$1"
}

# Find and edit files quickly
fe() {
    local file
    file=$(fzf --query="$1" --select-1 --exit-0)
    [ -n "$file" ] && ${EDITOR:-nvim} "$file"
}

# Git commit with message
gcm() {
    git add -A && git commit -m "$1"
}
```

### Dotfile Management
Consider using a dotfile manager:

```bash
# Using chezmoi
curl -sfL https://git.io/chezmoi | sh
chezmoi init --apply https://github.com/yourusername/dotfiles.git

# Using GNU Stow
git clone https://github.com/yourusername/dotfiles.git ~/.dotfiles
cd ~/.dotfiles
stow zsh tmux neovim
```

## 🚀 Performance Tips

### Startup Performance
```bash
# Time zsh startup
time zsh -i -c exit

# Profile plugin loading times
for plugin in $plugins; do
    timer=$(date +%s%N)
    source $ZSH/plugins/$plugin/$plugin.plugin.zsh
    now=$(date +%s%N)
    elapsed=$(($now-$timer))
    echo $elapsed / 1000000 | bc | awk '{printf "%.3f ms - %s\n", $1, "'$plugin'"}'
done
```

### Memory Usage
```bash
# Monitor tmux memory usage
ps aux | grep tmux

# Limit history in memory-constrained environments
# In .tmux.conf:
set -g history-limit 10000
```

## 📚 Resources

### Documentation
- [Oh My Zsh Documentation](https://github.com/ohmyzsh/ohmyzsh)
- [Tmux Manual](https://github.com/tmux/tmux/wiki)
- [Neovim Documentation](https://neovim.io/doc/)
- [FZF Wiki](https://github.com/junegunn/fzf)

### Inspiration
- [Awesome Zsh Plugins](https://github.com/unixorn/awesome-zsh-plugins)
- [Tmux Configuration Examples](https://github.com/gpakosz/.tmux)
- [Neovim Configurations](https://github.com/topics/neovim-configuration)

### Communities
- [r/commandline](https://reddit.com/r/commandline)
- [r/tmux](https://reddit.com/r/tmux)
- [r/neovim](https://reddit.com/r/neovim)

---

*Remember to backup your configurations before making major changes, and test customizations in a non-critical environment first.*