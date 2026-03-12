package permission

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// PermConfig holds default and override action mappings for operations.
type PermConfig struct {
	Defaults  map[string]string `yaml:"defaults"`
	Overrides map[string]string `yaml:"overrides"`
}

// LoadPermConfig reads a permissions config from the given path.
// If the file does not exist, it returns an empty config without error.
func LoadPermConfig(path string) (*PermConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &PermConfig{
				Defaults:  map[string]string{},
				Overrides: map[string]string{},
			}, nil
		}
		return nil, err
	}

	var cfg PermConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Defaults == nil {
		cfg.Defaults = map[string]string{}
	}
	if cfg.Overrides == nil {
		cfg.Overrides = map[string]string{}
	}

	return &cfg, nil
}

// DefaultPermConfigPath returns the default path for the permissions config file.
func DefaultPermConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".apideck-cli", "permissions.yaml")
}
