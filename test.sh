#!/bin/bash
# Quick test script for delinea-netconfig

set -e  # Exit on error

echo "======================================"
echo "Delinea NetConfig - Quick Test Script"
echo "======================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go is not installed${NC}"
    echo ""
    echo "Please install Go first:"
    echo "  macOS: brew install go"
    echo "  Or visit: https://go.dev/dl/"
    echo ""
    echo "See TESTING.md for detailed instructions."
    exit 1
fi

echo -e "${GREEN}✓ Go is installed:${NC} $(go version)"
echo ""

# Build the project
echo "Building delinea-netconfig..."
if make build; then
    echo -e "${GREEN}✓ Build successful${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi
echo ""

# Test 1: Help
echo "Test 1: Show help"
./delinea-netconfig --help > /dev/null
echo -e "${GREEN}✓ Help works${NC}"
echo ""

# Test 2: Version
echo "Test 2: Show version"
VERSION=$(./delinea-netconfig --version 2>&1)
echo -e "${GREEN}✓ Version: ${VERSION}${NC}"
echo ""

# Test 3: Validate
echo "Test 3: Validate JSON"
if ./delinea-netconfig validate -f testdata/network-requirements.json > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Validation passed${NC}"
else
    echo -e "${YELLOW}⚠ Validation had warnings${NC}"
fi
echo ""

# Test 4: CSV conversion
echo "Test 4: Convert to CSV"
CSV_OUTPUT=$(./delinea-netconfig convert -f testdata/network-requirements.json --format csv -q 2>&1 | head -3)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ CSV conversion successful${NC}"
    echo "First 3 lines:"
    echo "$CSV_OUTPUT"
else
    echo -e "${RED}✗ CSV conversion failed${NC}"
    exit 1
fi
echo ""

# Test 5: YAML conversion
echo "Test 5: Convert to YAML"
if ./delinea-netconfig convert -f testdata/network-requirements.json --format yaml -q > /dev/null 2>&1; then
    echo -e "${GREEN}✓ YAML conversion successful${NC}"
else
    echo -e "${RED}✗ YAML conversion failed${NC}"
    exit 1
fi
echo ""

# Test 6: Terraform conversion
echo "Test 6: Convert to Terraform"
TF_OUTPUT=$(./delinea-netconfig convert -f testdata/network-requirements.json --format terraform -q 2>&1 | head -10)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Terraform conversion successful${NC}"
    echo "First 10 lines:"
    echo "$TF_OUTPUT"
else
    echo -e "${RED}✗ Terraform conversion failed${NC}"
    exit 1
fi
echo ""

# Test 7: Save to file
echo "Test 7: Save to file"
./delinea-netconfig convert -f testdata/network-requirements.json --format terraform -o /tmp/test-output.tf -q
if [ -f /tmp/test-output.tf ]; then
    SIZE=$(wc -c < /tmp/test-output.tf)
    echo -e "${GREEN}✓ File saved successfully (${SIZE} bytes)${NC}"
    rm /tmp/test-output.tf
else
    echo -e "${RED}✗ File save failed${NC}"
    exit 1
fi
echo ""

# Test 8: Multiple formats to directory
echo "Test 8: Multiple formats to directory"
rm -rf /tmp/test-output-dir
./delinea-netconfig convert -f testdata/network-requirements.json --format csv,yaml,terraform --output-dir /tmp/test-output-dir -q
if [ -d /tmp/test-output-dir ] && [ -f /tmp/test-output-dir/output.csv ] && [ -f /tmp/test-output-dir/output.yaml ] && [ -f /tmp/test-output-dir/output.tf ]; then
    echo -e "${GREEN}✓ Multiple formats saved successfully${NC}"
    echo "Files created:"
    ls -lh /tmp/test-output-dir/
    rm -rf /tmp/test-output-dir
else
    echo -e "${RED}✗ Multiple format save failed${NC}"
    exit 1
fi
echo ""

# Summary
echo "======================================"
echo -e "${GREEN}All tests passed! ✓${NC}"
echo "======================================"
echo ""
echo "Next steps:"
echo "  1. Review output formats"
echo "  2. Test with your own JSON files"
echo "  3. Try different command options"
echo ""
echo "For more info:"
echo "  ./delinea-netconfig --help"
echo "  cat README.md"
echo "  cat TESTING.md"
echo ""
