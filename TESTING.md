# Testing Guide for delinea-netconfig

This guide will help you install Go, build the project, and test it.

## Step 1: Install Go

### macOS (using Homebrew)
```bash
# Install Homebrew if you don't have it
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install Go
brew install go

# Verify installation
go version
```

### macOS (using official installer)
1. Download from: https://go.dev/dl/
2. Download the `.pkg` file for macOS
3. Run the installer
4. Open a new terminal and run: `go version`

### Alternative: Using asdf (version manager)
```bash
# Install asdf
brew install asdf

# Add Go plugin
asdf plugin add golang

# Install Go 1.23
asdf install golang 1.23.0
asdf global golang 1.23.0

# Verify
go version
```

## Step 2: Download Dependencies

Once Go is installed:

```bash
cd delinea-netconfig

# Download all dependencies
go mod download

# Verify dependencies
go mod verify
```

## Step 3: Build the Project

```bash
# Build the binary
make build

# Or manually:
go build -o delinea-netconfig ./cmd/delinea-netconfig

# Verify the binary was created
ls -lh delinea-netconfig
```

## Step 4: Run Basic Tests

### Test 1: Show Help
```bash
./delinea-netconfig --help
```

**Expected output**: Help text showing available commands

### Test 2: Show Version
```bash
./delinea-netconfig --version
```

**Expected output**: `delinea-netconfig version 0.1.0`

### Test 3: Validate JSON
```bash
./delinea-netconfig validate -f testdata/network-requirements.json
```

**Expected output**:
```
✓ Valid JSON structure
✓ Schema version: 1.0.0
✓ All required fields present
✓ [number] IPv4 ranges validated
✓ [number] IPv6 ranges validated
✓ [number] hostnames validated
✓ [number] services validated ([number] outbound, [number] inbound)
✓ [number] regions found
```

### Test 4: Convert to CSV
```bash
./delinea-netconfig convert -f testdata/network-requirements.json --format csv | head -20
```

**Expected output**: CSV with headers and network entries

### Test 5: Convert to YAML
```bash
./delinea-netconfig convert -f testdata/network-requirements.json --format yaml | head -30
```

**Expected output**: YAML structure with network requirements

### Test 6: Convert to Terraform
```bash
./delinea-netconfig convert -f testdata/network-requirements.json --format terraform | head -30
```

**Expected output**: Terraform variables

### Test 7: Save to File
```bash
./delinea-netconfig convert -f testdata/network-requirements.json --format terraform -o output.tf
cat output.tf | head -30
```

**Expected output**: Terraform variables saved to output.tf

### Test 8: Multiple Formats to Directory
```bash
./delinea-netconfig convert -f testdata/network-requirements.json --format csv,yaml,terraform --output-dir ./test-output
ls -lh test-output/
```

**Expected output**: Three files created:
- test-output/output.csv
- test-output/output.yaml
- test-output/output.tf

### Test 9: Verbose Mode
```bash
./delinea-netconfig convert -f testdata/network-requirements.json --format csv -v | head -20
```

**Expected output**: Verbose logging messages plus CSV output

## Step 5: Run Go Tests (when we add them)

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests for specific package
go test -v ./internal/parser/
```

## Step 6: Use Make Targets

```bash
# Quick test all formats
make test-all-formats

# Test CSV conversion
make test-csv

# Test YAML conversion
make test-yaml

# Test Terraform conversion
make test-terraform

# Test validation
make test-validate
```

## Troubleshooting

### Issue: "command not found: go"
**Solution**: Go is not installed or not in your PATH. Follow Step 1.

### Issue: "cannot find package"
**Solution**: Run `go mod download` to download dependencies.

### Issue: "permission denied: ./delinea-netconfig"
**Solution**: Make the binary executable: `chmod +x delinea-netconfig`

### Issue: Build fails with errors
**Solution**:
1. Make sure you're in the project directory
2. Run `go mod tidy` to clean up dependencies
3. Try building again with `make build`

## What to Look For

When testing, verify:

✅ **No build errors** - The project compiles successfully
✅ **Help works** - `--help` shows documentation
✅ **Validation passes** - `validate` command shows all checks pass
✅ **CSV output** - Proper CSV format with headers
✅ **YAML output** - Valid YAML structure
✅ **Terraform output** - Valid HCL syntax
✅ **Files created** - Output files are created with correct extensions

## Sample Output Examples

### CSV Output (first few lines)
```csv
direction,service,region,type,value,protocol,ports,description,redundancy
outbound,platform_ssc_ips,global,ipv4,199.83.128.0/21,tcp,443,WAF IP ranges,
outbound,platform_ssc_ips,global,ipv4,198.143.32.0/19,tcp,443,WAF IP ranges,
```

### YAML Output (sample)
```yaml
delinea_network_requirements:
  outbound:
    platform_ssc_ips:
      global:
      - type: ipv4
        values:
        - 199.83.128.0/21
        - 198.143.32.0/19
```

### Terraform Output (sample)
```hcl
# Delinea Platform Network Requirements
# Generated by delinea-netconfig

variable "delinea_outbound_platform_ssc_ips_global_ipv4" {
  description = "platform_ssc_ips - WAF IP ranges (global)"
  type        = list(string)
  default = [
    "199.83.128.0/21",
    "198.143.32.0/19",
  ]
}
```

## Next Steps After Testing

Once all tests pass:

1. ✅ **Basic functionality works**
2. 🔄 **Add more converters** (Ansible, AWS SG, Cisco, PAN-OS)
3. 🔄 **Write unit tests**
4. 🔄 **Set up CI/CD**
5. 🔄 **Create release binaries**

## Getting Help

If you encounter issues:

1. Check this testing guide
2. Review the [README.md](README.md)
3. Check the [PLAN.md](PLAN.md) for design decisions
4. Open an issue on GitHub (once repo is public)

---

**Happy Testing!** 🚀
