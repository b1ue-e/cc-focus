#!/bin/bash
# Called by Claude Code Stop/PermissionRequest hooks. Forwards event JSON to cc-focus daemon.

SOCKET="${XDG_CACHE_HOME:-$HOME/.cache}/cc-focus/daemon.sock"

if [ ! -S "$SOCKET" ]; then
    exit 0
fi

input=$(cat 2>/dev/null)
if [ -n "$input" ]; then
    echo "$input" | nc -U "$SOCKET" 2>/dev/null || true
fi
