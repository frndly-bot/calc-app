// Package config resolves per-user storage locations and provides JSON
// load/save helpers for application state (history, settings).
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Dir returns the directory where calc-app stores its data, creating it if
// needed. XDG_CONFIG_HOME takes precedence, otherwise the platform storage
// root is used (~/.config, ~/Library/Application Support, %AppData%, or the
// app's private files directory on Android).
func Dir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return ensureDir(filepath.Join(xdg, "calc-app"))
	}
	base, err := storageRoot()
	if err != nil {
		return "", err
	}
	return ensureDir(filepath.Join(base, "calc-app"))
}

// Filepath joins name onto the config directory.
func Filepath(name string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

// LoadJSON reads path into v. A missing file is not an error; v is left at
// its zero value and found is false.
func LoadJSON(path string, v any) (found bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return false, err
	}
	return true, nil
}

// SaveJSON atomically writes v to path as indented JSON.
func SaveJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func ensureDir(path string) (string, error) {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", err
	}
	return path, nil
}
