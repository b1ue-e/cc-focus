# cc-focus Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a cross-platform daemon that detects Claude Code response completion via Stop hook and brings terminal to front.

**Architecture:** Go single binary with build-tag-isolated platform code. CLI dispatches `start`/`stop`/`status`/`install`/`uninstall`. Daemon listens on Unix domain socket, receives events from notify.sh (called by CC Stop hook), checks foreground app, activates terminal when appropriate.

**Tech Stack:** Go 1.22+, BurntSushi/toml, macOS osascript, Linux wmctrl/xdotool

---

### Task 1: Project scaffold

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `Makefile`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/shouyaoqi/cc-focus && go mod init github.com/b1ue-e/cc-focus
```

- [ ] **Step 2: Write main.go skeleton with CLI dispatch**

```go
package main

import (
	"fmt"
	"os"
)

const usage = `cc-focus — auto-focus terminal when Claude Code finishes responding

Usage:
  cc-focus start       Start the daemon
  cc-focus stop        Stop the daemon
  cc-focus status      Show daemon status and config
  cc-focus install     Install Stop hook into Claude Code settings
  cc-focus uninstall   Remove Stop hook from Claude Code settings
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "start":
		cmdStart()
	case "stop":
		cmdStop()
	case "status":
		cmdStatus()
	case "install":
		cmdInstall()
	case "uninstall":
		cmdUninstall()
	default:
		fmt.Print(usage)
		os.Exit(1)
	}
}

func cmdStart()    { fmt.Println("start: not implemented") }
func cmdStop()     { fmt.Println("stop: not implemented") }
func cmdStatus()   { fmt.Println("status: not implemented") }
func cmdInstall()  { fmt.Println("install: not implemented") }
func cmdUninstall(){ fmt.Println("uninstall: not implemented") }
```

- [ ] **Step 3: Write Makefile**

```makefile
.PHONY: build install clean

BINARY = cc-focus
PREFIX ?= $(HOME)/.local

build:
	go build -o $(BINARY) .

install: build
	mkdir -p $(PREFIX)/bin
	cp $(BINARY) $(PREFIX)/bin/
	cp notify.sh $(PREFIX)/bin/cc-focus-notify

clean:
	rm -f $(BINARY)
```

- [ ] **Step 4: Build and verify**

```bash
cd /Users/shouyaoqi/cc-focus && make build && ./cc-focus
```
Expected: prints usage text.

- [ ] **Step 5: Commit**

```bash
git add go.mod main.go Makefile
git commit -m "feat: project scaffold with CLI skeleton"
```

---

### Task 2: Config parsing

**Files:**
- Create: `config.go`

- [ ] **Step 1: Add toml dependency**

```bash
cd /Users/shouyaoqi/cc-focus && go get github.com/BurntSushi/toml
```

- [ ] **Step 2: Write config.go**

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Terminal struct {
		Name        string `toml:"name"`
		ActivateCmd string `toml:"activate_cmd"`
	} `toml:"terminal"`
	Hook struct {
		SocketPath string `toml:"socket_path"`
	} `toml:"hook"`
}

func defaultConfig() Config {
	cfg := Config{}
	cfg.Terminal.Name = "Ghostty"
	cfg.Hook.SocketPath = socketPath()
	return cfg
}

func configDir() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "cc-focus")
}

func configPath() string {
	return filepath.Join(configDir(), "config.toml")
}

func cacheDir() string {
	dir := os.Getenv("XDG_CACHE_HOME")
	if dir == "" {
		dir = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return filepath.Join(dir, "cc-focus")
}

func socketPath() string {
	return filepath.Join(cacheDir(), "daemon.sock")
}

func pidPath() string {
	return filepath.Join(cacheDir(), "daemon.pid")
}

func loadConfig() (Config, error) {
	cfg := defaultConfig()
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
```

- [ ] **Step 3: Build and verify**

```bash
cd /Users/shouyaoqi/cc-focus && go build ./...
```
Expected: compiles successfully.

- [ ] **Step 4: Commit**

```bash
git add config.go go.mod go.sum
git commit -m "feat: TOML config parsing with defaults"
```

---

### Task 3: Focus interface and macOS implementation

**Files:**
- Create: `focus.go`
- Create: `focus_darwin.go`

- [ ] **Step 1: Write focus.go (interface + common)**

```go
package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func isTerminalFocused(cfg Config) (bool, error) {
	name, err := foregroundAppName()
	if err != nil {
		return false, err
	}
	return strings.EqualFold(name, cfg.Terminal.Name), nil
}

func activateTerminal(cfg Config) error {
	if cfg.Terminal.ActivateCmd != "" {
		return runActivateCmd(cfg.Terminal.ActivateCmd)
	}
	return platformActivate(cfg.Terminal.Name)
}

func runActivateCmd(cmdStr string) error {
	cmd := exec.Command("sh", "-c", cmdStr)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("activate: %w: %s", err, string(out))
	}
	return nil
}
```

- [ ] **Step 2: Write focus_darwin.go**

```go
//go:build darwin

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func foregroundAppName() (string, error) {
	script := `tell application "System Events" to get name of first application process whose frontmost is true`
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("foreground check: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func platformActivate(appName string) error {
	script := fmt.Sprintf(`tell application "%s" to activate`, appName)
	cmd := exec.Command("osascript", "-e", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("activate %s: %w: %s", appName, err, string(out))
	}
	return nil
}
```

- [ ] **Step 3: Build and verify on macOS**

```bash
cd /Users/shouyaoqi/cc-focus && go build -tags darwin ./...
```
Expected: compiles on darwin.

- [ ] **Step 4: Commit**

```bash
git add focus.go focus_darwin.go
git commit -m "feat: focus interface with macOS implementation"
```

---

### Task 4: Linux focus implementation

**Files:**
- Create: `focus_linux.go`

- [ ] **Step 1: Write focus_linux.go**

```go
//go:build linux

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func foregroundAppName() (string, error) {
	cmd := exec.Command("xdotool", "getactivewindow", "getwindowname")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("foreground check (xdotool required): %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func platformActivate(appName string) error {
	cmd := exec.Command("wmctrl", "-a", appName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("activate %s (wmctrl required): %w: %s", appName, err, string(out))
	}
	return nil
}
```

- [ ] **Step 2: Verify conditional compilation**

```bash
cd /Users/shouyaoqi/cc-focus && GOOS=linux GOARCH=amd64 go build ./...
```
Expected: compiles for linux/amd64.

- [ ] **Step 3: Commit**

```bash
git add focus_linux.go
git commit -m "feat: Linux focus implementation via wmctrl/xdotool"
```

---

### Task 5: Daemon lifecycle and socket handling

**Files:**
- Create: `daemon.go`

- [ ] **Step 1: Write daemon.go**

```go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

type Event struct {
	Reason string `json:"stop_reason"`
}

func cmdStart() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	if daemonRunning() {
		fmt.Println("daemon is already running")
		os.Exit(0)
	}

	os.MkdirAll(cacheDir(), 0755)
	sockPath := cfg.Hook.SocketPath

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot find executable: %v\n", err)
		os.Exit(1)
	}

	attr := &os.ProcAttr{
		Files: []*os.File{nil, nil, nil},
	}
	proc, err := os.StartProcess(exe, []string{exe, "--daemon", sockPath}, attr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start daemon: %v\n", err)
		os.Exit(1)
	}

	os.WriteFile(pidPath(), []byte(strconv.Itoa(proc.Pid)), 0644)
	fmt.Printf("daemon started (pid: %d)\n", proc.Pid)
}

func cmdStop() {
	data, err := os.ReadFile(pidPath())
	if err != nil {
		fmt.Println("daemon is not running")
		os.Exit(0)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid pid file: %v\n", err)
		os.Exit(1)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "find process: %v\n", err)
		os.Exit(1)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		fmt.Fprintf(os.Stderr, "signal daemon: %v\n", err)
		os.Exit(1)
	}

	os.Remove(pidPath())
	os.Remove(socketPath())
	fmt.Println("daemon stopped")
}

func cmdStatus() {
	cfg, _ := loadConfig()
	if daemonRunning() {
		fmt.Println("daemon: running")
	} else {
		fmt.Println("daemon: stopped")
	}
	fmt.Printf("terminal: %s\n", cfg.Terminal.Name)
	fmt.Printf("socket: %s\n", cfg.Hook.SocketPath)
}

func daemonRunning() bool {
	data, err := os.ReadFile(pidPath())
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

func runDaemon(sockPath string) {
	cfg, _ := loadConfig()
	logPath := filepath.Join(cacheDir(), "events.log")

	os.Remove(sockPath)
	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	defer os.Remove(sockPath)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		listener.Close()
		os.Remove(pidPath())
		os.Exit(0)
	}()

	log.Printf("daemon listening on %s", sockPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Printf("accept: %v", err)
			}
			continue
		}
		go handleConnection(conn, cfg, logPath)
	}
}

func handleConnection(conn net.Conn, cfg Config, logPath string) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			log.Printf("parse event: %v", err)
			continue
		}

		focused, err := isTerminalFocused(cfg)
		if err != nil {
			log.Printf("foreground check: %v", err)
			continue
		}

		if !focused {
			if err := activateTerminal(cfg); err != nil {
				log.Printf("activate: %v", err)
				continue
			}
		}

		logEvent(logPath, event)
	}
}

func logEvent(logPath string, event Event) {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	entry, _ := json.Marshal(event)
	f.Write(append(entry, '\n'))
}
```

- [ ] **Step 2: Add --daemon flag handling to main.go**

In `main()` function, before the existing switch, add:

```go
// main.go — insert at the top of main(), before the existing switch:
if len(os.Args) >= 2 && os.Args[1] == "--daemon" {
	sockPath := defaultSocket
	if len(os.Args) >= 3 {
		sockPath = os.Args[2]
	}
	runDaemon(sockPath)
	return
}
```

And add at top of main.go:

```go
const defaultSocket = "" // filled at runtime from config
```

Actually, let me keep it simple — remove the `defaultSocket` const and just use the argument:

```go
// At the top of main(), before the len(os.Args) < 2 check:
if len(os.Args) >= 3 && os.Args[1] == "--daemon" {
	runDaemon(os.Args[2])
	return
}
```

- [ ] **Step 3: Add missing imports to daemon.go**

Ensure these imports are present at the top of daemon.go (in addition to what's shown above):

```go
import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)
```

- [ ] **Step 4: Build and smoke test**

```bash
cd /Users/shouyaoqi/cc-focus && go build -o cc-focus .
./cc-focus status
./cc-focus start
./cc-focus status
./cc-focus stop
```
Expected: status shows "stopped" → start shows pid → status shows "running" → stop shows "stopped".

- [ ] **Step 5: Commit**

```bash
git add daemon.go main.go
git commit -m "feat: daemon lifecycle with Unix socket listener"
```

---

### Task 6: Hook install/uninstall in Claude Code settings

**Files:**
- Create: `hook.go`

- [ ] **Step 1: Write hook.go**

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ccSettings struct {
	Hooks map[string][]hookEntry `json:"hooks"`
}

type hookEntry struct {
	Command string `json:"command"`
}

func ccSettingsPath() string {
	return filepath.Join(os.Getenv("HOME"), ".claude", "settings.local.json")
}

func notifyScriptPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), "cc-focus-notify"), nil
}

func cmdInstall() {
	notifyPath, err := notifyScriptPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot determine notify script path: %v\n", err)
		os.Exit(1)
	}

	settingsPath := ccSettingsPath()
	settings := loadSettings(settingsPath)

	entry := hookEntry{Command: notifyPath}
	existing := settings.Hooks["Stop"]
	for _, e := range existing {
		if e.Command == notifyPath {
			fmt.Println("hook already installed")
			return
		}
	}

	settings.Hooks["Stop"] = append(existing, entry)
	if settings.Hooks == nil {
		settings.Hooks = make(map[string][]hookEntry)
	}

	saveSettings(settingsPath, settings)
	fmt.Printf("hook installed: Stop → %s\n", notifyPath)
}

func cmdUninstall() {
	notifyPath, err := notifyScriptPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot determine notify script path: %v\n", err)
		os.Exit(1)
	}

	settingsPath := ccSettingsPath()
	settings := loadSettings(settingsPath)

	existing := settings.Hooks["Stop"]
	filtered := make([]hookEntry, 0, len(existing))
	for _, e := range existing {
		if e.Command != notifyPath {
			filtered = append(filtered, e)
		}
	}

	settings.Hooks["Stop"] = filtered
	saveSettings(settingsPath, settings)
	fmt.Println("hook uninstalled")
}

func loadSettings(path string) ccSettings {
	var settings ccSettings
	data, err := os.ReadFile(path)
	if err != nil {
		return settings
	}
	json.Unmarshal(data, &settings)
	return settings
}

func saveSettings(path string, settings ccSettings) {
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal settings: %v\n", err)
		os.Exit(1)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write settings: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Build and verify**

```bash
cd /Users/shouyaoqi/cc-focus && go build ./...
```
Expected: compiles successfully.

- [ ] **Step 3: Commit**

```bash
git add hook.go
git commit -m "feat: hook install/uninstall in CC settings.json"
```

---

### Task 7: notify.sh forwarding script

**Files:**
- Create: `notify.sh`

- [ ] **Step 1: Write notify.sh**

```bash
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
```

- [ ] **Step 2: Make executable**

```bash
chmod +x /Users/shouyaoqi/cc-focus/notify.sh
```

- [ ] **Step 3: Verify script parses correctly**

```bash
echo '{"stop_reason":"end_turn"}' | bash /Users/shouyaoqi/cc-focus/notify.sh
```
Expected: exits 0 (no socket, so just exits silently).

- [ ] **Step 4: Commit**

```bash
git add notify.sh
git commit -m "feat: notify.sh hook forwarding script"
```

---

### Task 8: End-to-end integration test

**Files:**
- Modify: `main.go` (add missing import for `runDaemon`)

- [ ] **Step 1: Build release binary**

```bash
cd /Users/shouyaoqi/cc-focus && go build -o cc-focus .
```
Expected: builds successfully.

- [ ] **Step 2: Start daemon and verify socket**

```bash
./cc-focus start && ls -la ~/.cache/cc-focus/daemon.sock
```
Expected: daemon starts, socket exists.

- [ ] **Step 3: Send test event to daemon**

```bash
echo '{"stop_reason":"end_turn"}' | nc -U ~/.cache/cc-focus/daemon.sock
```
Expected: daemon logs event. If terminal "Ghostty" is not focused, it should activate Ghostty.

- [ ] **Step 4: Check event log**

```bash
cat ~/.cache/cc-focus/events.log
```
Expected: contains `{"stop_reason":"end_turn"}`.

- [ ] **Step 5: Install hook**

```bash
./cc-focus install
cat ~/.claude/settings.local.json | python3 -m json.tool
```
Expected: settings.local.json contains Stop hook entry.

- [ ] **Step 6: Uninstall hook and stop daemon**

```bash
./cc-focus uninstall
./cc-focus stop
./cc-focus status
```
Expected: hook removed, daemon stopped.

- [ ] **Step 7: Commit any remaining changes**

```bash
git add -A && git commit -m "feat: integration test passed"
```
