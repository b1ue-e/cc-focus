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
