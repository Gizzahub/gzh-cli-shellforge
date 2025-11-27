#!/bin/bash
# Google Cloud SDK autocompletion

if [ -f "$HOMEBREW_PREFIX/share/google-cloud-sdk/path.bash.inc" ]; then
    . "$HOMEBREW_PREFIX/share/google-cloud-sdk/path.bash.inc"
fi

if [ -f "$HOMEBREW_PREFIX/share/google-cloud-sdk/completion.bash.inc" ]; then
    . "$HOMEBREW_PREFIX/share/google-cloud-sdk/completion.bash.inc"
fi
