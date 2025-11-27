#!/bin/bash
# Docker autocompletion

if command -v docker &> /dev/null; then
    if [ -f /usr/share/bash-completion/completions/docker ]; then
        . /usr/share/bash-completion/completions/docker
    fi
fi
