# 🔒 Security Basics Testing (Iterations 21-25)

Comprehensive security validation for core security features and configurations.

## 📋 Iteration 21: SSH Hardening Verification

### 🎯 Objective
Validate SSH security configuration and hardening measures across all deployment types.

### 📝 Pre-Test Setup
```bash
# Prepare test environment
./scripts/setup-test-env.sh --security-focus
cd /config/repos/doom-coding

# Backup current SSH config for comparison
sudo cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup
```

### ✅ Test Cases

#### TC-21.1: SSH Configuration Audit
**Deployment Types**: All
**Priority**: Critical

**Steps**:
1. [ ] Examine SSH daemon configuration
   ```bash
   sudo sshd -T | grep -E "(passwordauthentication|pubkeyauthentication|protocol|port)"
   ```

2. [ ] Verify password authentication is disabled
   ```bash
   sudo sshd -T | grep passwordauthentication
   # Expected: passwordauthentication no
   ```

3. [ ] Confirm public key authentication is enabled
   ```bash
   sudo sshd -T | grep pubkeyauthentication
   # Expected: pubkeyauthentication yes
   ```

4. [ ] Check SSH protocol version
   ```bash
   ssh -V
   # Expected: OpenSSH 8.0+ with protocol 2
   ```

**Expected Results**:
- [ ] Password authentication disabled
- [ ] Public key authentication enabled
- [ ] SSH protocol version 2 only
- [ ] No deprecated configuration options

#### TC-21.2: SSH Cipher Suite Validation
**Deployment Types**: All
**Priority**: High

**Steps**:
1. [ ] Test SSH connection with cipher information
   ```bash
   ssh -v user@localhost 2>&1 | grep -E "(cipher|kex|mac)"
   ```

2. [ ] Verify strong encryption algorithms
   ```bash
   sudo sshd -T | grep -E "(ciphers|kexalgorithms|macs)"
   ```

3. [ ] Check for weak ciphers (should be absent)
   ```bash
   sudo sshd -T | grep -i "arcfour\|des\|rc4"
   # Expected: No output (weak ciphers disabled)
   ```

**Expected Results**:
- [ ] Strong cipher suites configured (AES256, ChaCha20)
- [ ] Modern key exchange algorithms
- [ ] No weak/deprecated ciphers present
- [ ] HMAC algorithms are secure

#### TC-21.3: SSH Access Control
**Deployment Types**: All
**Priority**: Critical

**Steps**:
1. [ ] Test key-based authentication
   ```bash
   ssh -i ~/.ssh/doom-coding user@localhost "echo 'SSH key auth successful'"
   ```

2. [ ] Verify password authentication rejection
   ```bash
   ssh -o PubkeyAuthentication=no user@localhost
   # Expected: Connection refused or authentication failure
   ```

3. [ ] Check user access restrictions
   ```bash
   sudo sshd -T | grep -E "(allowusers|denyusers|allowgroups|denygroups)"
   ```

4. [ ] Verify root login restrictions
   ```bash
   sudo sshd -T | grep permitrootlogin
   # Expected: permitrootlogin no or permitrootlogin prohibit-password
   ```

**Expected Results**:
- [ ] Key-based authentication works
- [ ] Password authentication rejected
- [ ] Root login properly restricted
- [ ] User access controls configured

#### TC-21.4: SSH Security Enhancements
**Deployment Types**: All
**Priority**: Medium

**Steps**:
1. [ ] Check SSH idle timeout configuration
   ```bash
   sudo sshd -T | grep -E "(clientaliveinterval|clientalivecountmax)"
   ```

2. [ ] Verify maximum authentication attempts
   ```bash
   sudo sshd -T | grep maxauthtries
   # Expected: reasonable limit (3-6 attempts)
   ```

3. [ ] Check login banner configuration
   ```bash
   sudo sshd -T | grep banner
   ```

4. [ ] Verify X11 forwarding restrictions
   ```bash
   sudo sshd -T | grep x11forwarding
   # Expected: x11forwarding no (unless specifically required)
   ```

**Expected Results**:
- [ ] Idle timeout configured appropriately
- [ ] Authentication attempt limits set
- [ ] Security banner present (if configured)
- [ ] Unnecessary features disabled

### 📊 Test Results

| Test Case | Status | Notes | Evidence |
|-----------|--------|-------|----------|
| TC-21.1 | ⏳ | | |
| TC-21.2 | ⏳ | | |
| TC-21.3 | ⏳ | | |
| TC-21.4 | ⏳ | | |

### 🚨 Issues Found
*Record any security issues discovered during testing*

---

## 📋 Iteration 22: Container Security Scanning

### 🎯 Objective
Validate Docker container security configuration and scan for vulnerabilities.

### 📝 Pre-Test Setup
```bash
# Install security scanning tools
docker pull aquasec/trivy:latest

# Prepare container environment
docker compose up -d
docker ps
```

### ✅ Test Cases

#### TC-22.1: Base Image Vulnerability Scan
**Deployment Types**: All Docker variants
**Priority**: Critical

**Steps**:
1. [ ] Scan code-server container
   ```bash
   docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
     aquasec/trivy image code-server:latest
   ```

2. [ ] Scan Tailscale container
   ```bash
   docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
     aquasec/trivy image tailscale/tailscale:latest
   ```

3. [ ] Scan Claude Code container
   ```bash
   docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
     aquasec/trivy image doom-coding-claude:latest
   ```

4. [ ] Generate vulnerability report
   ```bash
   docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
     aquasec/trivy image --severity HIGH,CRITICAL --format table \
     code-server:latest > vulnerability-report.txt
   ```

**Expected Results**:
- [ ] No CRITICAL vulnerabilities in production images
- [ ] HIGH vulnerabilities documented and assessed
- [ ] Base images are recent and maintained
- [ ] Vulnerability report generated

#### TC-22.2: Container Runtime Security
**Deployment Types**: All Docker variants
**Priority**: High

**Steps**:
1. [ ] Verify non-root user execution
   ```bash
   docker exec code-server id
   # Expected: uid!=0 (not root)
   ```

2. [ ] Check container capabilities
   ```bash
   docker inspect code-server | jq '.[0].HostConfig.CapDrop'
   docker inspect code-server | jq '.[0].HostConfig.CapAdd'
   ```

3. [ ] Verify read-only filesystem where applicable
   ```bash
   docker exec code-server mount | grep "ro,"
   ```

4. [ ] Check security options
   ```bash
   docker inspect code-server | jq '.[0].HostConfig.SecurityOpt'
   ```

**Expected Results**:
- [ ] Containers run as non-root users
- [ ] Unnecessary capabilities dropped
- [ ] Security options configured appropriately
- [ ] No privileged containers unless required

#### TC-22.3: Container Isolation Verification
**Deployment Types**: All Docker variants
**Priority**: High

**Steps**:
1. [ ] Test process isolation
   ```bash
   docker exec code-server ps aux
   # Should only show container processes
   ```

2. [ ] Verify network isolation
   ```bash
   docker exec code-server netstat -tuln
   docker network ls
   ```

3. [ ] Check filesystem isolation
   ```bash
   docker exec code-server df -h
   # Verify only container filesystems visible
   ```

4. [ ] Test resource limitations
   ```bash
   docker stats --no-stream
   docker inspect code-server | jq '.[0].HostConfig.Memory'
   ```

**Expected Results**:
- [ ] Process isolation functioning
- [ ] Network segmentation effective
- [ ] Filesystem access restricted
- [ ] Resource limits enforced

#### TC-22.4: Secrets and Environment Security
**Deployment Types**: All Docker variants
**Priority**: Critical

**Steps**:
1. [ ] Check for exposed secrets in environment
   ```bash
   docker exec code-server env | grep -E "(PASSWORD|KEY|SECRET|TOKEN)"
   ```

2. [ ] Verify secrets mounting (if used)
   ```bash
   docker exec code-server ls -la /run/secrets/
   docker exec code-server ls -la /var/run/secrets/
   ```

3. [ ] Check file permissions on sensitive data
   ```bash
   docker exec code-server find /home -name "*.key" -o -name "*.pem" | xargs ls -la
   ```

4. [ ] Verify no credentials in image layers
   ```bash
   docker history code-server:latest | grep -E "(PASSWORD|KEY|SECRET)"
   ```

**Expected Results**:
- [ ] No plaintext secrets in environment variables
- [ ] Secrets properly mounted with correct permissions
- [ ] Sensitive files have restrictive permissions
- [ ] No credentials baked into image layers

### 📊 Test Results

| Test Case | Status | Notes | Evidence |
|-----------|--------|-------|----------|
| TC-22.1 | ⏳ | | |
| TC-22.2 | ⏳ | | |
| TC-22.3 | ⏳ | | |
| TC-22.4 | ⏳ | | |

---

## 📋 Iteration 23: Secrets Management Validation

### 🎯 Objective
Validate SOPS/age encryption and secrets management implementation.

### 📝 Pre-Test Setup
```bash
# Setup secrets management tools
./scripts/setup-secrets.sh generate-key
source .env
```

### ✅ Test Cases

#### TC-23.1: SOPS/age Encryption Functionality
**Deployment Types**: All
**Priority**: Critical

**Steps**:
1. [ ] Generate age key pair
   ```bash
   ./scripts/setup-secrets.sh generate-key
   ls -la ~/.config/age/
   ```

2. [ ] Test secret encryption
   ```bash
   echo "test-secret: sensitive-data" | ./scripts/setup-secrets.sh encrypt -
   ```

3. [ ] Test secret decryption
   ```bash
   ./scripts/setup-secrets.sh decrypt secrets.yaml
   ```

4. [ ] Verify key permissions
   ```bash
   ls -la ~/.config/age/keys.txt
   # Expected: 600 permissions (owner read/write only)
   ```

**Expected Results**:
- [ ] Age key pair generated successfully
- [ ] Encryption/decryption working
- [ ] Private keys have secure permissions
- [ ] No plaintext secrets in repository

#### TC-23.2: Environment Variable Security
**Deployment Types**: All
**Priority**: High

**Steps**:
1. [ ] Check for plaintext API keys in configs
   ```bash
   grep -r "sk-ant-" . --exclude-dir=.git
   # Expected: No matches (encrypted only)
   ```

2. [ ] Verify Tailscale key protection
   ```bash
   grep -r "tskey-" . --exclude-dir=.git
   # Expected: No matches (encrypted only)
   ```

3. [ ] Test environment variable loading
   ```bash
   source .env
   echo ${ANTHROPIC_API_KEY:0:10}...  # Show partial for verification
   ```

4. [ ] Check .env file permissions
   ```bash
   ls -la .env
   # Expected: 600 permissions minimum
   ```

**Expected Results**:
- [ ] No plaintext secrets in repository
- [ ] Environment variables loaded correctly
- [ ] Configuration files have secure permissions
- [ ] Secrets are properly encrypted at rest

#### TC-23.3: Container Secrets Management
**Deployment Types**: All Docker variants
**Priority**: High

**Steps**:
1. [ ] Verify secrets not in container environment
   ```bash
   docker exec code-server env | grep -v "^_" | grep -E "(sk-ant|tskey)"
   # Should show masked or no direct API keys
   ```

2. [ ] Check secrets mounting mechanism
   ```bash
   docker inspect code-server | jq '.[0].Mounts[] | select(.Type=="bind")'
   ```

3. [ ] Verify secret file permissions inside container
   ```bash
   docker exec code-server ls -la /run/secrets/ 2>/dev/null || echo "No mounted secrets dir"
   ```

4. [ ] Test secret access from application
   ```bash
   docker logs claude-code | grep -i "api key" | head -5
   # Should show masked/redacted keys only
   ```

**Expected Results**:
- [ ] API keys not exposed in container environment
- [ ] Secrets properly mounted if used
- [ ] Application can access secrets securely
- [ ] No secrets visible in container logs

### 📊 Test Results

| Test Case | Status | Notes | Evidence |
|-----------|--------|-------|----------|
| TC-23.1 | ⏳ | | |
| TC-23.2 | ⏳ | | |
| TC-23.3 | ⏳ | | |

---

## 📋 Iteration 24: Network Security Assessment

### 🎯 Objective
Validate network security controls and VPN configuration.

### ✅ Test Cases

#### TC-24.1: Firewall Configuration
**Deployment Types**: All
**Priority**: High

**Steps**:
1. [ ] Check UFW status and rules
   ```bash
   sudo ufw status verbose
   sudo ufw show raw
   ```

2. [ ] Verify only required ports are open
   ```bash
   netstat -tuln | grep LISTEN
   ss -tuln
   ```

3. [ ] Test external port accessibility
   ```bash
   nmap -sS localhost
   nmap -sS $(hostname -I | awk '{print $1}')
   ```

**Expected Results**:
- [ ] Firewall is active and configured
- [ ] Only necessary ports exposed
- [ ] External scan shows minimal attack surface

#### TC-24.2: Tailscale VPN Security
**Deployment Types**: Tailscale variants
**Priority**: Critical

**Steps**:
1. [ ] Verify Tailscale connection encryption
   ```bash
   tailscale status
   tailscale ping --verbose peer-node
   ```

2. [ ] Check Tailscale ACL configuration
   ```bash
   tailscale status | grep -E "(relay|direct)"
   ```

3. [ ] Test VPN isolation
   ```bash
   ip route show | grep 100.
   ```

**Expected Results**:
- [ ] VPN connection encrypted
- [ ] Direct connections preferred
- [ ] Network isolation maintained

### 📊 Test Results

| Test Case | Status | Notes | Evidence |
|-----------|--------|-------|----------|
| TC-24.1 | ⏳ | | |
| TC-24.2 | ⏳ | | |

---

## 📋 Iteration 25: Access Control Verification

### 🎯 Objective
Validate user permissions and service account security.

### ✅ Test Cases

#### TC-25.1: File Permission Audit
**Deployment Types**: All
**Priority**: High

**Steps**:
1. [ ] Check for world-writable files
   ```bash
   find /home/user -perm /o+w -type f 2>/dev/null
   # Expected: No world-writable files
   ```

2. [ ] Verify sensitive file permissions
   ```bash
   ls -la ~/.ssh/
   ls -la ~/.config/age/
   ```

3. [ ] Check Docker socket permissions
   ```bash
   ls -la /var/run/docker.sock
   groups user | grep docker
   ```

**Expected Results**:
- [ ] No world-writable files in user directories
- [ ] SSH keys have 600/700 permissions
- [ ] Docker access properly controlled

### 📊 Test Results

| Test Case | Status | Notes | Evidence |
|-----------|--------|-------|----------|
| TC-25.1 | ⏳ | | |

## 📋 Overall Security Phase Summary

### 🎯 Completion Status
- [ ] Iteration 21: SSH Hardening Verification
- [ ] Iteration 22: Container Security Scanning
- [ ] Iteration 23: Secrets Management Validation
- [ ] Iteration 24: Network Security Assessment
- [ ] Iteration 25: Access Control Verification

### 📊 Security Metrics
| Metric | Target | Actual | Status |
|--------|--------|--------|--------|
| Critical Vulns | 0 | TBD | ⏳ |
| High Vulns | <5 | TBD | ⏳ |
| SSH Config Score | >90% | TBD | ⏳ |
| Container Security | Pass | TBD | ⏳ |
| Secrets Management | Pass | TBD | ⏳ |

### 🚨 Critical Issues Found
*Record any critical security issues that need immediate attention*

### ✅ Recommendations
*List security improvements and recommendations*

### 🔄 Next Steps
*Outline follow-up actions and next iteration preparation*

---

<p align="center">
  <strong>Security First</strong><br>
  <em>Comprehensive validation for production-ready security</em>
</p>