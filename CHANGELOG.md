# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[1.0.4]: https://github.com/DelineaLaari/delinea-netconfig/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/DelineaLaari/delinea-netconfig/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/DelineaLaari/delinea-netconfig/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/DelineaLaari/delinea-netconfig/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/DelineaLaari/delinea-netconfig/releases/tag/v1.0.0
