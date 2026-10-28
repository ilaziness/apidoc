// Package config loads application settings that live outside the data directory
// (language, theme, data-dir path, global proxy). See docs/PRD.md §6.8–6.9.
package config

import (
	"os"
	"path/filepath"
)

// Language / theme / proxy mode values persisted in app config.
const (
	LangFollowOS = ""
	LangZH       = "zh"
	LangEN       = "en"

	ThemeLight  = "light"
	ThemeDark   = "dark"
	ThemeSystem = "system"

	ProxySystem = "system"
	ProxyOff    = "off"
	ProxyManual = "manual"
)

// Config is persisted under the OS app-config directory, not inside the data dir.
type Config struct {
	DataDir  string
	Language string // LangFollowOS | LangZH | LangEN
	Theme    string // ThemeLight | ThemeDark | ThemeSystem
	Proxy    ProxyConfig
}

// ProxyConfig is the global HTTP/HTTPS proxy.
type ProxyConfig struct {
	Mode   string // ProxySystem | ProxyOff | ProxyManual
	Manual string // URL when Mode == ProxyManual
}

// Default returns PRD defaults without reading disk.
func Default() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &Config{
		DataDir:  filepath.Join(home, ".apidoc"),
		Language: LangFollowOS,
		Theme:    ThemeSystem,
		Proxy: ProxyConfig{
			Mode: ProxySystem,
		},
	}, nil
}

// Load reads app config from the OS config dir. Missing file yields Default().
func Load() (*Config, error) {
	// Persistence (JSON/TOML) lands with the settings UI; defaults are enough for boot.
	return Default()
}

// AppConfigDir returns the OS-conventional path for language/theme/data-dir settings.
func AppConfigDir() (string, error) {
	return appConfigDir()
}
