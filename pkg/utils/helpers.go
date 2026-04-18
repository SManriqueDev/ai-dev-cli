package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func ReadFile(path string) (string, error) {
	cleanPath := filepath.Clean(path)

	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", cleanPath, err)
	}
	return string(content), nil
}

func WriteFile(path string, content []byte) error {
	cleanPath := filepath.Clean(path)
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return os.WriteFile(cleanPath, content, 0o600)
}

func FileExists(path string) bool {
	cleanPath := filepath.Clean(path)
	_, err := os.Stat(cleanPath)
	return err == nil
}

func GetFileName(path string) string {
	cleanPath := filepath.Clean(path)
	return filepath.Base(cleanPath)
}

func GetFileExt(path string) string {
	cleanPath := filepath.Clean(path)
	return filepath.Ext(cleanPath)
}
