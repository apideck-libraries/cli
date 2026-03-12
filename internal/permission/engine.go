package permission

import "github.com/apideck-io/cli/internal/spec"

// Action represents the permission action for an operation.
type Action int

const (
	// ActionAllow permits the operation without prompting (read-only ops).
	ActionAllow Action = iota
	// ActionPrompt requires user confirmation before proceeding (write ops).
	ActionPrompt
	// ActionBlock prevents the operation entirely (dangerous ops like DELETE).
	ActionBlock
)

// String returns a human-readable label for the action.
func (a Action) String() string {
	switch a {
	case ActionAllow:
		return "allow"
	case ActionPrompt:
		return "prompt"
	case ActionBlock:
		return "block"
	default:
		return "unknown"
	}
}

// actionFromString converts a string label to an Action value.
// Returns ActionPrompt as a safe default for unrecognised values.
func actionFromString(s string) (Action, bool) {
	switch s {
	case "allow":
		return ActionAllow, true
	case "prompt":
		return ActionPrompt, true
	case "block":
		return ActionBlock, true
	default:
		return ActionPrompt, false
	}
}

// Engine classifies operations into permission actions using configurable rules.
type Engine struct {
	config *PermConfig
}

// NewEngine creates an Engine, loading config from configPath.
// If configPath is empty, DefaultPermConfigPath is used.
// Config load errors are silently ignored and result in a default-only engine.
func NewEngine(configPath string) *Engine {
	if configPath == "" {
		configPath = DefaultPermConfigPath()
	}

	cfg, err := LoadPermConfig(configPath)
	if err != nil {
		cfg = &PermConfig{
			Defaults:  map[string]string{},
			Overrides: map[string]string{},
		}
	}

	return &Engine{config: cfg}
}

// Classify returns the Action for the given operation.
// Overrides (keyed by op.ID) are checked first; then the config defaults
// (keyed by permission level name); finally the built-in defaults apply.
func (e *Engine) Classify(op *spec.Operation) Action {
	// 1. Check per-operation overrides by operation ID.
	if raw, ok := e.config.Overrides[op.ID]; ok {
		if action, valid := actionFromString(raw); valid {
			return action
		}
	}

	// 2. Check config-level defaults keyed by permission level name.
	levelKey := op.Permission.String()
	if raw, ok := e.config.Defaults[levelKey]; ok {
		if action, valid := actionFromString(raw); valid {
			return action
		}
	}

	// 3. Built-in defaults: read→allow, write→prompt, dangerous→block.
	switch op.Permission {
	case spec.PermissionRead:
		return ActionAllow
	case spec.PermissionWrite:
		return ActionPrompt
	case spec.PermissionDangerous:
		return ActionBlock
	default:
		return ActionPrompt
	}
}
