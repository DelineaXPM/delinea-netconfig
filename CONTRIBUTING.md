# Contributing to delinea-netconfig

## Prerequisites

- Go 1.23 or later
- Make (optional, for Makefile targets)

## Building

```bash
# Build the binary
make build

# Install to $GOPATH/bin
make install
```

## Testing

```bash
# Run all tests (unit + integration)
make test

# Unit tests only (with race detection)
make test-unit

# Integration tests only (golden file comparison)
make test-integration

# With coverage report
make test-coverage
open coverage.html

# Run a specific test
go test -v ./internal/converter/ -run TestCSVConverter
```

### Golden Files

Integration tests compare converter output against reference files in `testdata/golden/`. If you intentionally change output format, regenerate them:

```bash
make build
./delinea-netconfig convert -f testdata/tenant-test.json --format csv -q > testdata/golden/tenant-test.csv
./delinea-netconfig convert -f testdata/tenant-test.json --format yaml -q > testdata/golden/tenant-test.yaml
./delinea-netconfig convert -f testdata/tenant-test.json --format terraform -q > testdata/golden/tenant-test.tf
./delinea-netconfig convert -f testdata/tenant-test.json --format ansible -q > testdata/golden/tenant-test.yml
./delinea-netconfig convert -f testdata/tenant-test.json --format aws-sg -q > testdata/golden/tenant-test-aws-sg.json
make test-integration
```

Only regenerate when output changes are intentional — golden files are the contract.

## Project Structure

```
delinea-netconfig/
├── cmd/delinea-netconfig/   # CLI entry point (main.go)
├── internal/
│   ├── cli/                 # Cobra commands (convert, validate, diff, info, tui)
│   ├── converter/           # Format converters (csv, yaml, terraform, ...)
│   ├── differ/              # Shared diff logic (used by cli/diff and tui)
│   ├── fetcher/             # Load from file or URL
│   ├── parser/              # JSON parsing and normalization
│   ├── tui/                 # Interactive terminal UI (Bubble Tea)
│   └── validator/           # JSON structure validation
├── pkg/types/               # Shared types (NetworkEntry, etc.)
├── testdata/
│   ├── network-requirements.json
│   ├── tenant-test.json
│   └── golden/              # Reference outputs for integration tests
└── test/integration/
    └── golden_test.sh       # Integration test runner
```

## Adding a New Converter

1. **Create `internal/converter/myformat.go`:**

```go
type MyFormatConverter struct{}

func (c *MyFormatConverter) Convert(entries []types.NetworkEntry) ([]byte, error) {
    // implementation
}

func (c *MyFormatConverter) Name() string          { return "MyFormat" }
func (c *MyFormatConverter) FileExtension() string { return "ext" }
```

2. **Register in `internal/converter/converter.go`** — add to the converters map.

3. **Add unit tests** in `internal/converter/myformat_test.go`.

4. **Generate a golden file** and add a test case to `test/integration/golden_test.sh`.

5. **Run `make test`** — all tests must pass with `-race`.

## Code Quality

```bash
make fmt    # gofmt
make vet    # go vet
make lint   # golangci-lint (requires installation)
```

All PRs must pass linting. No TODOs in committed code.

## Changelog

Update `CHANGELOG.md` for every meaningful change under `[Unreleased]`. Follow [Keep a Changelog](https://keepachangelog.com/) format (Added / Changed / Fixed / Removed).

## CI/CD

GitHub Actions runs on every push and PR to `main`:

- Unit tests with race detection (Go 1.23 and 1.24)
- Integration tests (golden file comparison)
- golangci-lint
- Multi-platform builds (Linux, macOS, Windows)

See [`.github/workflows/ci.yml`](.github/workflows/ci.yml).

## Releases

Releases are automated via GoReleaser. Create and push a version tag:

```bash
git tag -a v1.4.0 -m "Release v1.4.0"
git push origin v1.4.0
```

GitHub Actions builds multi-platform binaries, Docker images, and publishes to GitHub Releases. See [docs/RELEASING.md](docs/RELEASING.md) for the full process.
