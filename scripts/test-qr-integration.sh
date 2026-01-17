#!/usr/bin/env bash
# Doom Coding - QR Integration Test Script
# Tests QR code generation functionality

set -euo pipefail

# ===========================================
# COLORS
# ===========================================
readonly GREEN='\033[38;2;46;82;29m'
readonly RED='\033[0;31m'
readonly YELLOW='\033[0;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# ===========================================
# TEST COUNTERS
# ===========================================
TESTS_PASSED=0
TESTS_FAILED=0

# ===========================================
# TEST HELPERS
# ===========================================
test_pass() {
    ((TESTS_PASSED++))
    echo -e "${GREEN}✅${NC} $*"
}

test_fail() {
    ((TESTS_FAILED++))
    echo -e "${RED}❌${NC} $*"
}

test_skip() {
    echo -e "${YELLOW}⏭${NC}  $* (skipped)"
}

test_section() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $*${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# ===========================================
# QR CODE TESTS
# ===========================================

test_qrencode_installed() {
    test_section "Testing qrencode Installation"

    if command -v qrencode &>/dev/null; then
        local version
        version=$(qrencode --version 2>&1 | head -1)
        test_pass "qrencode is installed: $version"
        return 0
    else
        test_fail "qrencode is not installed"
        echo "    Install with: apt install qrencode"
        return 1
    fi
}

test_qr_generation_basic() {
    test_section "Testing Basic QR Generation"

    if ! command -v qrencode &>/dev/null; then
        test_skip "qrencode not installed"
        return 0
    fi

    # Test simple URL encoding
    local output
    if output=$(echo "https://example.com" | qrencode -t ansiutf8 2>&1); then
        if [[ -n "$output" ]]; then
            test_pass "Basic QR code generated successfully"
            echo "    Output length: ${#output} characters"
        else
            test_fail "QR code output is empty"
        fi
    else
        test_fail "qrencode failed to generate QR code"
    fi
}

test_qr_generation_url() {
    test_section "Testing URL QR Code Generation"

    if ! command -v qrencode &>/dev/null; then
        test_skip "qrencode not installed"
        return 0
    fi

    local test_urls=(
        "https://192.168.1.100:8443"
        "https://100.64.0.1:8443"
        "https://login.tailscale.com/admin/settings/keys"
        "https://console.anthropic.com/account/keys"
    )

    for url in "${test_urls[@]}"; do
        if echo "$url" | qrencode -t ansiutf8 &>/dev/null; then
            test_pass "QR generated for: $url"
        else
            test_fail "QR failed for: $url"
        fi
    done
}

test_qr_output_formats() {
    test_section "Testing QR Output Formats"

    if ! command -v qrencode &>/dev/null; then
        test_skip "qrencode not installed"
        return 0
    fi

    local test_data="https://test.example.com"

    # Test ANSI format (terminal)
    if echo "$test_data" | qrencode -t ansiutf8 &>/dev/null; then
        test_pass "ANSI UTF-8 format works"
    else
        test_fail "ANSI UTF-8 format failed"
    fi

    # Test ASCII format
    if echo "$test_data" | qrencode -t ASCII &>/dev/null; then
        test_pass "ASCII format works"
    else
        test_fail "ASCII format failed"
    fi

    # Test UTF-8 format
    if echo "$test_data" | qrencode -t UTF8 &>/dev/null; then
        test_pass "UTF-8 format works"
    else
        test_fail "UTF-8 format failed"
    fi
}

test_health_check_qr_flag() {
    test_section "Testing Health Check --qr Flag"

    local script_dir
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local health_script="${script_dir}/health-check.sh"

    if [[ ! -f "$health_script" ]]; then
        test_fail "health-check.sh not found at $health_script"
        return 1
    fi

    # Test help shows --qr option
    if "$health_script" --help 2>&1 | grep -q "\-\-qr"; then
        test_pass "health-check.sh supports --qr flag"
    else
        test_fail "health-check.sh does not show --qr in help"
    fi
}

test_install_script_qr_functions() {
    test_section "Testing Install Script QR Functions"

    local script_dir
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local install_script="${script_dir}/install.sh"

    if [[ ! -f "$install_script" ]]; then
        test_fail "install.sh not found at $install_script"
        return 1
    fi

    # Check for QR helper functions
    local functions=(
        "generate_qr"
        "show_access_qr"
        "show_service_qr"
        "show_troubleshoot_qr"
    )

    for func in "${functions[@]}"; do
        if grep -q "^${func}()" "$install_script" || grep -q "^${func} ()" "$install_script"; then
            test_pass "Function $func exists in install.sh"
        else
            test_fail "Function $func not found in install.sh"
        fi
    done

    # Check qrencode is in package list
    if grep -q "qrencode" "$install_script"; then
        test_pass "qrencode included in install packages"
    else
        test_fail "qrencode not in install packages"
    fi
}

test_go_qr_library() {
    test_section "Testing Go QR Library"

    local project_dir
    project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
    local qr_lib="${project_dir}/internal/qr/generator.go"
    local qr_test="${project_dir}/internal/qr/generator_test.go"

    if [[ -f "$qr_lib" ]]; then
        test_pass "Go QR library exists: internal/qr/generator.go"
    else
        test_fail "Go QR library not found"
        return 1
    fi

    if [[ -f "$qr_test" ]]; then
        test_pass "Go QR tests exist: internal/qr/generator_test.go"
    else
        test_fail "Go QR tests not found"
    fi

    # Check for key functions
    local expected_functions=(
        "GenerateASCII"
        "GenerateAccessQR"
        "GenerateExternalServiceQR"
        "GenerateURLWithFallback"
    )

    for func in "${expected_functions[@]}"; do
        if grep -q "func.*${func}" "$qr_lib"; then
            test_pass "Function $func exists in Go library"
        else
            test_fail "Function $func not found in Go library"
        fi
    done
}

test_mobile_documentation() {
    test_section "Testing Mobile Documentation"

    local project_dir
    project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
    local mobile_doc="${project_dir}/docs/mobile/smartphone-setup.md"

    if [[ -f "$mobile_doc" ]]; then
        test_pass "Mobile documentation exists"

        # Check for key sections
        local sections=(
            "Quick Start"
            "Mobile Apps"
            "Android"
            "iOS"
            "QR Code"
            "Troubleshooting"
        )

        for section in "${sections[@]}"; do
            if grep -qi "$section" "$mobile_doc"; then
                test_pass "Documentation has section: $section"
            else
                test_fail "Documentation missing section: $section"
            fi
        done
    else
        test_fail "Mobile documentation not found"
    fi
}

test_qr_visual_output() {
    test_section "Visual QR Code Test"

    if ! command -v qrencode &>/dev/null; then
        test_skip "qrencode not installed"
        return 0
    fi

    echo "Generating test QR code for: https://doom-coding.dev"
    echo ""

    if qrencode -t ansiutf8 -m 2 "https://doom-coding.dev"; then
        echo ""
        echo "    ↑ Scan this QR code with your phone"
        echo ""
        test_pass "Visual QR code displayed"
    else
        test_fail "Failed to display visual QR code"
    fi
}

# ===========================================
# MAIN
# ===========================================
main() {
    echo -e "${GREEN}🔬 Doom Coding QR Integration Tests${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    # Run all tests
    test_qrencode_installed
    test_qr_generation_basic
    test_qr_generation_url
    test_qr_output_formats
    test_health_check_qr_flag
    test_install_script_qr_functions
    test_go_qr_library
    test_mobile_documentation
    test_qr_visual_output

    # Summary
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "Test Results: ${GREEN}$TESTS_PASSED passed${NC}, ${RED}$TESTS_FAILED failed${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    if [[ $TESTS_FAILED -eq 0 ]]; then
        echo ""
        echo -e "${GREEN}🎉 All QR integration tests passed!${NC}"
        exit 0
    else
        echo ""
        echo -e "${RED}⚠ Some tests failed. Review the output above.${NC}"
        exit 1
    fi
}

main "$@"
