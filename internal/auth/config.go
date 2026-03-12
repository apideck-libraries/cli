package auth

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type FileConfig struct {
	APIKey     string `yaml:"api_key"`
	AppID      string `yaml:"app_id"`
	ConsumerID string `yaml:"consumer_id"`
	ServiceID  string `yaml:"service_id,omitempty"`
}

func LoadConfig(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg FileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(path string, cfg *FileConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path) // CORRECT — not string slicing
	os.MkdirAll(dir, 0755)
	return os.WriteFile(path, data, 0600)
}
