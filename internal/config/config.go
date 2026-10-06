package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	AppID   int    `json:"app_id"`
	AppHash string `json:"app_hash"`
}

func DefaultPath() string {
	dir, _ := os.UserConfigDir()
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "tele", "config.json")
}

func Load(path string) (Config, error) {
	if path == "" {
		path = DefaultPath()
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil // Return empty config if file doesn't exist
		}
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return Config{}, err
	}
	return c, nil
}

func Save(path string, c Config) error {
	if path == "" {
		path = DefaultPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

func (c Config) IsValid() bool {
	return c.AppID > 0 && c.AppHash != ""
}
