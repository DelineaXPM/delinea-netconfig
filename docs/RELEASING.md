# Release Process

This document describes how to create a new release of `delinea-netconfig`.

## Prerequisites

1. **Maintainer Access**: You need write access to the repository
2. **GitHub Token**: Ensure you have appropriate permissions for:
   - Creating releases
   - Pushing Docker images to GHCR
   - (Optional) Updating Homebrew tap

## Automated Releases with GoReleaser

Releases are fully automated using GoReleaser and GitHub Actions.

### Creating a New Release

1. **Ensure all changes are merged to main**
   ```bash
   git checkout main
   git pull origin main
   ```

2. **Update version in documentation** (if needed)
   - Update `CHANGELOG.md` with release notes
   - Ensure version matches what you'll tag

3. **Create and push a version tag**
   ```bash
   # For a new release (e.g., v0.4.0)
   git tag -a v0.4.0 -m "Release v0.4.0"
   git push origin v0.4.0
   ```

4. **GitHub Actions will automatically:**
   - Run all tests
   - Build binaries for all platforms:
     - Linux (amd64, arm64, 386, arm)
     - macOS (amd64, arm64)
     - Windows (amd64, 386)
     - FreeBSD (amd64, 386)
   - Create checksums
   - Build Docker images (multi-arch)
   - Create GitHub release with:
     - Release notes from CHANGELOG
     - All platform binaries
     - Checksums
     - Installation instructions
   - Update Homebrew tap (if configured)
   - Push Docker images to GHCR

5. **Verify the release**
   - Check [GitHub Releases](https://github.com/DelineaXPM/delinea-netconfig/releases)
   - Test installation methods:
     ```bash
     # Test install script
     curl -sfL https://raw.githubusercontent.com/DelineaXPM/delinea-netconfig/main/install.sh | sh

     # Test Docker image
     docker pull ghcr.io/delineaxpm/delinea-netconfig:v0.4.0

     # Test Homebrew (after tap is updated)
     brew upgrade delinea-netconfig
     ```

## Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version (v1.0.0): Incompatible API changes
- **MINOR** version (v0.1.0): New features, backwards compatible
- **PATCH** version (v0.0.1): Bug fixes, backwards compatible

### Pre-release versions

For testing releases:
```bash
git tag -a v0.4.0-rc1 -m "Release candidate 1 for v0.4.0"
git push origin v0.4.0-rc1
```

Pre-releases will be marked as "Pre-release" on GitHub.

## Manual Testing Before Release

Always test locally before creating a release tag:

```bash
# Run all tests
make test

# Test with GoReleaser locally (snapshot mode)
goreleaser release --snapshot --clean

# Check the dist/ directory for generated artifacts
ls -la dist/
```

## Rollback a Release

If a release has issues:

1. **Delete the tag** (if not published):
   ```bash
   git tag -d v0.4.0
   git push origin :refs/tags/v0.4.0
   ```

2. **Mark as draft** (if published):
   - Go to GitHub Releases
   - Edit the release
   - Check "This is a draft"

3. **Fix issues** and create a patch release:
   ```bash
   git tag -a v0.4.1 -m "Release v0.4.1 - Fix critical bug"
   git push origin v0.4.1
   ```

## Troubleshooting

### Build fails in GitHub Actions

- Check the [Actions tab](https://github.com/DelineaXPM/delinea-netconfig/actions)
- Common issues:
  - Tests failing: Fix tests before releasing
  - Docker build failing: Check Dockerfile
  - GoReleaser config error: Validate with `goreleaser check`

### Homebrew tap not updating

- Requires `HOMEBREW_TAP_GITHUB_TOKEN` secret with access to tap repo
- Check if the token has expired
- Verify tap repository exists: `DelineaXPM/homebrew-tap`

### Docker images not published

- Requires `packages: write` permission (already configured)
- Check if GHCR is enabled for the organization
- Verify Docker login step in workflow

## Configuration Files

- **`.goreleaser.yaml`**: Main GoReleaser configuration
- **`.github/workflows/release.yml`**: GitHub Actions workflow for releases
- **`Dockerfile`**: Container image definition
- **`install.sh`**: Installation script for curl install method

## Local Release Testing

To test the release process locally without publishing:

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Run snapshot release (doesn't publish)
goreleaser release --snapshot --clean

# Check the generated artifacts
ls -la dist/
```

## Post-Release Tasks

After a successful release:

1. **Update documentation** if needed
2. **Announce the release** (if major version):
   - Update project README
   - Notify users
3. **Close related issues/PRs**:
   - Reference the release in issue comments
   - Close fixed issues

## Getting Help

- **GoReleaser Docs**: https://goreleaser.com
- **GitHub Actions**: https://docs.github.com/en/actions
- **Project Issues**: https://github.com/DelineaXPM/delinea-netconfig/issues
