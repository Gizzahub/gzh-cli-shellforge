# .make/tools.mk - Tool installation
# Included by main Makefile

.PHONY: install-tools install-lint install-fumpt
.PHONY: print-golangci-version

install-tools: install-lint install-fumpt ## Install all development tools
	@echo "✅ All tools installed"

# Compare the version at the exact binary local lint runs. PATH cannot cause an
# unrelated binary to make this pin appear satisfied.
install-lint: ## Install golangci-lint (pinned, v2)
	@if [ -x "$(GOLANGCI_LINT_BIN)" ] && "$(GOLANGCI_LINT_BIN)" version 2>/dev/null | grep -qF "has version $(GOLANGCI_LINT_BARE) "; then \
		echo "✅ golangci-lint $(GOLANGCI_LINT_VERSION) already installed: $(GOLANGCI_LINT_BIN)"; \
	else \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION) to $(GOLANGCI_LINT_BIN)..."; \
		mkdir -p "$(GOLANGCI_LINT_DIR)"; \
		GOBIN="$(GOLANGCI_LINT_DIR)" $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
		"$(GOLANGCI_LINT_BIN)" version 2>/dev/null | grep -qF "has version $(GOLANGCI_LINT_BARE) " || { \
			echo "⚠️  golangci-lint installation did not produce $(GOLANGCI_LINT_VERSION): $(GOLANGCI_LINT_BIN)" >&2; \
			exit 1; \
		}; \
	fi

# Single source of the version for CI — the workflow reads this rather than
# carrying its own literal, so the two cannot drift apart.
print-golangci-version: ## Print the pinned golangci-lint version
	@echo $(GOLANGCI_LINT_VERSION)

install-fumpt: ## Install gofumpt
	@echo "Installing gofumpt..."
	@if ! command -v gofumpt >/dev/null 2>&1; then \
		$(GO) install mvdan.cc/gofumpt@latest; \
	else \
		echo "gofumpt already installed"; \
	fi

install-goreleaser: ## Install goreleaser
	@echo "Installing goreleaser..."
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		$(GO) install github.com/goreleaser/goreleaser@latest; \
	else \
		echo "goreleaser already installed"; \
	fi
