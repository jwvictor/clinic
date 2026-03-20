package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
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

		// Force-refresh registry to pick up latest tool definitions
		reg, _ := registry.ForceRefresh()
		if reg == nil {
			reg = registry.Load() // fall back to cache/embedded
		}

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

		// Multi-select which tools to install (all pre-selected)
		options := make([]huh.Option[string], 0, len(stack.Tools))
		var defaultSelected []string
		for _, toolName := range stack.Tools {
			tool, ok := reg.GetTool(toolName)
			if !ok {
				continue
			}
			options = append(options, huh.NewOption(
				fmt.Sprintf("%s — %s", tool.Name, tool.Description),
				tool.Name,
			))
			defaultSelected = append(defaultSelected, tool.Name)
		}

		var selectedTools []string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title(fmt.Sprintf("Stack: %s — select tools to install", stack.Name)).
					Description("Space to toggle, Enter to confirm").
					Options(options...).
					Value(&selectedTools),
			),
		)
		selectedTools = defaultSelected // pre-select all
		if err := form.Run(); err != nil {
			return nil // user cancelled
		}

		if len(selectedTools) == 0 {
			fmt.Println("No tools selected.")
			return nil
		}

		fmt.Printf("\nClinic — Setting up your agent workspace\n\n")

		lf, err := config.LoadLockfile()
		if err != nil {
			return err
		}
		lf.Project.Stack = stack.Name

		// Track tools that need auth and tools that got authed
		type unauthTool struct {
			name     string
			tool     registry.ToolDef
			status   installer.Status
			authCmd  string
			authHint string
		}
		var needsAuth []unauthTool

		for i, toolName := range selectedTools {
			tool, ok := reg.GetTool(toolName)
			if !ok {
				fmt.Printf("[%d/%d] %s — not found in registry, skipping\n", i+1, len(selectedTools), toolName)
				continue
			}

			fmt.Printf("[%d/%d] %s (%s)\n", i+1, len(selectedTools), tool.Name, tool.Description)

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
				status.InstalledVia = method
				fmt.Printf("  ✓ Installed v%s via %s\n", status.Version, method)
			}

			// Check auth
			health := doctor.Check(tool)
			noAuthNeeded := tool.Auth.InjectType == "" || tool.Auth.InjectType == "none"

			if noAuthNeeded {
				// No auth needed — generate skills immediately
			} else if health.AuthOK {
				fmt.Printf("  ✓ Authenticated (%s)\n", health.AuthUser)
			} else {
				fmt.Printf("  ⚠ Not authenticated\n")
				if tool.Auth.AuthCmd != "" {
					needsAuth = append(needsAuth, unauthTool{
						name:     tool.Name,
						tool:     tool,
						status:   status,
						authCmd:  tool.Auth.AuthCmd,
						authHint: tool.Auth.AuthHint,
					})
				}
			}

			// Generate skill (will be skipped if auth is needed but not done)
			if desc, err := skills.Generate(tool, status, health.AuthUser, health.AuthOK || noAuthNeeded); err != nil {
				fmt.Printf("  ⚠ Skills: %s\n", err)
			} else {
				fmt.Printf("  ✓ Skills: %s (%s)\n", skills.SkillPath(tool.Name), desc)
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

		fmt.Printf("Wrote %s\n\n", config.LockfilePath())

		// Offer to authenticate tools that need it
		if len(needsAuth) > 0 {
			authOptions := make([]huh.Option[string], len(needsAuth))
			for i, t := range needsAuth {
				authOptions[i] = huh.NewOption(t.name, t.name)
			}

			var selectedAuth []string
			authForm := huh.NewForm(
				huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("Which tools do you want to authenticate?").
						Description("Skills are only installed for authenticated tools. Space to toggle, Enter to confirm").
						Options(authOptions...).
						Value(&selectedAuth),
				),
			)

			if err := authForm.Run(); err != nil {
				fmt.Println()
			}

			if len(selectedAuth) > 0 {
				fmt.Println()
				authMap := map[string]unauthTool{}
				for _, t := range needsAuth {
					authMap[t.name] = t
				}
				for _, name := range selectedAuth {
					t := authMap[name]
					fmt.Printf("─── Authenticating %s ───\n", t.name)
					if t.authHint != "" {
						fmt.Printf("  ℹ %s\n", t.authHint)
					}
					fmt.Println()
					parts := strings.Fields(t.authCmd)
					c := exec.Command(parts[0], parts[1:]...)
					c.Stdin = os.Stdin
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					if err := c.Run(); err != nil {
						fmt.Printf("\n⚠ %s auth failed: %s\n\n", t.name, err)
					} else {
						fmt.Printf("\n✓ %s authenticated\n", t.name)
						// Now generate skills for the newly authed tool
						health := doctor.Check(t.tool)
						if desc, err := skills.Generate(t.tool, t.status, health.AuthUser, true); err != nil {
							fmt.Printf("  ⚠ Skills: %s\n\n", err)
						} else {
							fmt.Printf("  ✓ Skills installed: %s (%s)\n\n", skills.SkillPath(t.name), desc)
						}
					}
				}
			}
		}

		fmt.Println("Your workspace is agent-ready. 🤝")
		return nil
	},
}

func init() {
	initCmd.Flags().StringVarP(&initStack, "stack", "s", "", "Stack to initialize (e.g., saas-founder, devops, indie-hacker)")
}
