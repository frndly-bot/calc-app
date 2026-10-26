//go:build android

package config

import (
	"errors"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// storageRoot returns the app's private files directory on Android. The JVM
// process has no HOME, so os.UserConfigDir cannot resolve; the Wails bridge
// answers with getFilesDir. An empty path means the bridge was not ready yet.
func storageRoot() (string, error) {
	dir := application.Mobile.StoragePath()
	if dir == "" {
		return "", errors.New("android storage path unavailable (bridge not initialized)")
	}
	return dir, nil
}
