//go:build windows

package config

import (
	"os"
	"path/filepath"
)

func appConfigDir() (string, error) {
	base := os.Getenv("APPDATA")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, "AppData", "Roaming")
	}
	return filepath.Join(base, "apidoc"), nil
}
