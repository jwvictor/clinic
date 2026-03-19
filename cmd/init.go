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

var initStack string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize workspace with a stack of CLI tools",
	Long:  `Sets up your agent workspace by installing, authenticating, and generating skills for a curated stack of CLI tools.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.EnsureDirs(); err != nil {
			return fmt.Errorf("creating clinic directories: %w", err)
		}

		reg := registry.Load()

		if initStack == "" {
			fmt.Println("Available stacks:")
			fmt.Println()
			for _, s := range reg.Stacks {
				fmt.Printf("  %-16s %s (%d tools)\n", s.Name, s.Description, len(s.Tools))
			}
			fmt.Println()
			fmt.Println("Run: clinic init --stack <name>")
			return nil
		}

		stack, ok := reg.GetStack(initStack)
		if !ok {
			return fmt.Errorf("unknown stack: %s", initStack)
		}

		fmt.Printf("Clinic — Setting up your agent workspace\n\n")
		fmt.Printf("Stack: %s (%d tools)\n\n", stack.Name, len(stack.Tools))

		lf, err := config.LoadLockfile()
		if err != nil {
			return err
		}
		lf.Project.Stack = stack.Name

		for i, toolName := range stack.Tools {
			tool, ok := reg.GetTool(toolName)
			if !ok {
				fmt.Printf("[%d/%d] %s — not found in registry, skipping\n", i+1, len(stack.Tools), toolName)
				continue
			}

			fmt.Printf("[%d/%d] %s (%s)\n", i+1, len(stack.Tools), tool.Name, tool.Description)

			// Detect existing installation
			status := installer.Detect(tool)

			if status.Installed {
				fmt.Printf("  ✓ Already installed (v%s via %s)\n", status.Version, status.InstalledVia)
			} else {
				fmt.Printf("  → Installing...\n")
				method, err := installer.Install(tool)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  ✗ Install failed: %s\n", err)
					continue
				}
				status = installer.Detect(tool)
				status.InstalledVia = method // trust the method we just used
				fmt.Printf("  ✓ Installed v%s via %s\n", status.Version, method)
			}

			// Check auth
			health := doctor.Check(tool)
			if tool.Auth.InjectType == "" || tool.Auth.InjectType == "none" {
				// No auth needed, skip
			} else if health.AuthOK {
				fmt.Printf("  ✓ Authenticated (%s)\n", health.AuthUser)
			} else {
				fmt.Printf("  ⚠ Not authenticated — run: %s\n", tool.Auth.AuthCmd)
			}

			// Generate skill
			if err := skills.Generate(tool, status, health.AuthUser); err != nil {
				fmt.Fprintf(os.Stderr, "  ⚠ Skill generation failed: %s\n", err)
			} else {
				fmt.Printf("  ✓ Skill generated: %s\n", skills.SkillPath(tool.Name))
			}

			// Record in lockfile
			lf.Tools[tool.Name] = config.ToolLock{
				Version:      status.Version,
				InstalledVia: status.InstalledVia,
			}

			fmt.Println()
		}

		if err := lf.Save(); err != nil {
			return fmt.Errorf("saving lockfile: %w", err)
		}

		fmt.Printf("Wrote %s (commit this to your repo)\n\n", config.LockfilePath())
		fmt.Println("Your workspace is agent-ready. 🤝")
		return nil
	},
}

func init() {
	initCmd.Flags().StringVarP(&initStack, "stack", "s", "", "Stack to initialize (e.g., saas-founder, devops, indie-hacker)")
}
