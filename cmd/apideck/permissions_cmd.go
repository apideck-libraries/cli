package main

import (
	"fmt"

	"github.com/apideck-io/cli/internal/permission"
	"github.com/apideck-io/cli/internal/ui"
	"github.com/spf13/cobra"
)

func newPermissionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "permissions",
		Short: "Show permission configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := permission.DefaultPermConfigPath()
			cfg, err := permission.LoadPermConfig(path)
			if err != nil {
				fmt.Println(ui.Dim.Render("No permissions config found."))
				fmt.Println(ui.Dim.Render(fmt.Sprintf("Create one at: %s", path)))
				fmt.Println()
				fmt.Println("Default permissions:")
				fmt.Println("  read (GET):       auto-approved")
				fmt.Println("  write (POST/PUT): confirmation prompt")
				fmt.Println("  dangerous (DELETE): blocked")
				return nil
			}
			fmt.Println(ui.PrimaryBold.Render("Permission Defaults:"))
			for level, action := range cfg.Defaults {
				fmt.Printf("  %-12s %s\n", level+":", action)
			}
			if len(cfg.Overrides) > 0 {
				fmt.Println()
				fmt.Println(ui.PrimaryBold.Render("Overrides:"))
				for op, level := range cfg.Overrides {
					fmt.Printf("  %-40s %s\n", op, level)
				}
			}
			fmt.Printf("\nConfig: %s\n", ui.Dim.Render(path))
			return nil
		},
	}
}
