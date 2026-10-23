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

test-coverage: test ## Generate HTML coverage report
	@echo "Generating coverage report..."
	$(GO) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "Coverage report: $(COVERAGE_HTML)"
	@coverage_result=$$($(GO) tool cover -func=$(COVERAGE_OUT)) && \
		printf '%s\n' "$$coverage_result" | awk '/^total:/ { print; found=1; if (($$NF + 0) < 80) { print "Coverage below required 80%" > "/dev/stderr"; exit 1 } } END { if (!found) exit 1 }'

test-verbose: ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	$(GOTEST) -v $(RACE_FLAG) -timeout $(TEST_TIMEOUT) ./...

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem -run=^$$ ./... | tee bench.txt
