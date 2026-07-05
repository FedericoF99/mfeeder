package appdata

import (
	"os"
	"path/filepath"
)

const appName = "mfeeder"

func dataDir() (string, error) {
	if dir := os.Getenv("MFEEDER_DATA_DIR"); dir != "" {
		return dir, nil
	}

	if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
		return filepath.Join(dir, appName), nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, appName), nil
}

func ensureDataDir() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}

	if err = os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return dir, nil
}

func dataPath(name string) (string, error) {
	dir, err := ensureDataDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, name), nil
}

func ConfFile() (string, error) {
	return dataPath("mfeeder.conf")
}

func DbFile() (string, error) {
	return dataPath("mfeeder.db")
}

func LogDir() (string, error) {
	return dataPath("logs")
}
