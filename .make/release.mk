# Local release preparation never publishes artifacts or updates the tap.
GORELEASER_VERSION := v2.19.0
GORELEASER_DIR ?= $(CURDIR)/bin/tools
GORELEASER_BIN := $(GORELEASER_DIR)/goreleaser
GORELEASER_VERSION_OK = $(GO) version -m "$(GORELEASER_BIN)" 2>/dev/null | \
	awk '$$1 == "mod" && $$2 == "github.com/goreleaser/goreleaser/v2" && $$3 == "$(GORELEASER_VERSION)" { found = 1 } END { exit !found }'

.PHONY: install-goreleaser release-check release-snapshot
install-goreleaser: ## Install and verify the pinned repository-local release tool
	@if $(GORELEASER_VERSION_OK); then \
		echo "GoReleaser $(GORELEASER_VERSION) verified"; \
	else \
		mkdir -p "$(GORELEASER_DIR)" && \
		GOBIN="$(GORELEASER_DIR)" $(GO) install github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION) && \
		$(GORELEASER_VERSION_OK) || exit 1; \
	fi

release-check: install-goreleaser ## Validate release configuration without publishing
	RELEASE_VERSION="$(VERSION)" "$(GORELEASER_BIN)" check

release-snapshot: release-check ## Prepare local archives, checksums and cask without publishing
	RELEASE_VERSION="$(VERSION)" "$(GORELEASER_BIN)" release --snapshot --skip=publish
