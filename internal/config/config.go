// Package config loads application settings that live outside the data directory
// (language, theme, data-dir path, global proxy). See docs/PRD.md §6.8–6.10.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	LangEN     = "en"
	LangZHHans = "zh-Hans"
	LangZHHant = "zh-Hant"

	ThemeLight  = "light"
	ThemeDark   = "dark"
	ThemeSystem = "system"

	ProxySystem = "system"
	ProxyOff    = "off"
	ProxyManual = "manual"

	configFileName = "config.json"
)

// Config is persisted under the OS app-config directory, not inside the data dir.
type Config struct {
	DataDir           string      `json:"dataDir"`
	Language          string      `json:"language"`
	Theme             string      `json:"theme"`
	Proxy             ProxyConfig `json:"proxy"`
	LastCollectionID  int64       `json:"lastCollectionId,omitempty"`
	LastEnvironmentID int64       `json:"lastEnvironmentId,omitempty"`
	languageFromDisk  bool        `json:"-"`
}

// ProxyConfig is the global HTTP/HTTPS proxy.
type ProxyConfig struct {
	Mode   string `json:"mode"`
	Manual string `json:"manual,omitempty"`
}

// Default returns PRD defaults without reading disk.
func Default() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &Config{
		DataDir:  filepath.Join(home, ".apidoc"),
		Language: "",
		Theme:    ThemeSystem,
		Proxy: ProxyConfig{
			Mode: ProxySystem,
		},
	}, nil
}

// Load reads app config. On first run (no language saved), detects OS locale,
// persists a concrete language, then returns the config.
func Load() (*Config, error) {
	cfg, err := Default()
	if err != nil {
		return nil, err
	}
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config: read: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("config: parse: %w", err)
		}
		if cfg.Language != "" {
			cfg.languageFromDisk = true
		}
	}
	normalize(cfg)
	if !cfg.languageFromDisk || cfg.Language == "" {
		cfg.Language = DetectLanguage(ReadOSLocale())
		if err := Save(cfg); err != nil {
			return nil, fmt.Errorf("config: first-run language: %w", err)
		}
		cfg.languageFromDisk = true
	}
	return cfg, nil
}

// Save writes config to the OS app-config directory.
func Save(cfg *Config) error {
	if cfg == nil {
		return errors.New("config: nil")
	}
	normalize(cfg)
	dir, err := AppConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("config: mkdir: %w", err)
	}
	path := filepath.Join(dir, configFileName)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("config: write: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("config: rename: %w", err)
	}
	cfg.languageFromDisk = true
	return nil
}

// AppConfigDir returns the OS-conventional path for settings.
func AppConfigDir() (string, error) {
	return appConfigDir()
}

func configPath() (string, error) {
	dir, err := AppConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

func normalize(cfg *Config) {
	if cfg.Theme == "" {
		cfg.Theme = ThemeSystem
	}
	if cfg.Proxy.Mode == "" {
		cfg.Proxy.Mode = ProxySystem
	}
	if cfg.DataDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			cfg.DataDir = filepath.Join(home, ".apidoc")
		}
	}
	if cfg.Language != "" && !ValidLanguage(cfg.Language) {
		cfg.Language = LangEN
	}
}

// ValidLanguage reports whether s is a supported UI language.
func ValidLanguage(s string) bool {
	switch s {
	case LangEN, LangZHHans, LangZHHant:
		return true
	default:
		return false
	}
}
