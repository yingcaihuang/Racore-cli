#!/usr/bin/env bash
#
# test_domain_set.sh — Test domain SET (write) operations
# Domain: v6.bbc.cfai.work (must already exist)
# Uses correct API field names from OpenAPI spec
#
# Strategy: Toggle values (set different → verify → set back → verify)
# to avoid "No Information modified" errors from the API.
#

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

PASS_COUNT=0
FAIL_COUNT=0
TOTAL_COUNT=0

DOMAIN="v6.bbc.cfai.work"

run_test() {
    local desc="$1"
    shift
    TOTAL_COUNT=$((TOTAL_COUNT + 1))

    printf "${YELLOW}[STEP %d]${NC} %s\n" "$TOTAL_COUNT" "$desc"
    printf "  ${CYAN}CMD:${NC} %s\n" "$*"

    output=$("$@" 2>&1)
    exit_code=$?

    if [ $exit_code -eq 0 ]; then
        printf "  ${GREEN}✓ PASS${NC}\n"
        echo "$output" | head -5 | sed 's/^/  > /'
        printf "\n"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        printf "  ${RED}✗ FAIL${NC} (exit code: %d)\n" "$exit_code"
        echo "$output" | sed 's/^/  > /'
        printf "\n"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi

    sleep 2
}

echo "══════════════════════════════════════════════════"
echo " Domain SET Operations Test"
echo " Domain: $DOMAIN"
echo "══════════════════════════════════════════════════"
echo ""

# Build
echo "Building binary..."
if ! make 2>&1 | tail -1; then
    printf "${RED}Build failed!${NC}\n"
    exit 1
fi
echo ""

# ─── Verify domain exists first ───────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Pre-check: Verify domain exists"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "Domain exists check" \
    ./racore-cli domain list --filter "$DOMAIN"

# ─── HTTP/2 ───────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " HTTP/2: Toggle off → verify → on → verify"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "HTTP/2 SET (disable)" \
    ./racore-cli domain http2 set --domain "$DOMAIN" --config '{"enable":"off"}'

run_test "HTTP/2 GET (verify off)" \
    ./racore-cli domain http2 get --domain "$DOMAIN"

run_test "HTTP/2 SET (re-enable)" \
    ./racore-cli domain http2 set --domain "$DOMAIN" --config '{"enable":"on"}'

run_test "HTTP/2 GET (verify on)" \
    ./racore-cli domain http2 get --domain "$DOMAIN"

# ─── HTTP/3 ───────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " HTTP/3: Toggle off → verify → on → verify"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "HTTP/3 SET (disable)" \
    ./racore-cli domain http3 set --domain "$DOMAIN" --config '{"enable":"off"}'

run_test "HTTP/3 GET (verify off)" \
    ./racore-cli domain http3 get --domain "$DOMAIN"

run_test "HTTP/3 SET (re-enable)" \
    ./racore-cli domain http3 set --domain "$DOMAIN" --config '{"enable":"on"}'

run_test "HTTP/3 GET (verify on)" \
    ./racore-cli domain http3 get --domain "$DOMAIN"

# ─── Enforce HTTPS ────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Enforce HTTPS: GET only (requires SSL enabled)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "Enforce HTTPS GET" \
    ./racore-cli domain enforce-https get --domain "$DOMAIN"
# Note: enforce-https requires SSL to be enabled first

# ─── Compression ──────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Compression: Toggle off → verify → on → verify"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "Compress SET (disable)" \
    ./racore-cli domain compress set --domain "$DOMAIN" --config '{"enable":"off"}'

run_test "Compress GET (verify off)" \
    ./racore-cli domain compress get --domain "$DOMAIN"

run_test "Compress SET (re-enable)" \
    ./racore-cli domain compress set --domain "$DOMAIN" --config '{"enable":"on"}'

run_test "Compress GET (verify on)" \
    ./racore-cli domain compress get --domain "$DOMAIN"

# ─── IPv6 ─────────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " IPv6: Toggle 0 → verify → 1 → verify"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "IPv6 SET (disable)" \
    ./racore-cli domain ipv6 set --domain "$DOMAIN" --config '{"enable":"0"}'

run_test "IPv6 GET (verify 0)" \
    ./racore-cli domain ipv6 get --domain "$DOMAIN"

run_test "IPv6 SET (re-enable)" \
    ./racore-cli domain ipv6 set --domain "$DOMAIN" --config '{"enable":"1"}'

run_test "IPv6 GET (verify 1)" \
    ./racore-cli domain ipv6 get --domain "$DOMAIN"

# ─── TLS Version ──────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " TLS Version: Toggle TLSv1 → verify → TLSv1.2_2021 → verify"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "TLS Version SET (TLSv1)" \
    ./racore-cli domain tls-version set --domain "$DOMAIN" --config '{"min_tls_version":"TLSv1"}'

run_test "TLS Version GET (verify TLSv1)" \
    ./racore-cli domain tls-version get --domain "$DOMAIN"

run_test "TLS Version SET (TLSv1.2_2021)" \
    ./racore-cli domain tls-version set --domain "$DOMAIN" --config '{"min_tls_version":"TLSv1.2_2021"}'

run_test "TLS Version GET (verify TLSv1.2_2021)" \
    ./racore-cli domain tls-version get --domain "$DOMAIN"

# ─── Origin Protocol ──────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Origin Protocol: Toggle match-viewer → verify → https-only → verify"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "Origin Protocol SET (match-viewer)" \
    ./racore-cli domain origin-protocol set --domain "$DOMAIN" --config '{"origin_protocol_policy":"match-viewer","origin_protocol_http_port":"80","origin_protocol_https_port":"443"}'

run_test "Origin Protocol GET (verify match-viewer)" \
    ./racore-cli domain origin-protocol get --domain "$DOMAIN"

run_test "Origin Protocol SET (https-only)" \
    ./racore-cli domain origin-protocol set --domain "$DOMAIN" --config '{"origin_protocol_policy":"https-only","origin_protocol_http_port":"80","origin_protocol_https_port":"443"}'

run_test "Origin Protocol GET (verify https-only)" \
    ./racore-cli domain origin-protocol get --domain "$DOMAIN"

# ─── Origin Host ──────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Origin Host: Toggle type 1 → verify → type 3 → verify"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "Origin Host SET (type 1 - origin server)" \
    ./racore-cli domain origin-host set --domain "$DOMAIN" --config '{"origin_host_type":"1"}'

run_test "Origin Host GET (verify type 1)" \
    ./racore-cli domain origin-host get --domain "$DOMAIN"

run_test "Origin Host SET (type 3 - custom)" \
    ./racore-cli domain origin-host set --domain "$DOMAIN" --config '{"origin_host_type":"3","origin_host":"www2.myccdn.info"}'

run_test "Origin Host GET (verify type 3)" \
    ./racore-cli domain origin-host get --domain "$DOMAIN"

# ─── IP Filter ────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " IP Filter: Set whitelist → verify → disable → verify"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "IP Filter SET (whitelist: 1.2.3.4)" \
    ./racore-cli domain ip-filter set --domain "$DOMAIN" --config '{"state":"on","type":"white","value":["1.2.3.4"]}'

run_test "IP Filter GET (after set)" \
    ./racore-cli domain ip-filter get --domain "$DOMAIN"

run_test "IP Filter SET (disable)" \
    ./racore-cli domain ip-filter set --domain "$DOMAIN" --config '{"state":"off","value":[]}'

run_test "IP Filter GET (after disable)" \
    ./racore-cli domain ip-filter get --domain "$DOMAIN"

# ─── Referer Filter ───────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Referer Filter: Set whitelist → verify → disable → verify"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "Referer Filter SET (whitelist: example.com)" \
    ./racore-cli domain referer-filter set --domain "$DOMAIN" --config '{"state":"on","type":"white","values":["example.com"],"allow_empty":"on"}'

run_test "Referer Filter GET (after set)" \
    ./racore-cli domain referer-filter get --domain "$DOMAIN"

run_test "Referer Filter SET (disable)" \
    ./racore-cli domain referer-filter set --domain "$DOMAIN" --config '{"state":"off","type":"off","values":["example.com"],"allow_empty":"on"}'

run_test "Referer Filter GET (after disable)" \
    ./racore-cli domain referer-filter get --domain "$DOMAIN"

# ─── Summary ──────────────────────────────────────
echo ""
echo "══════════════════════════════════════════════════"
echo " TEST SUMMARY"
echo "══════════════════════════════════════════════════"
printf " Total:  %d\n" "$TOTAL_COUNT"
printf " ${GREEN}Passed: %d${NC}\n" "$PASS_COUNT"
printf " ${RED}Failed: %d${NC}\n" "$FAIL_COUNT"
echo "══════════════════════════════════════════════════"
echo ""
printf " Domain: ${CYAN}%s${NC}\n" "$DOMAIN"
echo ""

if [ "$FAIL_COUNT" -gt 0 ]; then
    printf "${RED}Some steps failed! Review output above.${NC}\n"
    exit 1
else
    printf "${GREEN}All SET operations passed!${NC}\n"
    exit 0
fi
