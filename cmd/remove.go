package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/togglemedia/cliq/internal/config"
	"github.com/togglemedia/cliq/internal/skills"
)

var removeCmd = &cobra.Command{
	Use:   "remove <tool>",
	Short: "Remove a CLI tool from your workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		toolName := args[0]

		lf, err := config.LoadLockfile()
		if err != nil {
			return err
		}

		if _, ok := lf.Tools[toolName]; !ok {
			return fmt.Errorf("%s is not in your cliq workspace", toolName)
		}

		// Remove skill file
		if err := skills.Remove(toolName); err != nil {
			fmt.Printf("⚠ Could not remove skill: %s\n", err)
		} else {
			fmt.Printf("✓ Removed skill for %s\n", toolName)
		}

		// Remove from lockfile
		delete(lf.Tools, toolName)
		if err := lf.Save(); err != nil {
			return err
		}

		fmt.Printf("✓ Removed %s from cliq workspace\n", toolName)
		fmt.Printf("\nNote: The CLI binary itself was not uninstalled.\n")
		fmt.Printf("To uninstall it, run the appropriate command for your package manager.\n")
		return nil
	},
}
