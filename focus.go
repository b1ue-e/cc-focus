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
