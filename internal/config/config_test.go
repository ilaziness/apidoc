package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	// Force AppConfigDir on Windows to use APPDATA; on Unix override via XDG.
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	cfg.Language = LangZHHans
	cfg.Theme = ThemeDark
	cfg.Proxy.Mode = ProxyOff
	cfg.DataDir = filepath.Join(dir, "data")
	cfg.LastCollectionID = 7

	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Language != LangZHHans {
		t.Fatalf("language=%q", got.Language)
	}
	if got.Theme != ThemeDark {
		t.Fatalf("theme=%q", got.Theme)
	}
	if got.Proxy.Mode != ProxyOff {
		t.Fatalf("proxy=%q", got.Proxy.Mode)
	}
	if got.LastCollectionID != 7 {
		t.Fatalf("lastCollection=%d", got.LastCollectionID)
	}
	if _, err := os.Stat(filepath.Join(appConfigDirMust(t), configFileName)); err != nil {
		t.Fatal(err)
	}
}

func appConfigDirMust(t *testing.T) string {
	t.Helper()
	d, err := AppConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestValidLanguage(t *testing.T) {
	if !ValidLanguage(LangEN) || !ValidLanguage(LangZHHans) || !ValidLanguage(LangZHHant) {
		t.Fatal("expected valid")
	}
	if ValidLanguage("zh") || ValidLanguage("") {
		t.Fatal("expected invalid")
	}
}
