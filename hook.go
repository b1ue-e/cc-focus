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

	if settings.Hooks == nil {
		settings.Hooks = make(map[string][]hookEntry)
	}
	settings.Hooks["Stop"] = append(settings.Hooks["Stop"], entry)

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
