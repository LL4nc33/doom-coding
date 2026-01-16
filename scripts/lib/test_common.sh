#!/usr/bin/env bash
# =============================================================================
# Doom Coding - Common Library Test Suite
# =============================================================================
# Tests for the shared common.sh library functions.
#
# Usage:
#   ./scripts/lib/test_common.sh
# =============================================================================

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

echo "Testing common.sh library..."
echo "=============================================="

# Test logging
echo ""
echo "Testing logging functions:"
log_info "Info message test"
log_success "Success message test"
log_warning "Warning message test"
log_step "Step message test"
log_pass "Pass message test"
log_fail "Fail message test (expected)"

# Test detection
echo ""
echo "Testing detection functions:"
echo "OS: $(detect_os)"
echo "Arch: $(detect_arch)"
echo "Package Manager: $(detect_package_manager)"
container_type=$(detect_container_type)
if [[ -n "$container_type" ]]; then
    echo "Container: $container_type"
else
    echo "Container: (none - bare metal or VM)"
fi

# Test utilities
echo ""
echo "Testing utility functions:"
if command_exists bash; then
    log_pass "command_exists works for 'bash'"
else
    log_fail "command_exists failed for 'bash'"
fi

if ! command_exists nonexistent_command_12345; then
    log_pass "command_exists correctly returns false for missing command"
else
    log_fail "command_exists incorrectly returned true for missing command"
fi

# Test is_root
echo ""
echo "Testing permission functions:"
if is_root; then
    log_info "Running as root"
else
    log_info "Running as non-root user"
fi

# Test ensure_directory (dry run - just check logic)
echo ""
echo "Testing file operation functions:"
log_info "ensure_directory and backup_file available"

echo ""
echo "=============================================="
log_success "All tests completed!"
