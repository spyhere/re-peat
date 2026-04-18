//go:build !windows && !darwin

package configs

import "fmt"

func getSystemLocale() (string, error) {
	return "", fmt.Errorf("Unsupported OS")
}
