#!/usr/bin/env bash
#
# test_domain_create.sh — Test domain creation and enable features
# Domain: v5.bbc.cfai.work | Source: www2.myccdn.info
# NOTE: Does NOT disable or delete the domain
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
SOURCE="www2.myccdn.info"
TYPE="oversea"

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
        echo "$output" | head -10 | sed 's/^/  > /'
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
echo " Domain Create & Enable Test"
echo " Domain: $DOMAIN"
echo " Source: $SOURCE"
echo "══════════════════════════════════════════════════"
echo ""

# Build
echo "Building binary..."
if ! make 2>&1 | tail -1; then
    printf "${RED}Build failed!${NC}\n"
    exit 1
fi
echo ""

# ─── Phase 1: Create Domain ───────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Phase 1: Create Domain (with auto cert matching)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "Create domain $DOMAIN (source: $SOURCE, auto-cert)" \
    ./racore-cli domain create --domain "$DOMAIN" --type "$TYPE" --source "$SOURCE"

# ─── Phase 2: Verify Domain Exists ────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Phase 2: Verify Domain Exists"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "List domains (filter: $DOMAIN)" \
    ./racore-cli domain list --filter "$DOMAIN"

# ─── Phase 3: Query Domain Configurations ─────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Phase 3: Query Domain Configurations"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "Query source configuration" \
    ./racore-cli domain source get --domain "$DOMAIN"

run_test "Query SSL configuration" \
    ./racore-cli domain ssl get --domain "$DOMAIN"

run_test "Query HTTP/2 configuration" \
    ./racore-cli domain http2 get --domain "$DOMAIN"

run_test "Query HTTP/3 configuration" \
    ./racore-cli domain http3 get --domain "$DOMAIN"

run_test "Query enforce-https configuration" \
    ./racore-cli domain enforce-https get --domain "$DOMAIN"

run_test "Query origin-protocol configuration" \
    ./racore-cli domain origin-protocol get --domain "$DOMAIN"

run_test "Query TLS version configuration" \
    ./racore-cli domain tls-version get --domain "$DOMAIN"

run_test "Query compression configuration" \
    ./racore-cli domain compress get --domain "$DOMAIN"

run_test "Query IPv6 configuration" \
    ./racore-cli domain ipv6 get --domain "$DOMAIN"

run_test "Query cache policy" \
    ./racore-cli domain cache-policy get --domain "$DOMAIN"

run_test "Query origin host" \
    ./racore-cli domain origin-host get --domain "$DOMAIN"

run_test "Query origin timeout" \
    ./racore-cli domain origin-timeout get --domain "$DOMAIN"

run_test "Query IP filter" \
    ./racore-cli domain ip-filter get --domain "$DOMAIN"

run_test "Query referer filter" \
    ./racore-cli domain referer-filter get --domain "$DOMAIN"

run_test "Query UA filter" \
    ./racore-cli domain ua-filter get --domain "$DOMAIN"

run_test "Query geo restriction" \
    ./racore-cli domain geo-restriction get --domain "$DOMAIN"

run_test "Query request headers" \
    ./racore-cli domain request-headers get --domain "$DOMAIN"

run_test "Query response headers" \
    ./racore-cli domain response-headers get --domain "$DOMAIN"

run_test "Query request header policy" \
    ./racore-cli domain request-header-policy get --domain "$DOMAIN"

run_test "Query response header policy" \
    ./racore-cli domain response-header-policy get --domain "$DOMAIN"

# ─── Phase 4: Enable Domain ───────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Phase 4: Enable Domain"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "Enable domain $DOMAIN" \
    ./racore-cli domain enable --domain "$DOMAIN"

# ─── Phase 5: Verify Domain State After Enable ────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Phase 5: Verify Domain State After Enable"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "List domain (verify enabled state)" \
    ./racore-cli domain list --filter "$DOMAIN"

run_test "Query SSL (verify cert bound after enable)" \
    ./racore-cli domain ssl get --domain "$DOMAIN"

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
echo " NOTE: Domain was NOT disabled or deleted."
echo " To clean up manually:"
echo "   ./racore-cli domain disable --domain $DOMAIN"
echo "   ./racore-cli domain delete --domain $DOMAIN"
echo ""

if [ "$FAIL_COUNT" -gt 0 ]; then
    printf "${RED}Some steps failed!${NC}\n"
    exit 1
else
    printf "${GREEN}All steps passed!${NC}\n"
    exit 0
fi
