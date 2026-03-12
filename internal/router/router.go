// internal/router/router.go
package router

import (
	"fmt"
	"strings"

	"github.com/apideck-io/cli/internal/spec"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ExecutorFunc is called when a dynamic command is run.
type ExecutorFunc func(op *spec.Operation, flags map[string]string) error

// BuildCommands generates Cobra subcommands from the API spec.
func BuildCommands(root *cobra.Command, apiSpec *spec.APISpec, executor ExecutorFunc) {
	for _, group := range apiSpec.APIGroups {
		groupCmd := &cobra.Command{
			Use:   group.Name,
			Short: fmt.Sprintf("Interact with Apideck %s API", group.Name),
		}

		// Capture for closure
		g := group

		// --list flag for the group
		var listResources bool
		groupCmd.Flags().BoolVar(&listResources, "list", false, "List all resources and operations")
		groupCmd.RunE = func(cmd *cobra.Command, args []string) error {
			if listResources {
				printResourceList(cmd, g)
				return nil
			}
			return cmd.Help()
		}

		for _, resource := range group.Resources {
			resourceCmd := &cobra.Command{
				Use:   resource.Name,
				Short: fmt.Sprintf("Manage %s", resource.Name),
			}

			for _, op := range resource.Operations {
				opCmd := buildOperationCommand(op, executor)
				resourceCmd.AddCommand(opCmd)
			}

			groupCmd.AddCommand(resourceCmd)
		}

		root.AddCommand(groupCmd)
	}
}

// buildOperationCommand creates a Cobra command for a single operation.
func buildOperationCommand(op *spec.Operation, executor ExecutorFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   op.Verb(),
		Short: op.Summary,
	}

	// Register flags from parameters, skipping header params
	for _, param := range op.Parameters {
		if param.In == "header" {
			continue
		}
		// Skip path params — they are handled via --id
		if param.In == "path" {
			continue
		}
		desc := param.Description
		if len(param.Enum) > 0 {
			desc = fmt.Sprintf("%s (one of: %s)", desc, strings.Join(param.Enum, ", "))
		}
		cmd.Flags().String(param.Name, "", desc)
	}

	// Add --id flag if the operation has a path ID
	if op.HasPathID {
		cmd.Flags().String("id", "", "Resource ID")
	}

	// Add --data flag for mutating methods
	method := strings.ToUpper(op.Method)
	if method == "POST" || method == "PUT" || method == "PATCH" {
		cmd.Flags().String("data", "", "Request body as JSON string")
	}

	// Capture op for closure
	operation := op

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if executor == nil {
			return fmt.Errorf("no executor configured")
		}

		flags := make(map[string]string)
		cmd.Flags().Visit(func(f *pflag.Flag) {
			flags[f.Name] = f.Value.String()
		})

		return executor(operation, flags)
	}

	return cmd
}

// printResourceList prints all resources and their available operations to stdout.
func printResourceList(cmd *cobra.Command, group *spec.APIGroup) {
	fmt.Fprintf(cmd.OutOrStdout(), "Resources in %s API:\n\n", group.Name)
	for _, resource := range group.Resources {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", resource.Name)
		for _, op := range resource.Operations {
			fmt.Fprintf(cmd.OutOrStdout(), "    %-10s %s %s\n", op.Verb(), op.Method, op.Path)
		}
	}
}
