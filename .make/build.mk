# .make/build.mk - Build targets
# Included by main Makefile

.PHONY: build build-all install run

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PKG)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PKG)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PKG)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PKG)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PKG)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PKG)
	@echo "Built binaries:"
	@ls -lh $(BUILD_DIR)/

install: build ## Install the binary to GOPATH/bin and ~/.local/bin
	@GOBIN=$$(go env GOBIN); \
	GOPATH=$$(go env GOPATH); \
	if [ -z "$$GOBIN" ]; then \
		BINDIR="$$GOPATH/bin"; \
	else \
		BINDIR="$$GOBIN"; \
	fi; \
	USERBIN="$$HOME/.local/bin"; \
	printf "Installing $(BINARY_NAME) to $$BINDIR\n"; \
	mkdir -p "$$BINDIR"; \
	mkdir -p "$$USERBIN"; \
	cp $(BUILD_DIR)/$(BINARY_NAME) "$$USERBIN/$(BINARY_NAME)"; \
	mv $(BUILD_DIR)/$(BINARY_NAME) "$$BINDIR/$(BINARY_NAME)"; \
	printf "✅ Installed $(BINARY_NAME) to $$BINDIR/$(BINARY_NAME)\n"; \
	printf "✅ Installed $(BINARY_NAME) to $$USERBIN/$(BINARY_NAME)\n"; \
	echo ""; \
	printf "Verifying installation...\n"; \
	"$$BINDIR/$(BINARY_NAME)" --version || echo "Note: Binary installed but --version flag not implemented"; \
	echo ""; \
	printf "🎉 Installation complete! Run '$(BINARY_NAME) --help' to get started.\n"

run: ## Run the application
	@echo "Running $(BINARY_NAME)..."
	$(GO) run $(MAIN_PKG)
