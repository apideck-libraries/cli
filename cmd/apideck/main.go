package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/apideck-io/cli/internal/apiclient"
	"github.com/apideck-io/cli/internal/auth"
	"github.com/apideck-io/cli/internal/output"
	"github.com/apideck-io/cli/internal/permission"
	"github.com/apideck-io/cli/internal/router"
	"github.com/apideck-io/cli/internal/spec"
	"github.com/apideck-io/cli/internal/tui"
	"github.com/apideck-io/cli/internal/ui"
	"github.com/apideck-io/cli/specs"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "apideck",
		Short: "Beautiful, agent-friendly CLI for the Apideck Unified API",
		Long:  "apideck turns the Apideck Unified API into a beautiful, secure, AI-agent-friendly command-line experience.",
	}
	rootCmd.Version = version

	// Global flags
	var outputFormat, fieldsFlag, serviceIDFlag string
	var quietFlag, rawFlag, yesFlag, forceFlag bool
	var timeoutFlag int

	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format: json|table|yaml|csv")
	rootCmd.PersistentFlags().StringVar(&fieldsFlag, "fields", "", "Comma-separated list of fields to display")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Suppress non-data output")
	rootCmd.PersistentFlags().BoolVar(&rawFlag, "raw", false, "Output raw API response")
	rootCmd.PersistentFlags().StringVar(&serviceIDFlag, "service-id", "", "Target a specific connector")
	rootCmd.PersistentFlags().BoolVar(&yesFlag, "yes", false, "Skip write confirmation prompts")
	rootCmd.PersistentFlags().BoolVar(&forceFlag, "force", false, "Override dangerous operation blocks")
	rootCmd.PersistentFlags().IntVar(&timeoutFlag, "timeout", 30, "Request timeout in seconds")

	// Suppress unused variable warning for quietFlag — it may be used later
	_ = quietFlag

	// Load spec: cache → embedded
	cache := spec.NewCache(spec.DefaultCacheDir())
	var apiSpec *spec.APISpec
	var loadErr error

	if cache.IsFresh() {
		apiSpec, loadErr = cache.Load()
		if loadErr != nil {
			apiSpec = nil
		}
	}
	if apiSpec == nil {
		apiSpec, loadErr = spec.ParseSpec(specs.EmbeddedSpec)
		if loadErr != nil {
			fmt.Fprintln(os.Stderr, ui.ErrorBox("Failed to parse OpenAPI spec", loadErr.Error(), "Try: apideck sync"))
			os.Exit(1)
		}
	}

	// --list flag on root
	var listAPIs bool
	rootCmd.Flags().BoolVar(&listAPIs, "list", false, "List available API groups")
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if listAPIs {
			for name := range apiSpec.APIGroups {
				fmt.Println(name)
			}
			return nil
		}
		return cmd.Help()
	}

	// Auth + Permission
	authMgr := auth.NewManager("")
	permEngine := permission.NewEngine(permission.DefaultPermConfigPath())

	// Build executor
	executor := func(op *spec.Operation, flags map[string]string) error {
		// Permission check
		action := permEngine.Classify(op)
		switch action {
		case permission.ActionPrompt:
			if !yesFlag {
				if !ui.IsTTY() {
					return fmt.Errorf("write operation %s requires --yes flag in non-interactive mode", op.ID)
				}
				var confirmed bool
				err := huh.NewConfirm().
					Title(fmt.Sprintf("%s %s", op.Method, op.Path)).
					Description("This write operation requires confirmation.").
					Affirmative("Yes").
					Negative("No").
					Value(&confirmed).
					Run()
				if err != nil || !confirmed {
					return fmt.Errorf("operation cancelled")
				}
			}
		case permission.ActionBlock:
			if !forceFlag {
				fmt.Fprintln(os.Stderr, ui.ErrorBox(
					fmt.Sprintf("Blocked: %s %s", op.Method, op.Path),
					"This operation is classified as dangerous.",
					"Use --force to override, or: apideck permissions",
				))
				os.Exit(1)
			}
		}

		// Resolve auth
		creds, err := authMgr.Resolve()
		if err != nil {
			return fmt.Errorf("auth: %w", err)
		}

		// Apply --service-id override
		if serviceIDFlag != "" {
			creds.ServiceID = serviceIDFlag
		}

		// Create HTTP client
		client := apiclient.NewClient(apiclient.ClientConfig{
			BaseURL:     apiSpec.BaseURL,
			Headers:     creds.Headers(),
			TimeoutSecs: timeoutFlag,
		})

		// Build query params from flags
		queryParams := url.Values{}
		var body any
		for k, v := range flags {
			if k == "__data" || k == "data" {
				rawData := v
				if strings.HasPrefix(rawData, "@") {
					fileData, err := os.ReadFile(rawData[1:])
					if err != nil {
						return fmt.Errorf("read data file %s: %w", rawData[1:], err)
					}
					rawData = string(fileData)
				}
				var parsed any
				if err := json.Unmarshal([]byte(rawData), &parsed); err != nil {
					return fmt.Errorf("invalid JSON body: %w", err)
				}
				body = parsed
				continue
			}
			if k == "id" {
				continue
			}
			queryParams.Set(k, v)
		}

		// Build path with ID substitution
		path := op.Path
		if id, ok := flags["id"]; ok {
			path = strings.Replace(path, "{id}", id, 1)
		}

		// Execute
		resp, err := client.Do(op.Method, path, queryParams, body)
		if err != nil {
			return err
		}

		// Handle error responses
		if !resp.Success && resp.Error != nil {
			return fmt.Errorf("%s (status %d)", resp.Error.Message, resp.StatusCode)
		}

		// Determine output format
		format := outputFormat
		if format == "" {
			if ui.IsTTY() {
				format = "table"
			} else {
				format = "json"
			}
		}

		if rawFlag {
			os.Stdout.Write(resp.RawBody)
			fmt.Println()
			return nil
		}

		// Parse fields
		var fields []string
		if fieldsFlag != "" {
			fields = strings.Split(fieldsFlag, ",")
		}

		formatter := output.NewFormatter(format, os.Stdout, fields)
		return formatter.Format(resp)
	}

	// Build dynamic commands from spec
	router.BuildCommands(rootCmd, apiSpec, executor)

	// Static commands
	rootCmd.AddCommand(newSyncCmd(cache))
	rootCmd.AddCommand(newAgentPromptCmd(apiSpec))
	rootCmd.AddCommand(newInfoCmd(apiSpec, cache))

	// Auth commands
	configPath := authMgr.ConfigPath
	authCmd := &cobra.Command{Use: "auth", Short: "Manage authentication credentials"}
	authCmd.AddCommand(&cobra.Command{
		Use:   "setup",
		Short: "Interactive credential setup wizard",
		RunE: func(cmd *cobra.Command, args []string) error {
			return auth.RunSetup(configPath)
		},
	})
	authCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show credential sources",
		Run: func(cmd *cobra.Command, args []string) {
			auth.RunStatus(configPath)
		},
	})
	rootCmd.AddCommand(authCmd)

	// Explore command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "explore",
		Short: "Launch interactive TUI API explorer",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Run(apiSpec, "")
		},
	})

	// Completion command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletion(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
