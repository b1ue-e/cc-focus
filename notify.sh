#!/bin/bash
# Called by Claude Code Stop hook. Forwards event JSON to cc-focus daemon.

SOCKET="${XDG_CACHE_HOME:-$HOME/.cache}/cc-focus/daemon.sock"

if [ ! -S "$SOCKET" ]; then
    exit 0
fi

# Build minimal event JSON from stdin (CC passes hook data via stdin)
event="{}"
if [ -p /dev/stdin ] || [ -t 0 ]; then
    : # no stdin
else
    input=$(cat)
    if [ -n "$input" ]; then
        stop_reason=$(echo "$input" | python3 -c "import sys,json; print(json.load(sys.stdin).get('stop_reason',''))" 2>/dev/null || true)
        if [ -n "$stop_reason" ]; then
            event="{\"stop_reason\":\"$stop_reason\"}"
        fi
    fi
fi

echo "$event" | nc -U "$SOCKET" 2>/dev/null || true
