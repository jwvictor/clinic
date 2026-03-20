package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

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

		// Track tools that need auth
		type unauthTool struct {
			name     string
			authCmd  string
			authHint string
		}
		var needsAuth []unauthTool

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
				fmt.Printf("  ⚠ Not authenticated\n")
				needsAuth = append(needsAuth, unauthTool{
					name:     tool.Name,
					authCmd:  tool.Auth.AuthCmd,
					authHint: tool.Auth.AuthHint,
				})
			}

			// Generate skill
			if desc, err := skills.Generate(tool, status, health.AuthUser); err != nil {
				fmt.Fprintf(os.Stderr, "  ⚠ Skill generation failed: %s\n", err)
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
			fmt.Printf("The following tools need authentication:\n\n")
			for i, t := range needsAuth {
				fmt.Printf("  %d. %s\n", i+1, t.name)
			}
			fmt.Println()
			fmt.Printf("Which tools to authenticate? (e.g. 1,3 / all / none): ")

			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			// Parse selection
			selected := parseSelection(answer, len(needsAuth))

			if len(selected) > 0 {
				fmt.Println()
				for _, idx := range selected {
					t := needsAuth[idx]
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
						fmt.Printf("\n✓ %s authenticated\n\n", t.name)
					}
				}
			}
		}

		fmt.Println("Your workspace is agent-ready. 🤝")
		return nil
	},
}

// parseSelection parses user input like "1,3,5", "all", or "none" into
// a slice of 0-based indices.
func parseSelection(input string, total int) []int {
	if input == "none" || input == "n" || input == "" {
		return nil
	}
	if input == "all" || input == "a" {
		indices := make([]int, total)
		for i := range indices {
			indices[i] = i
		}
		return indices
	}
	var indices []int
	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		n := 0
		for _, c := range part {
			if c >= '0' && c <= '9' {
				n = n*10 + int(c-'0')
			}
		}
		if n >= 1 && n <= total {
			indices = append(indices, n-1)
		}
	}
	return indices
}

func init() {
	initCmd.Flags().StringVarP(&initStack, "stack", "s", "", "Stack to initialize (e.g., saas-founder, devops, indie-hacker)")
}
