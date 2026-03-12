package auth

import (
	"fmt"
	"os"
)

const (
	envAPIKey     = "APIDECK_API_KEY"
	envAppID      = "APIDECK_APP_ID"
	envConsumerID = "APIDECK_CONSUMER_ID"
	envServiceID  = "APIDECK_SERVICE_ID"

	DefaultConfigPath = "~/.apideck-cli/config.yaml"
)

// Credentials holds the resolved Apideck authentication credentials.
type Credentials struct {
	APIKey     string
	AppID      string
	ConsumerID string
	ServiceID  string // optional
}

// Headers returns the HTTP headers required for Apideck API calls.
func (c *Credentials) Headers() map[string]string {
	headers := map[string]string{
		"Authorization":       "Bearer " + c.APIKey,
		"x-apideck-app-id":    c.AppID,
		"x-apideck-consumer-id": c.ConsumerID,
	}
	if c.ServiceID != "" {
		headers["x-apideck-service-id"] = c.ServiceID
	}
	return headers
}

// Manager resolves Apideck credentials from environment variables or a config file.
type Manager struct {
	ConfigPath string
}

// NewManager creates a Manager using the given config file path.
// If the path is empty, DefaultConfigPath is used.
// A leading ~ is expanded to the user's home directory.
func NewManager(configPath string) *Manager {
	if configPath == "" {
		configPath = DefaultConfigPath
	}
	return &Manager{ConfigPath: expandPath(configPath)}
}

// Resolve returns credentials using the resolution chain: env vars > config file > error.
// APIKey, AppID, and ConsumerID are required. ServiceID is optional.
func (m *Manager) Resolve() (*Credentials, error) {
	creds := &Credentials{}

	// Read environment variables first.
	creds.APIKey = os.Getenv(envAPIKey)
	creds.AppID = os.Getenv(envAppID)
	creds.ConsumerID = os.Getenv(envConsumerID)
	creds.ServiceID = os.Getenv(envServiceID)

	// Determine whether we need to consult the config file.
	// We always try the config file so that optional fields (ServiceID) can be
	// filled in even when required fields are already resolved from env vars.
	needsConfig := creds.APIKey == "" || creds.AppID == "" || creds.ConsumerID == "" || creds.ServiceID == ""

	if needsConfig {
		cfg, err := LoadConfig(m.ConfigPath)
		if err != nil {
			// Only treat a missing/unreadable config as fatal when required
			// fields are also missing from the environment.
			if creds.APIKey == "" || creds.AppID == "" || creds.ConsumerID == "" {
				return nil, fmt.Errorf("credentials not found in environment variables and failed to load config file %q: %w", m.ConfigPath, err)
			}
			// Required fields are present via env; ignore config load error.
		} else {
			// Merge: env vars take precedence over config file values.
			if creds.APIKey == "" {
				creds.APIKey = cfg.APIKey
			}
			if creds.AppID == "" {
				creds.AppID = cfg.AppID
			}
			if creds.ConsumerID == "" {
				creds.ConsumerID = cfg.ConsumerID
			}
			if creds.ServiceID == "" {
				creds.ServiceID = cfg.ServiceID
			}
		}
	}

	// Validate required fields.
	if creds.APIKey == "" {
		return nil, fmt.Errorf("missing required credential: api_key (set %s env var or add to config file)", envAPIKey)
	}
	if creds.AppID == "" {
		return nil, fmt.Errorf("missing required credential: app_id (set %s env var or add to config file)", envAppID)
	}
	if creds.ConsumerID == "" {
		return nil, fmt.Errorf("missing required credential: consumer_id (set %s env var or add to config file)", envConsumerID)
	}

	return creds, nil
}
