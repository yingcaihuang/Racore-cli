#!/usr/bin/env bash
#
# test_filter_debug.sh — Debug IP Filter and Referer Filter parameters
# Tests various parameter combinations to find the correct ones
#

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

DOMAIN="v6.bbc.cfai.work"
PASS_COUNT=0
FAIL_COUNT=0
TOTAL_COUNT=0

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
        echo "$output" | sed 's/^/  > /'
        printf "\n"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        printf "  ${RED}✗ FAIL${NC} (exit code: %d)\n" "$exit_code"
        echo "$output" | sed 's/^/  > /'
        printf "\n"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi

    sleep 1
}

echo "══════════════════════════════════════════════════"
echo " Filter Parameter Debug"
echo " Domain: $DOMAIN"
echo "══════════════════════════════════════════════════"
echo ""

make 2>&1 | tail -1
echo ""

# ─── IP Filter Debug ──────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " IP Filter: Trying different parameter combinations"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "IP Filter GET (current state)" \
    ./racore-cli domain ip-filter get --domain "$DOMAIN"

# Try: type field (from OpenAPI doc)
run_test "IP: type=white, value=[1.2.3.4]" \
    ./racore-cli domain ip-filter set --domain "$DOMAIN" --config '{"type":"white","value":["1.2.3.4"]}'

# Try: state + type together
run_test "IP: state=on, type=white, value=[1.2.3.4]" \
    ./racore-cli domain ip-filter set --domain "$DOMAIN" --config '{"state":"on","type":"white","value":["1.2.3.4"]}'

# Try: just state with different values
run_test "IP: state=on, value=[1.2.3.4]" \
    ./racore-cli domain ip-filter set --domain "$DOMAIN" --config '{"state":"on","value":["1.2.3.4"]}'

# Try: state=1 (numeric)
run_test "IP: state=1, type=white, value=[1.2.3.4]" \
    ./racore-cli domain ip-filter set --domain "$DOMAIN" --config '{"state":"1","type":"white","value":["1.2.3.4"]}'

# Try: is_open + type (from first test script attempt)
run_test "IP: is_open=1, type=white, value=[1.2.3.4]" \
    ./racore-cli domain ip-filter set --domain "$DOMAIN" --config '{"is_open":"1","type":"white","value":["1.2.3.4"]}'

# Check state after
run_test "IP Filter GET (after tests)" \
    ./racore-cli domain ip-filter get --domain "$DOMAIN"

# Reset to off
run_test "IP: state=off, value=[] (reset)" \
    ./racore-cli domain ip-filter set --domain "$DOMAIN" --config '{"state":"off","value":[]}'

echo ""

# ─── Referer Filter Debug ─────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Referer Filter: Trying different parameter combinations"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

run_test "Referer Filter GET (current state)" \
    ./racore-cli domain referer-filter get --domain "$DOMAIN"

# Try: type field only (from OpenAPI doc)
run_test "Ref: type=white, value=[example.com], allow_empty=on" \
    ./racore-cli domain referer-filter set --domain "$DOMAIN" --config '{"type":"white","value":["example.com"],"allow_empty":"on"}'

# Try: state + type together
run_test "Ref: state=on, type=white, value=[example.com], allow_empty=on" \
    ./racore-cli domain referer-filter set --domain "$DOMAIN" --config '{"state":"on","type":"white","value":["example.com"],"allow_empty":"on"}'

# Try: type=white with state=white
run_test "Ref: state=white, type=white, value=[example.com], allow_empty=on" \
    ./racore-cli domain referer-filter set --domain "$DOMAIN" --config '{"state":"white","type":"white","value":["example.com"],"allow_empty":"on"}'

# Try: state=1, type=white
run_test "Ref: state=1, type=white, value=[example.com], allow_empty=on" \
    ./racore-cli domain referer-filter set --domain "$DOMAIN" --config '{"state":"1","type":"white","value":["example.com"],"allow_empty":"on"}'

# Check state after
run_test "Referer Filter GET (after tests)" \
    ./racore-cli domain referer-filter get --domain "$DOMAIN"

# Try reset: type=off
run_test "Ref: type=off, value=[], allow_empty=on (reset)" \
    ./racore-cli domain referer-filter set --domain "$DOMAIN" --config '{"type":"off","value":[],"allow_empty":"on"}'

# Try reset: state=off, type=off
run_test "Ref: state=off, type=off, value=[], allow_empty=on (reset)" \
    ./racore-cli domain referer-filter set --domain "$DOMAIN" --config '{"state":"off","type":"off","value":[],"allow_empty":"on"}'

# Final check
run_test "Referer Filter GET (final)" \
    ./racore-cli domain referer-filter get --domain "$DOMAIN"

# ─── Summary ──────────────────────────────────────
echo ""
echo "══════════════════════════════════════════════════"
echo " DEBUG SUMMARY"
echo "══════════════════════════════════════════════════"
printf " Total:  %d\n" "$TOTAL_COUNT"
printf " ${GREEN}Passed: %d${NC}\n" "$PASS_COUNT"
printf " ${RED}Failed: %d${NC}\n" "$FAIL_COUNT"
echo "══════════════════════════════════════════════════"
