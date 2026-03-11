# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.1] - 2026-03-11

### Changed
- Replaced archived `gopkg.in/yaml.v3` with actively maintained `github.com/goccy/go-yaml`

### Added
- Tests for `parser.Parse` covering v1 format, v2 format, missing version, invalid JSON, and empty input
- Tests for `runDiff` command covering added/removed entries, summary-only flag, identical files, missing files, and invalid JSON

### Fixed
- Lint: unchecked `w.Write` return values in `info_test.go` httptest handlers

## [1.2.0] - 2026-03-11

### Added
- `--region` / `-r` flag on `convert` command to filter output to global + region-specific rules (e.g. `--region eu`, `--region au`)

## [1.1.2] - 2026-03-11

### Fixed
- Tenant substitution (`--tenant`) no longer replaces `<tenant>` in descriptions — only applied to values (hostnames/IPs)

## [1.1.1] - 2026-03-11

### Added
- Tests for `info --updates` and `info --latest` flags using httptest mock server
- Tests for fetch error handling on both flags

### Fixed
- Support for new network-requirements.json v2 format (`items` array with `id`, flat `ports`, `protocol` fields)
- Backward compatibility preserved for old v1 format (`tcp_ports`/`udp_ports`, nested `ports`, map-based services)
- Parsing of v2 format with empty `items` array (e.g., `"inbound": {"items": []}`)

## [1.1.0] - 2026-03-11

### Added
- `info --updates` flag to fetch and display the network requirements changelog from Delinea
- `info --latest` flag to check the latest published version of network requirements
- `info --tenant` flag for tenant-specific URL construction (e.g., `https://<tenant>.delinea.app`)

### Changed
- Updated default network requirements URL from `provisioning.delinea.app/.well-known/network-requirements.json` to `setup.delinea.app/network-requirements.json`
- Removed all "NEW" markers from README.md (features are now established)
- `info` command file argument is now optional when using `--updates` or `--latest`

### Fixed
- CI badge in README.md pointing to wrong repository

## [1.0.5] - 2026-02-11

### Fixed
- Install script repository references (updated from DelineaXPM/delinea-platform to DelineaLaari/delinea-netconfig)

## [1.0.4] - 2026-02-11

### Added
- OCI labels for Docker images (`org.opencontainers.image.description`, `source`, `licenses`)

### Fixed
- Docker image "No description provided" message on GitHub Container Registry

## [1.0.3] - 2026-02-11

### Changed
- Updated release notes: removed Homebrew tap, added Docker installation instructions
- Reverted incorrect rlcp field in GoReleaser config
- Fixed GoReleaser deprecation warnings

## [1.0.2] - 2026-02-10

### Fixed
- Reverted incorrect rlcp field in GoReleaser config

## [1.0.1] - 2026-02-10

### Fixed
- Disabled Homebrew tap publishing (not yet available)

## [1.0.0] - 2026-02-10

### Added
- **Cisco ACL converter** - Generate Cisco IOS ACL format with wildcard mask support
- **PAN-OS XML converter** - Generate Palo Alto Networks XML configuration
- **`diff` command** - Compare two versions of network requirements files
- **`info` command** - Show statistics and analysis of network requirements
- **Shell completion** - Auto-completion for bash, zsh, fish, and PowerShell
- **Ansible converter** - Generate Ansible variables for automation playbooks
- **AWS Security Group converter** - Generate AWS Security Group JSON
- **GoReleaser** - Automated multi-platform binary releases (Linux, macOS, Windows, FreeBSD)
- **Docker images** - Multi-arch container images (amd64, arm64) on GHCR
- **MIT License**
- 100+ unit tests with race detection
- 7 integration tests with golden file validation
- CI/CD pipeline with GitHub Actions
- Deterministic, sorted output for consistent diffs and testing

### Features
- CSV, YAML, Terraform, Ansible, AWS SG, Cisco ACL, PAN-OS XML output formats
- Tenant substitution (`--tenant` flag) to replace `<tenant>` placeholders
- Flexible input from local files (`-f`) or remote URLs (`-u`)
- JSON structure validation
- Multiple format output in a single command
- Output to file (`-o`) or directory (`--output-dir`)
- Verbose (`-v`) and quiet (`-q`) modes

### Fixed
- Terraform output ordering for consistent golden file tests
- CI failures with missing cmd directory and linting errors
- GoReleaser config: removed unsupported alternative_names field
- Docker image registry references
- Repository name references in release config

[1.1.1]: https://github.com/DelineaLaari/delinea-netconfig/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/DelineaLaari/delinea-netconfig/compare/v1.0.5...v1.1.0
[1.0.5]: https://github.com/DelineaLaari/delinea-netconfig/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/DelineaLaari/delinea-netconfig/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/DelineaLaari/delinea-netconfig/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/DelineaLaari/delinea-netconfig/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/DelineaLaari/delinea-netconfig/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/DelineaLaari/delinea-netconfig/releases/tag/v1.0.0
