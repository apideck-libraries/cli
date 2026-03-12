package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apideck-io/cli/internal/agent"
	"github.com/apideck-io/cli/internal/spec"
	"github.com/apideck-io/cli/internal/ui"
	"github.com/spf13/cobra"
)

func newSkillCmd(apiSpec *spec.APISpec) *cobra.Command {
	cmd := &cobra.Command{Use: "skill", Short: "Manage AI agent skills"}
	cmd.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Install Claude Code skill to ~/.claude/skills/",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			dir := filepath.Join(home, ".claude", "skills")
			os.MkdirAll(dir, 0755)
			path := filepath.Join(dir, "apideck.md")

			content := fmt.Sprintf("---\ndescription: Interact with Apideck Unified API via CLI\n---\n\n%s\n",
				agent.GlobalPrompt(apiSpec))

			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return err
			}
			fmt.Println(ui.SuccessMsg(fmt.Sprintf("Skill installed to %s", path)))
			return nil
		},
	})
	return cmd
}
