#!/bin/bash
# Conda/Mamba initialization

if [ -f "$HOME/mambaforge/etc/profile.d/conda.sh" ]; then
    . "$HOME/mambaforge/etc/profile.d/conda.sh"
fi
