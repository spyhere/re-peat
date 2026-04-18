//go:build darwin

package configs

import (
	"fmt"
	"os/exec"
	"strings"
)

func getSystemLocale() (string, error) {
	out, err := exec.Command("defaults", "read", "-g", "AppleLanguages").Output()
	if err != nil {
		return "", err
	}
	languages := strings.TrimSpace(string(out))
	splitted := strings.Split(languages, ",\n")
	if len(splitted) == 0 {
		return "", fmt.Errorf("No preffered languages discovered")
	}
	fLocale := splitted[0]
	idx := strings.Index(fLocale, "\"")
	if idx == -1 {
		return "", fmt.Errorf("Incorrect system locale format %v", fLocale)
	}
	fLocale = fLocale[idx+1 : idx+6]
	return fLocale, nil
}
