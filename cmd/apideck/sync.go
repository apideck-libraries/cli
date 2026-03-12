package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/apideck-io/cli/internal/spec"
	"github.com/apideck-io/cli/internal/ui"
	"github.com/spf13/cobra"
)

const specURL = "https://ci-spec-unify.s3.eu-central-1.amazonaws.com/speakeasy-spec.yml"

func newSyncCmd(cache *spec.Cache) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Download and cache the latest Apideck OpenAPI spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(ui.StepProgress(false, "Downloading spec from S3..."))
			resp, err := http.Get(specURL)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			fmt.Println(ui.StepProgress(true, "Downloaded"))
			fmt.Println(ui.StepProgress(false, "Parsing..."))
			apiSpec, err := spec.ParseSpec(data)
			if err != nil {
				return err
			}
			fmt.Println(ui.StepProgress(true, "Parsed"))
			fmt.Println(ui.StepProgress(false, "Caching..."))
			if err := cache.Save(apiSpec, data); err != nil {
				return err
			}
			fmt.Println(ui.StepProgress(true, "Cached"))
			fmt.Println(ui.SuccessMsg(fmt.Sprintf("Spec synced: %s v%s (%d API groups)", apiSpec.Name, apiSpec.Version, len(apiSpec.APIGroups))))
			return nil
		},
	}
}
