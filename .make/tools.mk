# .make/tools.mk - Tool installation
# Included by main Makefile

.PHONY: install-tools install-lint install-fumpt
.PHONY: print-golangci-version

install-tools: install-lint install-fumpt ## Install all development tools
	@echo "✅ All tools installed"

GOLANGCI_LINT_MODULE := github.com/golangci/golangci-lint/v2

# Exits 0 only when the binary really is the pinned module version AND was built
# with the Go toolchain that is active right now.
#
# The version half alone is not enough, and that gap was measured rather than
# guessed. `go install` builds with whatever toolchain is active and ignores
# go.mod's `toolchain` directive, and task worktrees live outside this tree so
# mise resolves the *global* Go there, not the repo's pin. A binary built by an
# older Go then reads a stdlib its go/types cannot parse and dies mid-analysis:
# "panic: file requires newer Go version go1.27 (application built with
# go1.26)". A `version`-string check passes that binary happily, because the
# pinned release number is correct -- it is the compiler that is wrong. The
# integration gate hit exactly this on 2026-09-04 with a correctly-pinned
# v2.13.1 binary in bin/tools.
#
# `go version -m` is what makes the check honest: it reads the module version
# recorded inside the binary, so no wrapper script or --version string can
# satisfy it, and it reports the building toolchain in the same output.
GOLANGCI_LINT_VERSION_OK = $(GO) version -m "$(GOLANGCI_LINT_BIN)" 2>/dev/null | \
	awk -v want="$$($(GO) env GOVERSION)" 'NR == 1 { built = $$NF } $$1 == "mod" && $$2 == "$(GOLANGCI_LINT_MODULE)" && $$3 == "$(GOLANGCI_LINT_VERSION)" { found = 1 } END { exit !(found && built == want) }'

# Compare at the exact binary local lint runs. PATH cannot cause an unrelated
# binary to make this pin appear satisfied.
install-lint: ## Install golangci-lint (pinned, v2)
	@if $(GOLANGCI_LINT_VERSION_OK); then \
		echo "✅ golangci-lint $(GOLANGCI_LINT_VERSION) already installed: $(GOLANGCI_LINT_BIN)"; \
	else \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION) to $(GOLANGCI_LINT_BIN)..."; \
		mkdir -p "$(GOLANGCI_LINT_DIR)"; \
		rm -f "$(GOLANGCI_LINT_BIN)"; \
		GOBIN="$(GOLANGCI_LINT_DIR)" $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
		$(GOLANGCI_LINT_VERSION_OK) || { \
			echo "⚠️  golangci-lint at $(GOLANGCI_LINT_BIN) is not $(GOLANGCI_LINT_MODULE) $(GOLANGCI_LINT_VERSION) built with $$($(GO) env GOVERSION)" >&2; \
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
