# cc-focus Design Spec

## Overview

A cross-platform daemon that monitors Claude Code response completion via hooks and automatically brings the terminal window to front when the model finishes responding — so you can browse elsewhere and be notified when results are ready.

## Use Case

1. User sends a request in Claude Code (Ghostty terminal)
2. User switches to Edge browser or other apps while waiting
3. Claude Code completes its response (thinking + tool use + output)
4. cc-focus daemon detects completion and brings terminal to front
5. User sees results without manual Alt+Tab

## Architecture

```
Claude Code  →  Stop hook  →  notify.sh  →  unix socket  →  daemon  →  platform API  →  Terminal in focus
                                                                  │
                                                            event log (extensible for stats)
```

### Components

| Component | Role |
|-----------|------|
| `notify.sh` | Called by Claude Code Stop hook; forwards event JSON to daemon via Unix socket |
| `daemon` | Core process: listens on Unix socket, detects foreground app, switches window, logs events |
| `cc-focus` CLI | User-facing: `start`, `stop`, `status`, `install`, `uninstall` |
| settings.json | Claude Code hook config pointing to notify.sh |

## Behavior

When daemon receives a Stop event:

1. Detect current foreground application
2. If terminal is already in foreground → no-op
3. If terminal is not in foreground → activate terminal window
4. Log event (timestamp, session info) for future statistics

## CLI Interface

```
cc-focus start              # Start daemon in background
cc-focus stop               # Stop daemon
cc-focus status             # Check if daemon is running + current config
cc-focus install            # Add Stop hook to Claude Code settings
cc-focus uninstall          # Remove Stop hook
```

## Configuration

`~/.config/cc-focus/config.toml`:

```toml
[terminal]
name = "Ghostty"

[hook]
socket_path = ""  # default: ~/.cache/cc-focus/daemon.sock

[stats]
enabled = false   # future: token usage tracking
```

## Platform Support

- **macOS**: `osascript -e 'tell application "<terminal>" to activate'`
- **Linux X11**: `wmctrl` or `xdotool`
- **Linux Wayland**: compositor-specific, deferred to later version

## Technology

- **Language**: Go (single binary, cross-compilation, good daemon/socket support)
- **IPC**: Unix domain socket
- **Config**: TOML (BurntSushi/toml)

## Future Extensions

- Token usage and cost tracking from hook data
- Session duration statistics
- Stats dashboard / CLI report
- Wayland compositor support
- Configurable delay before switching
