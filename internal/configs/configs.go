package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const fileName = "configs.json"

func getConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "re-peat", fileName), nil
}

type configs struct {
	Lang string `json:"lang"`
}

func loadI18nPreference() (string, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return "", err
	}
	f, err := os.Open(configPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	var c configs
	if err = decoder.Decode(&c); err != nil {
		return "", err
	}
	if c.Lang == "" {
		return "", fmt.Errorf("Configs file is being corrupted, defaulting.")
	}
	return c.Lang, nil
}

func GetLocale() (string, error) {
	lang, err := loadI18nPreference()
	if err != nil {
		return getSystemLocale()
	}
	return lang, nil
}

func SaveLocale(lang string) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(configPath)
	if err = os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	return encoder.Encode(configs{Lang: lang})
}
