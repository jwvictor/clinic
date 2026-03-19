package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/togglemedia/clinic/internal/config"
	"github.com/togglemedia/clinic/internal/installer"
	"github.com/togglemedia/clinic/internal/registry"
	"github.com/togglemedia/clinic/internal/skills"
)

var keepBinary bool

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

		toolLock, ok := lf.Tools[toolName]
		if !ok {
			return fmt.Errorf("%s is not in your clinic workspace", toolName)
		}

		// Uninstall the binary unless --keep-binary is set
		if !keepBinary {
			reg := registry.Load()
			if tool, found := reg.GetTool(toolName); found {
				if err := installer.Uninstall(tool, toolLock.InstalledVia); err != nil {
					fmt.Printf("⚠ Could not uninstall binary: %s\n", err)
				} else {
					fmt.Printf("✓ Uninstalled %s (was installed via %s)\n", toolName, toolLock.InstalledVia)
				}
			}
		}

		// Remove skill files
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

		fmt.Printf("✓ Removed %s from clinic workspace\n", toolName)
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVar(&keepBinary, "keep-binary", false, "Only remove from workspace, don't uninstall the CLI binary")
}
