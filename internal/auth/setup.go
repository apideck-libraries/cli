package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apideck-io/cli/internal/ui"
	"github.com/charmbracelet/huh"
)

// RunSetup launches an interactive form to collect Apideck credentials and
// persist them to the given config file path.
func RunSetup(configPath string) error {
	var apiKey, appID, consumerID, serviceID string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("API Key").
				Description("Your Apideck API key").
				Value(&apiKey).
				EchoMode(huh.EchoModePassword),
			huh.NewInput().
				Title("App ID").
				Description("Your Apideck application ID").
				Value(&appID),
			huh.NewInput().
				Title("Consumer ID").
				Description("The consumer ID to use").
				Value(&consumerID),
			huh.NewInput().
				Title("Service ID (optional)").
				Description("Default connector (e.g., quickbooks, xero)").
				Value(&serviceID),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("setup cancelled: %w", err)
	}

	cfg := &FileConfig{
		APIKey:     apiKey,
		AppID:      appID,
		ConsumerID: consumerID,
		ServiceID:  serviceID,
	}

	if err := SaveConfig(configPath, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Println(ui.SuccessMsg(fmt.Sprintf("Credentials saved to %s", configPath)))
	return nil
}

// RunStatus prints the current credential resolution status, showing where
// each credential is sourced from (environment variable or config file).
func RunStatus(configPath string) {
	mgr := NewManager(configPath)
	creds, err := mgr.Resolve()

	fmt.Println(ui.Bold.Render("Credential Status"))
	fmt.Println()

	fmt.Printf("  API Key:     %s\n", checkSource(envAPIKey, configPath, "api_key"))
	fmt.Printf("  App ID:      %s\n", checkSource(envAppID, configPath, "app_id"))
	fmt.Printf("  Consumer ID: %s\n", checkSource(envConsumerID, configPath, "consumer_id"))
	fmt.Printf("  Service ID:  %s\n", checkSource(envServiceID, configPath, "service_id"))
	fmt.Println()

	if err != nil {
		fmt.Println(ui.ErrorMsg(fmt.Sprintf("Credentials not fully resolved: %s", err)))
	} else {
		fmt.Println(ui.SuccessMsg("All required credentials resolved"))
		if creds.ServiceID != "" {
			fmt.Printf("  Default service: %s\n", ui.Primary.Render(creds.ServiceID))
		}
	}
}

// checkSource returns a human-readable string describing where a credential
// value is sourced from: environment variable, config file, or missing.
func checkSource(envVar, configPath, configKey string) string {
	if v := os.Getenv(envVar); v != "" {
		masked := maskValue(v)
		return ui.Success.Render(fmt.Sprintf("%s (from env %s)", masked, envVar))
	}

	cfg, err := LoadConfig(expandPath(configPath))
	if err == nil {
		var val string
		switch configKey {
		case "api_key":
			val = cfg.APIKey
		case "app_id":
			val = cfg.AppID
		case "consumer_id":
			val = cfg.ConsumerID
		case "service_id":
			val = cfg.ServiceID
		}
		if val != "" {
			masked := maskValue(val)
			return ui.Primary.Render(fmt.Sprintf("%s (from config)", masked))
		}
	}

	if configKey == "service_id" {
		return ui.Dim.Render("not set (optional)")
	}
	return ui.Error.Render("not set")
}

// maskValue masks a credential value for display, showing only the first and
// last two characters.
func maskValue(v string) string {
	if len(v) <= 6 {
		return "****"
	}
	return v[:2] + "****" + v[len(v)-2:]
}

// expandPath expands a leading ~ in a file path to the user's home directory.
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}
