package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDataDir(t *testing.T) {
	oldDir := t.TempDir()
	newDir := t.TempDir()
	sub := filepath.Join(oldDir, "nested")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "a.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CopyDataDir(oldDir, newDir); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(newDir, "nested", "a.txt"))
	if err != nil || string(got) != "hi" {
		t.Fatalf("got=%q err=%v", got, err)
	}
}
