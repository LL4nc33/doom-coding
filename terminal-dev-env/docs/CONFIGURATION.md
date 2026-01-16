# Konfigurationshandbuch

Detaillierte Beschreibung aller Konfigurations-Optionen fuer das Terminal Development Environment.

## Inhaltsverzeichnis

1. [nginx-Konfiguration](#nginx-konfiguration)
2. [ttyd-Konfiguration](#ttyd-konfiguration)
3. [tmux-Konfiguration](#tmux-konfiguration)
4. [Neovim-Konfiguration](#neovim-konfiguration)
5. [Zsh-Konfiguration](#zsh-konfiguration)
6. [SSL-Konfiguration](#ssl-konfiguration)
7. [Firewall-Konfiguration](#firewall-konfiguration)
8. [Umgebungsvariablen](#umgebungsvariablen)

---

## nginx-Konfiguration

**Datei**: `/opt/terminal-dev-env/config/nginx/terminal.conf`

### Wichtige Parameter

```nginx
# Upstream-Definition fuer ttyd
upstream ttyd_backend {
    server 127.0.0.1:7681;    # ttyd lauscht nur auf localhost
    keepalive 32;              # Persistente Verbindungen
}
```

### SSL-Einstellungen

```nginx
# Moderne SSL-Konfiguration (TLS 1.2+)
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:...;
ssl_prefer_server_ciphers off;
ssl_session_cache shared:SSL:10m;
ssl_session_timeout 1d;
ssl_session_tickets off;
```

### Timeout-Einstellungen

Fuer lange Terminal-Sessions muessen die Timeouts entsprechend hoch sein:

```nginx
proxy_connect_timeout 7d;   # 7 Tage
proxy_send_timeout 7d;
proxy_read_timeout 7d;
```

**Anpassen**: Fuer kuerzere maximale Session-Dauer:

```nginx
proxy_connect_timeout 24h;  # 24 Stunden
proxy_send_timeout 24h;
proxy_read_timeout 24h;
```

### Rate Limiting

Schutz vor Brute-Force-Angriffen:

```nginx
limit_req_zone $binary_remote_addr zone=terminal_limit:10m rate=10r/s;

location / {
    limit_req zone=terminal_limit burst=20 nodelay;
    # ...
}
```

**Anpassen**: Fuer strengeres Limiting:

```nginx
# Max 5 Requests pro Sekunde, Burst von 10
limit_req_zone $binary_remote_addr zone=terminal_limit:10m rate=5r/s;
limit_req zone=terminal_limit burst=10 nodelay;
```

### Authentifizierung aendern

Neuen Benutzer hinzufuegen:

```bash
# Mit htpasswd (falls installiert)
htpasswd /opt/terminal-dev-env/config/nginx/.htpasswd neueruser

# Mit OpenSSL
echo "neueruser:$(openssl passwd -apr1 neuespasswort)" >> /opt/terminal-dev-env/config/nginx/.htpasswd
```

Passwort aendern:

```bash
# Datei ueberschreiben mit neuem Passwort
echo "admin:$(openssl passwd -apr1 neuespasswort)" > /opt/terminal-dev-env/config/nginx/.htpasswd
sudo systemctl reload nginx
```

### IP-basierte Zugriffskontrolle

Zugriff auf bestimmte IPs beschraenken:

```nginx
location / {
    # Nur lokales Netzwerk erlauben
    allow 192.168.178.0/24;
    allow 10.0.0.0/8;
    allow 127.0.0.1;
    deny all;

    # Rest der Konfiguration...
}
```

---

## ttyd-Konfiguration

**Datei**: `/etc/systemd/system/ttyd.service`

### Standard-Parameter

```ini
ExecStart=/usr/local/bin/ttyd \
    --port 7681 \              # Port (nur localhost)
    --interface 127.0.0.1 \    # Nur localhost binden
    --max-clients 5 \          # Max gleichzeitige Verbindungen
    --ping-interval 30 \       # WebSocket-Ping alle 30 Sekunden
    /opt/terminal-dev-env/bin/terminal-session.sh
```

### Anpassbare Optionen

| Option | Beschreibung | Standard |
|--------|--------------|----------|
| `--port` | Listening-Port | 7681 |
| `--interface` | Binding-Interface | 127.0.0.1 |
| `--max-clients` | Max Clients | 5 |
| `--ping-interval` | Ping-Intervall (Sek) | 30 |
| `--once` | Nach erster Verbindung beenden | deaktiviert |
| `--readonly` | Nur-Lese-Modus | deaktiviert |

### Beispiel: Nur eine Session erlauben

```ini
ExecStart=/usr/local/bin/ttyd \
    --port 7681 \
    --interface 127.0.0.1 \
    --max-clients 1 \
    --once \
    /opt/terminal-dev-env/bin/terminal-session.sh
```

### Aenderungen uebernehmen

```bash
sudo systemctl daemon-reload
sudo systemctl restart ttyd
```

---

## tmux-Konfiguration

**Datei**: `/opt/terminal-dev-env/config/tmux/tmux.conf`

### Prefix-Taste aendern

Standard: `Ctrl+a` (Mobile-freundlicher als `Ctrl+b`)

```bash
# Zurueck auf Standard
unbind C-a
set -g prefix C-b
bind C-b send-prefix
```

### Maus-Unterstuetzung

```bash
# Maus aktivieren (Standard: an)
set -g mouse on

# Maus deaktivieren
set -g mouse off
```

### Farb-Schema anpassen

```bash
# Status-Bar-Farben (Tokyo Night Theme)
set -g status-style 'bg=#1a1b26 fg=#a9b1d6'

# Alternative: Gruenes Theme
set -g status-style 'bg=#1d2021 fg=#a89984'
set -g window-status-current-format '#[fg=#1d2021,bg=#98971a,bold] #I:#W '
```

### History-Limit

```bash
# Standard: 50000 Zeilen
set -g history-limit 50000

# Reduziert fuer weniger Speicherverbrauch
set -g history-limit 10000
```

### Session-Persistenz konfigurieren

```bash
# Auto-Save-Intervall (Minuten)
set -g @continuum-save-interval '5'    # Standard: alle 5 Minuten
set -g @continuum-save-interval '15'   # Alle 15 Minuten

# Auto-Restore beim Start
set -g @continuum-restore 'on'         # Standard: aktiviert
set -g @continuum-restore 'off'        # Deaktivieren
```

### Konfiguration neu laden

In tmux: `Prefix + r`

Oder im Terminal:

```bash
tmux source-file /opt/terminal-dev-env/config/tmux/tmux.conf
```

---

## Neovim-Konfiguration

**Datei**: `/opt/terminal-dev-env/config/neovim/init.lua` oder `~/.config/nvim/init.lua`

### Grundeinstellungen

```lua
-- Zeilennummern
opt.number = true           -- Absolute Zeilennummern
opt.relativenumber = true   -- Relative Zeilennummern

-- Nur absolute Zeilennummern
opt.number = true
opt.relativenumber = false
```

### Einrueckung

```lua
-- Standard: 4 Spaces
opt.tabstop = 4
opt.shiftwidth = 4
opt.expandtab = true

-- Fuer 2-Space-Projekte
opt.tabstop = 2
opt.shiftwidth = 2
```

### Farbschema aendern

```lua
-- Standard: Tokyo Night
vim.cmd.colorscheme("tokyonight")

-- Alternativen (nach Plugin-Installation):
-- vim.cmd.colorscheme("catppuccin")
-- vim.cmd.colorscheme("gruvbox")
-- vim.cmd.colorscheme("nord")
```

### LSP-Server hinzufuegen

```lua
require("mason-lspconfig").setup({
    ensure_installed = {
        "lua_ls",
        "pyright",
        "ts_ls",
        "rust_analyzer",
        -- Weitere hinzufuegen:
        "gopls",           -- Go
        "clangd",          -- C/C++
        "html",            -- HTML
        "cssls",           -- CSS
        "jsonls",          -- JSON
    },
})
```

### Plugins hinzufuegen

Neue Plugins in `require("lazy").setup({...})` einfuegen:

```lua
-- Beispiel: GitHub Copilot
{
    "github/copilot.vim",
    event = "InsertEnter",
},

-- Beispiel: Markdown-Preview
{
    "iamcco/markdown-preview.nvim",
    build = "cd app && npm install",
    ft = "markdown",
},
```

Nach Aenderungen in Neovim: `:Lazy sync`

---

## Zsh-Konfiguration

**Datei**: `/opt/terminal-dev-env/config/zsh/.zshrc` oder `~/.zshrc`

### Oh-My-Zsh Theme aendern

```bash
# Standard: robbyrussell
ZSH_THEME="robbyrussell"

# Alternativen:
# ZSH_THEME="agnoster"      # Powerline-Style (benoetigt Nerd Font)
# ZSH_THEME="simple"        # Minimalistisch
# ZSH_THEME="powerlevel10k/powerlevel10k"  # Sehr anpassbar
```

### Plugins aktivieren/deaktivieren

```bash
plugins=(
    git              # Git-Aliase und Funktionen
    docker           # Docker-Completion
    docker-compose   # Docker Compose-Completion
    npm              # npm-Completion
    node             # Node.js-Funktionen
    python           # Python-Funktionen
    rust             # Rust/Cargo-Completion
    tmux             # tmux-Integration
    fzf              # Fuzzy-Finder
    z                # Directory-Jumping
    history          # Erweiterte History-Suche
    sudo             # Doppel-ESC fuer sudo
    # Hinzufuegen oder entfernen nach Bedarf
)
```

### Eigene Aliase

Am Ende der `.zshrc` hinzufuegen:

```bash
# Projekt-spezifische Aliase
alias myproject="cd ~/projects/myproject && nvim ."
alias dev="cd ~/dev"

# Schnellzugriff auf haeufige Befehle
alias weather="curl wttr.in"
alias pubip="curl ifconfig.me"
```

### FZF-Optionen

```bash
# Standard-Optionen anpassen
export FZF_DEFAULT_OPTS="
    --height 40%
    --layout=reverse
    --border
    --preview-window=right:50%
"

# Preview-Befehl fuer Dateien
export FZF_CTRL_T_OPTS="--preview 'bat --color=always {}'"
```

---

## SSL-Konfiguration

**Verzeichnis**: `/opt/terminal-dev-env/ssl/`

### Zertifikat-Informationen anzeigen

```bash
openssl x509 -in /opt/terminal-dev-env/ssl/server.crt -text -noout
```

### Neues Zertifikat mit anderen Parametern

```bash
# openssl.cnf bearbeiten
vim /opt/terminal-dev-env/ssl/openssl.cnf
```

Wichtige Felder in `openssl.cnf`:

```ini
[dn]
C = DE                              # Land
ST = Bayern                         # Bundesland
L = Muenchen                        # Stadt
O = Meine Firma                     # Organisation
CN = mein-server.local              # Common Name

[alt_names]
DNS.1 = mein-server.local
DNS.2 = localhost
IP.1 = 127.0.0.1
IP.2 = 192.168.178.78               # Server-IP
IP.3 = 10.0.0.100                   # Weitere IP
```

Neues Zertifikat generieren:

```bash
cd /opt/terminal-dev-env/ssl
sudo openssl req -new -x509 -sha256 \
    -key server.key \
    -out server.crt \
    -days 365 \
    -config openssl.cnf

sudo systemctl restart nginx
```

---

## Firewall-Konfiguration

### UFW (Ubuntu/Debian)

```bash
# Status anzeigen
sudo ufw status verbose

# Standard-Regeln
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Ports oeffnen
sudo ufw allow 22/tcp comment 'SSH'
sudo ufw allow 80/tcp comment 'HTTP'
sudo ufw allow 443/tcp comment 'HTTPS'

# Nur bestimmte IP erlauben
sudo ufw allow from 192.168.178.0/24 to any port 443 comment 'LAN HTTPS'

# Firewall aktivieren
sudo ufw enable
```

### firewalld (RHEL/Fedora)

```bash
# Status anzeigen
sudo firewall-cmd --list-all

# Services hinzufuegen
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --permanent --add-service=ssh

# Port direkt hinzufuegen
sudo firewall-cmd --permanent --add-port=443/tcp

# Aenderungen uebernehmen
sudo firewall-cmd --reload
```

---

## Umgebungsvariablen

### System-weite Variablen

In `/etc/environment` oder `/etc/profile.d/terminal-dev-env.sh`:

```bash
export TERMINAL_DEV_ENV="/opt/terminal-dev-env"
export EDITOR="nvim"
export VISUAL="nvim"
```

### Benutzer-spezifische Variablen

In `~/.zshrc` oder `~/.bashrc`:

```bash
# Node.js
export NPM_CONFIG_PREFIX="$HOME/.npm-global"
export PATH="$HOME/.npm-global/bin:$PATH"

# Go
export GOPATH="$HOME/go"
export PATH="$GOPATH/bin:$PATH"

# Rust
export PATH="$HOME/.cargo/bin:$PATH"

# Python
export PATH="$HOME/.local/bin:$PATH"
```

### ttyd-spezifische Variablen

Im systemd-Service koennen Umgebungsvariablen gesetzt werden:

```ini
[Service]
Environment="TERM=xterm-256color"
Environment="LANG=de_DE.UTF-8"
```

---

## Zusammenfassung der wichtigsten Dateien

| Datei | Beschreibung | Neustart erforderlich |
|-------|--------------|----------------------|
| `/etc/nginx/sites-available/terminal` | nginx-Konfiguration | `systemctl reload nginx` |
| `/etc/systemd/system/ttyd.service` | ttyd-Service | `systemctl daemon-reload && systemctl restart ttyd` |
| `/opt/terminal-dev-env/config/tmux/tmux.conf` | tmux-Konfiguration | Prefix + r |
| `~/.config/nvim/init.lua` | Neovim-Konfiguration | `:source %` oder Neovim neu starten |
| `~/.zshrc` | Zsh-Konfiguration | `source ~/.zshrc` |
| `/opt/terminal-dev-env/ssl/server.crt` | SSL-Zertifikat | `systemctl restart nginx` |

---

Weitere Informationen: [README.md](README.md) | [TROUBLESHOOTING.md](TROUBLESHOOTING.md) | [MOBILE-TIPS.md](MOBILE-TIPS.md)
