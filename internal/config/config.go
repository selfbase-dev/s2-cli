package config

import (
	"os"
	"path/filepath"
)

// Dir returns the S2 config directory (~/.config/s2).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "s2"), nil
}

// EnsureDir creates the config directory if it doesn't exist.
func EnsureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}
