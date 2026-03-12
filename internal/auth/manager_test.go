package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{envAPIKey, envAppID, envConsumerID, envServiceID} {
		t.Setenv(key, "")
	}
}

func TestResolveFromEnv(t *testing.T) {
	t.Setenv(envAPIKey, "env-api-key")
	t.Setenv(envAppID, "env-app-id")
	t.Setenv(envConsumerID, "env-consumer-id")
	t.Setenv(envServiceID, "env-service-id")

	m := NewManager("/nonexistent/config.yaml")
	creds, err := m.Resolve()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if creds.APIKey != "env-api-key" {
		t.Errorf("expected APIKey %q, got %q", "env-api-key", creds.APIKey)
	}
	if creds.AppID != "env-app-id" {
		t.Errorf("expected AppID %q, got %q", "env-app-id", creds.AppID)
	}
	if creds.ConsumerID != "env-consumer-id" {
		t.Errorf("expected ConsumerID %q, got %q", "env-consumer-id", creds.ConsumerID)
	}
	if creds.ServiceID != "env-service-id" {
		t.Errorf("expected ServiceID %q, got %q", "env-service-id", creds.ServiceID)
	}
}

func TestResolveFromConfig(t *testing.T) {
	clearEnv(t)

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	cfg := &FileConfig{
		APIKey:     "file-api-key",
		AppID:      "file-app-id",
		ConsumerID: "file-consumer-id",
		ServiceID:  "file-service-id",
	}
	if err := SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	m := NewManager(configPath)
	creds, err := m.Resolve()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if creds.APIKey != "file-api-key" {
		t.Errorf("expected APIKey %q, got %q", "file-api-key", creds.APIKey)
	}
	if creds.AppID != "file-app-id" {
		t.Errorf("expected AppID %q, got %q", "file-app-id", creds.AppID)
	}
	if creds.ConsumerID != "file-consumer-id" {
		t.Errorf("expected ConsumerID %q, got %q", "file-consumer-id", creds.ConsumerID)
	}
	if creds.ServiceID != "file-service-id" {
		t.Errorf("expected ServiceID %q, got %q", "file-service-id", creds.ServiceID)
	}
}

func TestResolveEnvOverridesConfig(t *testing.T) {
	t.Setenv(envAPIKey, "override-api-key")
	t.Setenv(envAppID, "override-app-id")
	t.Setenv(envConsumerID, "override-consumer-id")
	t.Setenv(envServiceID, "")

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	cfg := &FileConfig{
		APIKey:     "file-api-key",
		AppID:      "file-app-id",
		ConsumerID: "file-consumer-id",
		ServiceID:  "file-service-id",
	}
	if err := SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	m := NewManager(configPath)
	creds, err := m.Resolve()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if creds.APIKey != "override-api-key" {
		t.Errorf("expected APIKey %q, got %q", "override-api-key", creds.APIKey)
	}
	if creds.AppID != "override-app-id" {
		t.Errorf("expected AppID %q, got %q", "override-app-id", creds.AppID)
	}
	if creds.ConsumerID != "override-consumer-id" {
		t.Errorf("expected ConsumerID %q, got %q", "override-consumer-id", creds.ConsumerID)
	}
	// ServiceID env was empty, so config file value should be used
	if creds.ServiceID != "file-service-id" {
		t.Errorf("expected ServiceID %q from config, got %q", "file-service-id", creds.ServiceID)
	}
}

func TestResolveNoCredentials(t *testing.T) {
	clearEnv(t)

	m := NewManager("/nonexistent/path/config.yaml")
	_, err := m.Resolve()
	if err == nil {
		t.Fatal("expected an error when no credentials are available, got nil")
	}
}

func TestCredentialsHeaders(t *testing.T) {
	creds := &Credentials{
		APIKey:     "test-api-key",
		AppID:      "test-app-id",
		ConsumerID: "test-consumer-id",
		ServiceID:  "test-service-id",
	}

	headers := creds.Headers()

	expected := map[string]string{
		"Authorization":         "Bearer test-api-key",
		"x-apideck-app-id":      "test-app-id",
		"x-apideck-consumer-id": "test-consumer-id",
		"x-apideck-service-id":  "test-service-id",
	}

	for key, want := range expected {
		got, ok := headers[key]
		if !ok {
			t.Errorf("missing header %q", key)
			continue
		}
		if got != want {
			t.Errorf("header %q: expected %q, got %q", key, want, got)
		}
	}

	if len(headers) != len(expected) {
		t.Errorf("expected %d headers, got %d", len(expected), len(headers))
	}
}

func TestCredentialsHeadersNoServiceID(t *testing.T) {
	creds := &Credentials{
		APIKey:     "test-api-key",
		AppID:      "test-app-id",
		ConsumerID: "test-consumer-id",
		ServiceID:  "",
	}

	headers := creds.Headers()

	if _, ok := headers["x-apideck-service-id"]; ok {
		t.Error("expected x-apideck-service-id header to be omitted when ServiceID is empty")
	}

	expectedKeys := []string{"Authorization", "x-apideck-app-id", "x-apideck-consumer-id"}
	for _, key := range expectedKeys {
		if _, ok := headers[key]; !ok {
			t.Errorf("missing expected header %q", key)
		}
	}

	if len(headers) != 3 {
		t.Errorf("expected 3 headers, got %d", len(headers))
	}
}

// Ensure clearEnv uses os package (avoids unused import warning).
var _ = os.Getenv
