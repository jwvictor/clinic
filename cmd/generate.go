package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/jwvictor/clinic/internal/config"
	"github.com/jwvictor/clinic/internal/doctor"
	"github.com/jwvictor/clinic/internal/installer"
	"github.com/jwvictor/clinic/internal/registry"
	"github.com/jwvictor/clinic/internal/skills"
)

var generatePlatform string

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Regenerate skill files for all installed tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		lf, err := config.LoadLockfile()
		if err != nil {
			return err
		}

		if len(lf.Tools) == 0 {
			fmt.Println("No tools in workspace. Run 'clinic init --stack <name>' first.")
			return nil
		}

		reg := registry.Load()
		generated := 0

		for toolName := range lf.Tools {
			tool, ok := reg.GetTool(toolName)
			if !ok {
				continue
			}

			status := installer.Detect(tool)
			if !status.Installed {
				fmt.Printf("%-16s skipped (not installed)\n", toolName)
				continue
			}

			health := doctor.Check(tool)
			noAuthNeeded := tool.Auth.InjectType == "" || tool.Auth.InjectType == "none"

			if desc, err := skills.Generate(tool, status, health.AuthUser, health.AuthOK || noAuthNeeded); err != nil {
				fmt.Printf("%-16s ✗ %s\n", toolName, err)
			} else {
				fmt.Printf("%-16s ✓ %s (%s)\n", toolName, skills.SkillPath(toolName), desc)
				generated++
			}
		}

		fmt.Printf("\n%d skill(s) generated\n", generated)
		return nil
	},
}

func init() {
	generateCmd.Flags().StringVar(&generatePlatform, "platform", "", "Target platform (claude, cursor, copilot)")
}
