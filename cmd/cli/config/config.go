package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Token string `json:"token"`
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".blackwater")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

func SaveToken(token string) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	cfg := Config{Token: token}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func LoadToken() (string, error) {
	path, err := getConfigPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No token yet
		}
		return "", err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", err
	}

	return cfg.Token, nil
}
