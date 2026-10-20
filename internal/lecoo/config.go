package lecoo

import (
	"os"
	"path/filepath"
)

var configDirOverride string

func configPath(name string) (string, error) {
	if configDirOverride != "" {
		return filepath.Join(configDirOverride, name), nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "lecoo-control-gui", name), nil
}
