#!/bin/bash
# macOS-specific configurations

if [ "$MACHINE" = "Mac" ]; then
    # Use GNU tools if available
    if [ -d "$HOMEBREW_PREFIX/opt/coreutils/libexec/gnubin" ]; then
        export PATH="$HOMEBREW_PREFIX/opt/coreutils/libexec/gnubin:$PATH"
    fi
fi
