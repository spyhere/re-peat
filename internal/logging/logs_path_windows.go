//go:build windows

package logging

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetLogsPath() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return "", fmt.Errorf("LOCALAPPDATA is not set")
	}
	return filepath.Join(base, appName, "Logs"), nil
}
