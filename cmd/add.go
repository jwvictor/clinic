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
		preExisting := status.Installed
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
			fmt.Printf("⚠ Not authenticated — run: clinic auth %s\n", tool.Name)
		}

		// Generate skill (only if authed or no auth needed)
		noAuthNeeded := tool.Auth.InjectType == "" || tool.Auth.InjectType == "none"
		if desc, err := skills.Generate(tool, status, health.AuthUser, health.AuthOK || noAuthNeeded); err != nil {
			fmt.Printf("⚠ Skills: %s\n", err)
		} else {
			fmt.Printf("✓ Skills: %s (%s)\n", skills.SkillPath(tool.Name), desc)
		}

		// Update lockfile
		lf, err := config.LoadLockfile()
		if err != nil {
			return err
		}
		lf.Tools[tool.Name] = config.ToolLock{
			Version:      status.Version,
			InstalledVia: status.InstalledVia,
			PreExisting:  preExisting,
		}
		return lf.Save()
	},
}
