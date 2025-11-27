#!/bin/bash
# Ruby version manager (rbenv)

if command -v rbenv &> /dev/null; then
    eval "$(rbenv init - bash)"
fi
