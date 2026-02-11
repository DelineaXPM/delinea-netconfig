.PHONY: help build install test test-unit test-integration test-coverage clean fmt vet lint run

# Variables
BINARY_NAME=delinea-netconfig
GO=go
GOFLAGS=-v
CMD_PATH=./cmd/delinea-netconfig

# Default target
help:
	@echo "Available targets:"
	@echo "  build              - Build the binary"
	@echo "  install            - Install the binary to GOPATH/bin"
	@echo "  test               - Run all tests (unit + integration)"
	@echo "  test-unit          - Run unit tests only"
	@echo "  test-integration   - Run integration tests only"
	@echo "  test-coverage      - Run tests with coverage"
	@echo "  clean              - Remove build artifacts"
	@echo "  fmt                - Format code"
	@echo "  vet                - Run go vet"
	@echo "  lint               - Run golangci-lint (requires golangci-lint installed)"
	@echo "  run                - Run the tool (pass args with ARGS=...)"
	@echo ""
	@echo "Quick format tests:"
	@echo "  test-csv           - Test CSV conversion"
	@echo "  test-yaml          - Test YAML conversion"
	@echo "  test-terraform     - Test Terraform conversion"
	@echo "  test-validate      - Test validation"
	@echo "  test-all-formats   - Test all formats"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test"
	@echo "  make run ARGS='convert -f testdata/network-requirements.json --format csv'"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) $(CMD_PATH)

# Install the binary
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(GOFLAGS) $(CMD_PATH)

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	$(GO) test $(GOFLAGS) -race ./...

# Run integration tests
test-integration: build
	@echo "Running integration tests..."
	@./test/integration/golden_test.sh

# Run all tests
test: test-unit test-integration
	@echo "All tests completed!"

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test $(GOFLAGS) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	$(GO) clean

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Run golangci-lint (requires installation)
lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

# Run the tool
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME) $(ARGS)

# Quick test - convert to CSV
test-csv: build
	@echo "Testing CSV conversion..."
	./$(BINARY_NAME) convert -f testdata/network-requirements.json --format csv | head -20

# Quick test - convert to YAML
test-yaml: build
	@echo "Testing YAML conversion..."
	./$(BINARY_NAME) convert -f testdata/network-requirements.json --format yaml | head -30

# Quick test - convert to Terraform
test-terraform: build
	@echo "Testing Terraform conversion..."
	./$(BINARY_NAME) convert -f testdata/network-requirements.json --format terraform | head -30

# Quick test - validate
test-validate: build
	@echo "Testing validation..."
	./$(BINARY_NAME) validate -f testdata/network-requirements.json

# Run all quick tests
test-all-formats: test-csv test-yaml test-terraform test-validate
	@echo "All format tests completed!"
