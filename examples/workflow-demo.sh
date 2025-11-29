#!/bin/bash
# Workflow Demo: Complete migrate→build→diff workflow
# This script demonstrates the end-to-end usage of gz-shellforge

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper function for step headers
step() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

# Helper function for success messages
success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Helper function for info messages
info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Find gz-shellforge binary
SHELLFORGE=""
if command -v gz-shellforge &> /dev/null; then
    SHELLFORGE="gz-shellforge"
elif [ -f "../build/gz-shellforge" ]; then
    SHELLFORGE="$(cd .. && pwd)/build/gz-shellforge"
    info "Using local binary: $SHELLFORGE"
elif [ -f "build/gz-shellforge" ]; then
    SHELLFORGE="$(pwd)/build/gz-shellforge"
    info "Using local binary: $SHELLFORGE"
else
    echo "Error: gz-shellforge is not installed or not in PATH"
    echo "Please run: make install"
    echo "Or run from project root after: make build"
    exit 1
fi

# Setup demo directories
DEMO_DIR="$(mktemp -d)/shellforge-demo"
mkdir -p "$DEMO_DIR"
cd "$DEMO_DIR"

info "Working directory: $DEMO_DIR"

# Step 1: Create a sample .zshrc file
step "Step 1: Create Sample .zshrc File"

cat > .zshrc << 'EOF'
# Sample .zshrc for demonstration
# This is a monolithic configuration file

# === OS Detection ===
case "$(uname -s)" in
  Darwin)
    export MACHINE="Mac"
    ;;
  Linux)
    export MACHINE="Linux"
    ;;
esac

# === PATH Setup ===
export PATH="/usr/local/bin:$PATH"

# --- Homebrew ---
if [[ "$MACHINE" == "Mac" ]]; then
    export PATH="/opt/homebrew/bin:$PATH"
fi

# === NVM Setup ===
export NVM_DIR="$HOME/.nvm"
if [[ -s "$NVM_DIR/nvm.sh" ]]; then
    source "$NVM_DIR/nvm.sh"
fi

# --- Git Aliases ---
alias gs='git status'
alias ga='git add'
alias gc='git commit'

# === Helper Functions ===
function mkcd() {
  mkdir -p "$1" && cd "$1"
}
EOF

success "Created sample .zshrc ($(wc -l < .zshrc) lines)"
info "Original file location: $DEMO_DIR/.zshrc"

# Step 2: Migrate to modular structure
step "Step 2: Migrate RC File to Modular Structure"

info "Running: $SHELLFORGE migrate .zshrc --output-dir modules --manifest manifest.yaml"
$SHELLFORGE migrate .zshrc --output-dir modules --manifest manifest.yaml

success "Migration complete!"
echo ""
echo "Generated files:"
ls -lh manifest.yaml
echo ""
echo "Module files:"
ls -lh modules/*/

# Step 3: Build configuration for Mac
step "Step 3: Build Configuration for Mac"

info "Running: $SHELLFORGE build --manifest manifest.yaml --config-dir modules --os Mac --output .zshrc.mac"
$SHELLFORGE build --manifest manifest.yaml --config-dir modules --os Mac --output .zshrc.mac

success "Build complete for Mac!"
echo ""
echo "Generated file:"
ls -lh .zshrc.mac
echo ""
echo "Preview (first 20 lines):"
head -20 .zshrc.mac

# Step 4: Build configuration for Linux
step "Step 4: Build Configuration for Linux (Optional)"

info "Running: $SHELLFORGE build --manifest manifest.yaml --config-dir modules --os Linux --output .zshrc.linux"
$SHELLFORGE build --manifest manifest.yaml --config-dir modules --os Linux --output .zshrc.linux

success "Build complete for Linux!"
echo ""
echo "Generated file:"
ls -lh .zshrc.linux

# Step 5: Compare original with generated
step "Step 5: Compare Original with Generated (Mac)"

info "Running: $SHELLFORGE diff .zshrc .zshrc.mac --format summary"
echo ""
$SHELLFORGE diff .zshrc .zshrc.mac --format summary

echo ""
echo ""
info "Running: $SHELLFORGE diff .zshrc .zshrc.mac --format unified"
echo ""
$SHELLFORGE diff .zshrc .zshrc.mac --format unified | head -30
echo "... (truncated)"

# Step 6: List modules
step "Step 6: List Modules in Manifest"

info "Running: $SHELLFORGE list --manifest manifest.yaml"
echo ""
$SHELLFORGE list --manifest manifest.yaml

# Step 7: Validate configuration
step "Step 7: Validate Configuration"

info "Running: $SHELLFORGE validate --manifest manifest.yaml --config-dir modules"
echo ""
$SHELLFORGE validate --manifest manifest.yaml --config-dir modules

# Summary
step "Workflow Complete!"

echo "All files are located in: $DEMO_DIR"
echo ""
echo "Files created:"
echo "  - .zshrc           (original monolithic file)"
echo "  - manifest.yaml    (module manifest)"
echo "  - modules/         (modular configuration files)"
echo "  - .zshrc.mac       (generated Mac configuration)"
echo "  - .zshrc.linux     (generated Linux configuration)"
echo ""
echo "To clean up this demo:"
echo "  rm -rf $DEMO_DIR"
echo ""
success "Demo completed successfully!"
