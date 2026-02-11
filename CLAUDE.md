# CLAUDE.md - Delinea Network Requirements Converter

> **Repository:** delinea-netconfig
> **Organization:** DelineaXPM / delinea-platform
> **Type:** CLI Tool
> **Language:** Go 1.23+
> **Status:** Phase 3 Complete ✅ (Production Ready with Distribution)

---

## Project Overview

**delinea-netconfig** is a CLI tool that converts Delinea's Platform IP/CIDR network requirements JSON into various firewall and infrastructure-as-code formats.

### Purpose

Enable customers and internal teams to easily integrate Delinea's network requirements into their security infrastructure by providing conversions to multiple formats: CSV, YAML, Terraform, Ansible, AWS Security Groups, and more.

### Key Features

- **Multiple Output Formats:** CSV, YAML, Terraform, Ansible, AWS SG
- **Tenant Substitution:** Replace `<tenant>` placeholders with actual tenant names
- **Flexible Input:** Load from local files or remote URLs
- **Validation:** Validate JSON structure and network entries
- **CLI-First Design:** Easy integration into automation pipelines

---

## Current Status: Phase 3 Complete ✅

### What's Implemented

#### Phase 1 (MVP) ✅
- [x] Core CLI structure (Cobra)
- [x] JSON parser and normalizer
- [x] CSV, YAML, Terraform converters
- [x] Validation command
- [x] URL fetching support
- [x] Tenant substitution (`--tenant` flag)

#### Phase 2 (Production Ready) ✅
- [x] **Unit tests** (40+ test cases with race detection)
- [x] **Ansible converter** - Generates Ansible variables
- [x] **AWS Security Group converter** - Generates AWS SG JSON
- [x] **Integration tests** - Golden file testing
- [x] **CI/CD pipeline** - GitHub Actions with multi-platform builds
- [x] **Deterministic output** - Sorted entries for consistent results

#### Phase 3 (Enterprise Features & Distribution) ✅
- [x] **Cisco ACL converter** - Generate Cisco IOS ACL format with wildcard masks
- [x] **PAN-OS XML converter** - Generate Palo Alto Networks XML configuration
- [x] **`diff` command** - Compare two versions of network requirements files
- [x] **`info` command** - Show statistics and analysis of network requirements
- [x] **Shell completion** - Auto-completion for bash, zsh, fish, and PowerShell
- [x] **GoReleaser** - Automated multi-platform binary releases
- [x] **Docker images** - Multi-arch container images (amd64, arm64)
- [x] **Installation methods** - Install script, Homebrew tap, pre-built binaries
- [x] **Release automation** - GitHub Actions workflow for releases
- [x] **Comprehensive testing** - 100+ unit tests, 7 integration tests

### Test Coverage

```
✓ Unit tests: 100+ tests passing
✓ Integration tests: 7 formats validated
✓ Race detection: enabled and clean
✓ Coverage: Comprehensive across all converters, parser, CLI
```

---

## Architecture

### Project Structure

```
delinea-netconfig/
├── cmd/
│   └── delinea-netconfig/     # CLI entry point (main.go)
├── internal/
│   ├── cli/                   # CLI commands (convert, validate, diff, info)
│   │   ├── convert.go         # Convert command
│   │   ├── convert_test.go    # Tenant substitution tests
│   │   ├── validate.go        # Validate command
│   │   ├── diff.go            # Compare files ✨ Phase 3
│   │   ├── info.go            # Show statistics ✨ Phase 3
│   │   ├── version.go         # Version command
│   │   └── root.go            # Root command
│   ├── parser/                # JSON parsing and normalization
│   │   └── normalize_test.go  # Parser unit tests
│   ├── fetcher/               # Fetch from URL/file
│   ├── validator/             # JSON validation
│   └── converter/             # Format converters
│       ├── csv.go             # CSV converter
│       ├── csv_test.go        # CSV tests
│       ├── yaml.go            # YAML converter
│       ├── terraform.go       # Terraform HCL converter
│       ├── ansible.go         # Ansible variables converter
│       ├── aws_sg.go          # AWS Security Group JSON
│       ├── cisco_acl.go       # Cisco ACL converter ✨ Phase 3
│       └── panos.go           # PAN-OS XML converter ✨ Phase 3
├── pkg/
│   └── types/                 # Shared types (NetworkEntry, etc.)
├── testdata/
│   ├── network-requirements.json    # Full test data
│   ├── tenant-test.json            # Tenant substitution test data
│   └── golden/                     # Golden files for integration tests
├── test/
│   └── integration/
│       └── golden_test.sh          # Integration test script
├── .github/
│   └── workflows/
│       ├── ci.yml                  # CI/CD pipeline
│       └── release.yml             # Release workflow ✨ Phase 3
├── docs/
│   └── RELEASING.md                # Release process documentation ✨ Phase 3
├── .goreleaser.yaml                # GoReleaser configuration ✨ Phase 3
├── Dockerfile                      # Container image definition ✨ Phase 3
├── install.sh                      # Installation script ✨ Phase 3
├── CHANGELOG.md                    # Release notes ✨ Phase 3
├── Makefile                        # Build and test targets
├── PLAN.md                         # Project roadmap
├── README.md                       # User documentation
└── CLAUDE.md                       # This file
```

### Supported Formats

| Format | Status | Extension | Converter | Use Case |
|--------|--------|-----------|-----------|----------|
| CSV | ✅ | `.csv` | `CSVConverter` | Spreadsheets, reporting |
| YAML | ✅ | `.yaml` | `YAMLConverter` | Configuration files |
| Terraform | ✅ | `.tf` | `TerraformConverter` | Infrastructure as Code |
| Ansible | ✅ | `.yml` | `AnsibleConverter` | Automation playbooks |
| AWS SG | ✅ | `.json` | `AWSSecurityGroupConverter` | AWS Security Groups |
| Cisco ACL | ✅ | `.acl` | `CiscoACLConverter` | Cisco firewall rules |
| PAN-OS | ✅ | `.xml` | `PANOSConverter` | Palo Alto Networks firewalls |

---

## Session Start Protocol

When starting a Claude Code session in delinea-netconfig:

```bash
# 1. Where am I?
pwd && git status && git branch

# 2. Check project status
make help

# 3. Run tests to ensure everything works
make test

# 4. Check recent changes
git log --oneline -10

# 5. Review current phase
cat PLAN.md | grep -A 10 "Phase 2"
```

### Quick Status Check

```bash
# See what's implemented
./delinea-netconfig --help

# Test a conversion
./delinea-netconfig convert -f testdata/network-requirements.json --format csv | head -10

# Run all tests
make test

# See coverage
make test-coverage
```

---

## Development Workflow

### Adding a New Converter

1. **Create converter file:** `internal/converter/myformat.go`
   ```go
   type MyFormatConverter struct{}

   func (c *MyFormatConverter) Convert(entries []types.NetworkEntry) ([]byte, error) {
       // Implementation
   }

   func (c *MyFormatConverter) Name() string { return "MyFormat" }
   func (c *MyFormatConverter) FileExtension() string { return "ext" }
   ```

2. **Register in converter.go:**
   ```go
   var converters = map[string]Converter{
       // ...
       "myformat": &MyFormatConverter{},
   }
   ```

3. **Create tests:** `internal/converter/myformat_test.go`

4. **Generate golden file:**
   ```bash
   ./delinea-netconfig convert -f testdata/tenant-test.json --format myformat -q > testdata/golden/tenant-test.ext
   ```

5. **Update integration test:** Add test case to `test/integration/golden_test.sh`

6. **Run tests:**
   ```bash
   make test
   ```

### Testing Strategy

#### Unit Tests
```bash
# Run all unit tests
make test-unit

# Run specific package tests
go test -v ./internal/converter/

# Run with coverage
make test-coverage
```

#### Integration Tests
```bash
# Run golden file tests
make test-integration

# Regenerate golden files (if intentional changes)
make build
./delinea-netconfig convert -f testdata/tenant-test.json --format csv -q > testdata/golden/tenant-test.csv
```

#### Test Data

- **`testdata/network-requirements.json`**: Full production-like test data
- **`testdata/tenant-test.json`**: Minimal test data with `<tenant>` placeholders
- **`testdata/golden/`**: Reference outputs for integration tests

---

## Key Implementation Details

### Tenant Substitution

The `--tenant` flag replaces `<tenant>` placeholders in:
- Hostname values (e.g., `<tenant>.delinea.app` → `mycompany.delinea.app`)
- Descriptions (e.g., "Service for `<tenant>`" → "Service for mycompany")

Implementation: [`internal/cli/convert.go:179-201`](internal/cli/convert.go)

### Deterministic Output

All converters produce **deterministic, sorted output** to ensure:
- Consistent test results
- Reliable diffs
- Reproducible builds

Sorting: [`internal/parser/normalize.go:21-35`](internal/parser/normalize.go)

Sort order:
1. Direction (outbound first, then inbound)
2. Service name (alphabetical)
3. Region (alphabetical)
4. Type (alphabetical)

### Network Entry Structure

Core data structure used by all converters:

```go
type NetworkEntry struct {
    Direction   string   // "outbound" or "inbound"
    Service     string   // Service name (e.g., "platform_ssc_ips")
    Region      string   // Region code (e.g., "us", "global")
    Type        string   // "ipv4", "ipv6", "hostname", etc.
    Values      []string // IP ranges or hostnames
    Protocol    string   // "tcp", "udp", "both"
    Ports       []int    // Port numbers
    Description string   // Service description
    Redundancy  string   // "primary", "dr", or empty
    Tags        []string // Optional tags
}
```

---

## Quality Standards

**We are building mission-critical software. Quality is king.**

### Testing Requirements

- **ALL new features must have unit tests**
- **ALL new converters must have integration tests**
- **Coverage must not decrease**
- **All tests must pass with race detection (`-race`)**
- **Golden files must be updated when output intentionally changes**

### Code Quality

- Run `make fmt` before committing
- Run `make vet` to catch common issues
- Run `make lint` (requires golangci-lint)
- No compiler warnings
- No TODOs in committed code (use issues instead)

### Security

- **NEVER** introduce security vulnerabilities
- Validate all user input
- Sanitize data in descriptions/outputs
- Follow OWASP top 10 guidelines
- If you notice insecure code, fix it immediately

### Changelog

- **ALWAYS update CHANGELOG.md when making any meaningful change** (features, fixes, refactors, config changes)
- Follow [Keep a Changelog](https://keepachangelog.com/) format with Added/Changed/Fixed/Removed sections
- Entry goes under an `[Unreleased]` section if no release tag is being created
- Entry goes under the version heading (e.g., `[1.0.4]`) if preparing a release
- Keep entries concise but descriptive enough to understand the change without reading code

### Simplicity

- **Avoid over-engineering**
- Only make changes that are directly requested
- Don't add features beyond what was asked
- Three similar lines > premature abstraction
- Keep converters focused and single-purpose

---

## Makefile Targets

```bash
# Building
make build              # Build the binary
make install            # Install to $GOPATH/bin

# Testing
make test               # Run all tests (unit + integration)
make test-unit          # Run unit tests only
make test-integration   # Run integration tests only
make test-coverage      # Run tests with coverage report

# Quality
make fmt                # Format code
make vet                # Run go vet
make lint               # Run golangci-lint

# Quick format tests
make test-csv           # Test CSV conversion
make test-yaml          # Test YAML conversion
make test-terraform     # Test Terraform conversion
make test-validate      # Test validation
make test-all-formats   # Test all formats

# Other
make clean              # Remove build artifacts
make help               # Show all targets
```

---

## CI/CD Pipeline

**GitHub Actions:** [`.github/workflows/ci.yml`](.github/workflows/ci.yml)

### Pipeline Jobs

1. **Test Job**
   - Runs on: `ubuntu-latest`
   - Go versions: 1.23, 1.24
   - Runs unit tests with race detection
   - Generates coverage report
   - Uploads to Codecov
   - Runs integration tests

2. **Lint Job**
   - Runs on: `ubuntu-latest`
   - Uses golangci-lint
   - Checks code quality

3. **Build Job**
   - Runs on: Linux, macOS, Windows
   - Builds binary for each platform
   - Tests binary runs (`--version`)

### Triggers

- Push to `main` or `develop`
- Pull requests to `main` or `develop`
- Manual trigger (`workflow_dispatch`)

---

## Release Process

**Full Documentation:** [docs/RELEASING.md](docs/RELEASING.md)

### Creating a Release

Releases are fully automated using GoReleaser and GitHub Actions:

```bash
# 1. Ensure main is up to date
git checkout main
git pull origin main

# 2. Update CHANGELOG.md with release notes

# 3. Create and push version tag
git tag -a v0.4.0 -m "Release v0.4.0"
git push origin v0.4.0
```

GitHub Actions will automatically:
- Build binaries for all platforms (Linux, macOS, Windows, FreeBSD)
- Create multi-arch Docker images
- Generate checksums
- Create GitHub release with artifacts
- Update Homebrew tap
- Publish Docker images to GHCR

### Installation Methods

After release, users can install via:

```bash
# Install script (recommended)
curl -sfL https://raw.githubusercontent.com/DelineaXPM/delinea-platform/main/delinea-netconfig/install.sh | sh

# Homebrew
brew install delinea/tap/delinea-netconfig

# Docker
docker pull ghcr.io/delineaxpm/delinea-netconfig:latest

# Pre-built binaries
# Download from GitHub Releases
```

### Version Injection

Version information is injected at build time via ldflags:
- `Version` - Semantic version (e.g., "0.4.0")
- `Commit` - Git commit hash
- `Date` - Build timestamp

See: [internal/cli/root.go:10-18](internal/cli/root.go)

---

## Common Commands

### Development

```bash
# Build and test
make build && make test

# Test a specific converter
./delinea-netconfig convert -f testdata/tenant-test.json --format ansible --tenant acme

# Test with real data
./delinea-netconfig convert -f testdata/network-requirements.json --format csv | head -20

# Validate JSON
./delinea-netconfig validate -f testdata/network-requirements.json

# Compare two versions
./delinea-netconfig diff old-requirements.json new-requirements.json
./delinea-netconfig diff old.json new.json --summary  # Summary only

# Show statistics
./delinea-netconfig info testdata/network-requirements.json

# Check version
./delinea-netconfig version

# Generate shell completion
./delinea-netconfig completion bash > /etc/bash_completion.d/delinea-netconfig
./delinea-netconfig completion zsh > "${fpath[1]}/_delinea-netconfig"

# Fetch from URL and convert
./delinea-netconfig convert -u https://provisioning.delinea.app/.well-known/network-requirements.json --format terraform
```

### Testing

```bash
# Run specific test
go test -v ./internal/cli/ -run TestSubstituteTenant

# Run with verbose output
go test -v ./...

# Check race conditions
go test -race ./...

# Generate coverage
make test-coverage && open coverage.html
```

### Debugging

```bash
# Verbose mode
./delinea-netconfig convert -f testdata/network-requirements.json --format csv -v

# Quiet mode (suppress logs)
./delinea-netconfig convert -f testdata/network-requirements.json --format csv -q

# Save to file
./delinea-netconfig convert -f testdata/network-requirements.json --format terraform -o output.tf

# Multiple formats to directory
./delinea-netconfig convert -f testdata/network-requirements.json --format csv,yaml,terraform --output-dir ./output
```

---

## Phase 3 Complete ✅

All Phase 3 items have been implemented:

- [x] **Cisco ACL Converter** - Generate Cisco ACL format with wildcard masks
- [x] **PAN-OS Converter** - Generate Palo Alto XML format
- [x] **`diff` Command** - Compare two versions of network-requirements.json
- [x] **`info` Command** - Show summary statistics
- [x] **Shell Completion** - bash, zsh, fish, PowerShell completions
- [x] **GoReleaser** - Multi-platform binary releases (Linux, macOS, Windows, FreeBSD)
- [x] **Docker Images** - Multi-arch container images
- [x] **Installation Methods** - Install script, Homebrew tap, binaries

## Future Enhancements (Phase 4+)

### Potential Features

- [ ] **Filtering Support** - Add `--region`, `--service`, `--direction` flags
- [ ] **`list-services` Command** - List available services
- [ ] **`list-regions` Command** - List available regions
- [ ] **Azure NSG Converter** - Generate Azure NSG JSON
- [ ] **GCP Firewall Converter** - Generate GCP firewall rules
- [ ] **Custom Templates** - User-defined Go templates
- [ ] **Caching Layer** - Cache URL fetches
- [ ] **Performance Optimization** - Parallel processing for large datasets

---

## Troubleshooting

### Tests Failing After Code Changes

1. **Check if output format changed intentionally:**
   ```bash
   # Regenerate golden files
   make build
   ./delinea-netconfig convert -f testdata/tenant-test.json --format csv -q > testdata/golden/tenant-test.csv
   # Repeat for other formats
   ```

2. **Check for race conditions:**
   ```bash
   go test -race ./...
   ```

3. **Check test coverage:**
   ```bash
   make test-coverage
   open coverage.html
   ```

### Build Failures

1. **Update dependencies:**
   ```bash
   go mod download
   go mod verify
   go mod tidy
   ```

2. **Check Go version:**
   ```bash
   go version  # Should be 1.23 or later
   ```

3. **Clean and rebuild:**
   ```bash
   make clean
   make build
   ```

### Integration Tests Failing

1. **Check golden files exist:**
   ```bash
   ls -la testdata/golden/
   ```

2. **Compare output manually:**
   ```bash
   ./delinea-netconfig convert -f testdata/tenant-test.json --format csv -q > /tmp/test.csv
   diff testdata/golden/tenant-test.csv /tmp/test.csv
   ```

3. **Check sorting (output must be deterministic):**
   - Verify `internal/parser/normalize.go` has sorting logic
   - Run test multiple times to ensure consistency

---

## Related Documentation

### In This Repository
- [README.md](./README.md) - User documentation and installation guide
- [PLAN.md](./PLAN.md) - Project roadmap and design
- [TESTING.md](./TESTING.md) - Testing guide
- [docs/RELEASING.md](./docs/RELEASING.md) - Release process documentation ✨ Phase 3
- [CHANGELOG.md](./CHANGELOG.md) - Version history and release notes ✨ Phase 3
- [Makefile](./Makefile) - Build targets
- [.goreleaser.yaml](./.goreleaser.yaml) - Release configuration ✨ Phase 3

### External Resources
- [Delinea Platform](https://delinea.com)
- [Network Requirements JSON](https://provisioning.delinea.app/.well-known/network-requirements.json)
- [Go Documentation](https://go.dev/doc/)
- [Cobra CLI Framework](https://cobra.dev/)

---

## Agent Orchestration Strategy

### When to Use Different Approaches

#### Work Directly (No Agent)
Use for simple, straightforward tasks:
- **Single file edits** with clear changes
- **Obvious fixes** (typos, formatting, simple bugs)
- **Following established patterns** (e.g., adding a new converter using existing template)
- **Quick reads** (1-3 files)
- **Running tests** or build commands

**Example:** "Fix typo in README.md" or "Add unit test for existing function"

#### Use Explore Agent (Task with subagent_type=Explore)
Use for codebase investigation:
- **Finding patterns** across multiple files ("how are converters registered?")
- **Understanding architecture** ("how does normalization work?")
- **Discovering all usages** of a function or pattern
- **Deep investigation** when you're unfamiliar with an area

**Example:** "Find all places where tenant substitution is mentioned" or "How do golden file tests work?"

Set thoroughness appropriately:
- **"quick"**: Basic search, 1-2 locations
- **"medium"**: Moderate exploration, multiple files
- **"very thorough"**: Comprehensive analysis (use sparingly - high token cost)

#### Use EnterPlanMode
Use for complex implementation requiring user approval:
- **New converters** with unclear output format
- **Architectural changes** (e.g., changing how normalization works)
- **Multi-file changes** affecting >3 files
- **Unclear requirements** needing exploration first

**Don't use for:** Research tasks, exploration, or understanding existing code

**Example:** "Plan implementation of Cisco ACL converter" or "Plan refactoring of parser module"

### Parallel Execution

**Always make independent tool calls in parallel:**

```bash
# ✅ GOOD - Single message with parallel calls
[Read converter/csv.go, Read converter/yaml.go, Read converter/terraform.go]

# ❌ BAD - Sequential when unnecessary
Read csv.go → wait → Read yaml.go → wait → Read terraform.go
```

**Token savings:** 3-4x fewer API calls = 3-4x lower cost

---

## Token Cost Optimization 💰

Keeping token costs low while maintaining quality:

### Use Precise Tool Parameters

```go
// ✅ GOOD - Read specific section
Read(file_path="internal/converter/csv.go", offset=50, limit=30)

// ❌ BAD - Read entire large file when you only need a function
Read(file_path="internal/converter/csv.go")
```

### Avoid Duplicate Work

```bash
# ❌ BAD - Spawning agent then doing same work yourself
1. Task(Explore agent: "find all converters")
2. Then: Glob("internal/converter/*.go")  # Duplicate!

# ✅ GOOD - Trust agent results
1. Task(Explore agent: "find all converters")
2. Act on agent's findings directly
```

### Be Specific in Requests

**Vague** (requires clarification loops):
> "Look at the CSV converter"

**Specific** (one-shot answer):
> "Read internal/converter/csv.go and explain how it handles multiple values per entry"

**Token savings:** Eliminates 2-3 clarification round-trips

### Batch Related Operations

```bash
# ✅ GOOD - Single message
Read file1, Read file2, Edit file1, Edit file2

# ❌ BAD - Multiple messages (4x cost)
Message 1: Read file1
Message 2: Read file2
Message 3: Edit file1
Message 4: Edit file2
```

### Don't Over-Explore

- Use **Glob** when you know the file pattern: `internal/converter/*.go`
- Use **Grep** for specific searches: `grep "Convert.*NetworkEntry"`
- Use **Read** for known files: `internal/converter/csv.go`
- Use **Explore** only when you truly don't know where to look

### Trust Established Patterns

Don't read 5 examples when 1 is clear:
- Read CSV converter, then implement Cisco converter following same pattern
- Reference existing test, then write similar test
- Check one golden file example, then create new one

---

## Common Scenarios & Patterns

### Scenario: "Add a New Converter"

**Efficient approach:**

1. **Read one existing converter** as template (e.g., `csv.go`)
2. **Create new converter file** following the pattern
3. **Register in converter.go** (add to map + interface check)
4. **Write unit tests** (reference `csv_test.go`)
5. **Generate golden file** for integration test
6. **Update integration test script**
7. **Run tests:** `make test`

**Total files to read:** 2-3 (csv.go, csv_test.go, converter.go)
**No exploration needed:** Pattern is clear

### Scenario: "Fix a Bug in Output"

**Efficient approach:**

1. **Reproduce:** Run the failing command
2. **Identify converter:** Which format is affected?
3. **Read relevant converter:** Just that one file
4. **Find bug:** Usually in `Convert()` method
5. **Fix:** Minimal change to address root cause
6. **Test:** Run `make test`
7. **Regenerate golden file** if output intentionally changed

**Don't:** Read all converters, explore entire codebase, refactor unrelated code

### Scenario: "Add Tests for Feature"

**Efficient approach:**

1. **Find existing tests:** `internal/*/test.go` files
2. **Read one relevant test file** as example
3. **Follow test structure** and naming conventions
4. **Write new test** mirroring existing patterns
5. **Run:** `go test -v ./internal/cli/ -run YourTest`

**Pattern to follow:**
- Test function names: `TestFunctionName`
- Table-driven tests with subtests
- Use `assert` from testify
- Test edge cases (nil, empty, multiple values)

### Scenario: "Understand How Something Works"

**Efficient approach:**

1. **Check CLAUDE.md first** - might already be documented
2. **Check code comments** in relevant files
3. **Read specific file** if you know location
4. **Use Explore agent** if truly unclear where to look
5. **Update MEMORY.md** with findings for next time

**Example questions:**
- "How does tenant substitution work?" → Read `internal/cli/convert.go:179-201`
- "How are entries sorted?" → Read `internal/parser/normalize.go:21-35`
- "What converters exist?" → Read `internal/converter/converter.go:16-26`

### Scenario: "Regenerate Golden Files After Change"

**When intentional output changes:**

```bash
# 1. Rebuild
make build

# 2. Regenerate golden files
./delinea-netconfig convert -f testdata/tenant-test.json --format csv -q 2>/dev/null > testdata/golden/tenant-test.csv
./delinea-netconfig convert -f testdata/tenant-test.json --format yaml -q 2>/dev/null > testdata/golden/tenant-test.yaml
./delinea-netconfig convert -f testdata/tenant-test.json --format terraform -q 2>/dev/null > testdata/golden/tenant-test.tf
./delinea-netconfig convert -f testdata/tenant-test.json --format ansible -q 2>/dev/null > testdata/golden/tenant-test.yml
./delinea-netconfig convert -f testdata/tenant-test.json --format aws-sg -q 2>/dev/null > testdata/golden/tenant-test-aws-sg.json

# 3. Verify tests pass
make test-integration
```

**Important:** Only regenerate when output format **intentionally** changed!

### Scenario: "Compare Network Requirements Versions"

**Use the diff command:**

```bash
# Show all differences
./delinea-netconfig diff old-requirements.json new-requirements.json

# Show summary only
./delinea-netconfig diff old-requirements.json new-requirements.json --summary

# Analyze changes
# - Added entries: New requirements in new file
# - Removed entries: Requirements removed from old file
# - Modified entries: Same service/region but different IPs/ports
```

**Implementation:** [internal/cli/diff.go](internal/cli/diff.go)

### Scenario: "Analyze Network Requirements"

**Use the info command:**

```bash
# Show comprehensive statistics
./delinea-netconfig info testdata/network-requirements.json

# Output includes:
# - Total entries count
# - Direction breakdown (inbound vs outbound)
# - Service distribution
# - Region distribution
# - Protocol usage
# - Port analysis
# - IP address types (IPv4, IPv6, hostnames)
```

**Implementation:** [internal/cli/info.go](internal/cli/info.go)

### Scenario: "Create a New Release"

**Follow the release process:**

```bash
# 1. Update CHANGELOG.md
# Add release notes for new version

# 2. Create and push tag
git tag -a v0.4.0 -m "Release v0.4.0"
git push origin v0.4.0

# 3. GitHub Actions automatically:
# - Builds multi-platform binaries
# - Creates Docker images
# - Publishes to GitHub Releases
# - Updates Homebrew tap

# 4. Verify release
# Check GitHub Releases page
# Test installation methods
```

**Documentation:** [docs/RELEASING.md](docs/RELEASING.md)

---

## Best Practices for This Project

### Before Making Changes

**Quick checklist:**
- [ ] Have I read similar code (converters, tests)?
- [ ] Am I following the established pattern?
- [ ] Do I understand how this fits into the architecture?
- [ ] Have I checked for edge cases (nil, empty, multiple)?

### Go-Specific Guidelines

**Code Organization:**
- Keep converters focused: one format per file
- Follow Go naming conventions: `CSVConverter`, `ConvertToCSV`
- Unexported helpers: `formatPorts`, `getProtocol`
- Exported interfaces: `Converter`, `NetworkEntry`

**Error Handling:**
- Return errors, don't panic (except in truly exceptional cases)
- Wrap errors with context: `fmt.Errorf("failed to marshal: %w", err)`
- Validate at boundaries (user input, file I/O)
- Trust internal function calls (don't over-validate)

**Testing:**
- Table-driven tests are preferred
- Use subtests: `t.Run(tt.name, func(t *testing.T) { ... })`
- Test edge cases: nil, empty arrays, single values, multiple values
- Use `assert` from testify for cleaner assertions
- Always run with `-race` flag

### Converter Development Pattern

**Every converter must:**
1. Implement `types.Converter` interface
2. Have `Convert()`, `Name()`, `FileExtension()` methods
3. Be registered in `converter.go` converters map
4. Have unit tests in `*_test.go` file
5. Have golden file in `testdata/golden/`
6. Be added to integration test script

**Converter template:**
```go
type MyFormatConverter struct{}

func (c *MyFormatConverter) Convert(entries []types.NetworkEntry) ([]byte, error) {
    // Implementation
}

func (c *MyFormatConverter) Name() string {
    return "MyFormat"
}

func (c *MyFormatConverter) FileExtension() string {
    return "ext"
}
```

### Testing Strategy

**Unit tests** (`*_test.go` files):
- Test converter logic in isolation
- Test helper functions
- Test edge cases
- Run: `make test-unit`

**Integration tests** (golden files):
- Test end-to-end conversion
- Ensure output format is correct
- Detect unintended changes
- Run: `make test-integration`

**When tests fail:**
1. **Understand why** - don't just fix the test
2. **Fix root cause** - not symptoms
3. **Regenerate golden files** only if change was intentional
4. **Verify determinism** - run tests multiple times

---

## Decision Matrix

Use this to decide your approach:

| Task | Complexity | Files | Recommended Approach | Est. Tokens |
|------|-----------|-------|---------------------|-------------|
| Add simple test | Low | 1 | Direct: Read example test, write new test | 💰 Low |
| Fix converter bug | Low-Med | 1-2 | Direct: Read converter, fix, test | 💰 Low |
| Add new converter | Medium | 4-5 | Direct: Read CSV example, implement, test | 💰💰 Medium |
| Understand architecture | Medium | Many | Explore agent (quick/medium) | 💰💰 Medium |
| Major refactoring | High | >5 | EnterPlanMode → implement | 💰💰💰 High |
| New feature (unclear) | High | Unknown | EnterPlanMode first | 💰💰💰 High |

**Golden rule:** Invest tokens in understanding patterns once, then apply efficiently multiple times.

---

## Contacts

### Project Lead
- **Name:** Adil Laari
- **Email:** adil.laari@delinea.com
- **Role:** Architect / Engineer / PM

### Support
- **GitHub Issues:** https://github.com/DelineaXPM/delinea-platform/issues
- **Documentation:** https://docs.delinea.com

---

## Session End Checklist

Before ending a Claude Code session:

- [ ] All code changes work correctly
- [ ] **All tests passing** (`make test`)
- [ ] No race conditions (`go test -race ./...`)
- [ ] Code formatted (`make fmt`)
- [ ] No lint warnings (`make lint`)
- [ ] Documentation updated (if applicable)
  - [ ] README.md updated for user-facing changes
  - [ ] **CHANGELOG.md updated for ALL meaningful changes** (not just releases)
  - [ ] CLAUDE.md updated for dev workflow changes
- [ ] No security vulnerabilities introduced
- [ ] Golden files updated (if output changed intentionally)
- [ ] Git commit with clear message
- [ ] Project builds successfully (`make build`)
- [ ] Version numbers updated (if preparing release)

---

## Quick Reference

### Most Common Tasks

```bash
# 1. Start development
make build && make test

# 2. Convert network requirements
./delinea-netconfig convert -f testdata/network-requirements.json --format csv
./delinea-netconfig convert -f file.json --format cisco-acl
./delinea-netconfig convert -f file.json --format panos

# 3. Compare versions
./delinea-netconfig diff old.json new.json
./delinea-netconfig diff old.json new.json --summary

# 4. Show statistics
./delinea-netconfig info testdata/network-requirements.json

# 5. Add new converter
# - Create internal/converter/myformat.go
# - Implement Convert(), Name(), FileExtension()
# - Register in converter.go
# - Add tests
# - Generate golden file
# - Update integration test

# 6. Test changes
make test

# 7. Generate coverage report
make test-coverage

# 8. Create a release
git tag -a v0.4.0 -m "Release v0.4.0"
git push origin v0.4.0
# GitHub Actions handles the rest

# 9. Shell completion
./delinea-netconfig completion bash > /etc/bash_completion.d/delinea-netconfig
```

### Testing Checklist

- [ ] Unit tests written and passing
- [ ] Integration test updated
- [ ] Golden file generated (if new format)
- [ ] Tests pass with `-race`
- [ ] Coverage maintained or improved

---

**Last Updated:** 2026-02-10 (Phase 3 Complete - Production Ready with Distribution)
