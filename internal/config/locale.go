package config

import (
	"os"
	"strings"
)

// DetectLanguage maps an OS locale tag to a supported UI language (PRD §6.9).
func DetectLanguage(locale string) string {
	s := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(locale, "_", "-")))
	if s == "" {
		return LangEN
	}
	// Strip encoding suffixes: zh-cn.utf-8
	if i := strings.IndexByte(s, '.'); i >= 0 {
		s = s[:i]
	}
	if i := strings.IndexByte(s, '@'); i >= 0 {
		s = s[:i]
	}

	switch {
	case strings.HasPrefix(s, "zh-hant"),
		strings.HasPrefix(s, "zh-tw"),
		strings.HasPrefix(s, "zh-hk"),
		strings.HasPrefix(s, "zh-mo"):
		return LangZHHant
	case strings.HasPrefix(s, "zh-hans"),
		strings.HasPrefix(s, "zh-cn"),
		strings.HasPrefix(s, "zh-sg"),
		s == "zh":
		return LangZHHans
	case strings.HasPrefix(s, "zh"):
		// Other zh-* without known traditional region → simplified (PRD: bare zh → Hans).
		return LangZHHans
	default:
		return LangEN
	}
}

// ReadOSLocale returns a best-effort locale string from the environment / OS.
func ReadOSLocale() string {
	if v := firstNonEmpty(os.Getenv("LC_ALL"), os.Getenv("LC_MESSAGES"), os.Getenv("LANG")); v != "" {
		return v
	}
	return readOSLocaleNative()
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
