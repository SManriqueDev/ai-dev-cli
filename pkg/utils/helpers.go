package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return string(content), nil
}

func WriteFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return os.WriteFile(path, content, 0644)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func GetFileName(path string) string {
	return filepath.Base(path)
}

func GetFileExt(path string) string {
	return filepath.Ext(path)
}
