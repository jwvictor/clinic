package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/jwvictor/clinic/internal/config"
	"github.com/jwvictor/clinic/internal/doctor"
	"github.com/jwvictor/clinic/internal/installer"
	"github.com/jwvictor/clinic/internal/registry"
	"github.com/jwvictor/clinic/internal/skills"
)

var authStatus bool
var authHeadless bool

var authCmd = &cobra.Command{
	Use:   "auth [tool]",
	Short: "Authenticate a CLI tool or check auth status",
	Long: `Authenticate a CLI tool interactively.

Use --headless when you don't have a browser on this machine (e.g., Docker,
SSH, CI). The tool will print a URL and code — visit it on your phone or
laptop to complete auth.

Headless mode is auto-detected when the DISPLAY env var is unset and
the system has no browser.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if authStatus || len(args) == 0 {
			return showAuthStatus()
		}
		return runAuth(args[0])
	},
}

func init() {
	authCmd.Flags().BoolVar(&authStatus, "status", false, "Show auth status for all tools")
	authCmd.Flags().BoolVar(&authHeadless, "headless", false, "Use device-code / no-browser auth flow")
}

func showAuthStatus() error {
	lf, err := config.LoadLockfile()
	if err != nil {
		return err
	}

	if len(lf.Tools) == 0 {
		fmt.Println("No tools in workspace.")
		return nil
	}

	reg := registry.Load()

	fmt.Printf("%-16s %-10s %s\n", "Tool", "Auth", "Details")
	fmt.Printf("%-16s %-10s %s\n", "────", "────", "───────")

	for toolName := range lf.Tools {
		tool, ok := reg.GetTool(toolName)
		if !ok {
			continue
		}
		if tool.Auth.InjectType == "" || tool.Auth.InjectType == "none" {
			fmt.Printf("%-16s %-10s %s\n", toolName, "n/a", "no auth needed")
			continue
		}

		health := doctor.Check(tool)
		if health.AuthOK {
			fmt.Printf("%-16s %-10s %s\n", toolName, "✓ ok", health.AuthUser)
		} else {
			fmt.Printf("%-16s %-10s run: clinic auth %s\n", toolName, "✗ no", toolName)
		}
	}
	return nil
}

func runAuth(toolName string) error {
	reg := registry.Load()
	tool, ok := reg.GetTool(toolName)
	if !ok {
		return fmt.Errorf("unknown tool: %s", toolName)
	}

	if tool.Auth.InjectType == "" || tool.Auth.InjectType == "none" {
		fmt.Printf("%s does not require authentication.\n", toolName)
		return nil
	}

	// If the tool has env prompts (no interactive auth command), use our own flow
	if len(tool.Auth.AuthEnvPrompts) > 0 {
		return runEnvAuth(toolName, tool)
	}

	// If there's no auth command at all but there's an env var, fall back to
	// prompting for the single env var
	if tool.Auth.AuthCmd == "" && tool.Auth.EnvVar != "" {
		return runEnvAuth(toolName, tool)
	}

	// Determine which auth command to use
	headless := authHeadless || detectHeadless()
	authCommand := tool.Auth.AuthCmd

	if headless && tool.Auth.AuthCmdHeadless != "" {
		authCommand = tool.Auth.AuthCmdHeadless
		fmt.Printf("Authenticating %s (headless mode)...\n", toolName)
		fmt.Println("A URL will be displayed — open it on any device with a browser.\n")
	} else if headless {
		fmt.Printf("Authenticating %s...\n", toolName)
		fmt.Printf("⚠ No specific headless auth flow known for %s — trying default auth.\n", toolName)
		if tool.Auth.EnvVar != "" {
			fmt.Printf("  Tip: you can also set %s directly.\n\n", tool.Auth.EnvVar)
		} else {
			fmt.Println()
		}
	} else {
		fmt.Printf("Authenticating %s...\n\n", toolName)
	}

	// Show hint if available
	if tool.Auth.AuthHint != "" {
		fmt.Printf("ℹ %s\n\n", tool.Auth.AuthHint)
	}

	// Run the auth command interactively via a shell so TUI tools get
	// proper terminal control (process group, job control, etc.).
	c := exec.Command("sh", "-c", authCommand)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	fmt.Printf("\n✓ %s authenticated\n", toolName)
	generateSkillsAfterAuth(tool)
	return nil
}

// runEnvAuth handles authentication for tools that only need env vars
// (no interactive auth command). Prompts the user for each value and
// saves to ~/.clinic/env/<tool>.env.
func runEnvAuth(toolName string, tool registry.ToolDef) error {
	fmt.Printf("Authenticating %s\n\n", toolName)

	if tool.Auth.AuthHint != "" {
		fmt.Printf("ℹ %s\n\n", tool.Auth.AuthHint)
	}

	reader := bufio.NewReader(os.Stdin)
	envVars := map[string]string{}

	// If there are explicit prompts, use them
	if len(tool.Auth.AuthEnvPrompts) > 0 {
		for _, prompt := range tool.Auth.AuthEnvPrompts {
			label := prompt.Label
			if !prompt.Required {
				label += " (optional)"
			}
			fmt.Printf("  %s: ", label)
			value, _ := reader.ReadString('\n')
			value = strings.TrimSpace(value)

			if value == "" && prompt.Required {
				return fmt.Errorf("%s is required", prompt.Label)
			}
			if value != "" {
				envVars[prompt.EnvVar] = value
				os.Setenv(prompt.EnvVar, value) // set for current process
			}
		}
	} else {
		// Fallback: single env var prompt
		fmt.Printf("  %s: ", tool.Auth.EnvVar)
		value, _ := reader.ReadString('\n')
		value = strings.TrimSpace(value)
		if value == "" {
			return fmt.Errorf("%s is required", tool.Auth.EnvVar)
		}
		envVars[tool.Auth.EnvVar] = value
		os.Setenv(tool.Auth.EnvVar, value)
	}

	// Save to ~/.clinic/env/<tool>.env
	if err := config.SaveToolEnv(toolName, envVars); err != nil {
		return fmt.Errorf("saving credentials: %w", err)
	}

	fmt.Printf("\n✓ %s authenticated\n", toolName)
	fmt.Printf("  Saved to ~/.clinic/env/%s.env\n", toolName)

	generateSkillsAfterAuth(tool)
	checkShellenvSetup()
	return nil
}

// generateSkillsAfterAuth generates skills for a tool after successful auth.
func generateSkillsAfterAuth(tool registry.ToolDef) {
	status := installer.Detect(tool)
	health := doctor.Check(tool)
	if desc, err := skills.Generate(tool, status, health.AuthUser, true); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Skills: %s\n", err)
	} else {
		fmt.Printf("✓ Skills installed: %s (%s)\n", skills.SkillPath(tool.Name), desc)
	}
}

var shellenvHintShown bool

// checkShellenvSetup checks if eval "$(clinic shellenv)" is in the user's
// shell RC file, and suggests adding it if not. Only prints once per session.
func checkShellenvSetup() {
	if shellenvHintShown || config.HasShellenvInRC() {
		return
	}
	shellenvHintShown = true
	rcFile := config.ShellRCFile()
	fmt.Printf("\n  To load credentials in new shells, add this to %s:\n", rcFile)
	fmt.Printf("    eval \"$(clinic shellenv)\"\n")
}

// detectHeadless returns true if we're likely in a headless environment
// (no browser available).
func detectHeadless() bool {
	// SSH session — no local browser
	if os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_TTY") != "" {
		return true
	}

	// CI environments
	if os.Getenv("CI") != "" {
		return true
	}

	// Docker / container (/.dockerenv or /run/.containerenv)
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	if _, err := os.Stat("/run/.containerenv"); err == nil {
		return true
	}

	// No display server on Linux
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		// On macOS, `open` always exists even with a GUI, so only flag
		// headless if neither xdg-open nor open can be found.
		if _, err := exec.LookPath("xdg-open"); err != nil {
			if _, err := exec.LookPath("open"); err != nil {
				return true
			}
		}
	}

	return false
}
