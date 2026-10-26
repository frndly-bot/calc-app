//go:build !android

package config

import "os"

// storageRoot returns the base directory holding application data on desktop
// platforms (~/.config, ~/Library/Application Support, %AppData%).
func storageRoot() (string, error) {
	return os.UserConfigDir()
}
