# Mobile-Nutzungstipps

Optimale Nutzung des Terminal Development Environment auf Smartphone-Browsern.

## Inhaltsverzeichnis

1. [Browser-Empfehlungen](#browser-empfehlungen)
2. [Bildschirm-Orientierung](#bildschirm-orientierung)
3. [Tastatur-Tipps](#tastatur-tipps)
4. [Touch-Gesten](#touch-gesten)
5. [tmux auf Mobile](#tmux-auf-mobile)
6. [Neovim auf Mobile](#neovim-auf-mobile)
7. [Verbindungs-Management](#verbindungs-management)
8. [Akku-Optimierung](#akku-optimierung)
9. [Haeufige Probleme](#haeufige-probleme)

---

## Browser-Empfehlungen

### Android

| Browser | Empfehlung | Anmerkungen |
|---------|------------|-------------|
| Chrome | Sehr gut | Beste WebSocket-Unterstuetzung |
| Firefox | Gut | Gute Touch-Unterstuetzung |
| Samsung Internet | Gut | Gute Performance auf Samsung-Geraeten |
| Brave | Gut | Chrome-basiert, schnell |

### iOS

| Browser | Empfehlung | Anmerkungen |
|---------|------------|-------------|
| Safari | Sehr gut | Beste Integration, native WebSocket |
| Chrome | Gut | Nutzt Safari WebKit Engine |
| Firefox | Gut | Nutzt Safari WebKit Engine |

**Empfehlung**: Chrome (Android) oder Safari (iOS) fuer beste Erfahrung.

### Browser-Einstellungen

1. **Desktop-Modus deaktivieren** - Mobile-Ansicht nutzen
2. **JavaScript aktiviert** - Erforderlich fuer ttyd
3. **Cookies erlauben** - Fuer Session-Management
4. **Pop-up-Blocker deaktivieren** - Falls Probleme auftreten

---

## Bildschirm-Orientierung

### Querformat (Landscape) - Empfohlen

Vorteile:
- Mehr horizontaler Platz fuer Code
- Bessere Lesbarkeit
- Aehnlicher zu Desktop-Erfahrung

```
+--------------------------------------------------+
| tmux-Statusbar                                    |
+--------------------------------------------------+
|                                                   |
|  Code / Terminal                                  |
|  (Mehr Platz fuer lange Zeilen)                  |
|                                                   |
+--------------------------------------------------+
| Tastatur (kleinerer Bereich)                      |
+--------------------------------------------------+
```

### Hochformat (Portrait)

Geeignet fuer:
- Schnelle Befehle
- Log-Anzeige
- Vertikale Splits

```
+------------------------+
| tmux-Statusbar         |
+------------------------+
|                        |
|  Terminal/Code         |
|  (Weniger Zeichen      |
|   pro Zeile)           |
|                        |
+------------------------+
| Tastatur               |
| (Mehr Platz)           |
+------------------------+
```

**Tipp**: Bildschirm-Rotation-Sperre aktivieren waehrend der Arbeit.

---

## Tastatur-Tipps

### Externe Bluetooth-Tastatur

Beste Option fuer laengere Coding-Sessions:
- Volle Tastatur mit allen Modifier-Tasten
- Keine Bildschirm-Verdeckung
- Schnelleres Tippen

Empfohlene kompakte Tastaturen:
- Logitech K380
- Apple Magic Keyboard
- Keychron K3

### On-Screen-Tastatur

#### Empfohlene Drittanbieter-Tastaturen

**Android**:
- **Hacker's Keyboard**: Volle Tastatur mit Ctrl, Alt, Esc, F-Tasten
- **Termux:API Keyboard**: Optimiert fuer Terminal
- **Microsoft SwiftKey**: Gute Autokorrektur, aber Standard-Layout

**iOS**:
- Standard-Tastatur funktioniert gut
- Bluetooth-Tastatur empfohlen fuer erweiterte Tasten

#### Hacker's Keyboard Setup (Android)

1. Aus Play Store installieren
2. Als Eingabemethode aktivieren
3. Einstellungen > Keyboard mode > 5-row compact
4. Einstellungen > Key behavior > Ctrl/Alt/Meta aktivieren

### Wichtige Sonderzeichen

| Zeichen | Verwendung | Eingabe-Tipps |
|---------|------------|---------------|
| `Ctrl` | tmux-Prefix, vim-Commands | Hacker's Keyboard |
| `Esc` | vim-Normal-Mode | `jk` oder `kj` in vim konfiguriert |
| `Tab` | Einrueckung, Completion | Standard-Taste |
| `|` (Pipe) | Unix-Pipes, tmux-Split | Oft unter Sonderzeichen |
| `~` | Home-Verzeichnis | Oft unter Sonderzeichen |
| `` ` `` (Backtick) | Shell-Substitution | Oft unter Sonderzeichen |

### Escape-Taste Alternativen

Da `Esc` auf Mobile schwer erreichbar ist:

1. **In Neovim**: `jk` oder `kj` schnell tippen (vorkonfiguriert)
2. **In tmux**: Copy-Mode mit `Prefix + Enter` statt `Prefix + [`
3. **Allgemein**: `Ctrl + [` ist aequivalent zu `Esc`

---

## Touch-Gesten

### ttyd Touch-Support

| Geste | Aktion |
|-------|--------|
| Tippen | Cursor positionieren |
| Wischen hoch/runter | Scrollen im Terminal |
| Pinch-Zoom | Schriftgroesse aendern |
| Doppeltippen | Wort auswaehlen |
| Langes Druecken | Kontextmenue (Copy/Paste) |

### tmux Touch-Support

| Geste | Aktion (mit Mouse-Support) |
|-------|---------------------------|
| Tippen auf Pane | Pane auswaehlen |
| Tippen auf Fenster | Fenster wechseln (Statusbar) |
| Wischen | Scrollen in History |
| Ziehen am Rand | Pane-Groesse aendern |

---

## tmux auf Mobile

### Mobile-optimierte Konfiguration (vorkonfiguriert)

Die mitgelieferte `tmux.conf` ist bereits fuer Mobile optimiert:

- **Prefix**: `Ctrl+a` (einfacher als `Ctrl+b`)
- **Maus**: Vollstaendig aktiviert
- **Pane-Wechsel**: `Alt+Pfeiltaste` (ohne Prefix)
- **Fenster-Wechsel**: `Shift+Pfeiltaste` (ohne Prefix)
- **Auto-Save**: Alle 5 Minuten

### Empfohlenes Layout fuer Mobile

Einfaches Layout ohne zu viele Splits:

```
# Maximal 2 Panes nebeneinander
+----------------------+----------------------+
|                      |                      |
|   Editor (nvim)      |   Terminal/Logs      |
|                      |                      |
+----------------------+----------------------+
```

Befehle:

```bash
# Horizontal teilen (nebeneinander)
Prefix + |

# Vertikal teilen (uebereinander)
Prefix + -

# Nur ein Pane maximieren
Prefix + z

# Zurueck zu normal
Prefix + z
```

### Schnelle Fenster-Navigation

```bash
# Fenster 1-5 direkt anspringen
Alt+1, Alt+2, Alt+3, Alt+4, Alt+5

# Naechstes/Vorheriges Fenster
Shift+Rechts, Shift+Links
```

### Session-Persistenz

Bei Verbindungsabbruch bleibt die Session erhalten:

```bash
# Session manuell speichern (vor kritischen Aktionen)
Prefix + Ctrl+s

# Session wiederherstellen (nach Reconnect)
Prefix + Ctrl+r

# Liste aller Sessions
Prefix + s
```

---

## Neovim auf Mobile

### Mobile-freundliche Keymaps (vorkonfiguriert)

| Aktion | Standard | Mobile-Alternative |
|--------|----------|-------------------|
| Normal-Mode | `Esc` | `jk` oder `kj` |
| Speichern | `:w` | `Space + w` oder `Ctrl+s` |
| Beenden | `:q` | `Space + q` |
| File-Explorer | `:NvimTreeToggle` | `Space + e` |
| Datei suchen | `:Telescope find_files` | `Space + ff` |

### Which-Key Plugin

Zeigt verfuegbare Keybindings nach Druecken von `Space`:

```
Nach "Space" druecken erscheint:
+------------------------------------+
| e  -> File Explorer                |
| ff -> Find Files                   |
| fg -> Find Text (Grep)             |
| w  -> Save                         |
| q  -> Quit                         |
| ...                                |
+------------------------------------+
```

### Empfohlene Einstellungen fuer Mobile

Falls noch nicht konfiguriert:

```lua
-- Groesserer Scroll-Offset (mehr Kontext)
opt.scrolloff = 10
opt.sidescrolloff = 10

-- Cursorline fuer bessere Orientierung
opt.cursorline = true

-- Relative Zeilennummern (optional deaktivieren auf Mobile)
opt.relativenumber = false  -- Weniger visuelles Rauschen
```

### Dateien schnell finden

Mit Telescope (Space + ff):

1. `Space + ff` druecken
2. Dateiname tippen (fuzzy matching)
3. Mit Pfeiltasten navigieren
4. Enter zum Oeffnen

### Text suchen im Projekt

Mit Telescope (Space + fg):

1. `Space + fg` druecken
2. Suchbegriff eingeben
3. Ergebnisse durchsuchen
4. Enter zum Springen

---

## Verbindungs-Management

### Automatische Wiederverbindung

ttyd versucht automatisch, die Verbindung wiederherzustellen.

Falls die Verbindung abbricht:
1. Warten (automatischer Reconnect-Versuch)
2. Seite neu laden (tmux-Session bleibt erhalten)
3. Browser-Tab schliessen und neu oeffnen

### Verbindung stabil halten

1. **WLAN**: Stabiles 5GHz-Netzwerk bevorzugen
2. **Mobile Daten**: 4G/LTE ausreichend, 5G optimal
3. **Energiesparmodus**: Kann Verbindung unterbrechen - deaktivieren
4. **Browser im Hintergrund**: Verbindung kann getrennt werden

### Vor wichtigen Aenderungen

```bash
# Session manuell speichern
Prefix + Ctrl+s

# Oder Dateien in vim speichern
:wa  (alle Dateien speichern)
```

---

## Akku-Optimierung

### Browser-Einstellungen

- Dunkles Theme verwenden (OLED-Displays)
- Hardware-Beschleunigung aktivieren
- Bildschirmhelligkeit reduzieren

### Terminal-Einstellungen

Das vorkonfigurierte Tokyo-Night-Theme ist bereits dunkel und Akku-schonend.

### Verbindungs-Intervall

ttyd sendet Pings alle 30 Sekunden. Bei sehr langen Sessions ohne Aktivitaet:

```bash
# In einer tmux-Pane: Uhr anzeigen (haelt Verbindung aktiv)
watch -n 60 date
```

---

## Haeufige Probleme

### Problem: Tastatur verdeckt Terminal

**Loesung**:
1. Bildschirm nach oben scrollen
2. Querformat nutzen
3. Externe Tastatur verwenden

### Problem: Sonderzeichen nicht verfuegbar

**Loesung**:
1. Hacker's Keyboard installieren (Android)
2. Bluetooth-Tastatur verwenden
3. Alternative Eingaben:
   - `Ctrl+[` statt `Esc`
   - Compose-Sequenzen nutzen

### Problem: Verbindung bricht staendig ab

**Loesungen**:
1. Stabiles WLAN nutzen
2. Energiesparmodus deaktivieren
3. Browser im Vordergrund halten
4. tmux-Ping-Intervall pruefen

### Problem: Touch-Scrollen funktioniert nicht

**Loesung**:
1. Maus-Support in tmux pruefen: `set -g mouse on`
2. Browser aktualisieren
3. Seite neu laden

### Problem: Schrift zu klein/gross

**Loesung**:
1. Pinch-Zoom im Browser
2. Browser-Einstellungen > Textgroesse
3. Terminal: `Ctrl+Plus` / `Ctrl+Minus`

### Problem: Copy/Paste funktioniert nicht

**Loesung Android**:
1. Langes Druecken fuer Kontextmenue
2. In tmux: `Prefix + [` fuer Copy-Mode
3. Text auswaehlen, `y` zum Kopieren
4. `Prefix + ]` zum Einfuegen

**Loesung iOS**:
1. Doppeltippen fuer Auswahl
2. Standard-iOS-Kontextmenue nutzen

---

## Quick Reference Card

Zum Ausdrucken oder Screenshot:

```
+------------------------------------------+
|    MOBILE TERMINAL QUICK REFERENCE       |
+------------------------------------------+
| TMUX                                     |
|   Prefix:        Ctrl+a                  |
|   Pane wechseln: Alt+Pfeiltaste         |
|   Fenster:       Shift+Pfeiltaste       |
|   Fenster 1-5:   Alt+1 bis Alt+5        |
|   Split |:       Prefix + |              |
|   Split -:       Prefix + -              |
|   Maximize:      Prefix + z              |
|   Save Session:  Prefix + Ctrl+s         |
|   Restore:       Prefix + Ctrl+r         |
+------------------------------------------+
| NEOVIM                                   |
|   Leader:        Space                   |
|   Escape:        jk oder kj              |
|   Save:          Space+w oder Ctrl+s     |
|   Quit:          Space+q                 |
|   File Explorer: Space+e                 |
|   Find File:     Space+ff                |
|   Find Text:     Space+fg                |
|   Buffers:       Shift+h/l               |
+------------------------------------------+
| ALLGEMEIN                                |
|   Escape-Alternative: Ctrl+[             |
|   Session pruefen: tmux list-sessions    |
|   Health Check: tde-status               |
+------------------------------------------+
```

---

Weitere Informationen: [README.md](README.md) | [CONFIGURATION.md](CONFIGURATION.md) | [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
