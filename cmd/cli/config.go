package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type config struct {
	Printer struct {
		IP string `toml:"ip"`
	} `toml:"printer"`
}

type app struct {
	config config
}

func newApp() (app, error) {
	path, err := configPath()
	if err != nil {
		return app{}, err
	}
	config, err := loadConfig(path)
	if err != nil {
		return app{}, err
	}
	return app{config: config}, nil
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve configuration directory: %w", err)
	}
	return filepath.Join(dir, "ticket-control", "config.toml"), nil
}

func loadConfig(path string) (config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return config{}, fmt.Errorf("decode config %s: %w", path, err)
	}
	if cfg.Printer.IP == "" {
		return config{}, fmt.Errorf("config %s: printer.ip is required", path)
	}
	return cfg, nil
}

func (a app) printerIP(flagIP string) string {
	if flagIP != "" {
		return flagIP
	}
	if envIP := os.Getenv("PRINTER_IP"); envIP != "" {
		return envIP
	}
	return a.config.Printer.IP
}
