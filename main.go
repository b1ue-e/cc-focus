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
func cmdUninstall() { fmt.Println("uninstall: not implemented") }
