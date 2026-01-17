# 🐛 Bug Report Template

**Bug ID**: BUG-[YYYY-MM-DD]-[Number]  
**Reported Date**: [Date]  
**Reporter**: [Name/Username]  
**Assigned To**: [Name]  
**Status**: Open / In Progress / Fixed / Closed  
**Priority**: Critical (P0) / High (P1) / Medium (P2) / Low (P3)  
**Severity**: Blocker / Critical / Major / Minor / Trivial  

## 📋 Bug Summary

**Title**: [Clear, descriptive title]

**Category**: 
- [ ] Installation
- [ ] Configuration  
- [ ] Security
- [ ] Performance
- [ ] UI/UX
- [ ] Documentation
- [ ] Integration
- [ ] Network
- [ ] Other: [Specify]

**Component(s) Affected**:
- [ ] Installation Scripts
- [ ] Docker Containers
- [ ] Tailscale Integration
- [ ] Web Interface
- [ ] SSH Configuration
- [ ] Health Checks
- [ ] Documentation
- [ ] TUI Interface

## 🔍 Detailed Description

**What happened**:
[Clear description of the issue]

**What was expected**:
[Expected behavior description]

**Impact on users**:
[How this affects end users]

**Business impact**:
[Impact on project goals/timeline]

## 🔄 Steps to Reproduce

**Prerequisites**:
- Environment: [OS, version, etc.]
- Setup: [Any specific setup required]

**Reproduction Steps**:
1. [Step 1]
2. [Step 2]
3. [Step 3]
4. [Continue as needed]

**Expected Result**: [What should happen]
**Actual Result**: [What actually happens]

**Reproduction Rate**: 
- [ ] Always (100%)
- [ ] Often (75-99%)
- [ ] Sometimes (25-74%)
- [ ] Rarely (1-24%)
- [ ] Unable to reproduce

## 💻 Environment Details

**System Information**:
- Operating System: [OS and version]
- Architecture: [x86_64/arm64/etc.]
- Kernel Version: [Version]
- Available Memory: [RAM amount]
- Available Storage: [Disk space]

**Software Versions**:
- Docker: [Version]
- Docker Compose: [Version]
- Git: [Version]
- Tailscale: [Version]
- Browser: [Browser and version if applicable]

**Deployment Configuration**:
- Deployment Type: [Standard/LXC-Tailscale/LXC-Local/Terminal/Native]
- Docker Compose File: [Which compose file used]
- Custom Configuration: [Any customizations]

**Network Configuration**:
- Network Type: [Local/Tailscale/Hybrid]
- Firewall: [Enabled/Disabled/Custom]
- Proxy: [Yes/No/Details]

## 📊 Evidence and Logs

**Error Messages**:
```
[Paste exact error messages here]
```

**Log Excerpts**:
```bash
# Installation logs (if applicable)
[Relevant installation log lines]

# Docker logs
[Output from docker logs [container-name]]

# System logs
[Relevant system log entries]

# Health check output
[Output from ./scripts/health-check.sh]
```

**Screenshots** (if applicable):
- [Attach or link to screenshots]
- [Annotate important areas]

**Configuration Files**:
```yaml
# docker-compose.yml (sanitized)
[Relevant portions of configuration]
```

```bash
# Environment variables (sanitized)
[Relevant environment variables - REMOVE SENSITIVE DATA]
```

## 🔧 Diagnostic Information

**Health Check Results**:
```bash
[Output from health check script]
```

**System Resource Usage**:
```bash
# Memory usage
[free -h output]

# Disk usage
[df -h output]

# CPU information
[top/htop snapshot]

# Docker stats
[docker stats output]
```

**Network Connectivity**:
```bash
# Network interfaces
[ip addr show output]

# Routing table
[ip route show output]

# DNS resolution
[nslookup/dig results]

# Tailscale status (if applicable)
[tailscale status output]
```

## 🛠️ Attempted Solutions

**What has been tried**:
- [ ] [Solution attempt 1] - Result: [Success/Failure/Partial]
- [ ] [Solution attempt 2] - Result: [Success/Failure/Partial]
- [ ] [Solution attempt 3] - Result: [Success/Failure/Partial]

**Workarounds found**:
- [Workaround 1]: [Description]
- [Workaround 2]: [Description]

## 📋 Additional Context

**When did this start happening**:
[Timeline of when the issue first appeared]

**What changed recently**:
[Any recent system changes, updates, etc.]

**Similar issues**:
[Links to similar bug reports or issues]

**User impact assessment**:
- Number of users affected: [Estimate]
- Frequency of occurrence: [How often it happens]
- Severity of user experience degradation: [1-10 scale]

## 🎯 Acceptance Criteria for Fix

**Definition of Done**:
- [ ] Bug no longer reproducible
- [ ] No regression in related functionality
- [ ] Fix verified across all supported platforms
- [ ] Documentation updated if needed
- [ ] Test case added to prevent regression

**Test Verification Steps**:
1. [Verification step 1]
2. [Verification step 2]
3. [Verification step 3]

## 🔄 Bug Resolution Tracking

### Investigation Phase
**Assigned To**: [Developer name]  
**Investigation Started**: [Date]  
**Root Cause**: [To be filled by developer]  
**Estimated Fix Time**: [Time estimate]  

### Development Phase
**Fix Started**: [Date]  
**Approach**: [Technical approach description]  
**Files Modified**: [List of modified files]  
**Testing Method**: [How the fix was tested]  

### Verification Phase
**Fix Completed**: [Date]  
**Verified By**: [Tester name]  
**Verification Date**: [Date]  
**Status**: [Verified/Failed verification]  
**Notes**: [Verification notes]  

### Closure Phase
**Closed By**: [Name]  
**Closed Date**: [Date]  
**Resolution**: Fixed / Won't Fix / Duplicate / Cannot Reproduce / By Design  
**Final Notes**: [Closure comments]  

## 📊 Related Information

**Related Bugs**: 
- [Link to related bug reports]

**Related Features**:
- [Link to related feature requests]

**Documentation References**:
- [Links to relevant documentation]

**Test Case References**:
- [Links to test cases that should catch this]

## 📝 Communication Log

**Stakeholder Notifications**:
- [Date]: [Who was notified and how]
- [Date]: [Status update sent to stakeholders]

**Team Discussions**:
- [Date]: [Summary of team discussion]
- [Date]: [Key decisions made]

## 📎 Attachments

**Files Attached**:
- [ ] Log files: [Filename(s)]
- [ ] Screenshots: [Filename(s)]
- [ ] Configuration files: [Filename(s)]
- [ ] Video recordings: [Filename(s)]
- [ ] Diagnostic outputs: [Filename(s)]

**External Links**:
- [Link 1]: [Description]
- [Link 2]: [Description]

---

## ⚠️ Important Notes

**Security Considerations**:
[Any security implications of this bug]

**Performance Impact**:
[Any performance implications]

**Data Loss Risk**:
[Any risk of data loss or corruption]

**Breaking Changes**:
[Any potential breaking changes in the fix]

---

**Report Created**: [Timestamp]  
**Last Updated**: [Timestamp]  
**Report Version**: [Version number]  
**Next Review**: [Date for next review]

---

**Reporter Signature**: [Name] - [Date]  
**Lead Reviewer**: [Name] - [Date]