package permission

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apideck-io/cli/internal/spec"
)

// makeOp is a helper that builds a minimal Operation for testing.
func makeOp(id, method string) *spec.Operation {
	return &spec.Operation{
		ID:         id,
		Method:     method,
		Permission: spec.PermissionLevelFromMethod(method),
	}
}

// TestClassifyDefaults verifies the built-in method→action defaults.
func TestClassifyDefaults(t *testing.T) {
	engine := NewEngine("nonexistent-config-path-that-does-not-exist.yaml")

	tests := []struct {
		name   string
		op     *spec.Operation
		want   Action
	}{
		{"GET allows", makeOp("invoices-list", "GET"), ActionAllow},
		{"HEAD allows", makeOp("invoices-head", "HEAD"), ActionAllow},
		{"OPTIONS allows", makeOp("invoices-options", "OPTIONS"), ActionAllow},
		{"POST prompts", makeOp("invoices-create", "POST"), ActionPrompt},
		{"PUT prompts", makeOp("invoices-update", "PUT"), ActionPrompt},
		{"PATCH prompts", makeOp("invoices-patch", "PATCH"), ActionPrompt},
		{"DELETE blocks", makeOp("invoices-delete", "DELETE"), ActionBlock},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.Classify(tt.op)
			if got != tt.want {
				t.Errorf("Classify(%s %s) = %s; want %s",
					tt.op.Method, tt.op.ID, got, tt.want)
			}
		})
	}
}

// TestClassifyOverrides verifies that per-operation overrides take precedence.
func TestClassifyOverrides(t *testing.T) {
	// Write a temporary permissions.yaml with overrides and custom defaults.
	dir := t.TempDir()
	configPath := filepath.Join(dir, "permissions.yaml")

	content := `
overrides:
  # Upgrade: allow a specific DELETE without prompting.
  invoices-delete-safe: allow
  # Downgrade: block a normally-allowed GET.
  invoices-list-blocked: block
  # Downgrade a POST from prompt to block.
  invoices-create-blocked: block
defaults:
  # Override built-in default for dangerous to prompt instead of block.
  dangerous: prompt
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	engine := NewEngine(configPath)

	tests := []struct {
		name   string
		op     *spec.Operation
		want   Action
	}{
		// Override upgrades DELETE to allow.
		{
			"DELETE with allow override",
			&spec.Operation{ID: "invoices-delete-safe", Method: "DELETE", Permission: spec.PermissionDangerous},
			ActionAllow,
		},
		// Override downgrades GET to block.
		{
			"GET with block override",
			&spec.Operation{ID: "invoices-list-blocked", Method: "GET", Permission: spec.PermissionRead},
			ActionBlock,
		},
		// Override downgrades POST to block.
		{
			"POST with block override",
			&spec.Operation{ID: "invoices-create-blocked", Method: "POST", Permission: spec.PermissionWrite},
			ActionBlock,
		},
		// No override for this DELETE; falls through to config default dangerous→prompt.
		{
			"DELETE with no override uses config default",
			&spec.Operation{ID: "other-delete", Method: "DELETE", Permission: spec.PermissionDangerous},
			ActionPrompt,
		},
		// No override and no config default for read; uses built-in allow.
		{
			"GET with no override uses built-in default",
			makeOp("invoices-list", "GET"),
			ActionAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.Classify(tt.op)
			if got != tt.want {
				t.Errorf("Classify(%s %s) = %s; want %s",
					tt.op.Method, tt.op.ID, got, tt.want)
			}
		})
	}
}

// TestLoadPermConfigMissingFile verifies that a missing config file is not an error.
func TestLoadPermConfigMissingFile(t *testing.T) {
	cfg, err := LoadPermConfig("/nonexistent/path/permissions.yaml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.Overrides) != 0 || len(cfg.Defaults) != 0 {
		t.Errorf("expected empty config, got overrides=%v defaults=%v", cfg.Overrides, cfg.Defaults)
	}
}

// TestWriteDefaultConfig verifies that WriteDefaultConfig creates a valid config
// and is idempotent (does not overwrite an existing file).
func TestWriteDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "permissions.yaml")

	// Creates file and parent directory.
	if err := WriteDefaultConfig(path); err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	cfg, err := LoadPermConfig(path)
	if err != nil {
		t.Fatalf("failed to load generated config: %v", err)
	}
	if cfg.Defaults["read"] != "allow" || cfg.Defaults["write"] != "prompt" || cfg.Defaults["dangerous"] != "block" {
		t.Errorf("unexpected defaults: %v", cfg.Defaults)
	}

	// Overwrite file with custom content, then call again -- should not overwrite.
	os.WriteFile(path, []byte("defaults:\n  read: block\n"), 0600)
	if err := WriteDefaultConfig(path); err != nil {
		t.Fatalf("idempotent call failed: %v", err)
	}
	cfg2, _ := LoadPermConfig(path)
	if cfg2.Defaults["read"] != "block" {
		t.Error("WriteDefaultConfig overwrote existing file")
	}
}

// TestActionString verifies Action.String() returns correct labels.
func TestActionString(t *testing.T) {
	cases := []struct {
		action Action
		want   string
	}{
		{ActionAllow, "allow"},
		{ActionPrompt, "prompt"},
		{ActionBlock, "block"},
	}
	for _, c := range cases {
		if got := c.action.String(); got != c.want {
			t.Errorf("Action(%d).String() = %q; want %q", c.action, got, c.want)
		}
	}
}
