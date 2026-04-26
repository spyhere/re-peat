package configs

import (
	"encoding/json"
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

func Load() (*Configs, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Configs{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var configs Configs
	if err = json.NewDecoder(f).Decode(&configs); err != nil {
		return nil, err
	}
	return &configs, nil
}

type Configs struct {
	Lang string `json:"lang"`
}

func (c *Configs) Save() error {
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

	return json.NewEncoder(f).Encode(c)
}

func (c *Configs) GetLocale() (string, error) {
	if c.Lang == "" {
		return getSystemLocale()
	}
	return c.Lang, nil
}
