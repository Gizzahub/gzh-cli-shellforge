# .make/tools.mk - Tool installation
# Included by main Makefile

.PHONY: install-tools install-lint install-fumpt install-mdformat
.PHONY: print-golangci-version

install-tools: install-lint install-fumpt install-mdformat ## Install all development tools
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

# ISSUE-352: 패밀리 고정 묶음. gfm이 표를 처리하므로 tables 플러그인은 쓰지 않는다.
MDFORMAT_VERSION := 1.0.0
MDFORMAT_GFM_VERSION := 1.0.0
MDFORMAT_FRONTMATTER_VERSION := 2.1.2
MDFORMAT := $(CURDIR)/bin/tools/mdformat$(shell $(GO) env GOEXE)
MDFORMAT_TOOL_DIR := $(CURDIR)/bin/tools/python
MDFORMAT_PYTHON := $(MDFORMAT_TOOL_DIR)/mdformat/$(if $(shell $(GO) env GOEXE),Scripts/python.exe,bin/python)
MDFORMAT_CONSTRAINTS := $(CURDIR)/.make/mdformat-constraints.txt
# --version 출력을 awk에 그대로 파이프하지 않고 변수에 담는다: producer가 죽으면
# && 체인이 awk 이전에 멈추므로 파싱 실패가 producer 실패를 가리지 않는다.
MDFORMAT_VERSION_OK = { mdformat_version_output=$$("$(MDFORMAT)" --version 2>/dev/null) && printf '%s\n' "$$mdformat_version_output" | \
	awk -v core="mdformat $(MDFORMAT_VERSION)" -v a="mdformat-gfm $(MDFORMAT_GFM_VERSION)" -v b="mdformat_frontmatter $(MDFORMAT_FRONTMATTER_VERSION)" \
	'NR == 1 && sub(/\)$$/, "") { n = split($$0, p, / \(|, /); ok = n == 3 && p[1] == core && ((p[2] == a && p[3] == b) || (p[2] == b && p[3] == a)) } END { exit !ok }'; }
MDFORMAT_DEPS_OK = "$(MDFORMAT_PYTHON)" -I -c 'import sys, importlib.metadata as m; lines = [l for l in map(str.strip, open(sys.argv[1])) if l and l[0] != chr(35)]; got = lambda n: next((d.version for d in m.distributions(name=n)), None); pairs = [(l, l.split("==")) for l in lines]; bad = [l + (" (installed %s)" % got(p[0]) if len(p) == 2 else " (not name==version)") for l, p in pairs if len(p) != 2 or got(p[0]) != p[1]]; bad and print("mdformat dependency mismatch: " + "; ".join(bad), file=sys.stderr); sys.exit(not lines or bool(bad))' "$(MDFORMAT_CONSTRAINTS)"

install-mdformat: ## install the pinned Markdown bundle into bin/tools
	@mkdir -p "$(dir $(MDFORMAT))"
	@if ! { $(MDFORMAT_VERSION_OK) && $(MDFORMAT_DEPS_OK) 2>/dev/null; }; then \
		command -v uv >/dev/null 2>&1 || { echo "uv is required to install mdformat" >&2; exit 1; }; \
		UV_TOOL_DIR="$(MDFORMAT_TOOL_DIR)" UV_TOOL_BIN_DIR="$(dir $(MDFORMAT))" \
			uv tool install --quiet --force --constraints "$(MDFORMAT_CONSTRAINTS)" \
				--with "mdformat-gfm==$(MDFORMAT_GFM_VERSION)" \
				--with "mdformat-frontmatter==$(MDFORMAT_FRONTMATTER_VERSION)" \
				"mdformat==$(MDFORMAT_VERSION)" || exit 1; \
	fi
	@$(MDFORMAT_VERSION_OK) || { echo "mdformat version/plugin set mismatch: $(MDFORMAT)" >&2; exit 1; }
	@$(MDFORMAT_DEPS_OK) || { echo "mdformat dependencies do not match $(MDFORMAT_CONSTRAINTS)" >&2; exit 1; }

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
