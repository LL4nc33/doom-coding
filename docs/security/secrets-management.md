# Secrets Management

Guide to managing secrets securely with SOPS and age encryption.

## Overview

Doom Coding uses **SOPS** (Secrets OPerationS) with **age** encryption for managing sensitive data:

- API keys
- Passwords
- Authentication tokens
- Private certificates

## Why SOPS + age?

| Feature | Benefit |
|---------|---------|
| **Encryption at rest** | Secrets are never stored in plaintext |
| **Git-friendly** | Encrypted files can be safely committed |
| **Diff-able** | See what changed in encrypted files |
| **No cloud dependency** | Works entirely offline |
| **Simple key management** | Single key file to backup |

## Initial Setup

### 1. Initialize Secrets Management

```bash
./scripts/setup-secrets.sh init
```

This will:
- Install SOPS and age if not present
- Generate an encryption key
- Update `.sops.yaml` with your public key

### 2. Backup Your Key

**CRITICAL**: Backup your encryption key immediately!

```bash
# Key location
~/.config/sops/age/keys.txt

# Backup command
cp ~/.config/sops/age/keys.txt ~/backup/doom-coding-age-key.txt
```

Store this backup securely (password manager, secure cloud storage, etc.)

### 3. Create Secrets Template

```bash
./scripts/setup-secrets.sh template
```

This creates `secrets/secrets.yaml` with a template structure.

## Managing Secrets

### Creating Secrets

1. Edit the template:
   ```bash
   vim secrets/secrets.yaml
   ```

2. Add your actual values:
   ```yaml
   tailscale:
       auth_key: "tskey-auth-YOUR-ACTUAL-KEY"

   code_server:
       password: "your-actual-password"
       sudo_password: "your-sudo-password"

   anthropic:
       api_key: "sk-ant-YOUR-API-KEY"
   ```

3. Encrypt the file:
   ```bash
   ./scripts/setup-secrets.sh encrypt secrets/secrets.yaml
   ```

4. Delete the unencrypted file:
   ```bash
   rm secrets/secrets.yaml
   ```

### Viewing Secrets

```bash
./scripts/setup-secrets.sh decrypt secrets/secrets.enc.yaml
```

### Editing Secrets

```bash
./scripts/setup-secrets.sh edit secrets/secrets.enc.yaml
```

This opens the decrypted file in your editor, then re-encrypts on save.

### Exporting for Docker

Extract individual secrets for Docker:

```bash
./scripts/setup-secrets.sh export
```

This creates plain text files in `secrets/`:
- `anthropic_api_key.txt`
- `tailscale_auth_key.txt`
- `code_server_password.txt`

## Docker Integration

### Using Docker Secrets

Secrets are mounted as files, not environment variables:

```yaml
# docker-compose.yml
services:
  claude:
    secrets:
      - anthropic_api_key

secrets:
  anthropic_api_key:
    file: ./secrets/anthropic_api_key.txt
```

### Reading Secrets in Containers

```bash
# In container
cat /run/secrets/anthropic_api_key
```

Or in application code:
```python
with open('/run/secrets/anthropic_api_key') as f:
    api_key = f.read().strip()
```

## Key Management

### Viewing Your Public Key

```bash
./scripts/setup-secrets.sh show-key
```

### Rotating Keys

1. Generate new key:
   ```bash
   age-keygen -o ~/.config/sops/age/keys-new.txt
   ```

2. Decrypt all secrets with old key
3. Update `.sops.yaml` with new public key
4. Re-encrypt all secrets
5. Replace old key file

### Multiple Recipients

To allow multiple users to decrypt:

```yaml
# .sops.yaml
creation_rules:
  - path_regex: secrets/.*\.yaml$
    age: >-
      age1user1publickey,
      age1user2publickey,
      age1user3publickey
```

## Best Practices

### DO

- Backup encryption keys securely
- Use different keys for different environments
- Commit encrypted files to git
- Use Docker secrets (files) instead of environment variables
- Rotate keys periodically

### DON'T

- Commit unencrypted secrets
- Store keys in the repository
- Use environment variables for sensitive data
- Share keys via insecure channels
- Forget to delete unencrypted files

## File Structure

```
secrets/
├── .gitignore           # Ignores unencrypted files
├── secrets.enc.yaml     # Encrypted master secrets
├── anthropic_api_key.txt  # Exported for Docker
├── tailscale_auth_key.txt # Exported for Docker
└── code_server_password.txt # Exported for Docker
```

### .gitignore for Secrets

```
# Ignore unencrypted secrets
secrets/*.yaml
!secrets/*.enc.yaml
secrets/*.txt
secrets/*.json
!secrets/*.enc.json
```

## Troubleshooting

### Cannot Decrypt

```
Error: no matching keys found
```

**Solution**: Ensure `SOPS_AGE_KEY_FILE` points to your key:
```bash
export SOPS_AGE_KEY_FILE=~/.config/sops/age/keys.txt
```

### Wrong Public Key in .sops.yaml

```bash
# Update with your actual public key
./scripts/setup-secrets.sh show-key
vim .sops.yaml
```

### Lost Encryption Key

If you lose your key, encrypted secrets cannot be recovered. You must:
1. Regenerate all secrets from source (API consoles, etc.)
2. Create new encryption key
3. Re-encrypt everything

## Integration Examples

### CI/CD Pipeline

```yaml
# GitHub Actions
- name: Decrypt secrets
  env:
    SOPS_AGE_KEY: ${{ secrets.SOPS_AGE_KEY }}
  run: |
    echo "$SOPS_AGE_KEY" > /tmp/age-key.txt
    export SOPS_AGE_KEY_FILE=/tmp/age-key.txt
    sops -d secrets/secrets.enc.yaml > secrets/secrets.yaml
```

### Ansible Integration

```yaml
- name: Decrypt secrets
  command: sops -d secrets/secrets.enc.yaml
  register: secrets_raw

- name: Parse secrets
  set_fact:
    secrets: "{{ secrets_raw.stdout | from_yaml }}"
```

## Next Steps

- [Security Hardening](hardening.md)
- [SSH Configuration](ssh-config.md)
- [Docker Security](../docker/security.md)