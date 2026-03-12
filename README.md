# Delinea Network Requirements Converter

[![CI](https://github.com/DelineaLaari/delinea-netconfig/workflows/CI/badge.svg)](https://github.com/DelineaLaari/delinea-netconfig/actions)
[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue.svg)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A CLI tool that converts Delinea's Platform IP/CIDR network requirements JSON into various firewall and infrastructure-as-code formats.

## Features

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

### Pre-built Binaries (Recommended)
Download the latest release for your platform from [GitHub Releases](https://github.com/DelineaLaari/delinea-netconfig/releases).

**Quick install (Linux/macOS):**
```bash
curl -sfL https://raw.githubusercontent.com/DelineaLaari/delinea-netconfig/main/install.sh | sh
```

**Manual download:**
1. Visit [Releases](https://github.com/DelineaLaari/delinea-netconfig/releases)
2. Download the appropriate archive for your OS/architecture
3. Extract and move to your PATH:
```bash
# Example for Linux amd64
tar -xzf delinea-netconfig_*_Linux_x86_64.tar.gz
sudo mv delinea-netconfig /usr/local/bin/
```

### Docker
```bash
# Pull the image
docker pull ghcr.io/delinealaari/delinea-netconfig:latest

# Run with local files
docker run --rm -v $(pwd):/data ghcr.io/delinealaari/delinea-netconfig:latest \
  convert -f /data/network-requirements.json --format csv
```

### Using Go Install

```bash
go install github.com/DelineaLaari/delinea-netconfig/cmd/delinea-netconfig@latest
```

### From Source

```bash
# Clone the repository
git clone https://github.com/DelineaLaari/delinea-netconfig.git
cd delinea-netconfig

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
  -u https://setup.delinea.app/network-requirements.json \
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
| `diff` | Compare two versions of network requirements files |
| `info` | Show statistics about network requirements |
| `completion` | Generate shell completion scripts (bash/zsh/fish/powershell) |
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

### Ansible
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

### AWS Security Groups
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

### Cisco ACL
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

### PAN-OS XML
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

### Compare Versions (diff)
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

### Statistics (info)
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

### Shell Completion
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
│       ├── cisco.go / cisco_test.go
│       └── panos.go / panos_test.go
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
        run: |
          curl -sfL https://raw.githubusercontent.com/DelineaLaari/delinea-netconfig/main/install.sh | sh
          echo "/usr/local/bin" >> $GITHUB_PATH

      - name: Fetch and convert
        run: |
          delinea-netconfig convert \
            -u https://setup.delinea.app/network-requirements.json \
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

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- Built with [Cobra](https://cobra.dev/) CLI framework
- Uses [goccy/go-yaml](https://github.com/goccy/go-yaml) for YAML processing
- Interactive TUI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and [Huh](https://github.com/charmbracelet/huh) from [Charm](https://charm.sh/)
- Clipboard support via [atotto/clipboard](https://github.com/atotto/clipboard)
- Inspired by the need for easy network requirement management

---

**Made with ❤️ by the Delinea Platform Team**
