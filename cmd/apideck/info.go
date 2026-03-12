package main

import (
	"fmt"

	"github.com/apideck-io/cli/internal/spec"
	"github.com/apideck-io/cli/internal/ui"
	"github.com/spf13/cobra"
)

func newInfoCmd(apiSpec *spec.APISpec, cache *spec.Cache) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show API spec version and cache status",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(ui.PrimaryBold.Render("Apideck CLI"))
			fmt.Printf("  Version:     %s\n", version)
			fmt.Printf("  Spec:        %s v%s\n", apiSpec.Name, apiSpec.Version)
			fmt.Printf("  API Groups:  %d\n", len(apiSpec.APIGroups))
			totalOps := 0
			for _, g := range apiSpec.APIGroups {
				for _, r := range g.Resources {
					totalOps += len(r.Operations)
				}
			}
			fmt.Printf("  Operations:  %d\n", totalOps)
			if cache.IsFresh() {
				meta, err := cache.LoadMeta()
				if err == nil {
					fmt.Printf("  Cache:       fresh (fetched %s)\n", meta.FetchedAt.Format("2006-01-02 15:04"))
				}
			} else {
				fmt.Printf("  Cache:       %s\n", ui.Dim.Render("stale (using embedded spec)"))
			}
		},
	}
}
