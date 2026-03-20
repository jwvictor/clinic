package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/togglemedia/clinic/internal/config"
	"github.com/togglemedia/clinic/internal/installer"
	"github.com/togglemedia/clinic/internal/registry"
	"github.com/togglemedia/clinic/internal/skills"
)

var nukeCmd = &cobra.Command{
	Use:   "nuke",
	Short: "Uninstall all managed tools and remove all traces of Clinic",
	Long: `Removes every tool Clinic has installed, deletes all generated skill files,
and removes the ~/.clinic directory. After running this, it's as if Clinic
was never installed (except the clinic binary itself — delete it manually
or run: rm $(which clinic))`,
	RunE: func(cmd *cobra.Command, args []string) error {
		lf, err := config.LoadLockfile()
		if err != nil {
			return err
		}

		if len(lf.Tools) == 0 {
			fmt.Println("No tools managed by Clinic.")
			fmt.Println("Nothing to nuke.")
			return nil
		}

		// Show what will be destroyed
		fmt.Println("This will:")
		fmt.Printf("  • Uninstall %d tools (%s)\n", len(lf.Tools), toolNames(lf))
		fmt.Println("  • Delete all generated skill files")
		fmt.Println("  • Remove ~/.clinic and all config")
		fmt.Println()
		fmt.Print("Are you sure? Type 'nuke' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)

		if answer != "nuke" {
			fmt.Println("Aborted.")
			return nil
		}

		fmt.Println()
		reg := registry.Load()

		// 1. Uninstall each tool
		for toolName, toolLock := range lf.Tools {
			tool, ok := reg.GetTool(toolName)
			if !ok {
				fmt.Printf("  ⚠ %s not in registry, skipping uninstall\n", toolName)
				continue
			}

			if err := installer.Uninstall(tool, toolLock.InstalledVia); err != nil {
				fmt.Printf("  ⚠ %s: could not uninstall (%s)\n", toolName, err)
			} else {
				fmt.Printf("  ✓ Uninstalled %s\n", toolName)
			}
		}

		// 2. Remove all skill files
		fmt.Println()
		for toolName := range lf.Tools {
			if err := skills.Remove(toolName); err != nil {
				fmt.Printf("  ⚠ Could not remove skills for %s: %s\n", toolName, err)
			}
		}
		fmt.Println("  ✓ Removed all skill files")

		// 3. Remove ~/.clinic entirely
		clinicHome := config.ClinicHome()
		if err := os.RemoveAll(clinicHome); err != nil {
			fmt.Printf("  ⚠ Could not remove %s: %s\n", clinicHome, err)
		} else {
			fmt.Printf("  ✓ Removed %s\n", clinicHome)
		}

		fmt.Println()
		fmt.Println("Nuked. To remove the clinic binary itself:")
		fmt.Println("  rm $(which clinic)")
		return nil
	},
}

func toolNames(lf *config.Lockfile) string {
	names := make([]string, 0, len(lf.Tools))
	for name := range lf.Tools {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}
