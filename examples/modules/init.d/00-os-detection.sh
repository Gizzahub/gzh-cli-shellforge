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
    FreeBSD)
        export MACHINE="FreeBSD"
        ;;
    OpenBSD)
        export MACHINE="OpenBSD"
        ;;
    NetBSD)
        export MACHINE="NetBSD"
        ;;
    *)
        export MACHINE="Unknown"
        ;;
esac
