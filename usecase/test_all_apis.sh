#!/usr/bin/env bash
#
# test_all_apis.sh — Integration test script for racore-cli
# Tests all CLI commands against the domain v1.bbc.cfai.work
# Skips destructive operations (create/delete/update/cancel/close/reopen/send-message)
#

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASS_COUNT=0
FAIL_COUNT=0
TOTAL_COUNT=0

DOMAIN="v1.bbc.cfai.work"
START_TIME="2026-06-29"
END_TIME="2026-07-06"

# run_test executes a command and reports PASS/FAIL
# Usage: run_test "description" command [args...]
run_test() {
    local desc="$1"
    shift
    TOTAL_COUNT=$((TOTAL_COUNT + 1))

    printf "${YELLOW}[TEST %d]${NC} %s\n" "$TOTAL_COUNT" "$desc"
    printf "  CMD: %s\n" "$*"

    output=$("$@" 2>&1)
    exit_code=$?

    if [ $exit_code -eq 0 ]; then
        printf "  ${GREEN}PASS${NC}\n\n"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        printf "  ${RED}FAIL${NC} (exit code: %d)\n" "$exit_code"
        printf "  OUTPUT:\n%s\n\n" "$output"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
}

echo "=============================================="
echo " racore-cli Integration Test Suite"
echo "=============================================="
echo ""

# Step 1: Build the binary
echo "Building binary with make..."
if ! make; then
    echo "${RED}ERROR: Build failed. Aborting tests.${NC}"
    exit 1
fi
echo ""

# ──────────────────────────────────────────────
# Auth group
# ──────────────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Auth Commands"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
run_test "whoami" ./racore-cli whoami

# ──────────────────────────────────────────────
# Domain group
# ──────────────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Domain Commands"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
run_test "domain list" ./racore-cli domain list
run_test "domain list --filter" ./racore-cli domain list --filter "$DOMAIN"
run_test "domain source get" ./racore-cli domain source get --domain "$DOMAIN"
run_test "domain ssl get" ./racore-cli domain ssl get --domain "$DOMAIN"
run_test "domain enforce-https get" ./racore-cli domain enforce-https get --domain "$DOMAIN"
run_test "domain ip-filter get" ./racore-cli domain ip-filter get --domain "$DOMAIN"
run_test "domain referer-filter get" ./racore-cli domain referer-filter get --domain "$DOMAIN"
run_test "domain ua-filter get" ./racore-cli domain ua-filter get --domain "$DOMAIN"
run_test "domain origin-protocol get" ./racore-cli domain origin-protocol get --domain "$DOMAIN"
run_test "domain http2 get" ./racore-cli domain http2 get --domain "$DOMAIN"
run_test "domain http3 get" ./racore-cli domain http3 get --domain "$DOMAIN"
run_test "domain tls-version get" ./racore-cli domain tls-version get --domain "$DOMAIN"
run_test "domain compress get" ./racore-cli domain compress get --domain "$DOMAIN"
run_test "domain ipv6 get" ./racore-cli domain ipv6 get --domain "$DOMAIN"
run_test "domain cache-policy get" ./racore-cli domain cache-policy get --domain "$DOMAIN"
run_test "domain origin-host get" ./racore-cli domain origin-host get --domain "$DOMAIN"
run_test "domain origin-timeout get" ./racore-cli domain origin-timeout get --domain "$DOMAIN"
run_test "domain geo-restriction get" ./racore-cli domain geo-restriction get --domain "$DOMAIN"
run_test "domain request-headers get" ./racore-cli domain request-headers get --domain "$DOMAIN"
run_test "domain response-headers get" ./racore-cli domain response-headers get --domain "$DOMAIN"
run_test "domain request-header-policy get" ./racore-cli domain request-header-policy get --domain "$DOMAIN"
run_test "domain response-header-policy get" ./racore-cli domain response-header-policy get --domain "$DOMAIN"

# ──────────────────────────────────────────────
# Cache group
# ──────────────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Cache Commands"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
run_test "cache list-policies" ./racore-cli cache list-policies --domain "$DOMAIN"
run_test "cache list-policies --type custom" ./racore-cli cache list-policies --domain "$DOMAIN" --type custom
run_test "cache list-origin-request-policies" ./racore-cli cache list-origin-request-policies --domain "$DOMAIN"
run_test "cache list-response-header-policies" ./racore-cli cache list-response-header-policies --domain "$DOMAIN"
run_test "cache prewarm-regions" ./racore-cli cache prewarm-regions --url "https://${DOMAIN}/"
run_test "cache prewarm-pop --region" ./racore-cli cache prewarm-pop --region us-east-1

# ──────────────────────────────────────────────
# Stats group
# ──────────────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Stats Commands"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
run_test "stats flow" ./racore-cli stats flow --start-time "$START_TIME" --end-time "$END_TIME" --domains "$DOMAIN"
run_test "stats request" ./racore-cli stats request --start-time "$START_TIME" --end-time "$END_TIME" --domains "$DOMAIN"
run_test "stats hit-flow" ./racore-cli stats hit-flow --start-time "$START_TIME" --end-time "$END_TIME" --domains "$DOMAIN"
run_test "stats hit-request" ./racore-cli stats hit-request --start-time "$START_TIME" --end-time "$END_TIME" --domains "$DOMAIN"
run_test "stats http-code" ./racore-cli stats http-code --start-time "$START_TIME" --end-time "$END_TIME" --domains "$DOMAIN"
run_test "stats http-code-detail" ./racore-cli stats http-code-detail --start-time "$START_TIME" --end-time "$END_TIME" --domains "$DOMAIN"
run_test "stats district" ./racore-cli stats district --start-time "$START_TIME" --end-time "$END_TIME" --domains "$DOMAIN"
run_test "stats iso-country" ./racore-cli stats iso-country
run_test "stats top-domain" ./racore-cli stats top-domain --start-time "$START_TIME" --end-time "$END_TIME"
run_test "stats top-url" ./racore-cli stats top-url --scope yesterday
run_test "stats top-referer" ./racore-cli stats top-referer --scope yesterday
run_test "stats top-ua" ./racore-cli stats top-ua --scope yesterday

# ──────────────────────────────────────────────
# Cert group
# ──────────────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Cert Commands"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
run_test "cert list" ./racore-cli cert list

# ──────────────────────────────────────────────
# Workorder group (read-only)
# ──────────────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Workorder Commands (read-only)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
run_test "workorder list" ./racore-cli workorder list
run_test "workorder types" ./racore-cli workorder types

# ──────────────────────────────────────────────
# Log group
# ──────────────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Log Commands"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
run_test "log list" ./racore-cli log list --domain "$DOMAIN"

# ──────────────────────────────────────────────
# Summary
# ──────────────────────────────────────────────
echo ""
echo "=============================================="
echo " TEST SUMMARY"
echo "=============================================="
printf " Total:  %d\n" "$TOTAL_COUNT"
printf " ${GREEN}Passed: %d${NC}\n" "$PASS_COUNT"
printf " ${RED}Failed: %d${NC}\n" "$FAIL_COUNT"
echo "=============================================="

if [ "$FAIL_COUNT" -gt 0 ]; then
    printf "\n${RED}Some tests failed!${NC}\n"
    exit 1
else
    printf "\n${GREEN}All tests passed!${NC}\n"
    exit 0
fi
