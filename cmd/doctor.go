package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/togglemedia/clinic/internal/config"
	"github.com/togglemedia/clinic/internal/doctor"
	"github.com/togglemedia/clinic/internal/registry"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Health check — verify all tools are installed, authenticated, and have skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		lf, err := config.LoadLockfile()
		if err != nil {
			return err
		}

		if len(lf.Tools) == 0 {
			fmt.Println("No tools in workspace. Run 'clinic init --stack <name>' to get started.")
			return nil
		}

		reg := registry.Load()

		fmt.Printf("%-16s %-12s %-10s %-8s %s\n", "Tool", "Version", "Auth", "Skill", "Status")
		fmt.Printf("%-16s %-12s %-10s %-8s %s\n", "────", "───────", "────", "─────", "──────")

		issues := 0
		for toolName := range lf.Tools {
			tool, ok := reg.GetTool(toolName)
			if !ok {
				fmt.Printf("%-16s %-12s %-10s %-8s %s\n", toolName, "?", "?", "?", "✗ not in registry")
				issues++
				continue
			}

			health := doctor.Check(tool)

			// Format installed status
			versionStr := health.Version
			if !health.Installed {
				versionStr = "missing"
				issues++
			}

			// Format auth status
			authStr := "✓ ok"
			if !health.Installed {
				authStr = "—"
			} else if tool.Auth.InjectType == "none" {
				authStr = "n/a"
			} else if !health.AuthOK {
				authStr = "✗ no"
				issues++
			}

			// Format skill status
			skillStr := "✓"
			if !health.HasSkill {
				skillStr = "✗"
				issues++
			}

			// Overall status
			statusStr := "✓ ok"
			if !health.Installed {
				statusStr = "✗ not installed"
			} else if !health.AuthOK && tool.Auth.InjectType != "none" {
				statusStr = "⚠ auth needed"
			} else if !health.HasSkill {
				statusStr = "⚠ run clinic generate"
			}

			fmt.Printf("%-16s %-12s %-10s %-8s %s\n", toolName, versionStr, authStr, skillStr, statusStr)
		}

		fmt.Println()
		if issues == 0 {
			fmt.Println("All tools healthy.")
		} else {
			fmt.Printf("%d issue(s) found.\n", issues)
		}
		return nil
	},
}
