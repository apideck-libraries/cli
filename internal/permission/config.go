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

// defaultPermConfigTemplate is the YAML content written by WriteDefaultConfig.
const defaultPermConfigTemplate = `# Apideck CLI permission configuration
# Controls how different API operations are handled.
#
# Permission levels:
#   read      - GET requests
#   write     - POST/PUT/PATCH requests
#   dangerous - DELETE requests
#
# Actions:
#   allow  - execute without prompting
#   prompt - ask for confirmation before executing
#   block  - prevent execution entirely

defaults:
  read: allow
  write: prompt
  dangerous: block

# Per-operation overrides (keyed by operation ID):
# overrides:
#   invoices-delete: allow    # skip confirmation for this specific DELETE
#   invoices-list: block      # block a specific GET
`

// WriteDefaultConfig creates a permissions config file with built-in defaults
// at the given path. It does nothing if the file already exists.
func WriteDefaultConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(defaultPermConfigTemplate), 0600)
}
