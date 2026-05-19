package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func startMonitor(cfg Config) {
	if !cfg.Monitor.Enabled {
		return
	}

	go func() {
		var prevState string
		tick := time.NewTicker(time.Duration(cfg.Monitor.PollInterval) * time.Second)
		defer tick.Stop()

		for range tick.C {
			pid, err := findCCProcess(cfg.Monitor.ProcessName)
			if err != nil {
				prevState = ""
				continue
			}

			state, err := processState(pid)
			if err != nil {
				prevState = ""
				continue
			}

			if prevState == "R" && (state == "S" || state == "T") {
				log.Printf("CC transition: %s→%s activating terminal", prevState, state)
				if err := activateTerminal(cfg); err != nil {
					log.Printf("activate: %v", err)
				}
			}

			prevState = state
		}
	}()
}

func findCCProcess(name string) (int, error) {
	cmd := exec.Command("pgrep", "-x", name)
	out, err := cmd.Output()
	if err == nil {
		pid, err := strconv.Atoi(strings.TrimSpace(strings.Split(string(out), "\n")[0]))
		if err == nil {
			return pid, nil
		}
	}

	cmd = exec.Command("pgrep", "-f", fmt.Sprintf("%s", name))
	out, err = cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("process %s not found", name)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(strings.Split(string(out), "\n")[0]))
	if err != nil {
		return 0, fmt.Errorf("parse pid: %w", err)
	}
	return pid, nil
}

func processState(pid int) (string, error) {
	cmd := exec.Command("ps", "-o", "state=", "-p", strconv.Itoa(pid))
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ps: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
