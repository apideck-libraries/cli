package main

import (
	"fmt"

	"github.com/apideck-io/cli/internal/agent"
	"github.com/apideck-io/cli/internal/spec"
	"github.com/spf13/cobra"
)

func newAgentPromptCmd(apiSpec *spec.APISpec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent-prompt [api-group]",
		Short: "Output token-optimized prompt for AI agents",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				group, ok := apiSpec.APIGroups[args[0]]
				if !ok {
					return fmt.Errorf("unknown API group: %s", args[0])
				}
				fmt.Println(agent.ScopedPrompt(args[0], group))
			} else {
				fmt.Println(agent.GlobalPrompt(apiSpec))
			}
			return nil
		},
	}
	return cmd
}
