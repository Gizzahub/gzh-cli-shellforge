#!/bin/bash
# Homebrew PATH Initialization
# Adds Homebrew to PATH on macOS

if [[ "$MACHINE" == "Mac" ]]; then
    if [[ -d "/opt/homebrew/bin" ]]; then
        # Apple Silicon Mac
        export PATH="/opt/homebrew/bin:$PATH"
    elif [[ -d "/usr/local/bin" ]]; then
        # Intel Mac
        export PATH="/usr/local/bin:$PATH"
    fi
fi
