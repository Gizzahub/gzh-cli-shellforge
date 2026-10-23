# .make/test.mk - Testing targets
# Included by main Makefile

.PHONY: test test-unit test-integration test-coverage test-verbose bench

test: ## Run all tests with race detection
	@echo "Running tests..."
	$(GOTEST) $(RACE_FLAG) -timeout $(TEST_TIMEOUT) -coverprofile=$(COVERAGE_OUT) ./...

test-unit: ## Run unit tests only (skip integration)
	@echo "Running unit tests..."
	$(GOTEST) -v -short -timeout $(TEST_TIMEOUT) ./...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	$(GOTEST) -v -run Integration -timeout $(TEST_TIMEOUT) ./...

test-coverage: test test-coverage-gate ## Generate HTML coverage report
	@echo "Generating coverage report..."
	$(GO) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "Coverage report: $(COVERAGE_HTML)"
	@$(GO) tool cover -func=$(COVERAGE_OUT) >/dev/null
	@awk -f scripts/check-coverage.awk "$(COVERAGE_OUT)"

.PHONY: test-coverage-gate
test-coverage-gate: ## Verify exact coverage threshold enforcement
	@bash scripts/check-coverage_test.sh

test-verbose: ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	$(GOTEST) -v $(RACE_FLAG) -timeout $(TEST_TIMEOUT) ./...

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem -run=^$$ ./... | tee bench.txt

.PHONY: perf-startup
perf-startup: build ## Observe macOS startup, ten-module build and memory without changing acceptance thresholds
	python3 scripts/startup-profile.py --binary "$(BUILD_DIR)/$(BINARY_NAME)" --output tmp/startup-profile.json
