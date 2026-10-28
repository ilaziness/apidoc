package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyDataDir copies all files from oldDir into newDir (overwrite on conflict).
// Does not delete oldDir. Caller persists DataDir only after success (PRD §6.8).
func CopyDataDir(oldDir, newDir string) error {
	if oldDir == "" || newDir == "" {
		return fmt.Errorf("config: empty data dir")
	}
	absOld, err := filepath.Abs(oldDir)
	if err != nil {
		return err
	}
	absNew, err := filepath.Abs(newDir)
	if err != nil {
		return err
	}
	if absOld == absNew {
		return nil
	}
	if err := os.MkdirAll(absNew, 0o755); err != nil {
		return err
	}
	return filepath.Walk(absOld, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(absOld, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(absNew, rel)
		if info.IsDir() {
			return os.MkdirAll(dst, info.Mode())
		}
		return copyFile(path, dst, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
