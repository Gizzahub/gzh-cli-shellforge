#!/usr/bin/env bash
set -euo pipefail
# Synthetic profiles exercise the exact boundary without changing source coverage.
script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
expect_status() {
    local expected=$1 profile=$2 actual
    if printf '%s\n' "$profile" | awk -f "$script_dir/check-coverage.awk"; then
        actual=0
    else
        actual=$?
    fi
    if [[ "$actual" != "$expected" ]]; then
        printf 'coverage fixture expected %s, got %s\n' "$expected" "$actual" >&2
        exit 1
    fi
}
expect_status 1 $'mode: set\nfixture.go:1.1,1.2 7999 1\nfixture.go:2.1,2.2 2001 0'
expect_status 0 $'mode: set\nfixture.go:1.1,1.2 80 1\nfixture.go:2.1,2.2 20 0'
expect_status 1 ''
expect_status 1 $'mode: set\nfixture.go:1.1,1.2 invalid 1'
expect_status 1 $'mode: set\nfixture.go:1.1,1.2 0 0'
if awk -f "$script_dir/check-coverage.awk" "$script_dir/coverage-profile-that-does-not-exist.out"; then
    echo 'missing profile unexpectedly passed' >&2
    exit 1
fi
printf 'Coverage boundary fixtures passed\n'
