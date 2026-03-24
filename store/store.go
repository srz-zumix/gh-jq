/*
Copyright © 2025 srz_zumix
*/
package store

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultFileName = "last-input.json"
	dirPerm         = 0700
	filePerm        = 0600
)

// DataDir returns the directory used to store saved JSON data.
// It uses the user's config directory (e.g., ~/.config/gh-jq).
func DataDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine config directory: %w", err)
	}
	return filepath.Join(configDir, "gh-jq"), nil
}

// Save writes data to the store.
func Save(data []byte) error {
	dir, err := DataDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return fmt.Errorf("cannot create data directory: %w", err)
	}
	path := filepath.Join(dir, defaultFileName)
	if err := os.WriteFile(path, data, filePerm); err != nil {
		return fmt.Errorf("cannot save input: %w", err)
	}
	return nil
}

// Load reads data from the store.
func Load() ([]byte, error) {
	dir, err := DataDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, defaultFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no saved input found: %w", err)
	}
	return data, nil
}

// Clear removes the saved JSON file from the store.
func Clear() error {
	dir, err := DataDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, defaultFileName)
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot clear saved input: %w", err)
	}
	return nil
}
