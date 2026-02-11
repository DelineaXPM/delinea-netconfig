# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2024-02-10

### Added
- **Cisco ACL converter** - Generate Cisco IOS ACL format with wildcard mask support
- **PAN-OS XML converter** - Generate Palo Alto Networks XML configuration
- **`diff` command** - Compare two versions of network requirements files
- **`info` command** - Show statistics and analysis of network requirements
- **Shell completion** - Auto-completion for bash, zsh, fish, and PowerShell
- 100+ comprehensive unit tests covering all converters and commands
- 7 output formats now supported

### Changed
- Version bumped to 0.3.0
- Updated documentation with Phase 3 features
- Enhanced README with new converter examples

### Fixed
- Improved test coverage and reliability

## [0.2.0] - 2024-01-15

### Added
- Ansible converter for automation playbooks
- AWS Security Group JSON converter
- Comprehensive unit tests (40+ test cases)
- Integration tests with golden files
- CI/CD pipeline with GitHub Actions
- Deterministic output with consistent sorting

### Fixed
- Non-deterministic output issues
- Map iteration order in parser

### Changed
- Enhanced documentation (README, CLAUDE.md, TESTING.md)
- Added race detection to all tests
- Multi-platform CI builds (Linux, macOS, Windows)

## [0.1.0] - 2024-01-01

### Added
- Initial release
- Basic CLI with convert and validate commands
- CSV, YAML, and Terraform converters
- URL and file fetching
- Tenant substitution with `--tenant` flag
- Basic documentation

[0.3.0]: https://github.com/DelineaXPM/delinea-platform/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/DelineaXPM/delinea-platform/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/DelineaXPM/delinea-platform/releases/tag/v0.1.0
