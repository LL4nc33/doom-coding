# Smartphone Setup Guide

Access your Doom Coding development environment from your phone or tablet.

## Quick Start

After installation, run this command to display a QR code:

```bash
./scripts/health-check.sh --qr
```

Scan the QR code with your phone's camera to open code-server in your mobile browser.

## Mobile Apps (Recommended)

### Android

**Termux** - Full Linux terminal on Android
- Download: [Google Play Store](https://play.google.com/store/apps/details?id=com.termux)
- Use for: SSH access, running commands, local development

**JuiceSSH** - SSH client with good mobile UI
- Download: [Google Play Store](https://play.google.com/store/apps/details?id=com.sonelli.juicessh)
- Use for: Quick SSH connections

### iOS

**Blink Shell** - Professional SSH and Mosh client
- Download: [App Store](https://apps.apple.com/app/blink-shell-mosh-ssh-client/id1594898306)
- Use for: SSH access with persistent connections

**Termius** - Cross-platform SSH client
- Download: [App Store](https://apps.apple.com/app/termius-terminal-ssh-client/id549039908)
- Use for: SSH with sync across devices

## Access Methods

### 1. Web Browser (Easiest)

1. Run `./scripts/health-check.sh --qr` on your server
2. Scan the QR code with your phone's camera
3. Open the link in Safari/Chrome
4. Log in with your code-server password

**Browser Tips:**
- Add to home screen for app-like experience
- Use landscape mode for better coding
- Connect a Bluetooth keyboard for productivity

### 2. SSH Access

Connect directly to your server via SSH:

```bash
ssh user@your-server-ip
# or with Tailscale
ssh user@100.x.x.x
```

**Recommended SSH apps:**
- Android: JuiceSSH, Termux
- iOS: Blink Shell, Termius

### 3. Tailscale VPN (Recommended for Remote Access)

If you set up with Tailscale:

1. Install Tailscale on your phone
   - [Android](https://play.google.com/store/apps/details?id=com.tailscale.ipn)
   - [iOS](https://apps.apple.com/app/tailscale/id1470499037)
2. Log in to your Tailscale account
3. Access code-server at `https://100.x.x.x:8443`

## Mobile Keyboard Tips

### Using code-server on Mobile

1. **Landscape mode**: Always use landscape for more screen space
2. **Bluetooth keyboard**: Highly recommended for serious coding
3. **External display**: Connect to a larger screen when possible

### Touch Gestures in code-server

- **Pinch to zoom**: Adjust font size
- **Swipe left/right**: Navigate between tabs
- **Long press**: Context menu (right-click)

### Mobile-Friendly VS Code Extensions

Consider installing these extensions for better mobile experience:
- **Settings Sync**: Keep settings across devices
- **Remote Development**: Connect to containers
- **Bracket Pair Colorizer**: Visual code structure

## QR Code Reference

### Generate Access QR

```bash
# Full health check with QR
./scripts/health-check.sh --qr

# Quick QR generation (if qrencode is installed)
echo "https://$(tailscale ip -4):8443" | qrencode -t ansiutf8
```

### Service Setup QR Codes

During installation, scan these QR codes for quick access:

| Service | What it does |
|---------|--------------|
| Tailscale Keys | Create auth key for VPN setup |
| Anthropic Console | Get API key for Claude |
| GitHub Repo | Project documentation |

## Troubleshooting

### Can't Connect from Phone

1. **Same network?** Make sure your phone is on the same WiFi as your server
2. **Tailscale connected?** Check if both devices are on Tailscale
3. **Firewall?** Port 8443 must be open

Run diagnostics:
```bash
./scripts/health-check.sh
```

### Slow Performance

- Use a wired connection on your server
- Close unused browser tabs on your phone
- Consider using SSH instead of web browser for text-only work

### Connection Drops

- Enable "keep alive" in your SSH client
- Use Mosh instead of SSH (more stable on mobile)
- Check Tailscale connection status

## Security Notes

- **Never share QR codes** containing your access URLs publicly
- **Use strong passwords** for code-server
- **Enable 2FA** on your Tailscale account
- **Certificate warnings** are normal for self-signed certificates

## Next Steps

1. [Install mobile apps](#mobile-apps-recommended)
2. [Set up Tailscale](#3-tailscale-vpn-recommended-for-remote-access)
3. [Configure SSH access](#2-ssh-access)
4. Start coding from anywhere!
