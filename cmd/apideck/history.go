package main

import (
	"fmt"

	"github.com/apideck-io/cli/internal/history"
	"github.com/apideck-io/cli/internal/ui"
	"github.com/spf13/cobra"
)

func newHistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "history",
		Short: "Show recent API calls",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries := history.Load()
			if len(entries) == 0 {
				fmt.Println(ui.Dim.Render("No API calls recorded yet."))
				return nil
			}
			fmt.Printf("%-22s %-7s %-40s %s  %s\n",
				ui.Dim.Render("Timestamp"),
				ui.Dim.Render("Method"),
				ui.Dim.Render("Path"),
				ui.Dim.Render("Status"),
				ui.Dim.Render("Duration"))
			for _, e := range entries {
				fmt.Printf("%-22s %-7s %-40s %d     %dms\n",
					e.Timestamp[:19], e.Method, e.Path, e.Status, e.DurationMs)
			}
			return nil
		},
	}
}
