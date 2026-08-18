package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	path, err := configPath()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "ticket-control", "config.toml"); path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestConfigPathDefaultsToDotConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", home)

	path, err := configPath()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".config", "ticket-control", "config.toml"); path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("[printer]\nip = \"192.168.1.50\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	config, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if config.Printer.IP != "192.168.1.50" {
		t.Fatalf("printer IP = %q", config.Printer.IP)
	}
}

func TestLoadConfigReportsMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	_, err := loadConfig(path)
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("error = %q", err)
	}
}

func TestLoadConfigRejectsInvalidConfig(t *testing.T) {
	for name, contents := range map[string]string{
		"malformed":  "[printer\n",
		"missing-ip": "[printer]\n",
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.toml")
			if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := loadConfig(path)
			if err == nil {
				t.Fatal("loadConfig succeeded")
			}
			if name == "missing-ip" && !strings.Contains(err.Error(), "printer.ip is required") {
				t.Fatalf("error = %q", err)
			}
		})
	}
}

func TestPrinterIPPrecedence(t *testing.T) {
	a := app{config: config{}}
	a.config.Printer.IP = "config-ip"
	t.Setenv("PRINTER_IP", "environment-ip")

	if got := a.printerIP("flag-ip"); got != "flag-ip" {
		t.Fatalf("flag IP = %q", got)
	}
	if got := a.printerIP(""); got != "environment-ip" {
		t.Fatalf("environment IP = %q", got)
	}
	t.Setenv("PRINTER_IP", "")
	if got := a.printerIP(""); got != "config-ip" {
		t.Fatalf("config IP = %q", got)
	}
}
