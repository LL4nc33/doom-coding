# Contributing to Doom Coding

Thank you for your interest in contributing to Doom Coding! This guide will help you get started with contributing to the project.

## 🎯 Ways to Contribute

### 🐛 Bug Reports
- Use the [GitHub Issues](https://github.com/LL4nc33/doom-coding/issues) page
- Search existing issues before creating new ones
- Use the bug report template
- Include system information and reproduction steps

### ✨ Feature Requests
- Use GitHub Issues with the "enhancement" label
- Describe the use case and expected behavior
- Consider contributing the implementation yourself

### 📝 Documentation
- Fix typos and unclear instructions
- Add missing documentation
- Improve existing guides
- Translate documentation to other languages

### 🔧 Code Contributions
- Bug fixes
- New features
- Performance improvements
- Security enhancements
- Test coverage improvements

## 🚀 Getting Started

### Prerequisites
- Git installed
- Docker and Docker Compose
- Basic understanding of Bash scripting
- Linux environment (or WSL2)

### Development Setup

1. **Fork the Repository**
   ```bash
   # Fork via GitHub UI, then clone your fork
   git clone https://github.com/YOUR-USERNAME/doom-coding.git
   cd doom-coding
   ```

2. **Set Up Development Environment**
   ```bash
   # Copy environment template
   cp .env.example .env

   # Edit with your development settings
   vim .env
   ```

3. **Test Installation**
   ```bash
   # Test the installer
   ./scripts/install.sh --dry-run

   # Or test specific components
   ./scripts/setup-terminal.sh --dry-run
   ```

4. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/issue-description
   ```

## 📋 Development Guidelines

### Code Style

#### Bash Scripts
- Use `#!/usr/bin/env bash` shebang
- Set `set -euo pipefail` for error handling
- Use meaningful variable names
- Add comments for complex logic
- Follow Google Shell Style Guide

**Example:**
```bash
#!/usr/bin/env bash
set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly LOG_FILE="/var/log/install.log"

log_info() {
    echo "[INFO] $*" | tee -a "$LOG_FILE"
}

install_package() {
    local package="$1"

    if command -v "$package" &>/dev/null; then
        log_info "$package already installed"
        return 0
    fi

    # Installation logic here
}
```

#### Docker Configuration
- Use official base images
- Run as non-root user
- Multi-stage builds for smaller images
- Proper secrets handling
- Health checks for services

**Example Dockerfile:**
```dockerfile
FROM ubuntu:22.04

# Create non-root user
RUN useradd -m -u 1000 appuser

# Install dependencies
RUN apt-get update && apt-get install -y \
    package1 \
    package2 \
    && rm -rf /var/lib/apt/lists/*

USER appuser
HEALTHCHECK --interval=30s --timeout=10s \
    CMD curl -f http://localhost:8080/health || exit 1
```

#### Documentation
- Use clear headings and structure
- Include code examples
- Add screenshots for UI elements
- Write for different skill levels
- Keep language simple and accessible

### Testing

#### Manual Testing
- Test on clean Ubuntu 22.04 VM
- Test both Docker and terminal deployment modes
- Test with and without Tailscale
- Verify documentation accuracy

#### Automated Testing (Future)
```bash
# Run shellcheck on all scripts
find scripts/ -name "*.sh" -exec shellcheck {} \;

# Test Docker builds
docker build -f Dockerfile.claude .

# Validate docker-compose files
docker compose config
docker compose -f docker-compose.lxc.yml config
```

### Commit Guidelines

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `style:` Code style changes
- `refactor:` Code refactoring
- `test:` Test additions/changes
- `chore:` Maintenance tasks

**Examples:**
```
feat(install): add support for Alpine Linux
fix(docker): resolve health check timeout issue
docs(readme): update installation instructions
refactor(scripts): improve error handling in setup
```

## 🔄 Pull Request Process

### 1. Pre-Submission Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated if needed
- [ ] Manual testing performed
- [ ] Commit messages follow convention

### 2. Pull Request Template
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Tested on Ubuntu 22.04
- [ ] Tested Docker deployment
- [ ] Tested terminal deployment
- [ ] Tested with/without Tailscale

## Screenshots (if applicable)
Add screenshots to help explain your changes

## Additional Notes
Any additional information about the change
```

### 3. Review Process
1. Automated checks (shellcheck, markdown lint)
2. Manual code review by maintainers
3. Testing by reviewers
4. Final approval and merge

## 🐛 Bug Fixes

### Critical Bugs (Security/Data Loss)
- Create issue immediately
- Tag with `critical` label
- Consider creating hotfix branch directly

### Standard Bug Fix Process
1. Reproduce the issue
2. Create test case (if applicable)
3. Implement minimal fix
4. Test thoroughly
5. Update documentation if needed

## ✨ Feature Development

### Major Features
1. **Discuss First**: Create GitHub Discussion or Issue
2. **Design Document**: For complex features, write a brief design doc
3. **Incremental Development**: Break into smaller, reviewable chunks
4. **Documentation**: Update docs alongside code changes

### Minor Features
1. Check existing issues/discussions
2. Implement and test
3. Create pull request with detailed description

## 📚 Documentation Contributions

### Types of Documentation
- **Installation guides** - Step-by-step instructions
- **Configuration references** - Detailed option explanations
- **Troubleshooting guides** - Common issues and solutions
- **Examples and tutorials** - Real-world usage scenarios

### Documentation Standards
- Write for beginners and experts
- Include working code examples
- Use consistent formatting
- Add table of contents for long documents
- Cross-reference related topics

## 🔐 Security Considerations

### Reporting Security Issues
- **DO NOT** create public GitHub issues for security vulnerabilities
- Email maintainers directly with details
- Allow time for patches before public disclosure

### Security Guidelines
- Never hardcode secrets in scripts
- Use proper file permissions (600 for secrets)
- Validate all user inputs
- Follow principle of least privilege
- Keep dependencies updated

## 🤝 Community Guidelines

### Code of Conduct
- Be respectful and inclusive
- Help newcomers learn
- Focus on constructive feedback
- Assume good intentions

### Communication
- Use GitHub Issues for bugs and feature requests
- Use GitHub Discussions for questions and ideas
- Be patient with response times (this is volunteer-maintained)

## 🎖️ Recognition

Contributors are recognized in:
- README acknowledgments
- Git commit co-author tags
- Release notes for significant contributions
- Special thanks in documentation

## 📞 Getting Help

### For Contributors
- [GitHub Discussions](https://github.com/LL4nc33/doom-coding/discussions) for questions
- Reach out to maintainers for guidance on complex contributions
- Join community chat (if available)

### For Maintainers
- Review pull requests promptly
- Provide constructive feedback
- Help new contributors succeed
- Keep documentation updated

## 🚦 Release Process

### Version Numbering
- Follow [Semantic Versioning](https://semver.org/)
- `MAJOR.MINOR.PATCH` format
- Pre-release versions use `-alpha`, `-beta`, `-rc` suffixes

### Release Checklist
1. Update version numbers
2. Update CHANGELOG.md
3. Test release candidate
4. Create GitHub release
5. Update documentation
6. Announce release

---

Thank you for contributing to Doom Coding! Every contribution, no matter how small, helps make this project better for everyone. 🚀