# Delinea Network Requirements Converter

[![CI](https://github.com/DelineaLaari/delinea-netconfig/workflows/CI/badge.svg)](https://github.com/DelineaLaari/delinea-netconfig/actions)
[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue.svg)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Convert Delinea's Platform IP/CIDR network requirements JSON into firewall rules and infrastructure-as-code formats — via CLI or an interactive terminal UI.

## Features

- **Interactive TUI** — browse, filter, inspect, export, and diff entries without remembering flags
- **7 Output Formats** — CSV, YAML, Terraform, Ansible, AWS Security Groups, Cisco ACL, PAN-OS XML
- **Diff & Info** — compare versions and show statistics
- **Tenant Substitution** — replace `<tenant>` placeholders with your actual tenant name
- **Flexible Input** — load from local files or remote URLs
- **Shell Completion** — bash, zsh, fish, and PowerShell

## Installation

### Pre-built Binaries (Recommended)

```bash
# Linux / macOS
curl -sfL https://raw.githubusercontent.com/DelineaLaari/delinea-netconfig/main/install.sh | sh
```

Or download the archive for your platform from [GitHub Releases](https://github.com/DelineaLaari/delinea-netconfig/releases), extract, and move to your PATH.

### Docker

```bash
docker pull ghcr.io/delinealaari/delinea-netconfig:latest

docker run --rm -v $(pwd):/data ghcr.io/delinealaari/delinea-netconfig:latest \
  convert -f /data/network-requirements.json --format csv
```

### Go Install

```bash
go install github.com/DelineaLaari/delinea-netconfig/cmd/delinea-netconfig@latest
```

### From Source

```bash
git clone https://github.com/DelineaLaari/delinea-netconfig.git
cd delinea-netconfig
make build
```

## Quick Start

```bash
# Convert to CSV
delinea-netconfig convert -f network-requirements.json --format csv

# Convert with tenant substitution
delinea-netconfig convert -f network-requirements.json --format terraform --tenant mycompany

# Fetch from URL and convert
delinea-netconfig convert \
  -u https://setup.delinea.app/network-requirements.json \
  --format terraform --tenant mycompany

# Save multiple formats to a directory
delinea-netconfig convert -f network-requirements.json \
  --format csv,yaml,terraform,ansible,aws-sg \
  --output-dir ./configs
```

## Interactive TUI

Launch an interactive terminal UI to browse, filter, and export — no flags to remember.

```bash
# Open file picker
delinea-netconfig tui

# Load a file directly
delinea-netconfig tui -f network-requirements.json

# Load from a remote URL
delinea-netconfig tui -u https://setup.delinea.app/network-requirements.json

# Compare two versions interactively
delinea-netconfig tui --diff old.json new.json
```

### Key Bindings

| Key | Action |
|-----|--------|
| `↑` / `k`, `↓` / `j` | Navigate entries |
| `Tab` | Toggle All / Outbound / Inbound |
| `/` | Live text filter |
| `r` | Filter by region (enter to confirm, esc to cancel) |
| `x` | Clear region filter |
| `Enter` | Open entry detail |
| `e` | Export (from browser or detail) |
| `c` | Copy IPs to clipboard (detail screen) |
| `Tab` | Cycle diff tabs: All / Added / Removed / Modified |
| `q` / `Ctrl+C` | Quit |

## Commands

| Command | Description |
|---------|-------------|
| `tui` | Interactive terminal UI |
| `convert` | Convert to a supported output format |
| `validate` | Validate network requirements JSON |
| `diff` | Compare two versions |
| `info` | Show statistics |
| `completion` | Generate shell completion scripts |
| `version` | Show version information |

### Common Flags

```
-f, --file       Path to network-requirements.json
-u, --url        URL to fetch network-requirements.json
-o, --output     Output file (default: stdout)
    --output-dir Output directory for multiple formats
    --format      Output format(s): csv, yaml, terraform, ansible, aws-sg, cisco, panos
-t, --tenant     Substitute <tenant> placeholder
-v, --verbose    Verbose logging
-q, --quiet      Suppress non-error output
```

## Tenant Substitution

Entries containing `<tenant>` placeholders are replaced with your actual tenant name:

```
<tenant>.secretservercloud.com  →  mycompany.secretservercloud.com
<tenant>.delinea.app            →  mycompany.delinea.app
```

```bash
delinea-netconfig convert -f network-requirements.json --format csv --tenant mycompany
```

Works across all output formats and in the TUI export form.

## Output Formats

### CSV

```csv
direction,service,region,type,value,protocol,ports,description,redundancy
outbound,platform_ssc_ips,global,ipv4,199.83.128.0/21,tcp,443,WAF IP ranges,
```

### YAML

```yaml
delinea_network_requirements:
  outbound:
    platform_ssc_ips:
      global:
        - type: ipv4
          values: [199.83.128.0/21]
          protocol: tcp
          ports: [443]
```

### Terraform

```hcl
variable "delinea_outbound_platform_ssc_ips_global_ipv4" {
  description = "platform_ssc_ips - WAF IP ranges (global)"
  type        = list(string)
  default     = ["199.83.128.0/21", "198.143.32.0/19"]
}
```

### Ansible

```yaml
delinea_firewall_rules:
  outbound:
    - name: platform_ssc_ips_global_ipv4
      destinations: [199.83.128.0/21]
      protocol: tcp
      ports: [443]
```

### AWS Security Groups

Generates `IpPermissions` / `IpPermissionsEgress` JSON ready for use with `aws ec2 authorize-security-group-*` or CloudFormation.

### Cisco ACL

```
ip access-list extended DELINEA-OUTBOUND
 10 remark platform_ssc_ips - WAF IP ranges
 11 permit tcp any 199.83.128.0 0.0.7.255 eq 443
```

CIDR notation is converted to wildcard masks; `/32` addresses use the `host` keyword.

### PAN-OS XML

Generates address objects, service objects, and security rules for Palo Alto Networks firewalls.

## Diff

```bash
delinea-netconfig diff old.json new.json
delinea-netconfig diff --summary old.json new.json
```

```
Added (2 entries):
  + [outbound] new_service/us: 10.0.0.0/24 (tcp:[443])

Removed (1 entries):
  - [outbound] old_service/us: 192.168.1.0/24 (tcp:[80])

Summary: Added: 2  Removed: 1  Modified: 0  Total: 3
```

Or use `delinea-netconfig tui --diff old.json new.json` for the interactive tabbed view.

## Info

```bash
delinea-netconfig info network-requirements.json
```

Shows total entries, direction breakdown, service distribution, protocol usage, and port frequency.

## Shell Completion

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

**Install permanently (examples):**

```bash
# Bash (Linux)
delinea-netconfig completion bash > /etc/bash_completion.d/delinea-netconfig

# Zsh
delinea-netconfig completion zsh > "${fpath[1]}/_delinea-netconfig"

# Fish
delinea-netconfig completion fish > ~/.config/fish/completions/delinea-netconfig.fish
```

## Examples

### Terraform Integration

```bash
delinea-netconfig convert -f network-requirements.json --format terraform -o delinea.tf
```

Reference the generated variables in your Terraform:

```hcl
resource "aws_security_group_rule" "delinea_platform" {
  type              = "egress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = var.delinea_outbound_platform_ssc_ips_global_ipv4
  security_group_id = aws_security_group.main.id
}
```

### Ansible Integration

```bash
delinea-netconfig convert -f network-requirements.json \
  --format ansible --tenant mycompany \
  -o group_vars/all/delinea_network.yml
```

### Automated Updates with GitHub Actions

Keep firewall rules current by fetching the latest requirements on a schedule:

```yaml
# .github/workflows/update-firewall-rules.yml
name: Update Firewall Rules
on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly
  workflow_dispatch:

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install delinea-netconfig
        run: curl -sfL https://raw.githubusercontent.com/DelineaLaari/delinea-netconfig/main/install.sh | sh

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
          branch: update-delinea-network-reqs
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, testing, and how to add new converters.

## License

MIT License — see [LICENSE](LICENSE) for details.

## Acknowledgments

- Built with [Cobra](https://cobra.dev/) CLI framework
- Uses [goccy/go-yaml](https://github.com/goccy/go-yaml) for YAML processing
- Interactive TUI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and [Huh](https://github.com/charmbracelet/huh) from [Charm](https://charm.sh/)
- Clipboard support via [atotto/clipboard](https://github.com/atotto/clipboard)
- Inspired by the need for easy network requirement management

---

**Made with ❤️ by the Delinea Platform Team**
