/*
Copyright © 2025 srz_zumix
*/

// Package store persists JSON input for the --save, --load, --clear and --list options.
package store

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	// DefaultChannel is the reserved channel name that maps to the legacy single-file store.
	DefaultChannel = "default"
	// AllChannels is the sentinel passed to Clear to remove every channel.
	AllChannels = "*"

	defaultFileName = "last-input.json"
	channelDirName  = "channels"
	channelExt      = ".json"
	dirPerm         = 0700
	filePerm        = 0600
)

var channelNamePattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

// userConfigDir is a variable so that tests can redirect the config directory.
var userConfigDir = os.UserConfigDir

// DataDir returns the directory used to store saved JSON data.
// It uses the user's config directory (e.g., ~/.config/gh-jq).
func DataDir() (string, error) {
	configDir, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine config directory: %w", err)
	}
	return filepath.Join(configDir, "gh-jq"), nil
}

// ValidateChannel reports whether name is a usable channel name.
// Names are non-empty and consist only of lowercase letters, digits, '-' and '_'.
// Lowercase-only avoids aliasing on case-insensitive file systems.
func ValidateChannel(name string) error {
	if !channelNamePattern.MatchString(name) {
		return fmt.Errorf("invalid channel name %q: use lowercase letters, digits, '-' or '_'", name)
	}
	return nil
}

// channelPath returns the file path for a validated channel name.
// The DefaultChannel maps to the legacy single-file location.
func channelPath(dir, channel string) string {
	if channel == DefaultChannel {
		return filepath.Join(dir, defaultFileName)
	}
	return filepath.Join(dir, channelDirName, channel+channelExt)
}

// Save writes data to the given channel.
func Save(channel string, data []byte) error {
	if err := ValidateChannel(channel); err != nil {
		return err
	}
	dir, err := DataDir()
	if err != nil {
		return err
	}
	path := channelPath(dir, channel)
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return fmt.Errorf("cannot create data directory: %w", err)
	}
	if err := os.WriteFile(path, data, filePerm); err != nil {
		return fmt.Errorf("cannot save input to channel %q: %w", channel, err)
	}
	return nil
}

// Load reads data from the given channel.
func Load(channel string) ([]byte, error) {
	if err := ValidateChannel(channel); err != nil {
		return nil, err
	}
	dir, err := DataDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(channelPath(dir, channel))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no saved input found for channel %q: %w", channel, err)
		}
		return nil, fmt.Errorf("cannot load input from channel %q: %w", channel, err)
	}
	return data, nil
}

// Clear removes the saved data of the given channel.
// If channel is AllChannels, every channel including the default is removed.
// Missing data is not an error.
func Clear(channel string) error {
	if channel == AllChannels {
		return clearAll()
	}
	if err := ValidateChannel(channel); err != nil {
		return err
	}
	dir, err := DataDir()
	if err != nil {
		return err
	}
	if err := removeIfExists(channelPath(dir, channel)); err != nil {
		return fmt.Errorf("cannot clear channel %q: %w", channel, err)
	}
	return nil
}

// List returns the sorted names of all channels that hold saved data.
// The DefaultChannel is included only when the legacy file exists.
func List() ([]string, error) {
	dir, err := DataDir()
	if err != nil {
		return nil, err
	}
	names := []string{}
	if _, err := os.Stat(filepath.Join(dir, defaultFileName)); err == nil {
		names = append(names, DefaultChannel)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("cannot list channels: %w", err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, channelDirName))
	if err != nil {
		if os.IsNotExist(err) {
			return names, nil
		}
		return nil, fmt.Errorf("cannot list channels: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), channelExt) {
			continue
		}
		name := strings.TrimSuffix(e.Name(), channelExt)
		if name == DefaultChannel || ValidateChannel(name) != nil {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// clearAll removes the legacy file and every valid channel file, leaving other config untouched.
func clearAll() error {
	dir, err := DataDir()
	if err != nil {
		return err
	}
	if err := removeIfExists(filepath.Join(dir, defaultFileName)); err != nil {
		return fmt.Errorf("cannot clear channels: %w", err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, channelDirName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("cannot clear channels: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), channelExt) {
			continue
		}
		if ValidateChannel(strings.TrimSuffix(e.Name(), channelExt)) != nil {
			continue
		}
		if err := removeIfExists(filepath.Join(dir, channelDirName, e.Name())); err != nil {
			return fmt.Errorf("cannot clear channels: %w", err)
		}
	}
	return nil
}

func removeIfExists(path string) error {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
