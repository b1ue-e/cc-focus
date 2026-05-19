package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ccSettings struct {
	Hooks map[string][]stopHookGroup `json:"hooks"`
}

type stopHookGroup struct {
	Matcher string        `json:"matcher"`
	Hooks   []hookCommand `json:"hooks"`
}

type hookCommand struct {
	Type    string `json:"type"`
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

	cmd := hookCommand{Type: "command", Command: notifyPath}
	group := stopHookGroup{Matcher: "", Hooks: []hookCommand{cmd}}

	if settings.Hooks == nil {
		settings.Hooks = make(map[string][]stopHookGroup)
	}

	for _, g := range settings.Hooks["Stop"] {
		for _, h := range g.Hooks {
			if h.Command == notifyPath {
				fmt.Println("hook already installed")
				return
			}
		}
	}

	settings.Hooks["Stop"] = append(settings.Hooks["Stop"], group)
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

	var kept []stopHookGroup
	for _, g := range settings.Hooks["Stop"] {
		var keptHooks []hookCommand
		for _, h := range g.Hooks {
			if h.Command != notifyPath {
				keptHooks = append(keptHooks, h)
			}
		}
		if len(keptHooks) > 0 {
			g.Hooks = keptHooks
			kept = append(kept, g)
		}
	}

	settings.Hooks["Stop"] = kept
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
