#!/bin/bash
# Pre-Release Validation Script
# Validates everything before creating a new release
# Usage: ./scripts/pre-release.sh [--skip-benchmarks]

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
CHECKS_PASSED=0
CHECKS_FAILED=0
WARNINGS=0

# Flags
SKIP_BENCHMARKS=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-benchmarks)
            SKIP_BENCHMARKS=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--skip-benchmarks]"
            exit 1
            ;;
    esac
done

# Helper functions
print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

print_check() {
    echo -e "${BLUE}[CHECK]${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
    ((CHECKS_PASSED++))
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
    ((CHECKS_FAILED++))
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
    ((WARNINGS++))
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Check if running from project root
if [[ ! -f "go.mod" ]]; then
    print_error "Must run from project root directory"
    exit 1
fi

print_header "Pre-Release Validation"
print_info "Starting validation checks..."

# ============================================
# 1. Git Status Check
# ============================================
print_header "1. Git Status"

print_check "Checking for uncommitted changes..."
if [[ -n $(git status --porcelain) ]]; then
    print_warning "There are uncommitted changes:"
    git status --short
    echo ""
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_error "Aborting due to uncommitted changes"
        exit 1
    fi
else
    print_success "No uncommitted changes"
fi

print_check "Checking current branch..."
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$CURRENT_BRANCH" != "master" && "$CURRENT_BRANCH" != "main" ]]; then
    print_warning "Not on master/main branch (current: $CURRENT_BRANCH)"
else
    print_success "On $CURRENT_BRANCH branch"
fi

# ============================================
# 2. Version Consistency Check
# ============================================
print_header "2. Version Consistency"

print_check "Extracting version from sources..."

# Get version from main.go
MAIN_VERSION=$(grep 'Version.*=' cmd/shellforge/main.go | sed -E 's/.*"([0-9]+\.[0-9]+\.[0-9]+)".*/\1/')

# Get version from CHANGELOG.md
CHANGELOG_VERSION=$(grep -E '## \[[0-9]+\.[0-9]+\.[0-9]+\]' CHANGELOG.md | head -1 | sed -E 's/.*\[([0-9]+\.[0-9]+\.[0-9]+)\].*/\1/')

# Get version from git tags
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "none")

print_info "main.go version:      $MAIN_VERSION"
print_info "CHANGELOG.md version: $CHANGELOG_VERSION"
print_info "Latest git tag:       $LATEST_TAG"

if [[ "$MAIN_VERSION" == "$CHANGELOG_VERSION" ]]; then
    print_success "Version consistency check passed"
else
    print_error "Version mismatch between main.go ($MAIN_VERSION) and CHANGELOG.md ($CHANGELOG_VERSION)"
fi

# ============================================
# 3. Build Check
# ============================================
print_header "3. Build Verification"

print_check "Cleaning previous builds..."
make clean > /dev/null 2>&1
print_success "Build artifacts cleaned"

print_check "Building binary..."
if make build > /dev/null 2>&1; then
    print_success "Build succeeded"
else
    print_error "Build failed"
    exit 1
fi

print_check "Verifying binary..."
if [[ -f "build/gz-shellforge" ]]; then
    BINARY_SIZE=$(ls -lh build/gz-shellforge | awk '{print $5}')
    print_success "Binary created (size: $BINARY_SIZE)"

    # Test binary execution
    print_check "Testing binary execution..."
    if ./build/gz-shellforge --version > /dev/null 2>&1; then
        BINARY_VERSION=$(./build/gz-shellforge --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')
        print_success "Binary executes successfully (version: $BINARY_VERSION)"

        if [[ "$BINARY_VERSION" == "$MAIN_VERSION" ]]; then
            print_success "Binary version matches source"
        else
            print_error "Binary version ($BINARY_VERSION) doesn't match source ($MAIN_VERSION)"
        fi
    else
        print_error "Binary execution failed"
    fi
else
    print_error "Binary not found at build/gz-shellforge"
fi

# ============================================
# 4. Code Quality Check
# ============================================
print_header "4. Code Quality"

print_check "Running go fmt..."
UNFORMATTED=$(gofmt -l . 2>/dev/null | grep -v vendor || true)
if [[ -z "$UNFORMATTED" ]]; then
    print_success "All files are properly formatted"
else
    print_warning "Some files need formatting:"
    echo "$UNFORMATTED"
fi

print_check "Running go vet..."
if go vet ./... > /dev/null 2>&1; then
    print_success "go vet passed"
else
    print_error "go vet found issues"
fi

print_check "Checking for TODO/FIXME comments..."
TODO_COUNT=$(grep -r "TODO\|FIXME" --include="*.go" . | grep -v vendor | wc -l | tr -d ' ')
if [[ "$TODO_COUNT" -gt 0 ]]; then
    print_warning "Found $TODO_COUNT TODO/FIXME comments"
else
    print_success "No TODO/FIXME comments found"
fi

# ============================================
# 5. Test Suite
# ============================================
print_header "5. Test Suite"

print_check "Running all tests..."
if go test ./... -v > /tmp/test-output.txt 2>&1; then
    TEST_COUNT=$(grep -c "^ok" /tmp/test-output.txt || echo "0")
    print_success "All tests passed ($TEST_COUNT packages)"
else
    print_error "Some tests failed"
    echo ""
    tail -50 /tmp/test-output.txt
    exit 1
fi

print_check "Running tests with race detector..."
if go test ./... -race > /dev/null 2>&1; then
    print_success "Race detector found no issues"
else
    print_warning "Race detector found potential issues"
fi

print_check "Generating coverage report..."
if go test ./... -coverprofile=/tmp/coverage.out > /dev/null 2>&1; then
    COVERAGE=$(go tool cover -func=/tmp/coverage.out | grep total | awk '{print $3}')
    print_success "Code coverage: $COVERAGE"

    # Check coverage threshold
    COVERAGE_NUM=$(echo "$COVERAGE" | sed 's/%//')
    if (( $(echo "$COVERAGE_NUM >= 70.0" | bc -l) )); then
        print_success "Coverage meets 70% threshold"
    else
        print_warning "Coverage below 70% threshold"
    fi
else
    print_warning "Coverage report generation failed"
fi

# ============================================
# 6. Benchmark Tests (Optional)
# ============================================
if [[ "$SKIP_BENCHMARKS" == "false" ]]; then
    print_header "6. Benchmark Tests"

    print_check "Running benchmarks..."
    if go test -bench=. -benchmem ./internal/infra/diffcomparator/ > /tmp/bench-output.txt 2>&1; then
        BENCH_COUNT=$(grep -c "^Benchmark" /tmp/bench-output.txt || echo "0")
        print_success "All benchmarks completed ($BENCH_COUNT benchmarks)"

        # Show summary
        print_info "Benchmark summary (first 10):"
        grep "^Benchmark" /tmp/bench-output.txt | head -10
    else
        print_warning "Some benchmarks failed (non-critical)"
    fi
else
    print_info "Skipping benchmarks (--skip-benchmarks flag)"
fi

# ============================================
# 7. Documentation Check
# ============================================
print_header "7. Documentation"

print_check "Checking required documentation files..."
REQUIRED_DOCS=("README.md" "CHANGELOG.md" "LICENSE" "ARCHITECTURE.md" "TECH_STACK.md")
for doc in "${REQUIRED_DOCS[@]}"; do
    if [[ -f "$doc" ]]; then
        print_success "$doc exists"
    else
        print_error "$doc is missing"
    fi
done

print_check "Checking CHANGELOG.md has entry for current version..."
if grep -q "## \[$MAIN_VERSION\]" CHANGELOG.md; then
    print_success "CHANGELOG.md has entry for v$MAIN_VERSION"
else
    print_warning "CHANGELOG.md missing entry for v$MAIN_VERSION"
fi

print_check "Checking documentation examples..."
if [[ -f "examples/workflow-demo.sh" ]]; then
    if bash -n examples/workflow-demo.sh > /dev/null 2>&1; then
        print_success "workflow-demo.sh syntax is valid"
    else
        print_error "workflow-demo.sh has syntax errors"
    fi
else
    print_warning "workflow-demo.sh not found"
fi

# ============================================
# 8. Dependencies Check
# ============================================
print_header "8. Dependencies"

print_check "Checking go.mod consistency..."
if go mod tidy -v > /dev/null 2>&1; then
    if [[ -n $(git diff go.mod go.sum) ]]; then
        print_warning "go.mod/go.sum not up to date (run 'go mod tidy')"
    else
        print_success "go.mod and go.sum are up to date"
    fi
else
    print_error "go mod tidy failed"
fi

print_check "Checking for security vulnerabilities..."
if command -v govulncheck &> /dev/null; then
    if govulncheck ./... > /dev/null 2>&1; then
        print_success "No known vulnerabilities found"
    else
        print_warning "govulncheck found potential issues"
    fi
else
    print_info "govulncheck not installed (skipping vulnerability check)"
fi

# ============================================
# 9. Integration Test
# ============================================
print_header "9. Integration Test"

print_check "Running workflow demo in test mode..."
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

# Create minimal test
cat > .zshrc << 'EOF'
# Test config
export PATH="/usr/local/bin:$PATH"
alias gs='git status'
EOF

if "$OLDPWD/build/gz-shellforge" migrate .zshrc --output-dir modules --manifest manifest.yaml > /dev/null 2>&1; then
    if [[ -f "manifest.yaml" && -d "modules" ]]; then
        print_success "Migration integration test passed"

        # Test build
        if "$OLDPWD/build/gz-shellforge" build --manifest manifest.yaml --config-dir modules --os Mac --output .zshrc.new > /dev/null 2>&1; then
            print_success "Build integration test passed"

            # Test diff
            if "$OLDPWD/build/gz-shellforge" diff .zshrc .zshrc.new --format summary > /dev/null 2>&1; then
                print_success "Diff integration test passed"
            else
                print_error "Diff integration test failed"
            fi
        else
            print_error "Build integration test failed"
        fi
    else
        print_error "Migration didn't create expected files"
    fi
else
    print_error "Migration integration test failed"
fi

cd "$OLDPWD"
rm -rf "$TEMP_DIR"

# ============================================
# Summary
# ============================================
print_header "Validation Summary"

echo ""
echo -e "${GREEN}Checks passed:  $CHECKS_PASSED${NC}"
echo -e "${RED}Checks failed:  $CHECKS_FAILED${NC}"
echo -e "${YELLOW}Warnings:       $WARNINGS${NC}"
echo ""

if [[ $CHECKS_FAILED -eq 0 ]]; then
    print_success "All critical checks passed! Ready for release."
    echo ""
    echo "Next steps:"
    echo "  1. Review CHANGELOG.md"
    echo "  2. Create git tag: git tag -a v$MAIN_VERSION -m 'Release v$MAIN_VERSION'"
    echo "  3. Push tag: git push origin v$MAIN_VERSION"
    echo "  4. Create GitHub release with build artifacts"
    echo ""
    exit 0
else
    print_error "Some critical checks failed. Please fix before release."
    echo ""
    exit 1
fi
