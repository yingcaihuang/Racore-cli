#!/usr/bin/env bash
#
# test_domain_lifecycle.sh — Full domain lifecycle test
# Tests: create → list → query config → enable → disable → delete
#

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS_COUNT=0
FAIL_COUNT=0
TOTAL_COUNT=0

DOMAIN="v5.bbc.cfai.work"
SOURCE="www2.myccdn.info"
TYPE="oversea"

run_test() {
    local desc="$1"
    shift
    TOTAL_COUNT=$((TOTAL_COUNT + 1))

    printf "${YELLOW}[STEP %d]${NC} %s\n" "$TOTAL_COUNT" "$desc"
    printf "  CMD: %s\n" "$*"

    output=$("$@" 2>&1)
    exit_code=$?

    if [ $exit_code -eq 0 ]; then
        printf "  ${GREEN}PASS${NC}\n"
        # Show first 5 lines of output for context
        echo "$output" | head -5 | sed 's/^/  > /'
        printf "\n"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        printf "  ${RED}FAIL${NC} (exit code: %d)\n" "$exit_code"
        printf "  OUTPUT:\n%s\n\n" "$output"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
    
    # Small delay between API calls to be nice to the server
    sleep 1
}

echo "================================================"
echo " Domain Lifecycle Test: $DOMAIN"
echo " Origin: $SOURCE"
echo "================================================"
echo ""

# Build
echo "Building binary..."
if ! make 2>&1 | tail -1; then
    printf "${RED}Build failed!${NC}\n"
    exit 1
fi
echo ""

# Step 1: Create domain
run_test "Create domain $DOMAIN (source: $SOURCE)" \
    ./racore-cli domain create --domain "$DOMAIN" --type "$TYPE" --source "$SOURCE"

# Step 2: List domains and check it appears
run_test "List domains (verify $DOMAIN exists)" \
    ./racore-cli domain list --filter "$DOMAIN"

# Step 3: Query source config
run_test "Query source configuration" \
    ./racore-cli domain source get --domain "$DOMAIN"

# Step 4: Query SSL config
run_test "Query SSL configuration" \
    ./racore-cli domain ssl get --domain "$DOMAIN"

# Step 5: Query HTTP/2 config
run_test "Query HTTP/2 configuration" \
    ./racore-cli domain http2 get --domain "$DOMAIN"

# Step 6: Query cache-policy
run_test "Query cache policy" \
    ./racore-cli domain cache-policy get --domain "$DOMAIN"

# Step 7: Enable domain
run_test "Enable domain" \
    ./racore-cli domain enable --domain "$DOMAIN"

# Step 8: Disable domain
run_test "Disable domain" \
    ./racore-cli domain disable --domain "$DOMAIN"

# Step 9: Delete domain
run_test "Delete domain" \
    ./racore-cli domain delete --domain "$DOMAIN"

# Step 10: Verify domain is gone
run_test "Verify domain deleted (list should not contain $DOMAIN)" \
    ./racore-cli domain list --filter "$DOMAIN"

# Summary
echo ""
echo "================================================"
echo " LIFECYCLE TEST SUMMARY"
echo "================================================"
printf " Total:  %d\n" "$TOTAL_COUNT"
printf " ${GREEN}Passed: %d${NC}\n" "$PASS_COUNT"
printf " ${RED}Failed: %d${NC}\n" "$FAIL_COUNT"
echo "================================================"

if [ "$FAIL_COUNT" -gt 0 ]; then
    printf "\n${RED}Some steps failed!${NC}\n"
    exit 1
else
    printf "\n${GREEN}All steps passed!${NC}\n"
    exit 0
fi
