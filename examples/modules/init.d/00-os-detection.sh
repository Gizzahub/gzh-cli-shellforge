#!/bin/bash
# OS Detection
# Detects the operating system and sets MACHINE variable

case "$(uname -s)" in
    Darwin)
        export MACHINE="Mac"
        ;;
    Linux)
        export MACHINE="Linux"
        ;;
    *)
        export MACHINE="Unknown"
        ;;
esac
