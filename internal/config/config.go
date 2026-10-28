// Package config loads application settings that live outside the data directory
// (language, theme, data-dir path, global proxy). See docs/PRD.md §6.8–6.10.
package config

import (
	"os"
	"path/filepath"
)

// Language / theme / proxy mode values persisted in app config.
const (
	// UI languages (PRD §6.9). First launch maps OS locale into one of these and saves it.
	LangEN     = "en"      // English
	LangZHHans = "zh-Hans" // 简体中文
	LangZHHant = "zh-Hant" // 繁體中文

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
	Language string // LangEN | LangZHHans | LangZHHant (always concrete after first launch)
	Theme    string // ThemeLight | ThemeDark | ThemeSystem
	Proxy    ProxyConfig
}

// ProxyConfig is the global HTTP/HTTPS proxy.
type ProxyConfig struct {
	Mode   string // ProxySystem | ProxyOff | ProxyManual
	Manual string // URL when Mode == ProxyManual
}

// Default returns PRD defaults without reading disk.
// Language is left empty so Load/first-run can detect OS locale and persist it (PRD §6.9).
func Default() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &Config{
		DataDir:  filepath.Join(home, ".apidoc"),
		Language: "", // unset until first-run detection writes LangEN | LangZHHans | LangZHHant
		Theme:    ThemeSystem,
		Proxy: ProxyConfig{
			Mode: ProxySystem,
		},
	}, nil
}

// Load reads app config from the OS config dir. Missing file yields Default().
func Load() (*Config, error) {
	// Persistence (JSON/TOML) and first-run language detection land with settings UI.
	return Default()
}

// AppConfigDir returns the OS-conventional path for language/theme/data-dir settings.
func AppConfigDir() (string, error) {
	return appConfigDir()
}
