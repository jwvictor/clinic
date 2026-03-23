package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/jwvictor/clinic/internal/config"
	"github.com/jwvictor/clinic/internal/installer"
	"github.com/jwvictor/clinic/internal/registry"
	"github.com/jwvictor/clinic/internal/skills"
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

		// Count pre-existing vs clinic-installed
		var clinicInstalled, preExisting int
		for _, tl := range lf.Tools {
			if tl.PreExisting {
				preExisting++
			} else {
				clinicInstalled++
			}
		}

		// Show what will be destroyed
		fmt.Println("This will:")
		if clinicInstalled > 0 {
			fmt.Printf("  • Uninstall %d tool(s) installed by clinic\n", clinicInstalled)
		}
		if preExisting > 0 {
			fmt.Printf("  • Keep %d pre-existing tool(s) (only remove clinic skills/config)\n", preExisting)
		}
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

		// 1. Uninstall each tool (skip pre-existing ones)
		for toolName, toolLock := range lf.Tools {
			if toolLock.PreExisting {
				fmt.Printf("  ✓ Keeping %s (was already installed before clinic)\n", toolName)
				continue
			}

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

		// 2. Remove all skill files (including vendor skill directories)
		fmt.Println()
		toolList := make([]string, 0, len(lf.Tools))
		for name := range lf.Tools {
			toolList = append(toolList, name)
		}
		if err := skills.RemoveAllSkills(toolList); err != nil {
			fmt.Printf("  ⚠ Some skills could not be removed: %s\n", err)
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

