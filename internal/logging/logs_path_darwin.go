//go:build darwin

package logging

import (
	"os"
	"path/filepath"
)

func GetLogsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Logs", appName), nil
}
