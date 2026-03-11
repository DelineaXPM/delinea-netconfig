# Delinea Network Requirements Converter - Project Plan

**Project Name**: `delinea-netconfig`
**Full Name**: Delinea Network Requirements Converter
**Repository**: https://github.com/DelineaXPM/delinea-platform (part of larger platform repo)
**License**: MIT
**Primary Language**: Go

---

## 1. Project Overview

### Purpose
A CLI tool that converts Delinea's Platform IP/CIDR network requirements JSON into various firewall and infrastructure-as-code formats, making it easier for customers to integrate Delinea's network requirements into their security infrastructure.

### Target Users
- Security Engineers configuring firewall rules
- DevOps/Platform Engineers using Infrastructure-as-Code
- Network Administrators managing ACLs
- Compliance teams documenting network requirements

### Key Value Propositions
1. **Single Source of Truth**: Parse Delinea's canonical network requirements JSON
2. **Multi-Format Support**: Output to 9+ different formats
3. **Filtering Capabilities**: Extract only relevant rules (by region, service, direction)
4. **Automation-Friendly**: Easy to integrate into CI/CD pipelines
5. **Open Source**: Customers can extend/modify for their needs

---

## 2. Supported Output Formats

### Phase 1 (MVP)
- [x] CSV
- [x] YAML
- [x] Terraform (HCL)
- [x] Ansible

### Phase 2
- [ ] AWS Security Group JSON
- [ ] Cisco ACL
- [ ] PAN-OS (Palo Alto) XML

### Phase 3 (Future)
- [ ] TOML
- [ ] XML (generic)
- [ ] Azure NSG ARM/Bicep
- [ ] GCP Firewall Rules
- [ ] Custom templates (user-defined)

---

## 3. CLI Design

### Command Structure

```bash
delinea-netconfig [command] [flags]
```

### Core Commands

#### `convert` - Convert to different formats
```bash
# Basic usage
delinea-netconfig convert --file network-requirements.json --format terraform
delinea-netconfig convert --url https://setup.delinea.app/network-requirements.json --format ansible

# Output options
delinea-netconfig convert -f network-requirements.json -o output.tf --format terraform
delinea-netconfig convert -f network-requirements.json --format csv,yaml,terraform --output-dir ./configs
```

#### `validate` - Validate JSON structure
```bash
delinea-netconfig validate --file network-requirements.json
# Output:
# ✓ Valid JSON structure
# ✓ Schema version: 1.0.0
# ✓ All required fields present
# ✓ 145 IPv4 ranges validated
# ✓ 1 IPv6 range validated
```

#### `diff` - Compare two versions (Phase 2)
```bash
delinea-netconfig diff --old old.json --new new.json
# Output:
# ADDED IPs:
#   - outbound.platform_ssc_ips.global.ipv4: 203.0.113.0/24
# REMOVED IPs:
#   - outbound.webhooks.us.ipv4: 192.0.2.0/29
# CHANGED:
#   - version: 1.0.0 → 1.1.0
```

### Global Flags
```bash
--file, -f         Path to local network-requirements.json file
--url, -u          URL to fetch network-requirements.json
--output, -o       Output file path (default: stdout)
--output-dir       Output directory for multiple formats
--format           Output format(s): csv, yaml, toml, xml, terraform, ansible, aws-sg, cisco, panos
--verbose, -v      Verbose logging
--quiet, -q        Suppress non-error output
--version          Show version information
--help, -h         Show help
```

---

## 4. Project Structure

**Note**: This will be part of the `delinea-platform` repository. Structure below assumes a subdirectory within the larger repo (e.g., `tools/delinea-netconfig/` or `clients/delinea-netconfig/`).

```
delinea-netconfig/
├── cmd/
│   └── delinea-netconfig/
│       └── main.go                    # CLI entry point
├── internal/
│   ├── cli/                           # CLI commands (Cobra)
│   │   ├── root.go
│   │   ├── convert.go
│   │   ├── validate.go
│   │   └── diff.go                    # Phase 3
│   ├── parser/                        # JSON parsing
│   │   ├── parser.go                  # Main parser logic
│   │   └── normalize.go               # Normalize to common format
│   ├── fetcher/                       # Fetch from URL/file
│   │   ├── fetcher.go
│   │   └── cache.go                   # Optional caching
│   ├── validator/                     # JSON validation
│   │   └── validator.go
│   └── converter/                     # Format converters
│       ├── converter.go               # Converter interface
│       ├── csv.go
│       ├── yaml.go
│       ├── toml.go
│       ├── xml.go
│       ├── terraform.go
│       ├── ansible.go
│       ├── aws.go                     # AWS Security Group JSON
│       ├── cisco.go                   # Cisco ACL
│       └── panos.go                   # PAN-OS XML
├── pkg/
│   └── types/                         # Shared types/models
│       ├── types.go                   # Data structures
│       └── constants.go
├── testdata/                          # Test fixtures
│   ├── network-requirements.json
│   ├── network-requirements-v2.json   # For diff testing
│   └── invalid.json                   # For validation testing
├── examples/                          # Example outputs
│   ├── README.md
│   ├── output.csv
│   ├── output.yaml
│   ├── output.tf
│   ├── ansible.yml
│   ├── aws-sg.json
│   ├── cisco-acl.txt
│   └── panos.xml
├── docs/                              # Documentation
│   ├── installation.md
│   ├── usage.md
│   ├── formats.md                     # Format specifications
│   └── filtering.md
├── .github/
│   └── workflows/
│       ├── ci.yml                     # Tests, linting, build
│       ├── release.yml                # goreleaser for multi-platform binaries
│       └── codeql.yml                 # Security scanning
├── .goreleaser.yml                    # Release configuration
├── Makefile
├── go.mod
├── go.sum
├── README.md
├── LICENSE                            # MIT License
├── PLAN.md                            # This file
└── CHANGELOG.md
```

---

## 5. Data Model

### Core Types

```go
// pkg/types/types.go

// Root structure
type NetworkRequirements struct {
    Version     string                 `json:"version"`
    UpdatedAt   string                 `json:"updated_at"`
    Description string                 `json:"description"`
    Outbound    map[string]Service     `json:"outbound"`
    Inbound     map[string]Service     `json:"inbound"`
    RegionCodes map[string]string      `json:"region_codes"`
}

// Service definition
type Service struct {
    Description string                 `json:"description"`
    TCPPorts    []int                  `json:"tcp_ports,omitempty"`
    UDPPorts    []int                  `json:"udp_ports,omitempty"`
    Ports       *PortsNested           `json:"ports,omitempty"`
    Regions     map[string]RegionData  `json:"regions,omitempty"`
}

// Nested ports structure (for ad_connector_tcp_relays)
type PortsNested struct {
    External       *PortSpec `json:"external,omitempty"`
    InternalToADDC *PortSpec `json:"internal_to_ad_dc,omitempty"`
}

type PortSpec struct {
    TCPPorts []int `json:"tcp_ports,omitempty"`
    UDPPorts []int `json:"udp_ports,omitempty"`
}

// Regional data
type RegionData struct {
    Domain              string   `json:"domain,omitempty"`
    IPv4                []string `json:"ipv4,omitempty"`
    IPv6                []string `json:"ipv6,omitempty"`
    IPv4Primary         []string `json:"ipv4_primary,omitempty"`
    IPv4DR              []string `json:"ipv4_dr,omitempty"`
    Hostnames           []string `json:"hostnames,omitempty"`
    HostnamesSelfSigned []string `json:"hostnames_self_signed,omitempty"`
    HostnamesCASigned   []string `json:"hostnames_ca_signed,omitempty"`
    AWSSES              []string `json:"aws_ses,omitempty"`
}

// Normalized entry for processing
type NetworkEntry struct {
    Direction   string   // "outbound" or "inbound"
    Service     string   // "platform_ssc_ips", "webhooks", etc.
    Region      string   // "us", "global", etc.
    Type        string   // "ipv4", "ipv6", "hostname", etc.
    Values      []string // IPs or hostnames
    Protocol    string   // "tcp", "udp", "both"
    Ports       []int
    Description string
    Redundancy  string   // "primary", "dr", "" (for non-redundant)
    Tags        []string // Optional tags for grouping
}

// Converter interface
type Converter interface {
    Convert(entries []NetworkEntry) ([]byte, error)
    Name() string
    FileExtension() string
}
```

---

## 6. Technical Decisions

### Language: Go
**Rationale**:
- Single binary distribution (easy for customers)
- Excellent cross-platform compilation
- Great standard library for JSON, networking
- Fast execution
- Industry-standard for CLI tools (kubectl, terraform, gh)

### Dependencies

```go
// CLI Framework
github.com/spf13/cobra          // Command structure
github.com/spf13/viper          // Configuration management

// Output Formats
gopkg.in/yaml.v3                // YAML
github.com/BurntSushi/toml      // TOML
github.com/hashicorp/hcl/v2     // Terraform HCL (if needed)
encoding/json                   // JSON (stdlib)
encoding/xml                    // XML (stdlib)
encoding/csv                    // CSV (stdlib)

// Utilities
github.com/olekukonko/tablewriter  // Pretty tables
github.com/fatih/color             // Colored terminal output
golang.org/x/net/publicsuffix      // URL/hostname validation (optional)

// Testing
github.com/stretchr/testify        // Testing utilities
```

### Release Strategy
- **goreleaser** for multi-platform binary releases
- Build targets:
  - Linux: amd64, arm64
  - macOS: amd64 (Intel), arm64 (Apple Silicon)
  - Windows: amd64, arm64
- Distribute via:
  - GitHub Releases
  - Homebrew tap (optional)
  - Docker image (optional)

---

## 7. Implementation Phases

### Phase 1: MVP (Week 1-2)
**Goal**: Basic working CLI with core formats

- [ ] Project scaffolding
- [ ] Go modules setup
- [ ] Cobra CLI skeleton
- [ ] JSON parser for network-requirements.json
- [ ] Normalization layer (Service → []NetworkEntry)
- [ ] CSV converter
- [ ] YAML converter
- [ ] Terraform converter (simple variables)
- [ ] `validate` command - Validate JSON structure
- [ ] URL fetching support (`--url`) - Fetch from remote URL
- [ ] HTTP fetcher with basic error handling
- [ ] Unit tests for core logic
- [ ] README with usage examples
- [ ] Makefile for building

**Deliverables**:
- `delinea-netconfig convert -f network-requirements.json --format csv,yaml,terraform`
- `delinea-netconfig convert -u https://setup.delinea.app/network-requirements.json --format terraform`
- `delinea-netconfig validate -f network-requirements.json`

### Phase 2: Production Ready (Week 3-4)
**Goal**: All formats + production quality

- [ ] Ansible converter
- [ ] AWS Security Group JSON converter
- [ ] Cisco ACL converter
- [ ] PAN-OS XML converter
- [ ] TOML converter
- [ ] Integration tests (all formats)
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] goreleaser configuration (multi-platform binaries)
- [ ] Comprehensive error handling and user-friendly messages
- [ ] Examples directory with sample outputs for all formats
- [ ] Documentation (installation, usage, format specs)

**Deliverable**: Production-ready v1.0.0 release

### Phase 3: Advanced Features (Week 4+)
**Goal**: Community-driven enhancements

- [ ] Filtering support (--direction, --region, --service, --type)
- [ ] Tenant substitution (--tenant flag to replace <tenant> placeholders)
- [ ] Advanced Terraform module option (--terraform-style=module)
- [ ] `diff` command (compare versions)
- [ ] `info` command (show summary statistics)
- [ ] `list-services` command (list available services)
- [ ] `list-regions` command (list available regions)
- [ ] Caching layer for URL fetches
- [ ] Custom template support (Go templates)
- [ ] Azure NSG converter
- [ ] GCP Firewall Rules converter
- [ ] TOML converter
- [ ] Shell completion (bash, zsh, fish)
- [ ] Performance optimization
- [ ] Comprehensive documentation site

**Deliverable**: v1.1.0+ with community feedback

---

## 8. Testing Strategy

### Testing Approach

We'll use **multi-layered testing** to ensure reliability:

1. **Unit Tests** - Test individual components in isolation
2. **Integration Tests** - Test complete workflows end-to-end
3. **Golden File Testing** - Compare outputs against known-good examples
4. **CI/CD Testing** - Automated testing on every commit

### Unit Tests

**What to test**:
```
internal/parser/
  ├── parser_test.go          # Test JSON parsing
  └── normalize_test.go       # Test Service → NetworkEntry conversion

internal/converter/
  ├── csv_test.go             # Test CSV output
  ├── yaml_test.go            # Test YAML output
  ├── terraform_test.go       # Test Terraform output
  ├── ansible_test.go         # Test Ansible output
  ├── aws_test.go             # Test AWS SG JSON output
  ├── cisco_test.go           # Test Cisco ACL output
  └── panos_test.go           # Test PAN-OS XML output

internal/validator/
  └── validator_test.go       # Test JSON validation

internal/fetcher/
  └── fetcher_test.go         # Test URL fetching (mocked)
```

**Example Unit Test** (CSV converter):
```go
func TestCSVConverter(t *testing.T) {
    tests := []struct {
        name     string
        entries  []types.NetworkEntry
        expected string
        wantErr  bool
    }{
        {
            name: "single IPv4 entry",
            entries: []types.NetworkEntry{
                {
                    Direction:   "outbound",
                    Service:     "platform_ssc_ips",
                    Region:      "global",
                    Type:        "ipv4",
                    Values:      []string{"199.83.128.0/21"},
                    Protocol:    "tcp",
                    Ports:       []int{443},
                    Description: "WAF IP ranges",
                },
            },
            expected: "direction,service,region,type,value,protocol,port,description\noutbound,platform_ssc_ips,global,ipv4,199.83.128.0/21,tcp,443,WAF IP ranges\n",
            wantErr:  false,
        },
        // More test cases...
    }

    converter := &CSVConverter{}
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := converter.Convert(tt.entries)
            if (err != nil) != tt.wantErr {
                t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if string(got) != tt.expected {
                t.Errorf("Convert() = %v, want %v", string(got), tt.expected)
            }
        })
    }
}
```

### Integration Tests

**End-to-end CLI testing**:
```go
// internal/cli/convert_test.go
func TestConvertCommand_E2E(t *testing.T) {
    tests := []struct {
        name       string
        args       []string
        inputFile  string
        wantErr    bool
        checkFunc  func(t *testing.T, output string)
    }{
        {
            name:      "convert to CSV",
            args:      []string{"convert", "-f", "testdata/network-requirements.json", "--format", "csv"},
            inputFile: "testdata/network-requirements.json",
            wantErr:   false,
            checkFunc: func(t *testing.T, output string) {
                // Verify CSV has correct headers
                assert.Contains(t, output, "direction,service,region")
                // Verify at least one data row
                assert.Contains(t, output, "outbound,platform_ssc_ips")
            },
        },
        {
            name:      "convert to Terraform",
            args:      []string{"convert", "-f", "testdata/network-requirements.json", "--format", "terraform"},
            wantErr:   false,
            checkFunc: func(t *testing.T, output string) {
                assert.Contains(t, output, "variable \"delinea_")
                assert.Contains(t, output, "type        = list(string)")
            },
        },
        // More E2E tests...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create root command
            cmd := NewRootCmd()

            // Capture output
            buf := new(bytes.Buffer)
            cmd.SetOut(buf)
            cmd.SetErr(buf)

            // Set args
            cmd.SetArgs(tt.args)

            // Execute
            err := cmd.Execute()

            // Check error
            if (err != nil) != tt.wantErr {
                t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            // Run custom checks
            if tt.checkFunc != nil {
                tt.checkFunc(t, buf.String())
            }
        })
    }
}
```

### Golden File Testing

**Approach**: Store expected outputs and compare against actual outputs

```
testdata/
├── network-requirements.json       # Full test input
├── network-requirements-mini.json  # Minimal test input
└── golden/                         # Expected outputs
    ├── output.csv
    ├── output.yaml
    ├── output.tf
    ├── ansible.yml
    ├── aws-sg.json
    ├── cisco-acl.txt
    └── panos.xml
```

**Golden file test**:
```go
func TestGoldenFiles(t *testing.T) {
    formats := []string{"csv", "yaml", "terraform", "ansible", "aws-sg", "cisco", "panos"}

    for _, format := range formats {
        t.Run(format, func(t *testing.T) {
            // Parse input
            nr, err := parser.ParseFile("testdata/network-requirements.json")
            require.NoError(t, err)

            // Normalize
            entries := parser.Normalize(nr)

            // Convert
            converter := converter.Get(format)
            output, err := converter.Convert(entries)
            require.NoError(t, err)

            // Read golden file
            goldenPath := filepath.Join("testdata", "golden", "output."+converter.FileExtension())
            golden, err := os.ReadFile(goldenPath)

            if os.IsNotExist(err) {
                // Golden file doesn't exist, create it
                t.Logf("Creating golden file: %s", goldenPath)
                err = os.WriteFile(goldenPath, output, 0644)
                require.NoError(t, err)
                return
            }
            require.NoError(t, err)

            // Compare
            if !bytes.Equal(output, golden) {
                t.Errorf("Output does not match golden file.\nGot:\n%s\n\nWant:\n%s", output, golden)
            }
        })
    }
}
```

### Mock HTTP Server for Fetcher Tests

```go
func TestFetcher_FetchFromURL(t *testing.T) {
    // Create mock HTTP server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Serve test JSON
        data, _ := os.ReadFile("testdata/network-requirements.json")
        w.Header().Set("Content-Type", "application/json")
        w.Write(data)
    }))
    defer server.Close()

    // Test fetcher
    fetcher := NewFetcher()
    data, err := fetcher.FetchFromURL(server.URL)
    require.NoError(t, err)
    assert.NotEmpty(t, data)

    // Parse to verify it's valid JSON
    var nr types.NetworkRequirements
    err = json.Unmarshal(data, &nr)
    require.NoError(t, err)
    assert.Equal(t, "1.0.0", nr.Version)
}

func TestFetcher_ErrorHandling(t *testing.T) {
    tests := []struct {
        name       string
        statusCode int
        wantErr    bool
    }{
        {"success", http.StatusOK, false},
        {"not found", http.StatusNotFound, true},
        {"server error", http.StatusInternalServerError, true},
        {"unauthorized", http.StatusUnauthorized, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(tt.statusCode)
            }))
            defer server.Close()

            fetcher := NewFetcher()
            _, err := fetcher.FetchFromURL(server.URL)

            if (err != nil) != tt.wantErr {
                t.Errorf("FetchFromURL() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Test Coverage Goals

- **Overall**: 80%+ coverage
- **Parser**: 90%+ coverage (critical path)
- **Converters**: 85%+ coverage (each format)
- **Validator**: 90%+ coverage
- **CLI**: 70%+ coverage (harder to test, but integration tests help)

### CI/CD Testing (GitHub Actions)

```yaml
# .github/workflows/test.yml
name: Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    name: Test
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go: ['1.22', '1.23']

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: ${{ matrix.go }}

    - name: Run tests
      run: |
        go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.txt

    - name: Run integration tests
      run: |
        go test -v -tags=integration ./...

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: golangci/golangci-lint-action@v4
      with:
        version: latest
```

### Manual Testing Checklist

Before each release:
```markdown
- [ ] Test all formats with full network-requirements.json
- [ ] Test with minimal JSON (single service, single region)
- [ ] Test URL fetching with real URL
- [ ] Test error handling (invalid JSON, network errors, etc.)
- [ ] Test on macOS, Linux, Windows
- [ ] Verify outputs are valid (Terraform validate, YAML lint, etc.)
- [ ] Test with --output and --output-dir flags
- [ ] Test multiple format conversion in single run
- [ ] Verify file permissions (output files should be readable)
```

### Test Organization

```
delinea-netconfig/
├── internal/
│   ├── parser/
│   │   ├── parser.go
│   │   ├── parser_test.go         # Unit tests
│   │   └── testdata/
│   │       ├── valid.json
│   │       ├── invalid.json
│   │       └── minimal.json
│   └── converter/
│       ├── csv.go
│       ├── csv_test.go
│       └── testdata/
│           └── golden/
│               └── output.csv
├── testdata/                       # Shared test data
│   ├── network-requirements.json  # Full real file
│   ├── network-requirements-mini.json
│   └── golden/                     # Expected outputs
│       ├── output.csv
│       ├── output.yaml
│       ├── output.tf
│       └── ...
└── test/                           # Integration tests
    └── integration/
        └── cli_test.go
```

### Testing Tools

**Go packages**:
```go
github.com/stretchr/testify/assert   // Assertions
github.com/stretchr/testify/require  // Require (stop on failure)
github.com/stretchr/testify/mock     // Mocking (if needed)
```

**Make targets**:
```makefile
.PHONY: test
test:
	go test -v -race ./...

.PHONY: test-coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: test-integration
test-integration:
	go test -v -tags=integration ./test/integration/...

.PHONY: test-all
test-all: test test-integration

.PHONY: test-watch
test-watch:
	# Requires: go install github.com/cosmtrek/air@latest
	air -c .air.test.toml
```

---

## 9. Output Format Specifications

### CSV
**Columns**: `direction,service,region,type,value,protocol,port,description,redundancy`

```csv
direction,service,region,type,value,protocol,port,description,redundancy
outbound,platform_ssc_ips,global,ipv4,199.83.128.0/21,tcp,443,WAF IP ranges,
outbound,platform_ssc_ips,global,ipv6,2a02:e980::/29,tcp,443,WAF IP ranges,
inbound,ssc_radius,us,ipv4,20.228.138.112/29,udp,1812;1813,SSC RADIUS authentication egress IPs,primary
inbound,ssc_radius,us,ipv4,52.190.184.16/29,udp,1812;1813,SSC RADIUS authentication egress IPs,dr
```

### YAML
**Structure**: Preserve original JSON structure, with tenant substitution if requested

### Terraform
**Approach**: Variables + Resources

Options:
1. **Simple variables** (Phase 1)
2. **Module-based** (Phase 2)
3. **Complete module with security group resources** (Phase 3)

### Ansible
**Structure**: Variables file for use in playbooks

### AWS Security Group JSON
**Format**: Direct import into AWS CLI or CloudFormation

```json
{
  "IpPermissions": [...],
  "IpPermissionsEgress": [...]
}
```

### Cisco ACL
**Format**: Standard Cisco IOS ACL syntax

### PAN-OS
**Format**: XML configuration for import into Palo Alto firewalls

---

## 9. Open Questions & Decisions Needed

### 1. Terraform Module Design
**Question**: Should we generate a reusable Terraform module or just variables?

**Decision**: ✅ **Start with simple variables (Phase 1), add module option (Phase 2)**
- Phase 1: Simple variables only - maximum flexibility for customers
- Phase 2: Add `--terraform-style=[vars|module]` flag for opinionated module option
- Rationale: Less opinionated, easier to implement, works with existing Terraform setups

### 2. Hostname Resolution
**Question**: Should we resolve hostnames to IPs for firewall rules?

**Options**:
- A) No resolution, output hostnames as-is (customer handles DNS)
- B) Optionally resolve with `--resolve-hostnames` flag
- C) Always resolve and include IPs

**Recommendation**: A (no resolution), DNS is customer responsibility

### 3. Template Hostname Handling
**Question**: How to handle `<tenant>` placeholders?

**Decision**: ✅ **Leave as-is for MVP, add substitution in Phase 3**
- Phase 1 & 2: Output `<tenant>` placeholders as-is in the converted formats
- Phase 3: Add `--tenant` flag for optional substitution
  - If `--tenant mycompany` provided: Replace `<tenant>` with `mycompany`
  - `<tenant>.secretservercloud.com` → `mycompany.secretservercloud.com`
- Rationale: Keeps MVP simple, customers can do find/replace themselves if needed

### 4. Port Ranges
**Question**: How to represent multiple ports in formats that don't support arrays?

**Options**:
- A) Create separate rules per port
- B) Use format-specific syntax (e.g., "1812,1813" for Cisco)
- C) Port ranges where possible (e.g., "1812-1813")

**Recommendation**: B (format-specific), with A as fallback

### 5. Wildcard Hostnames
**Question**: How to handle wildcards like `*.digicert.com`?

**Options**:
- A) Output as-is (may not work in all formats)
- B) Skip wildcard entries (too restrictive)
- C) Add warning, output as comment in supported formats

**Recommendation**: C (warn + comment)

### 6. IPv6 Support
**Question**: Not all firewall formats handle IPv6 well. How to handle?

**Options**:
- A) Skip IPv6 if format doesn't support it
- B) Create separate IPv6 rules/files
- C) Always include, let format converter handle it

**Recommendation**: C, with format-specific handling

### 7. License
**Question**: What license should we use?

**Decision**: ✅ **MIT License**
- Permissive, customer-friendly
- Widely adopted and understood
- Allows commercial use without restrictions
- Aligns with open source community expectations

### 8. Release Versioning
**Question**: Should JSON schema version match CLI version?

**Decision**: ✅ **Independent versioning with compatibility matrix**
- CLI evolves independently from JSON schema
- CLI v1.0.0 supports JSON schema v1.x.x (forward compatible within major version)
- CLI v2.0.0 for breaking CLI changes (rare)
- JSON schema v2.0.0 would require CLI update for breaking schema changes
- Tool warns if JSON schema version is unsupported/too new
- Compatibility matrix documented in README

**Implementation**:
```go
const (
    CLIVersion              = "1.0.0"
    SupportedSchemaVersions = []string{"1.0.0", "1.1.0"}
    MinimumSchemaVersion    = "1.0.0"
)

// Parser validates JSON version
func ValidateSchemaVersion(version string) error {
    if !isSupportedVersion(version) {
        return fmt.Errorf("unsupported schema version: %s", version)
    }
    return nil
}
```

---

## 10. Success Metrics

### Technical Metrics
- ✅ Parse 100% of network-requirements.json without errors
- ✅ Generate valid output for all 9 formats
- ✅ 80%+ test coverage
- ✅ Cross-platform binary builds (6 platforms)
- ✅ Binary size < 20MB
- ✅ Execution time < 2s for full conversion

### User Metrics (Post-Launch)
- 📊 GitHub stars (target: 50+ in first 3 months)
- 📊 Download count (target: 500+ in first 3 months)
- 📊 Community contributions (PRs, issues)
- 📊 Customer feedback (support tickets mentioning the tool)

### Adoption Metrics
- 📝 Mention in Delinea docs
- 📝 Blog post / tutorial
- 📝 Community Slack/Discord mentions

---

## 11. Documentation Plan

### README.md
- Project overview
- Quick start
- Installation (homebrew, binary download, go install)
- Basic usage examples
- Links to detailed docs

### docs/installation.md
- Multiple installation methods
- Platform-specific instructions
- Verification steps

### docs/usage.md
- Detailed command reference
- All flags explained
- Advanced examples

### docs/formats.md
- Format-specific details
- Limitations of each format
- When to use which format

### docs/filtering.md
- How filtering works
- Filter combinations
- Performance tips

### CONTRIBUTING.md
- How to contribute
- Development setup
- Testing guidelines
- Code style

---

## 12. Next Steps

1. **Review & Iterate on Plan** (You are here!)
   - Review open questions
   - Make decisions on options
   - Add/remove items as needed

2. **Project Scaffolding**
   - Initialize Go module
   - Set up directory structure
   - Configure Cobra CLI

3. **Core Implementation**
   - Implement parser
   - Implement normalizer
   - Implement first converter (CSV)

4. **Iteration**
   - Add more converters
   - Add filtering
   - Add tests

---

## 13. Questions for You

1. **Timeline**: What's your target timeline for v1.0.0?
2. **Priority Formats**: Which 3 formats are most important for your customers?
3. **Repository Structure**: Where in `delinea-platform` should this live?
   - `tools/delinea-netconfig/`?
   - `clients/delinea-netconfig/`?
   - `cli/delinea-netconfig/`?
   - Root level `delinea-netconfig/`?
4. **Open Questions**: Review section 9 - any preferences on the remaining items?
5. **Features**: Any must-have features I missed?
6. **Branding**: Should we use Delinea branding/colors in CLI output?

---

**Last Updated**: 2026-02-10
**Status**: 📝 Draft - Awaiting Review
