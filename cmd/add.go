package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/togglemedia/clinic/internal/config"
	"github.com/togglemedia/clinic/internal/doctor"
	"github.com/togglemedia/clinic/internal/installer"
	"github.com/togglemedia/clinic/internal/registry"
	"github.com/togglemedia/clinic/internal/skills"
)

var addCmd = &cobra.Command{
	Use:   "add <tool>",
	Short: "Add a CLI tool to your workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		toolName := args[0]
		reg := registry.Load()

		tool, ok := reg.GetTool(toolName)
		if !ok {
			return fmt.Errorf("unknown tool: %s\n\nRun 'clinic stacks' to see available tools", toolName)
		}

		if err := config.EnsureDirs(); err != nil {
			return err
		}

		fmt.Printf("Adding %s (%s)\n\n", tool.Name, tool.Description)

		// Detect or install
		status := installer.Detect(tool)
		if status.Installed {
			fmt.Printf("✓ Already installed (v%s via %s)\n", status.Version, status.InstalledVia)
		} else {
			fmt.Printf("→ Installing...\n")
			method, err := installer.Install(tool)
			if err != nil {
				return fmt.Errorf("install failed: %w", err)
			}
			status = installer.Detect(tool)
			status.InstalledVia = method // trust the method we just used
			fmt.Printf("✓ Installed v%s via %s\n", status.Version, method)
		}

		// Check auth
		health := doctor.Check(tool)
		if tool.Auth.InjectType == "" || tool.Auth.InjectType == "none" {
			// No auth needed, skip
		} else if health.AuthOK {
			fmt.Printf("✓ Authenticated (%s)\n", health.AuthUser)
		} else {
			fmt.Printf("⚠ Not authenticated — run: %s\n", tool.Auth.AuthCmd)
		}

		// Generate skill
		if err := skills.Generate(tool, status, health.AuthUser); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ Skill generation failed: %s\n", err)
		} else {
			fmt.Printf("✓ Skill generated: %s\n", skills.SkillPath(tool.Name))
		}

		// Update lockfile
		lf, err := config.LoadLockfile()
		if err != nil {
			return err
		}
		lf.Tools[tool.Name] = config.ToolLock{
			Version:      status.Version,
			InstalledVia: status.InstalledVia,
		}
		return lf.Save()
	},
}
