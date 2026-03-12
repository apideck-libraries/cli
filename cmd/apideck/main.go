package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "apideck",
		Short: "Beautiful, agent-friendly CLI for the Apideck Unified API",
		Long:  "apideck turns the Apideck Unified API into a beautiful, secure, AI-agent-friendly command-line experience.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.Version = version

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
