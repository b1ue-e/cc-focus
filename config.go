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
	Monitor struct {
		Enabled      bool   `toml:"enabled"`
		ProcessName  string `toml:"process_name"`
		PollInterval int    `toml:"poll_interval"`
	} `toml:"monitor"`
}

func defaultConfig() Config {
	cfg := Config{}
	cfg.Terminal.Name = "Ghostty"
	cfg.Hook.SocketPath = socketPath()
	cfg.Monitor.Enabled = true
	cfg.Monitor.ProcessName = "claude"
	cfg.Monitor.PollInterval = 1
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
