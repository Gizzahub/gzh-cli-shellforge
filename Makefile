# Shellforge Modular Makefile
# Includes .make/*.mk for organized build targets

# Include all modular makefiles
include .make/vars.mk
include .make/build.mk
include .make/test.mk
include .make/quality.mk
include .make/deps.mk
include .make/tools.mk
include .make/dev.mk
include .make/docker.mk

.PHONY: help clean all validate validate-full demo pre-release

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST) .make/*.mk | sort

clean: ## Clean build artifacts and test outputs
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html bench.txt
	@echo "✅ Cleaned"

all: clean quality build ## Full build pipeline (clean, quality, build)
	@echo "✅ Complete build pipeline finished"

# Legacy targets for backward compatibility
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
