# .make/vars.mk - Common variables
# Included by main Makefile

# Project settings
BINARY_NAME := gz-shellforge
BUILD_DIR := build
MAIN_PKG := ./cmd/shellforge

# Version information
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildDate=$(BUILD_DATE)"

# Go commands
GO := go
GOBUILD := $(GO) build
GOTEST := $(GO) test
GOINSTALL := $(GO) install
GOMOD := $(GO) mod
GOFMT := $(GO) fmt
GOVET := $(GO) vet

# Test settings
COVERAGE_OUT := coverage.out
COVERAGE_HTML := coverage.html
TEST_TIMEOUT := 5m
RACE_FLAG := -race

# Linter settings
# .golangci.yml is the v2 schema, so this must be a v2 release — v1 cannot
# parse it. GOLANGCI_LINT_BARE drops the leading `v` because
# `golangci-lint version` prints "has version 2.13.1 built with ...".
#
# v2.13.1 rather than v2.12.2: 2.12.2 replays cached findings for source files
# that no longer exist. Measured 2026-09-04 right after this repo's own task
# worktree was reclaimed -- `make lint` reported 2 tagliatelle issues against
# ../../../worktrees/gzh-cli/gzh-cli-shellforge/claude__mbp__quality__golangci-lint-bin-tools/internal/domain/module.go,
# a path that had just been deleted, alongside "no such file or directory"
# warnings from the linter itself. The repo is clean; the findings were ghosts,
# and they turned devbox `lint-all` red. 2.13.1 keys its cache differently: on
# the same unmodified shared cache it reported 0 issues and named no worktree
# path. That is why the pin moved instead of the cache being wiped -- a cache
# wipe is machine-wide, one-shot, and would come back with the next reclaimed
# worktree. gzh-cli, gzh-cli-quality and gzh-cli-package-manager already pin
# v2.13.1, so this aligns with them rather than choosing a new version.
# (TASK-159, TASK-162)
GOLANGCI_LINT_VERSION := v2.13.1
GOLANGCI_LINT_BARE := $(GOLANGCI_LINT_VERSION:v%=%)
# All local lint paths use this exact managed binary. The directory is
# repo-owned rather than the shared $(GOPATH)/bin -- every Go project on a
# machine competed for one binary there, so two repos pinning different
# golangci-lint versions could not both be green (TASK-159). It stays
# overridable for isolated verification, while the filename remains fixed so
# install-lint and lint always agree on the destination.
GOLANGCI_LINT_DIR ?= $(CURDIR)/bin/tools
GOLANGCI_LINT_BIN := $(GOLANGCI_LINT_DIR)/golangci-lint
