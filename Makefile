# Shellforge Makefile

.PHONY: help build test lint clean install validate bench coverage-html demo pre-release

# Variables
BINARY_NAME=gz-shellforge
BUILD_DIR=build
MAIN_PATH=cmd/shellforge/main.go

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "Built binaries:"
	@ls -lh $(BUILD_DIR)/

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -cover ./...
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linters
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "Running go vet..."
	$(GOVET) ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Cleaned"

install: ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

run: ## Run the application
	$(GOCMD) run $(MAIN_PATH)

bench: ## Run benchmark tests
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./internal/infra/diffcomparator/
	@echo "Benchmark complete"

bench-all: ## Run all benchmarks in the project
	@echo "Running all benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...
	@echo "All benchmarks complete"

coverage-html: test-coverage ## Generate and open HTML coverage report
	@echo "Opening coverage report..."
	@command -v open >/dev/null 2>&1 && open coverage.html || \
	 command -v xdg-open >/dev/null 2>&1 && xdg-open coverage.html || \
	 echo "Please open coverage.html manually"

validate: ## Run pre-release validation (fast)
	@echo "Running pre-release validation..."
	@bash scripts/pre-release.sh --skip-benchmarks

validate-full: ## Run complete pre-release validation with benchmarks
	@echo "Running full pre-release validation..."
	@bash scripts/pre-release.sh

demo: ## Run the workflow demonstration script
	@echo "Running workflow demo..."
	@bash examples/workflow-demo.sh

pre-release: validate-full ## Prepare for release (full validation)
	@echo ""
	@echo "Pre-release validation complete!"
	@echo "If all checks passed, you can proceed with:"
	@echo "  1. Review CHANGELOG.md"
	@echo "  2. git tag -a v\$$(grep 'version.*=' internal/cli/root.go | sed -E 's/.*\"([0-9]+\.[0-9]+\.[0-9]+)\".*/\1/') -m 'Release v\$$(grep 'version.*=' internal/cli/root.go | sed -E 's/.*\"([0-9]+\.[0-9]+\.[0-9]+)\".*/\1/')'"
	@echo "  3. git push origin v\$$(grep 'version.*=' internal/cli/root.go | sed -E 's/.*\"([0-9]+\.[0-9]+\.[0-9]+)\".*/\1/')"

.DEFAULT_GOAL := help
