# Delinea Network Requirements Converter

[![CI](https://github.com/DelineaXPM/delinea-platform/workflows/CI/badge.svg)](https://github.com/DelineaXPM/delinea-platform/actions)
[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue.svg)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A CLI tool that converts Delinea's Platform IP/CIDR network requirements JSON into various firewall and infrastructure-as-code formats.

## ✨ Features

- **Multiple Output Formats**: CSV, YAML, Terraform, Ansible, AWS Security Groups, Cisco ACL, PAN-OS XML ✅
- **Diff & Info Commands**: Compare versions and show statistics ✅
- **Shell Completion**: Auto-completion for bash, zsh, fish, and PowerShell ✅
- **Tenant Substitution**: Replace `<tenant>` placeholders with your actual tenant name
- **Flexible Input**: Load from local files or remote URLs
- **Validation**: Validate JSON structure and network entries
- **Easy Integration**: Simple CLI interface for automation and CI/CD pipelines
- **Deterministic Output**: Consistent, sorted output for reliable diffs and testing
- **Well-Tested**: 100+ unit tests with race detection + integration tests

## 📦 Installation

### Pre-built Binaries (Recommended) ✨ NEW

Download the latest release for your platform from [GitHub Releases](https://github.com/DelineaXPM/delinea-platform/releases).

**Quick install (Linux/macOS):**
```bash
curl -sfL https://raw.githubusercontent.com/DelineaXPM/delinea-platform/main/delinea-netconfig/install.sh | sh
```

**Manual download:**
1. Visit [Releases](https://github.com/DelineaXPM/delinea-platform/releases)
2. Download the appropriate archive for your OS/architecture
3. Extract and move to your PATH:
```bash
# Example for Linux amd64
tar -xzf delinea-netconfig_*_Linux_x86_64.tar.gz
sudo mv delinea-netconfig /usr/local/bin/
```

### Homebrew (macOS/Linux) ✨ NEW

```bash
# Add Delinea tap
brew tap delinea/tap

# Install
brew install delinea-netconfig

# Update
brew upgrade delinea-netconfig
```

### Docker ✨ NEW

```bash
# Pull the image
docker pull ghcr.io/delineaxpm/delinea-netconfig:latest

# Run
docker run --rm -v $(pwd):/data ghcr.io/delineaxpm/delinea-netconfig:latest \
  convert -f /data/network-requirements.json --format csv
```

### Using Go Install

```bash
go install github.com/DelineaXPM/delinea-platform/delinea-netconfig/cmd/delinea-netconfig@latest
```

### From Source

```bash
# Clone the repository
git clone https://github.com/DelineaXPM/delinea-platform.git
cd delinea-platform/delinea-netconfig

# Build
make build

# Or install to $GOPATH/bin
make install
```

## 🚀 Quick Start

```bash
# Convert to CSV
delinea-netconfig convert -f network-requirements.json --format csv

# Convert with tenant substitution
delinea-netconfig convert -f network-requirements.json --format csv --tenant mycompany

# Convert to multiple formats
delinea-netconfig convert -f network-requirements.json --format csv,yaml,terraform,ansible

# Fetch from URL and convert
delinea-netconfig convert \
  -u https://provisioning.delinea.app/.well-known/network-requirements.json \
  --format terraform \
  --tenant mycompany

# Save to file
delinea-netconfig convert -f network-requirements.json --format terraform -o delinea.tf

# Save multiple formats to directory
delinea-netconfig convert -f network-requirements.json \
  --format csv,yaml,terraform,ansible,aws-sg \
  --output-dir ./configs

# Validate JSON
delinea-netconfig validate -f network-requirements.json
```

## 📖 Usage

### Commands

| Command | Description |
|---------|-------------|
| `convert` | Convert network requirements to different formats |
| `validate` | Validate network requirements JSON structure |
| `diff` | Compare two versions of network requirements files ✨ NEW |
| `info` | Show statistics about network requirements ✨ NEW |
| `completion` | Generate shell completion scripts (bash/zsh/fish/powershell) ✨ NEW |
| `help` | Show help for any command |
| `version` | Show version information |

### Global Flags

```
--file, -f       Path to network-requirements.json file
--url, -u        URL to fetch network-requirements.json
--output, -o     Output file path (default: stdout)
--output-dir     Output directory for multiple formats
--format         Output format(s): csv, yaml, terraform, ansible, aws-sg (comma-separated)
--tenant, -t     Substitute <tenant> placeholder with this value
--verbose, -v    Verbose logging
--quiet, -q      Suppress non-error output
--version        Show version information
--help, -h       Show help
```

## 🎯 Tenant Substitution

Some network entries contain `<tenant>` placeholders that need to be replaced with your actual tenant name.

**Example:**
- `<tenant>.secretservercloud.com` → `mycompany.secretservercloud.com`
- `<tenant>.delinea.app` → `mycompany.delinea.app`

Use the `--tenant` flag to automatically substitute these placeholders:

```bash
# Without substitution (default)
delinea-netconfig convert -f network-requirements.json --format csv
# Output: <tenant>.secretservercloud.com

# With substitution
delinea-netconfig convert -f network-requirements.json --format csv --tenant mycompany
# Output: mycompany.secretservercloud.com
```

This works across all output formats.

## 📄 Output Formats

### CSV

Simple CSV format with all network entries:

```csv
direction,service,region,type,value,protocol,ports,description,redundancy
outbound,platform_ssc_ips,global,ipv4,199.83.128.0/21,tcp,443,WAF IP ranges,
outbound,webhooks,us,ipv4,13.68.202.64/29,tcp,443,Webhook egress IPs,
```

**Use cases:** Spreadsheets, reporting, documentation

### YAML

Structured YAML format organized by direction and service:

```yaml
delinea_network_requirements:
  outbound:
    platform_ssc_ips:
      global:
        - type: ipv4
          values:
            - 199.83.128.0/21
            - 198.143.32.0/19
          protocol: tcp
          ports: [443]
```

**Use cases:** Configuration files, documentation

### Terraform

Terraform variables for easy integration:

```hcl
variable "delinea_outbound_platform_ssc_ips_global_ipv4" {
  description = "platform_ssc_ips - WAF IP ranges (global)"
  type        = list(string)
  default = [
    "199.83.128.0/21",
    "198.143.32.0/19",
  ]
}
```

**Use cases:** Infrastructure as Code, AWS/Azure/GCP deployments

### Ansible ✨ NEW

Ansible variables for playbook integration:

```yaml
_comment: Delinea Network Requirements - Ansible Variables
delinea_firewall_rules:
  outbound:
    - name: platform_ssc_ips_global_ipv4
      service: platform_ssc_ips
      region: global
      type: ipv4
      destinations:
        - 199.83.128.0/21
        - 198.143.32.0/19
      protocol: tcp
      ports: [443]
      description: WAF IP ranges
```

**Use cases:** Ansible automation, configuration management

### AWS Security Groups ✨ NEW

AWS Security Group JSON format:

```json
{
  "Description": "Delinea Platform Network Requirements",
  "GroupName": "delinea-platform-sg",
  "IpPermissions": [],
  "IpPermissionsEgress": [
    {
      "IpProtocol": "tcp",
      "FromPort": 443,
      "ToPort": 443,
      "IpRanges": [
        {
          "CidrIp": "199.83.128.0/21",
          "Description": "platform_ssc_ips - WAF IP ranges (global)"
        }
      ]
    }
  ],
  "Tags": [
    {"Key": "Name", "Value": "delinea-platform-sg"},
    {"Key": "ManagedBy", "Value": "delinea-netconfig"}
  ]
}
```

**Use cases:** AWS Security Groups, AWS CloudFormation

### Cisco ACL ✨ NEW

Cisco IOS Access Control List format:

```
! Delinea Platform Network Requirements
! Generated by delinea-netconfig
!
! Outbound Rules (Egress)
ip access-list extended DELINEA-OUTBOUND
 10 remark platform_ssc_ips - WAF IP ranges
 11 permit tcp any 199.83.128.0 0.0.7.255 eq 443
 20 remark platform_ssc_ips - WAF IP ranges
 21 permit tcp any 198.143.32.0 0.0.31.255 eq 443
!
! End of ACL
```

**Features:**
- Converts CIDR notation to wildcard masks
- Uses `host` keyword for /32 addresses
- Handles hostnames as remark comments
- Supports port ranges and multiple ports
- Separates inbound and outbound ACLs

**Use cases:** Cisco routers and switches, network automation

### PAN-OS XML ✨ NEW

Palo Alto Networks XML configuration format:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<config>
  <devices>
    <entry name="localhost.localdomain">
      <vsys>
        <entry name="vsys1">
          <address>
            <entry name="ip-199-83-128-0-21">
              <ip-netmask>199.83.128.0/21</ip-netmask>
              <description>WAF IP ranges</description>
            </entry>
          </address>
          <service>
            <entry name="svc-platform-tcp-443">
              <protocol>
                <tcp>
                  <port>443</port>
                </tcp>
              </protocol>
            </entry>
          </service>
          <rulebase>
            <security>
              <rules>
                <entry name="rule-1-platform-outbound">
                  <from><member>trust</member></from>
                  <to><member>untrust</member></to>
                  <source><member>any</member></source>
                  <destination><member>ip-199-83-128-0-21</member></destination>
                  <service><member>svc-platform-tcp-443</member></service>
                  <action>allow</action>
                </entry>
              </rules>
            </security>
          </rulebase>
        </entry>
      </vsys>
    </entry>
  </devices>
</config>
```

**Features:**
- Creates address objects (IP-netmask and FQDN)
- Creates service objects with TCP/UDP protocols
- Generates security rules with proper zones
- Sanitizes names to meet PAN-OS constraints
- Supports port ranges and comma-separated lists

**Use cases:** Palo Alto firewalls, network automation

### Coming in Future Releases

- **Azure NSG** - Azure Network Security Groups
- **GCP Firewall** - Google Cloud Platform firewall rules
- **Custom templates** - User-defined Go templates

### Compare Versions (diff) ✨ NEW

Compare two versions of network requirements to see what changed:

```bash
# Compare two versions
delinea-netconfig diff old-requirements.json new-requirements.json

# Show only summary
delinea-netconfig diff --summary v1.json v2.json

# Quiet mode (less verbose)
delinea-netconfig diff -q old.json new.json
```

**Output:**
```
Comparing:
  Old: old-requirements.json
  New: new-requirements.json

Added (3 entries):
  + [outbound] new_service/us: 10.0.0.0/24 (tcp:[443])
    → New service endpoints

Removed (1 entries):
  - [outbound] old_service/us: 192.168.1.0/24 (tcp:[80])
    → Deprecated service

Summary:
  Added:    3 entries
  Removed:  1 entries
  Modified: 0 entries
  Total changes: 4
```

### Statistics (info) ✨ NEW

Show detailed statistics about network requirements:

```bash
# Show statistics
delinea-netconfig info network-requirements.json

# Verbose mode (more details)
delinea-netconfig info -v network-requirements.json
```

**Output:**
```
Network Requirements Statistics
File: network-requirements.json

Overview:
  Total Entries:    61
  Total Values:     120
  Unique Values:    95

By Direction:
  outbound:         60
  inbound:          1

By Service:
  platform:         15
  ad_connector:     12
  secret_server:    8
  ...

By Protocol:
  tcp:              58
  udp:              2
  both:             1

Ports Used:
  443    (used 55 times)
  5671   (used 3 times)
  123    (used 2 times)
```

### Shell Completion ✨ NEW

Enable auto-completion for your shell:

```bash
# Bash
source <(delinea-netconfig completion bash)

# Zsh
source <(delinea-netconfig completion zsh)

# Fish
delinea-netconfig completion fish | source

# PowerShell
delinea-netconfig completion powershell | Out-String | Invoke-Expression
```

**Install permanently:**

```bash
# Bash (Linux)
delinea-netconfig completion bash > /etc/bash_completion.d/delinea-netconfig

# Bash (macOS)
delinea-netconfig completion bash > $(brew --prefix)/etc/bash_completion.d/delinea-netconfig

# Zsh
delinea-netconfig completion zsh > "${fpath[1]}/_delinea-netconfig"

# Fish
delinea-netconfig completion fish > ~/.config/fish/completions/delinea-netconfig.fish
```

## 💻 Development

### Prerequisites

- Go 1.23 or later
- Make (optional, for using Makefile targets)

### Building

```bash
# Build the binary
make build

# Run all tests (unit + integration)
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Clean build artifacts
make clean
```

### Testing

We maintain high test coverage with both unit and integration tests:

```bash
# Quick format tests
make test-csv           # Test CSV conversion
make test-yaml          # Test YAML conversion
make test-terraform     # Test Terraform conversion
make test-validate      # Test validation
make test-all-formats   # Test all formats

# Run specific tests
go test -v ./internal/converter/
go test -v ./internal/cli/ -run TestSubstituteTenant

# Check for race conditions
go test -race ./...
```

### Project Structure

```
delinea-netconfig/
├── cmd/
│   └── delinea-netconfig/      # CLI entry point
├── internal/
│   ├── cli/                    # CLI commands (Cobra)
│   │   └── convert_test.go     # CLI tests
│   ├── parser/                 # JSON parsing and normalization
│   │   └── normalize_test.go   # Parser tests
│   ├── fetcher/                # Fetch from URL/file
│   ├── validator/              # JSON validation
│   └── converter/              # Format converters
│       ├── csv.go / csv_test.go
│       ├── yaml.go
│       ├── terraform.go
│       ├── ansible.go
│       ├── aws_sg.go
│       ├── cisco.go / cisco_test.go     # ✨ Phase 3
│       └── panos.go / panos_test.go     # ✨ Phase 3
├── pkg/
│   └── types/                  # Shared types
├── testdata/
│   ├── network-requirements.json
│   ├── tenant-test.json
│   └── golden/                 # Golden files for integration tests
├── test/
│   └── integration/
│       └── golden_test.sh      # Integration test script
├── .github/
│   └── workflows/
│       └── ci.yml              # CI/CD pipeline
├── go.mod
├── Makefile
├── CLAUDE.md                   # Claude Code instructions
├── PLAN.md                     # Project roadmap
├── TESTING.md                  # Testing guide
└── README.md                   # This file
```

## 📚 Examples

### Use with Terraform

```bash
# Generate Terraform variables
delinea-netconfig convert -f network-requirements.json --format terraform -o delinea.tf

# Use in your Terraform configuration
cat > main.tf <<EOF
# Include Delinea network requirements
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Import the variables
# (Place delinea.tf in the same directory)

# Use the generated variables
resource "aws_security_group_rule" "delinea_platform" {
  type              = "egress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = var.delinea_outbound_platform_ssc_ips_global_ipv4
  security_group_id = aws_security_group.main.id
}
EOF

# Apply
terraform init
terraform plan
terraform apply
```

### Use with Ansible

```bash
# Generate Ansible variables
delinea-netconfig convert -f network-requirements.json \
  --format ansible \
  --tenant mycompany \
  -o group_vars/all/delinea_network.yml

# Use in your playbook
cat > playbook.yml <<EOF
---
- name: Configure firewall for Delinea
  hosts: firewalls
  vars_files:
    - group_vars/all/delinea_network.yml
  tasks:
    - name: Allow Delinea outbound traffic
      firewalld:
        rich_rule: "rule family=ipv4 destination address={{ item.destinations | join(',') }} port port={{ item.ports | join(',') }} protocol={{ item.protocol }} accept"
        permanent: yes
        state: enabled
      loop: "{{ delinea_firewall_rules.outbound }}"
      when: item.type == 'ipv4'
EOF

# Run playbook
ansible-playbook -i inventory.ini playbook.yml
```

### Use with AWS CLI

```bash
# Generate AWS Security Group JSON
delinea-netconfig convert -f network-requirements.json \
  --format aws-sg \
  --tenant mycompany \
  -o delinea-sg.json

# Create security group using AWS CLI
aws ec2 create-security-group \
  --group-name delinea-platform \
  --description "Delinea Platform Network Requirements" \
  --vpc-id vpc-xxxxx

# Apply security group rules
# (Manually extract rules from delinea-sg.json or use a script)
```

### Automation with CI/CD

Keep your firewall rules up-to-date automatically:

```yaml
# .github/workflows/update-firewall-rules.yml
name: Update Firewall Rules

on:
  schedule:
    - cron: '0 0 * * 0' # Weekly on Sunday
  workflow_dispatch:

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install delinea-netconfig
        run: go install github.com/DelineaXPM/delinea-platform/delinea-netconfig/cmd/delinea-netconfig@latest

      - name: Fetch and convert
        run: |
          delinea-netconfig convert \
            -u https://provisioning.delinea.app/.well-known/network-requirements.json \
            --format terraform \
            --tenant ${{ secrets.DELINEA_TENANT }} \
            -o delinea.tf

      - name: Create Pull Request
        uses: peter-evans/create-pull-request@v5
        with:
          commit-message: Update Delinea network requirements
          title: 'chore: Update Delinea network requirements'
          body: Automated update of Delinea network requirements
          branch: update-delinea-network-reqs
```

## 🧪 Testing

### Unit Tests

```bash
# Run all unit tests
make test-unit

# Run with coverage
make test-coverage
open coverage.html

# Test specific package
go test -v ./internal/converter/
```

### Integration Tests

```bash
# Run golden file tests
make test-integration

# Regenerate golden files (if output format changes intentionally)
make build
./delinea-netconfig convert -f testdata/tenant-test.json --format csv -q > testdata/golden/tenant-test.csv
```

### CI/CD

GitHub Actions automatically runs:
- Unit tests with race detection
- Integration tests
- Multi-platform builds (Linux, macOS, Windows)
- Code linting

See [`.github/workflows/ci.yml`](.github/workflows/ci.yml) for details.

## 📋 Roadmap

### ✅ Phase 1 - MVP (Complete)
- [x] Core CLI structure (Cobra framework)
- [x] CSV, YAML, Terraform converters
- [x] Validation command
- [x] URL fetching support
- [x] Tenant substitution (`--tenant` flag)
- [x] Basic documentation

### ✅ Phase 2 - Production Ready (Complete)
- [x] **Ansible converter**
- [x] **AWS Security Group JSON converter**
- [x] **Unit tests** (40+ test cases with race detection)
- [x] **Integration tests** (golden file testing)
- [x] **CI/CD pipeline** (GitHub Actions)
- [x] **Deterministic output** (sorted entries)
- [x] Enhanced documentation

### ✅ Phase 3 - Advanced Features (Complete)
- [x] **Cisco ACL converter** - Generate Cisco ACL format with wildcard masks
- [x] **PAN-OS converter** - Generate Palo Alto XML configuration
- [x] **`diff` command** - Compare two versions of network-requirements.json
- [x] **`info` command** - Show summary statistics and breakdowns
- [x] **Shell completion** - bash, zsh, fish, and PowerShell auto-completion
- [x] **100+ unit tests** - Comprehensive test coverage for all converters
- [x] **7 output formats** - CSV, YAML, Terraform, Ansible, AWS-SG, Cisco, PAN-OS

### ✅ Phase 4 - Distribution (In Progress)
- [x] **GoReleaser** - Multi-platform binary releases automation ✨ NEW
- [x] **Docker images** - Multi-arch container images (amd64, arm64) ✨ NEW
- [x] **Homebrew tap** - Easy installation on macOS/Linux ✨ NEW
- [x] **Installation script** - One-line curl install ✨ NEW
- [ ] **Filtering support** - Add `--region`, `--service`, `--direction` flags
- [ ] **Azure NSG converter** - Azure Network Security Groups
- [ ] **GCP Firewall converter** - Google Cloud Platform firewall rules

### 📋 Future Enhancements
- [ ] **Custom templates** - User-defined Go templates
- [ ] **Performance optimizations** - Parallel processing for large files
- [ ] **Plugin system** - Extensible converter architecture

## 📝 Changelog

### v0.3.0 (Current)
- ✨ Added **Cisco ACL converter** with wildcard mask support
- ✨ Added **PAN-OS XML converter** with full object and rule generation
- ✨ Added **`diff` command** to compare network requirement versions
- ✨ Added **`info` command** for statistics and analysis
- ✨ Added **shell completion** for bash, zsh, fish, and PowerShell
- ✨ Added **GoReleaser** configuration for automated multi-platform releases
- ✨ Added **Docker images** with multi-arch support (amd64, arm64)
- ✨ Added **Homebrew tap** for easy macOS/Linux installation
- ✨ Added **installation script** for one-line curl install
- ✨ 100+ unit tests covering all converters and commands
- ✨ 7 output formats now supported
- 📚 Comprehensive documentation updates
- 🚀 Professional distribution with pre-built binaries
- ✅ All Phase 3 goals achieved + initial Phase 4 distribution features

### v0.2.0 (Phase 2 Complete)
- ✨ Added Ansible converter for automation playbooks
- ✨ Added AWS Security Group JSON converter
- ✨ Added comprehensive unit tests (40+ test cases)
- ✨ Added integration tests with golden files
- ✨ Added CI/CD pipeline with GitHub Actions
- ✨ Added deterministic output with consistent sorting
- 🐛 Fixed non-deterministic output issues
- 📚 Enhanced documentation (README, CLAUDE.md, TESTING.md)
- 🧪 Added race detection to all tests
- ✅ Multi-platform CI builds (Linux, macOS, Windows)

### v0.1.0 (Phase 1 - MVP)
- 🎉 Initial release
- ✨ Basic CLI with convert and validate commands
- ✨ CSV, YAML, and Terraform converters
- ✨ URL and file fetching
- ✨ Tenant substitution with `--tenant` flag
- 📚 Basic documentation

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Run tests (`make test`)
6. Format code (`make fmt`)
7. Commit your changes (`git commit -m 'feat: add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

### Code Quality Requirements

- All tests must pass (`make test`)
- Code must be formatted (`make fmt`)
- No linter warnings (`make lint`)
- Coverage should not decrease
- New features require tests

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 💬 Support

For issues and questions:
- **GitHub Issues**: https://github.com/DelineaXPM/delinea-platform/issues
- **Documentation**: https://docs.delinea.com
- **Discussions**: https://github.com/DelineaXPM/delinea-platform/discussions

## 🙏 Acknowledgments

- Built with [Cobra](https://cobra.dev/) CLI framework
- Uses [yaml.v3](https://github.com/go-yaml/yaml) for YAML processing
- Inspired by the need for easy network requirement management

---

**Made with ❤️ by the Delinea Platform Team**
