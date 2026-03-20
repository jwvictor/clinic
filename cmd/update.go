package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/jwvictor/clinic/internal/config"
	"github.com/jwvictor/clinic/internal/installer"
	"github.com/jwvictor/clinic/internal/registry"
)

var updateCmd = &cobra.Command{
	Use:   "update [tool]",
	Short: "Update one or all tools in your workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		lf, err := config.LoadLockfile()
		if err != nil {
			return err
		}
		reg := registry.Load()

		var toolNames []string
		if len(args) > 0 {
			toolNames = args
		} else {
			for name := range lf.Tools {
				toolNames = append(toolNames, name)
			}
		}

		for _, toolName := range toolNames {
			tool, ok := reg.GetTool(toolName)
			if !ok {
				fmt.Printf("%-16s skipped (not in registry)\n", toolName)
				continue
			}

			lock, inLockfile := lf.Tools[toolName]
			status := installer.Detect(tool)

			if !status.Installed {
				fmt.Printf("%-16s not installed — run 'clinic add %s'\n", toolName, toolName)
				continue
			}

			fmt.Printf("%-16s v%s via %s", toolName, status.Version, status.InstalledVia)

			if inLockfile && lock.Version == status.Version {
				fmt.Printf(" — up to date\n")
			} else {
				fmt.Printf(" — updated lockfile\n")
				lf.Tools[toolName] = config.ToolLock{
					Version:      status.Version,
					InstalledVia: status.InstalledVia,
				}
			}
		}

		return lf.Save()
	},
}
