# Troubleshooting Guide

## Häufige Probleme und Lösungen

### Installation

#### "ttyd: command not found"

ttyd wurde nicht korrekt installiert. Manuell nachinstallieren:

```bash
# Debian/Ubuntu
sudo apt-get install -y libwebsockets-dev libjson-c-dev libssl-dev cmake

cd /tmp
git clone https://github.com/tsl0922/ttyd.git
cd ttyd
mkdir build && cd build
cmake ..
make
sudo make install
```

#### "nginx: [emerg] bind() to 0.0.0.0:443 failed"

Port 443 bereits belegt:

```bash
# Prozess auf Port 443 finden
sudo ss -tlnp | grep :443
sudo lsof -i :443

# Anderen Dienst stoppen
sudo systemctl stop apache2  # falls Apache läuft
```

### Services

#### ttyd-Service startet nicht

```bash
# Status und Logs
sudo systemctl status ttyd
sudo journalctl -u ttyd -n 100

# Häufige Ursachen:
# 1. ttyd-Binary nicht gefunden
which ttyd
sudo ln -s /usr/local/bin/ttyd /usr/bin/ttyd

# 2. terminal-session.sh nicht ausführbar
chmod +x /opt/terminal-dev-env/bin/terminal-session.sh

# 3. Falscher Pfad in Service-Datei
cat /etc/systemd/system/ttyd.service
```

#### nginx zeigt 502 Bad Gateway

ttyd läuft nicht oder ist nicht erreichbar:

```bash
# ttyd-Status prüfen
systemctl status ttyd

# Port prüfen
ss -tlnp | grep 7681

# ttyd manuell testen
ttyd --port 7681 bash
```

### SSL/TLS

#### Browser zeigt "NET::ERR_CERT_INVALID"

Bei self-signed Zertifikaten normal. Umgehung:
- Chrome: "Advanced" → "Proceed to site"
- Firefox: "Advanced" → "Accept the Risk"

#### Zertifikat abgelaufen

```bash
# Ablaufdatum prüfen
openssl x509 -enddate -noout -in /opt/terminal-dev-env/ssl/server.crt

# Neues Zertifikat generieren
cd /opt/terminal-dev-env/ssl
openssl req -new -x509 -sha256 \
    -key server.key -out server.crt \
    -days 365 -config openssl.cnf

sudo systemctl restart nginx
```

### Verbindungsprobleme

#### Keine Verbindung von externem Gerät

1. Firewall-Regeln prüfen:
```bash
sudo ufw status verbose
# Port 443 muss ALLOW sein
```

2. IP-Adresse prüfen:
```bash
ip addr show | grep inet
# 192.168.178.78 sollte erscheinen
```

3. nginx bindet auf richtige IP:
```bash
ss -tlnp | grep nginx
# Sollte 0.0.0.0:443 oder *:443 zeigen
```

#### WebSocket-Verbindung bricht ab

1. nginx Timeout erhöhen (bereits konfiguriert auf 7 Tage)
2. Netzwerk-Qualität prüfen
3. Browser-Konsole auf Fehler prüfen

### tmux

#### "no server running on /tmp/tmux-1000/default"

tmux-Server nicht gestartet:

```bash
# Neue Session starten
tmux new -s dev

# Oder über terminal-session.sh
/opt/terminal-dev-env/bin/terminal-session.sh
```

#### Session-Persistenz funktioniert nicht

tmux-resurrect Plugin prüfen:

```bash
# Plugin installieren
git clone https://github.com/tmux-plugins/tpm ~/.tmux/plugins/tpm

# In tmux: Prefix + I (Plugins installieren)
# Prefix + Ctrl+s (Session speichern)
# Prefix + Ctrl+r (Session wiederherstellen)
```

### WSL2-spezifisch

#### Port-Forwarding funktioniert nicht

```powershell
# Als Administrator in PowerShell
# WSL2-IP abrufen
wsl hostname -I

# Port-Proxy hinzufügen
netsh interface portproxy add v4tov4 listenport=443 listenaddress=0.0.0.0 connectport=443 connectaddress=<WSL2-IP>

# Firewall-Regel hinzufügen
New-NetFirewallRule -DisplayName "WSL2 HTTPS" -Direction Inbound -Action Allow -Protocol TCP -LocalPort 443
```

#### systemd nicht verfügbar in WSL2

```bash
# WSL-Version prüfen (muss >= 0.67.6 sein)
wsl --version

# systemd aktivieren
sudo nano /etc/wsl.conf
# [boot]
# systemd=true

# WSL neu starten (in PowerShell)
wsl --shutdown
wsl
```

### Performance

#### Hohe CPU-Auslastung

```bash
# Prozesse prüfen
top -c

# ttyd mit weniger Clients
# In /etc/systemd/system/ttyd.service:
# --max-clients 2 statt 5
```

#### Hoher Speicherverbrauch

```bash
# Speicherverbrauch prüfen
ps aux --sort=-%mem | head -10

# Logs rotieren
sudo logrotate -f /etc/logrotate.conf

# tmux-History reduzieren
# In tmux.conf: set -g history-limit 10000 statt 50000
```

### Authentifizierung

#### Passwort vergessen

```bash
# Neues Passwort setzen
htpasswd -c /opt/terminal-dev-env/config/nginx/.htpasswd admin

# Oder mit OpenSSL
echo "admin:$(openssl passwd -apr1 neuespasswort)" > /opt/terminal-dev-env/config/nginx/.htpasswd
```

#### Basic Auth funktioniert nicht

```bash
# htpasswd-Datei prüfen
cat /opt/terminal-dev-env/config/nginx/.htpasswd

# Berechtigungen prüfen
ls -la /opt/terminal-dev-env/config/nginx/.htpasswd
# Sollte 600 und root:root sein

# nginx-Config prüfen
grep -A2 "auth_basic" /etc/nginx/sites-enabled/terminal.conf
```

## Diagnose-Befehle

```bash
# Vollständiger System-Check
/opt/terminal-dev-env/bin/health-check.sh

# Service-Status
systemctl status ttyd nginx

# Netzwerk-Verbindungen
ss -tlnp

# Logs in Echtzeit
tail -f /opt/terminal-dev-env/logs/*.log

# nginx-Konfiguration testen
nginx -t

# SSL-Zertifikat-Info
openssl x509 -in /opt/terminal-dev-env/ssl/server.crt -text -noout
```

## Support

Bei weiteren Problemen:
1. Health Check ausführen
2. Logs sammeln
3. Issue mit Logs erstellen
