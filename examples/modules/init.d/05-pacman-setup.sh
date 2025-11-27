#!/bin/bash
# Pacman setup for Arch-based Linux systems

if command -v pacman &> /dev/null; then
    alias update='sudo pacman -Syu'
    alias install='sudo pacman -S'
fi
