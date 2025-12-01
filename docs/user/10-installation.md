# Installation Guide

Complete installation guide for Shellforge on all supported platforms.

---

## Prerequisites

### Required
- **Go 1.21 or later** - [Download](https://go.dev/dl/)
- **Git** - For backup/restore features

### Optional
- **Make** - For building from source
- **Shell** - zsh 5.8+ or bash 4.0+ (usually pre-installed)

---

## Quick Installation

### Option 1: Install via Go (Recommended)

**Fastest method - works on all platforms:**

```bash
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest
```

This installs `gz-shellforge` to `$(go env GOPATH)/bin/`.

### Option 2: Download Pre-built Binary

**Coming soon** - Pre-built binaries for macOS, Linux, and BSD.

### Option 3: Build from Source

**For contributors and customization:**

```bash
git clone https://github.com/gizzahub/gzh-cli-shellforge.git
cd gzh-cli-shellforge
make install
```

---

## Platform-Specific Installation

### macOS

#### Method 1: Go Install (Recommended)

```bash
# Install Go if not installed
brew install go

# Install Shellforge
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest

# Add Go bin to PATH (add to ~/.zshrc)
export PATH="$PATH:$(go env GOPATH)/bin"

# Reload shell
source ~/.zshrc
```

#### Method 2: Homebrew (Coming Soon)

```bash
brew tap gizzahub/tap
brew install shellforge
```

#### Method 3: Build from Source

```bash
# Install prerequisites
brew install go git

# Clone and build
git clone https://github.com/gizzahub/gzh-cli-shellforge.git
cd gzh-cli-shellforge
make install

# Verify
gz-shellforge --version
```

#### macOS-Specific Notes

**M1/M2 (Apple Silicon) Macs:**
- Homebrew is in `/opt/homebrew/`
- Use Rosetta 2 for x86_64 binaries if needed

**Intel Macs:**
- Homebrew is in `/usr/local/`
- Standard installation works

**Security Warning:**
If you get "developer cannot be verified" warning:
```bash
xattr -d com.apple.quarantine $(which gz-shellforge)
```

---

### Linux

#### Debian/Ubuntu

```bash
# Install Go
sudo apt update
sudo apt install golang-go git

# Or install latest Go manually
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Add to PATH (add to ~/.bashrc)
export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin

# Install Shellforge
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest

# Verify
gz-shellforge --version
```

#### CentOS/RHEL/Fedora

```bash
# Install Go
sudo dnf install golang git
# or
sudo yum install golang git

# Or install latest Go manually
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Add to PATH (add to ~/.bashrc)
export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin

# Install Shellforge
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest

# Verify
gz-shellforge --version
```

#### Arch Linux

```bash
# Install Go
sudo pacman -S go git

# Install Shellforge
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest

# Add to PATH (add to ~/.bashrc)
export PATH=$PATH:$(go env GOPATH)/bin

# Verify
gz-shellforge --version
```

#### Linux-Specific Notes

**Systemd user services:**
For auto-backup on login, create `~/.config/systemd/user/shellforge-backup.service`

**SELinux:**
If SELinux is enabled, you may need to adjust permissions:
```bash
chcon -t bin_t $(which gz-shellforge)
```

---

### BSD (Experimental)

#### FreeBSD 13+

```bash
# Install Go
pkg install go git

# Install Shellforge
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest

# Add to PATH (add to ~/.cshrc or ~/.shrc)
setenv PATH ${PATH}:`go env GOPATH`/bin

# Verify
gz-shellforge --version
```

**Note**: BSD support is experimental. Some features may not work as expected.

---

## Verifying Installation

### Check Version

```bash
gz-shellforge --version
# Expected output: shellforge version 0.2.0-alpha
```

### Check Binary Location

```bash
which gz-shellforge
# Expected: /Users/you/go/bin/gz-shellforge (macOS)
# Expected: /home/you/go/bin/gz-shellforge (Linux)
```

### Test Basic Command

```bash
gz-shellforge --help
# Should show help text with available commands
```

### Verify Go Environment

```bash
go env GOPATH
go env GOROOT
```

---

## Post-Installation Setup

### 1. Add to PATH (If Not Already Done)

#### zsh (macOS default)

Add to `~/.zshrc`:

```bash
# Add Go binaries to PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

Reload:
```bash
source ~/.zshrc
```

#### bash (Linux default)

Add to `~/.bashrc`:

```bash
# Add Go binaries to PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

Reload:
```bash
source ~/.bashrc
```

#### fish

Add to `~/.config/fish/config.fish`:

```fish
set -gx PATH $PATH (go env GOPATH)/bin
```

Reload:
```fish
source ~/.config/fish/config.fish
```

### 2. Enable Shell Completion (Optional)

#### Bash

```bash
gz-shellforge completion bash > /etc/bash_completion.d/gz-shellforge
source /etc/bash_completion.d/gz-shellforge
```

#### Zsh

```bash
gz-shellforge completion zsh > "${fpath[1]}/_gz-shellforge"
```

#### Fish

```bash
gz-shellforge completion fish > ~/.config/fish/completions/gz-shellforge.fish
```

### 3. Create Backup Directory

```bash
mkdir -p ~/.backup/shellforge
```

### 4. Test Installation

```bash
# Create test directory
mkdir -p ~/test-shellforge
cd ~/test-shellforge

# Create simple manifest
cat > manifest.yaml <<EOF
modules:
  - name: test
    file: modules/test.sh
    requires: []
    description: Test module
EOF

# Create test module
mkdir -p modules
echo '# Test module' > modules/test.sh
echo 'echo "Shellforge works!"' >> modules/test.sh

# Validate
gz-shellforge validate

# Build
gz-shellforge build --os Mac --dry-run

# Success! Clean up
cd ~
rm -rf ~/test-shellforge
```

---

## Updating Shellforge

### Update to Latest Version

```bash
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest
```

### Update to Specific Version

```bash
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@v0.2.0
```

### Check for Updates

```bash
# Current version
gz-shellforge --version

# Latest version (check GitHub)
curl -s https://api.github.com/repos/gizzahub/gzh-cli-shellforge/releases/latest | grep '"tag_name"'
```

---

## Uninstalling Shellforge

### Remove Binary

```bash
rm $(which gz-shellforge)
# or
rm $(go env GOPATH)/bin/gz-shellforge
```

### Remove Backup Data (Optional)

```bash
rm -rf ~/.backup/shellforge
```

### Remove Shell Completion (Optional)

```bash
# Bash
rm /etc/bash_completion.d/gz-shellforge

# Zsh
rm "${fpath[1]}/_gz-shellforge"

# Fish
rm ~/.config/fish/completions/gz-shellforge.fish
```

### Remove from PATH

Remove the export line from your shell config:
```bash
# Remove this line from ~/.zshrc or ~/.bashrc
export PATH="$PATH:$(go env GOPATH)/bin"
```

---

## Building from Source (Advanced)

### Clone Repository

```bash
git clone https://github.com/gizzahub/gzh-cli-shellforge.git
cd gzh-cli-shellforge
```

### Install Dependencies

```bash
go mod download
```

### Build Binary

```bash
# Build to build/ directory
make build

# Or use go directly
go build -o build/gz-shellforge cmd/shellforge/main.go
```

### Run Tests

```bash
make test
```

### Install to System

```bash
# Install to $GOPATH/bin
make install

# Or copy manually
cp build/gz-shellforge $(go env GOPATH)/bin/
```

### Build for Multiple Platforms

```bash
# Build for all platforms
make build-all

# Outputs:
# build/gz-shellforge-darwin-amd64   (macOS Intel)
# build/gz-shellforge-darwin-arm64   (macOS M1/M2)
# build/gz-shellforge-linux-amd64    (Linux x86_64)
# build/gz-shellforge-linux-arm64    (Linux ARM)
```

### Development Build

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o build/gz-shellforge cmd/shellforge/main.go

# Run without installing
go run cmd/shellforge/main.go --version
```

---

## Installation Troubleshooting

### Go Version Issues

**Problem:** Go version too old

**Solution:**
```bash
# Check current version
go version

# Update Go
# macOS: brew upgrade go
# Linux: Download from golang.org
```

### PATH Issues

**Problem:** Command not found after installation

**Solution:**
```bash
# Check if binary exists
ls -la $(go env GOPATH)/bin/gz-shellforge

# Add to PATH temporarily
export PATH="$PATH:$(go env GOPATH)/bin"

# Add permanently (see Post-Installation Setup)
```

### Permission Issues

**Problem:** Permission denied when installing

**Solution:**
```bash
# Don't use sudo with go install
# Instead, ensure GOPATH is writable
chmod 755 $(go env GOPATH)/bin

# If still fails, check GOPATH ownership
ls -la $(go env GOPATH)
```

### Network Issues

**Problem:** Cannot download packages

**Solution:**
```bash
# Use Go proxy
export GOPROXY=https://proxy.golang.org,direct

# Or use mirror (China)
export GOPROXY=https://goproxy.cn,direct

# Retry installation
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest
```

### Build Failures

**Problem:** Build fails with errors

**Solution:**
```bash
# Clean Go cache
go clean -modcache

# Clean build artifacts
make clean

# Rebuild
make build
```

---

## System Requirements

### Minimum Requirements
- **OS**: macOS 10.15+, Linux (any recent distro), FreeBSD 13+
- **CPU**: Any 64-bit processor (amd64 or arm64)
- **RAM**: 10MB (binary is ~8MB)
- **Disk**: 50MB (including backups)
- **Go**: 1.21+ (for installation)

### Recommended Requirements
- **OS**: macOS 13+, Ubuntu 22.04+, Fedora 38+
- **CPU**: Multi-core (for faster builds)
- **RAM**: 50MB (for large configs)
- **Disk**: 500MB (for extensive backups)
- **Shell**: zsh 5.8+ or bash 5.0+

---

## Docker Installation (Alternative)

### Run in Docker

```bash
# Pull image (coming soon)
docker pull ghcr.io/gizzahub/shellforge:latest

# Run with volume mount
docker run -v ~/.zshrc:/input/.zshrc \
           -v ./output:/output \
           ghcr.io/gizzahub/shellforge:latest \
           migrate /input/.zshrc
```

### Build Docker Image

```bash
# Clone repository
git clone https://github.com/gizzahub/gzh-cli-shellforge.git
cd gzh-cli-shellforge

# Build image
docker build -t shellforge:local .

# Run
docker run shellforge:local --version
```

---

## Next Steps

After installation:

1. **[Quick Start Guide](00-quick-start.md)** - Get started in 5 minutes
2. **[Basic Usage](20-basic-usage.md)** - Learn essential commands
3. **[Complete Workflows](30-workflows.md)** - Full workflow examples

---

## Getting Help

- **Installation Issues**: [Troubleshooting Guide](60-troubleshooting.md)
- **Command Help**: `gz-shellforge --help`
- **GitHub Issues**: [Report a problem](https://github.com/gizzahub/gzh-cli-shellforge/issues)
- **Discussions**: [Ask a question](https://github.com/gizzahub/gzh-cli-shellforge/discussions)

---

**Last Updated**: 2025-12-01
**Tested On**: macOS 14 (Sonoma), Ubuntu 22.04, Arch Linux, FreeBSD 13
