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
