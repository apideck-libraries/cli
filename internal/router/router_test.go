// internal/router/router_test.go
package router

import (
	"testing"

	"github.com/apideck-io/cli/internal/spec"
	"github.com/spf13/cobra"
)

func testSpec() *spec.APISpec {
	return &spec.APISpec{
		Name:    "Apideck",
		Version: "1.0.0",
		BaseURL: "https://unify.apideck.com",
		APIGroups: map[string]*spec.APIGroup{
			"accounting": {
				Name: "accounting",
				Resources: map[string]*spec.Resource{
					"invoices": {
						Name: "invoices",
						Operations: []*spec.Operation{
							{ID: "invoicesAll", Method: "GET", Path: "/accounting/invoices", Summary: "List invoices", Permission: spec.PermissionRead},
							{ID: "invoicesOne", Method: "GET", Path: "/accounting/invoices/{id}", Summary: "Get invoice", Permission: spec.PermissionRead, HasPathID: true},
							{ID: "invoicesAdd", Method: "POST", Path: "/accounting/invoices", Summary: "Create invoice", Permission: spec.PermissionWrite},
							{ID: "invoicesDelete", Method: "DELETE", Path: "/accounting/invoices/{id}", Summary: "Delete invoice", Permission: spec.PermissionDangerous, HasPathID: true},
						},
					},
				},
			},
		},
	}
}

func TestBuildCommands(t *testing.T) {
	rootCmd := &cobra.Command{Use: "apideck"}
	apiSpec := testSpec()
	BuildCommands(rootCmd, apiSpec, nil)

	// Should have "accounting" subcommand
	acctCmd, _, err := rootCmd.Find([]string{"accounting"})
	if err != nil || acctCmd.Use != "accounting" {
		t.Fatal("accounting command not found")
	}

	// Should have verb commands under "accounting invoices"
	for _, verb := range []string{"list", "get", "create", "delete"} {
		cmd, _, err := rootCmd.Find([]string{"accounting", "invoices", verb})
		if err != nil || cmd == nil {
			t.Errorf("accounting invoices %s command not found", verb)
		}
	}
}

func TestBuildCommands_OperationFlags(t *testing.T) {
	rootCmd := &cobra.Command{Use: "apideck"}
	apiSpec := testSpec()
	BuildCommands(rootCmd, apiSpec, nil)

	// "get" command should have --id flag
	getCmd, _, err := rootCmd.Find([]string{"accounting", "invoices", "get"})
	if err != nil || getCmd == nil {
		t.Fatal("get command not found")
	}
	if getCmd.Flags().Lookup("id") == nil {
		t.Error("get command should have --id flag")
	}

	// "create" command should have --data flag
	createCmd, _, err := rootCmd.Find([]string{"accounting", "invoices", "create"})
	if err != nil || createCmd == nil {
		t.Fatal("create command not found")
	}
	if createCmd.Flags().Lookup("data") == nil {
		t.Error("create command should have --data flag")
	}

	// "list" command should NOT have --id flag
	listCmd, _, err := rootCmd.Find([]string{"accounting", "invoices", "list"})
	if err != nil || listCmd == nil {
		t.Fatal("list command not found")
	}
	if listCmd.Flags().Lookup("id") != nil {
		t.Error("list command should not have --id flag")
	}
}

func TestBuildCommands_ExecutorCalled(t *testing.T) {
	rootCmd := &cobra.Command{Use: "apideck"}
	apiSpec := testSpec()

	var calledOp *spec.Operation
	var calledFlags map[string]string

	executor := func(op *spec.Operation, flags map[string]string) error {
		calledOp = op
		calledFlags = flags
		return nil
	}

	BuildCommands(rootCmd, apiSpec, executor)

	rootCmd.SetArgs([]string{"accounting", "invoices", "list"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if calledOp == nil {
		t.Fatal("executor was not called")
	}
	if calledOp.ID != "invoicesAll" {
		t.Errorf("expected invoicesAll, got %s", calledOp.ID)
	}
	_ = calledFlags
}

func TestBuildCommands_GroupListFlag(t *testing.T) {
	rootCmd := &cobra.Command{Use: "apideck"}
	apiSpec := testSpec()
	BuildCommands(rootCmd, apiSpec, nil)

	acctCmd, _, err := rootCmd.Find([]string{"accounting"})
	if err != nil || acctCmd == nil {
		t.Fatal("accounting command not found")
	}
	if acctCmd.Flags().Lookup("list") == nil {
		t.Error("accounting command should have --list flag")
	}
}
