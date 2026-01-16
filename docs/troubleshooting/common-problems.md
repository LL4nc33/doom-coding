# Troubleshooting Guide

Solutions for common issues with Doom Coding.

## Quick Diagnostics

Run the health check first:

```bash
./scripts/health-check.sh
```

Check logs:

```bash
# Installation log
tail -f /var/log/doom-coding-install.log

# Docker logs
docker compose logs -f
```

## Docker Issues

### Docker Not Starting

**Symptoms**: `Cannot connect to the Docker daemon`

**Solutions**:

1. Start Docker service:
   ```bash
   sudo systemctl start docker
   ```

2. Check Docker status:
   ```bash
   sudo systemctl status docker
   ```

3. Fix permissions:
   ```bash
   sudo usermod -aG docker $USER
   newgrp docker
   ```

### Container Won't Start

**Symptoms**: Container exits immediately

**Diagnose**:
```bash
docker compose logs <service-name>
```

**Common causes**:

1. **Missing secrets file**:
   ```bash
   # Create placeholder
   echo "placeholder" > secrets/anthropic_api_key.txt
   ```

2. **Port already in use**:
   ```bash
   # Find process using port
   sudo lsof -i :8443

   # Kill or change port in docker-compose.yml
   ```

3. **Insufficient permissions**:
   ```bash
   # Fix ownership
   sudo chown -R $(id -u):$(id -g) ./workspace ./secrets
   ```

### Build Failures

**Symptoms**: `docker compose build` fails

**Solutions**:

1. Clear build cache:
   ```bash
   docker compose build --no-cache
   ```

2. Pull fresh base images:
   ```bash
   docker compose pull
   docker compose build
   ```

3. Check disk space:
   ```bash
   df -h
   docker system prune -af
   ```

## Tailscale Issues

### Not Connecting

**Symptoms**: Tailscale shows "Stopped" or no IP

**Solutions**:

1. Re-authenticate:
   ```bash
   sudo tailscale up --auth-key=YOUR_KEY
   ```

2. Check service status:
   ```bash
   sudo systemctl status tailscaled
   ```

3. View Tailscale logs:
   ```bash
   journalctl -u tailscaled -f
   ```

### DNS Resolution Fails

**Symptoms**: Cannot resolve hostnames inside containers

**Solution**: Disable Tailscale DNS:

```bash
# In .env
TS_ACCEPT_DNS=false
```

Then restart:
```bash
docker compose down
docker compose up -d
```

### Container Can't Reach Internet

**Symptoms**: Containers have Tailscale IP but no internet

**Solutions**:

1. Check Tailscale exit node:
   ```bash
   tailscale status
   ```

2. Verify Docker network:
   ```bash
   docker network ls
   docker network inspect doom-coding-network
   ```

3. Test from container:
   ```bash
   docker exec doom-tailscale ping -c 3 google.com
   ```

## code-server Issues

### Can't Access Web UI

**Symptoms**: Browser shows connection refused

**Solutions**:

1. Verify container is running:
   ```bash
   docker compose ps code-server
   ```

2. Check if using correct IP:
   ```bash
   tailscale status
   # Use the Tailscale IP, not localhost
   ```

3. Check logs:
   ```bash
   docker compose logs code-server
   ```

4. Verify network mode:
   ```bash
   docker inspect doom-code-server --format '{{.HostConfig.NetworkMode}}'
   # Should show: service:tailscale
   ```

### Password Not Working

**Symptoms**: Password rejected at login

**Solutions**:

1. Check password is set:
   ```bash
   docker compose exec code-server cat /config/.config/code-server/config.yaml
   ```

2. Reset password:
   ```bash
   # Update .env
   CODE_SERVER_PASSWORD=new-password

   # Restart
   docker compose restart code-server
   ```

### Extensions Not Installing

**Symptoms**: Extension installation fails or not found

**Note**: code-server uses Open VSX, not Microsoft marketplace.

**Solutions**:

1. Search at https://open-vsx.org
2. Install manually:
   ```bash
   docker exec doom-code-server code-server --install-extension publisher.extension
   ```
3. Use VSIX file:
   ```bash
   # Download .vsix and mount to container
   docker exec doom-code-server code-server --install-extension /path/to/extension.vsix
   ```

## Claude Code Issues

### Command Not Found

**Symptoms**: `claude: command not found`

**Solutions**:

1. Check installation:
   ```bash
   docker exec doom-claude which claude
   docker exec doom-claude ls -la ~/.claude/bin/
   ```

2. Reinstall:
   ```bash
   docker exec doom-claude curl -fsSL https://claude.ai/install.sh | bash
   ```

3. Check PATH:
   ```bash
   docker exec doom-claude echo $PATH
   ```

### API Key Not Working

**Symptoms**: Authentication errors

**Solutions**:

1. Verify secret is mounted:
   ```bash
   docker exec doom-claude cat /run/secrets/anthropic_api_key
   ```

2. Check key format:
   - Should start with `sk-ant-`
   - No trailing whitespace

3. Regenerate key at https://console.anthropic.com

## SSH Issues

### Can't Connect via SSH

**Symptoms**: Connection refused or timeout

**Solutions**:

1. Check SSH service:
   ```bash
   sudo systemctl status sshd
   ```

2. Verify your key is authorized:
   ```bash
   cat ~/.ssh/authorized_keys
   ```

3. Test with verbose mode:
   ```bash
   ssh -vvv user@tailscale-ip
   ```

### Locked Out After Hardening

**Symptoms**: Can't login after applying SSH hardening

**Emergency access**:

1. Use console access (if available)
2. Remove hardening config:
   ```bash
   sudo rm /etc/ssh/sshd_config.d/99-doom-hardening.conf
   sudo systemctl restart sshd
   ```

## Terminal Tools Issues

### Oh My Zsh Not Loading

**Symptoms**: Zsh starts but no Oh My Zsh

**Solutions**:

1. Check installation:
   ```bash
   ls -la ~/.oh-my-zsh
   ```

2. Verify .zshrc:
   ```bash
   grep "ZSH=" ~/.zshrc
   ```

3. Reinstall:
   ```bash
   sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"
   ```

### Tmux Plugins Not Working

**Symptoms**: TPM plugins not loading

**Solutions**:

1. Install TPM:
   ```bash
   git clone https://github.com/tmux-plugins/tpm ~/.tmux/plugins/tpm
   ```

2. Install plugins (in tmux):
   Press `prefix + I` (Ctrl+A, then I)

3. Reload config:
   ```bash
   tmux source ~/.tmux.conf
   ```

## Performance Issues

### Slow Container Startup

**Solutions**:

1. Check system resources:
   ```bash
   htop
   docker stats
   ```

2. Reduce health check frequency in docker-compose.yml

3. Use SSD storage

### High Memory Usage

**Solutions**:

1. Limit container memory:
   ```yaml
   # docker-compose.yml
   services:
     code-server:
       deploy:
         resources:
           limits:
             memory: 2G
   ```

2. Prune unused resources:
   ```bash
   docker system prune -af
   ```

## Getting More Help

1. **Check logs thoroughly**:
   ```bash
   docker compose logs --tail=100
   ```

2. **Enable verbose mode**:
   ```bash
   ./scripts/install.sh --verbose
   ```

3. **Search issues**: Check GitHub issues for similar problems

4. **Collect diagnostics**:
   ```bash
   ./scripts/health-check.sh --json > diagnostics.json
   docker compose config > compose-config.txt
   ```