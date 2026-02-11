#!/bin/bash
# Golden file integration tests
# Compares actual output against known-good reference files

set -e

BINARY="./delinea-netconfig"
TEST_FILE="testdata/tenant-test.json"
GOLDEN_DIR="testdata/golden"
TEMP_DIR="/tmp/delinea-netconfig-test-$$"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create temp directory
mkdir -p "$TEMP_DIR"

# Cleanup function
cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

echo "======================================"
echo "Delinea NetConfig - Golden File Tests"
echo "======================================"
echo ""

PASSED=0
FAILED=0

# Test function
test_format() {
    local format=$1
    local golden_file=$2
    local extra_flags=${3:-""}

    echo -n "Testing $format format... "

    # Generate output
    if ! $BINARY convert -f "$TEST_FILE" --format "$format" $extra_flags -q > "$TEMP_DIR/output" 2>&1; then
        echo -e "${RED}FAILED${NC} (conversion error)"
        FAILED=$((FAILED + 1))
        return 1
    fi

    # Compare with golden file
    if diff -q "$GOLDEN_DIR/$golden_file" "$TEMP_DIR/output" > /dev/null 2>&1; then
        echo -e "${GREEN}PASSED${NC}"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo -e "${RED}FAILED${NC} (output differs from golden file)"
        echo ""
        echo "Differences:"
        diff "$GOLDEN_DIR/$golden_file" "$TEMP_DIR/output" || true
        echo ""
        FAILED=$((FAILED + 1))
        return 1
    fi
}

# Test CSV format
test_format "csv" "tenant-test.csv"

# Test YAML format
test_format "yaml" "tenant-test.yaml"

# Test Terraform format
test_format "terraform" "tenant-test.tf"

# Test Ansible format
test_format "ansible" "tenant-test.yml"

# Test AWS Security Group format
test_format "aws-sg" "tenant-test-aws-sg.json"

# Test Cisco ACL format
test_format "cisco" "tenant-test.acl"

# Test PAN-OS XML format
test_format "panos" "tenant-test.xml"

echo ""
echo "======================================"
echo "Results: $PASSED passed, $FAILED failed"
echo "======================================"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All golden file tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed${NC}"
    exit 1
fi
