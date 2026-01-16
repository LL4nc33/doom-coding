# Basic Configuration

Guide to configuring your Doom Coding environment.

## Environment Variables

All configuration is managed through the `.env` file.

### Creating Your Configuration

```bash
# Copy the template
cp .env.example .env

# Edit with your values
vim .env
```

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `TS_AUTHKEY` | Tailscale authentication key | `tskey-auth-xxx` |
| `CODE_SERVER_PASSWORD` | Web IDE login password | `secure-password-123` |
| `SUDO_PASSWORD` | Sudo password in containers | `sudo-password-456` |
| `PUID` | User ID for file permissions | `1000` |
| `PGID` | Group ID for file permissions | `1000` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TZ` | Timezone | `Europe/Berlin` |
| `TS_ACCEPT_DNS` | Use Tailscale DNS | `false` |
| `TS_EXTRA_ARGS` | Additional Tailscale arguments | `--advertise-tags=tag:doom-coding` |
| `WORKSPACE_PATH` | Projects directory | `./workspace` |

## Tailscale Configuration

### Getting an Auth Key

1. Go to https://login.tailscale.com/admin/settings/keys
2. Click "Generate auth key"
3. Configure options:
   - **Reusable**: Enable for multiple devices
   - **Ephemeral**: Disable for persistent devices
   - **Tags**: Add `tag:doom-coding`
4. Copy the key to your `.env` file

### OAuth vs Auth Keys

For production, consider using OAuth:

```yaml
# docker-compose.yml
environment:
  - TS_OAUTH_SECRET=your-oauth-secret
```

### DNS Configuration

If you experience DNS issues:

```bash
# In .env
TS_ACCEPT_DNS=false
```

## code-server Configuration

### Password Authentication

Set a strong password in `.env`:

```bash
CODE_SERVER_PASSWORD=your-very-secure-password
```

### Extensions

code-server uses Open VSX marketplace. To install extensions:

1. Search at https://open-vsx.org
2. Install via UI or command line:
   ```bash
   docker exec doom-code-server code-server --install-extension publisher.extension
   ```

### Custom Settings

Mount a custom settings file:

```yaml
# docker-compose.yml
volumes:
  - ./config/vscode/settings.json:/config/.local/share/code-server/User/settings.json
```

## Claude Code Configuration

### API Key Setup

1. Get your API key from https://console.anthropic.com
2. Add to secrets:
   ```bash
   echo "sk-ant-xxx" > secrets/anthropic_api_key.txt
   ```

### Automation Mode

For unattended operations:

```yaml
# docker-compose.yml
command: ["claude", "--dangerously-skip-permissions"]
```

**Warning**: Only use automation mode in trusted environments.

## Volume Configuration

### Workspace Directory

Configure your projects directory:

```bash
# In .env
WORKSPACE_PATH=/home/user/projects
```

### Persistent Data

All persistent data is stored in Docker volumes:

- `doom-tailscale-state`: Tailscale state
- `doom-code-server-config`: code-server configuration
- `doom-claude-config`: Claude Code configuration

### Backup Volumes

```bash
# Backup
docker run --rm -v doom-code-server-config:/data -v $(pwd):/backup alpine tar czf /backup/code-server-backup.tar.gz /data

# Restore
docker run --rm -v doom-code-server-config:/data -v $(pwd):/backup alpine tar xzf /backup/code-server-backup.tar.gz -C /
```

## Network Configuration

### Ports

Default ports used (via Tailscale):

| Service | Port |
|---------|------|
| code-server | 8443 |
| SSH | 22 |

### Custom Ports

To use different ports, modify `docker-compose.yml`:

```yaml
code-server:
  environment:
    - PORT=9000
```

## File Permissions

### PUID/PGID

Find your user IDs:

```bash
id -u  # PUID
id -g  # PGID
```

Set in `.env`:

```bash
PUID=1000
PGID=1000
```

### Fixing Permissions

If you encounter permission issues:

```bash
# Fix workspace permissions
sudo chown -R $(id -u):$(id -g) ./workspace

# Restart containers
docker compose restart
```

## Timezone

Set your timezone:

```bash
# In .env
TZ=Europe/Berlin
```

Valid timezones: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

## Configuration Validation

After making changes, validate your configuration:

```bash
# Check Docker Compose syntax
docker compose config

# Run health check
./scripts/health-check.sh
```

## Next Steps

- [Advanced Configuration](advanced-config.md)
- [Security Hardening](../security/hardening.md)
- [Secrets Management](../security/secrets-management.md)