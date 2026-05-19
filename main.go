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
  cc-focus stats       Show token usage statistics
  cc-focus install     Install Stop + PermissionRequest hooks into CC settings
  cc-focus uninstall   Remove Stop + PermissionRequest hooks from CC settings
`

func main() {
	if len(os.Args) >= 3 && os.Args[1] == "--daemon" {
		runDaemon(os.Args[2])
		return
	}

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
	case "stats":
		cmdStats()
	case "install":
		cmdInstall()
	case "uninstall":
		cmdUninstall()
	default:
		fmt.Print(usage)
		os.Exit(1)
	}
}

