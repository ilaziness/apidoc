//go:build !windows

package config

func readOSLocaleNative() string {
	return ""
}
