# Terminal Development Environment v0.0.2

Browser-zugaengliche Terminal-Entwicklungsumgebung basierend auf ttyd + tmux + neovim + Claude CLI.

## Uebersicht

Diese Loesung bietet eine leichtgewichtige Alternative zu VS Code/code-server mit folgenden Vorteilen:

- **Ressourcenverbrauch**: ~200MB RAM (vs. 1GB+ bei code-server)
- **Mobile-Optimiert**: Touch-freundliche Konfiguration fuer Smartphone-Browser
- **Bare-Metal**: systemd-Services, kein Docker erforderlich
- **Sicherheit**: HTTPS, Basic Auth, Firewall-Integration

## Architektur

```
Smartphone Browser (HTTPS)
         |
         v
   [nginx:443]         -- SSL-Terminierung, Basic Auth, Rate Limiting
         |
         v
   [ttyd:7681]         -- WebSocket Terminal-Daemon (nur localhost)
         |
         v
   [tmux]              -- Session-Multiplexer mit Persistenz
         |
    +----+----+
    |         |
    v         v
  [zsh]    [neovim]    -- Shell und Editor
```

## Voraussetzungen

### Linux (Debian/Ubuntu)

- Debian 11+ oder Ubuntu 20.04+
- Root-Zugriff (sudo)
- Mindestens 512MB RAM
- Mindestens 1GB freier Speicherplatz
- Netzwerkzugriff auf Port 80 und 443

### WSL2 (Windows)

- Windows 10 Version 2004+ oder Windows 11
- WSL2 mit Ubuntu 20.04+ oder Debian 11+
- WSL Version 0.67.6+ (fuer systemd-Support)
- Administrator-Rechte fuer Port-Forwarding

## Installation

### Verfuegbare Setup-Skripte

| Skript | Verwendung |
|--------|------------|
| `install.sh` | **Haupt-Installer** - Vollstaendige Installation mit CLI-Optionen |
| `setup-linux.sh` | Linux-spezifisches Setup (wird von install.sh aufgerufen) |
| `setup-windows.sh` | WSL2-spezifisches Setup inkl. Windows Port-Forwarding |
| `health-check.sh` | System-Diagnose und Statusueberpruefung |

### Option A: Linux (Bare-Metal / VM)

#### Schnellinstallation (empfohlen)

```bash
# 1. Repository klonen
git clone https://github.com/LL4nc33/doom-coding.git /tmp/doom-coding
sudo cp -r /tmp/doom-coding/terminal-dev-env /opt/terminal-dev-env

# 2. Haupt-Installer ausfuehren
sudo bash /opt/terminal-dev-env/bin/install.sh
```

#### Installation mit benutzerdefinierten Optionen

```bash
# Mit eigenem Benutzernamen und Passwort
sudo bash /opt/terminal-dev-env/bin/install.sh -u meinuser -p meinpasswort

# Mit eigener Domain fuer SSL-Zertifikat
sudo bash /opt/terminal-dev-env/bin/install.sh -d meine-domain.local

# Ohne Firewall-Konfiguration (falls bereits konfiguriert)
sudo bash /opt/terminal-dev-env/bin/install.sh --skip-firewall
```

#### Nur Linux-Setup (ohne CLI-Optionen)

Falls nur das Linux-spezifische Setup benoetigt wird:

```bash
sudo bash /opt/terminal-dev-env/bin/setup-linux.sh
```

#### Installer-Optionen (install.sh)

| Option | Beschreibung |
|--------|--------------|
| `-u, --user USER` | Benutzername fuer Web-Authentifizierung (Standard: admin) |
| `-p, --pass PASS` | Passwort (wird generiert wenn nicht angegeben) |
| `-d, --domain DOMAIN` | Domain fuer SSL-Zertifikat (Standard: localhost) |
| `--no-ssl` | SSL deaktivieren (nicht empfohlen) |
| `--skip-firewall` | Firewall-Konfiguration ueberspringen |
| `--force` | Neuinstallation erzwingen |
| `--health-check` | Nur Health Check ausfuehren |

### Option B: WSL2 (Windows)

#### Schritt 1: Installation in WSL2

```bash
# In WSL2-Terminal (Ubuntu/Debian)
git clone https://github.com/LL4nc33/doom-coding.git /tmp/doom-coding
sudo cp -r /tmp/doom-coding/terminal-dev-env /opt/terminal-dev-env

# WSL2-Setup ausfuehren (konfiguriert auch Windows-Skripte)
sudo bash /opt/terminal-dev-env/bin/setup-windows.sh
```

#### Schritt 2: Windows Port-Forwarding einrichten

Nach der Installation in WSL2 muss Port-Forwarding auf Windows konfiguriert werden:

```powershell
# In Windows PowerShell (als Administrator!)
C:\terminal-dev-env\setup-portforward.bat
```

Das Skript:
- Ermittelt automatisch die WSL2-IP-Adresse
- Richtet Port-Forwarding fuer Port 80 und 443 ein
- Konfiguriert Windows Firewall-Regeln

#### Schritt 3: WSL2 neu starten (nur beim ersten Setup)

```powershell
# In PowerShell
wsl --shutdown
wsl
```

#### Automatisches Port-Forwarding bei Windows-Start (optional)

```powershell
# Als Administrator ausfuehren
C:\terminal-dev-env\create-scheduled-task.ps1
```

Dies erstellt eine geplante Aufgabe, die das Port-Forwarding automatisch bei jedem Windows-Start aktualisiert (wichtig, da sich die WSL2-IP aendern kann).

## Zugriff

Nach erfolgreicher Installation:

| Zugriff | URL |
|---------|-----|
| Lokales Netzwerk | https://192.168.178.78/ |
| Localhost | https://localhost/ |
| Vom Windows-Host (WSL2) | https://localhost/ |

**Credentials**: Die Zugangsdaten werden bei der Installation angezeigt und gespeichert in:
```
/opt/terminal-dev-env/config/nginx/.htpasswd.plain
```

**Hinweis**: Das selbst-signierte SSL-Zertifikat erzeugt eine Browser-Warnung. Diese kann ignoriert werden ("Advanced" -> "Proceed to site").

## Komponenten

| Komponente | Version | Funktion |
|------------|---------|----------|
| ttyd | 1.7.4+ | Web-Terminal-Daemon (WebSocket) |
| nginx | 1.18+ | Reverse Proxy, SSL, Auth |
| tmux | 3.0+ | Terminal-Multiplexer mit Session-Persistenz |
| neovim | 0.8+ | Text-Editor (Lua-Config) |
| zsh | 5.8+ | Shell mit Oh-My-Zsh |

## Verzeichnisstruktur

```
/opt/terminal-dev-env/
|-- bin/
|   |-- install.sh           # Haupt-Installer
|   |-- setup-linux.sh       # Linux-spezifisches Setup
|   |-- setup-windows.sh     # WSL2-spezifisches Setup
|   |-- health-check.sh      # System-Diagnose
|   +-- terminal-session.sh  # tmux-Session-Wrapper
|-- config/
|   |-- nginx/
|   |   |-- terminal.conf    # nginx-Konfiguration
|   |   |-- .htpasswd        # Authentifizierungs-Datei
|   |   +-- .htpasswd.plain  # Klartext-Passwort (nur lokal!)
|   |-- tmux/
|   |   +-- tmux.conf        # tmux-Konfiguration (Mobile-optimiert)
|   |-- neovim/
|   |   +-- init.lua         # Neovim-Konfiguration
|   +-- zsh/
|       +-- .zshrc           # Zsh-Konfiguration
|-- ssl/
|   |-- server.crt           # SSL-Zertifikat
|   |-- server.key           # SSL-Private-Key
|   +-- openssl.cnf          # OpenSSL-Konfiguration
|-- logs/
|   |-- nginx-access.log     # Zugriffs-Log
|   |-- nginx-error.log      # Fehler-Log
|   |-- ttyd.log             # ttyd-Log
|   +-- install.log          # Installations-Log
+-- docs/
    |-- README.md            # Diese Datei
    |-- CONFIGURATION.md     # Konfigurations-Details
    |-- MOBILE-TIPS.md       # Mobile-Nutzungs-Tipps
    +-- TROUBLESHOOTING.md   # Problemloesung
```

## Service-Management

```bash
# Status pruefen
sudo systemctl status ttyd nginx

# Services starten
sudo systemctl start ttyd nginx

# Services stoppen
sudo systemctl stop ttyd nginx

# Services neustarten
sudo systemctl restart ttyd nginx

# Boot-Autostart aktivieren
sudo systemctl enable ttyd nginx

# Logs anzeigen (systemd)
sudo journalctl -u ttyd -f
sudo journalctl -u nginx -f

# Logs anzeigen (Dateien)
tail -f /opt/terminal-dev-env/logs/*.log

# Health Check ausfuehren
/opt/terminal-dev-env/bin/health-check.sh
```

## Schnellreferenz: Tastenkombinationen

### tmux

| Aktion | Tastenkombination |
|--------|-------------------|
| Prefix (statt Ctrl+b) | `Ctrl+a` |
| Pane horizontal teilen | `Prefix + |` |
| Pane vertikal teilen | `Prefix + -` |
| Pane wechseln | `Alt + Pfeiltaste` (ohne Prefix) |
| Fenster wechseln | `Shift + Pfeiltaste` (ohne Prefix) |
| Fenster 1-5 direkt | `Alt + 1-5` |
| Config neu laden | `Prefix + r` |
| Session speichern | `Prefix + Ctrl+s` |
| Session laden | `Prefix + Ctrl+r` |

### Neovim

| Aktion | Tastenkombination |
|--------|-------------------|
| Leader-Taste | `Space` |
| Speichern | `Space + w` oder `Ctrl+s` |
| Beenden | `Space + q` |
| Datei-Explorer | `Space + e` |
| Datei suchen | `Space + ff` |
| Text suchen | `Space + fg` |
| Buffer wechseln | `Shift + h/l` |
| Split vertikal | `Space + v` |
| Split horizontal | `Space + s` |
| Escape (Insert-Mode) | `jk` oder `kj` |

## SSL-Zertifikate

### Standard (Self-Signed)

Wird automatisch bei der Installation generiert (365 Tage gueltig).

### Zertifikat erneuern

```bash
cd /opt/terminal-dev-env/ssl
sudo openssl req -new -x509 -sha256 \
    -key server.key \
    -out server.crt \
    -days 365 \
    -config openssl.cnf

sudo systemctl restart nginx
```

### Let's Encrypt (optional)

Fuer oeffentlich zugaengliche Server mit Domain:

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

## Sicherheitshinweise

1. **Netzwerk-Isolation**: Der Dienst sollte nur im lokalen Netzwerk oder ueber VPN erreichbar sein
2. **Passwoerter**: Starkes Passwort verwenden und regelmaessig aendern
3. **Updates**: System und Komponenten regelmaessig aktualisieren
4. **Logs**: Zugriffs-Logs regelmaessig auf verdaechtige Aktivitaeten pruefen
5. **Firewall**: Nur notwendige Ports oeffnen (80, 443, 22)
6. **Self-Signed SSL**: Nur fuer lokales Netzwerk geeignet

## Deinstallation

```bash
# Services stoppen
sudo systemctl stop ttyd nginx
sudo systemctl disable ttyd nginx

# Service-Dateien entfernen
sudo rm /etc/systemd/system/ttyd.service
sudo systemctl daemon-reload

# nginx-Konfiguration entfernen
sudo rm /etc/nginx/sites-enabled/terminal
sudo rm /etc/nginx/sites-available/terminal
sudo systemctl restart nginx

# Installationsverzeichnis entfernen
sudo rm -rf /opt/terminal-dev-env

# (Optional) Benutzer-Konfigurationen entfernen
rm -rf ~/.config/nvim
rm -rf ~/.oh-my-zsh
rm ~/.zshrc
rm -rf ~/.tmux
```

## Weitere Dokumentation

- [CONFIGURATION.md](CONFIGURATION.md) - Detaillierte Konfigurations-Optionen
- [MOBILE-TIPS.md](MOBILE-TIPS.md) - Tipps fuer Mobile-Nutzung
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Problemloesung

## Version History

| Version | Datum | Aenderungen |
|---------|-------|-------------|
| v0.0.2 | 2025-01 | Initiale Implementation mit Linux + WSL2 Support |
| v0.0.1 | - | Konzept und Planung |

## Lizenz

MIT License
