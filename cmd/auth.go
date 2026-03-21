package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
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

	// Special guided flows for tools that need multi-step setup
	if toolName == "gws" {
		return runGwsAuth(tool)
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

// runGwsAuth handles the multi-step Google Workspace CLI auth flow.
func runGwsAuth(tool registry.ToolDef) error {
	// Styles
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render
	step := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")).Render
	success := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")).Render
	warn := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).Render
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render
	warningBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("11")).
		Padding(1, 2).
		Width(62)

	fmt.Println()
	fmt.Println(title("Google Workspace CLI Authentication"))
	fmt.Println()
	fmt.Println(dim("Gmail, Drive, Calendar, Docs, Sheets, Chat, Tasks, and more"))
	fmt.Println()

	// Check if client_secret.json already exists
	home, _ := os.UserHomeDir()
	clientSecretPath := home + "/.config/gws/client_secret.json"
	needsSetup := false
	if _, err := os.Stat(clientSecretPath); err != nil {
		needsSetup = true
	}

	if needsSetup {
		fmt.Println(step("Step 1 of 2") + "  OAuth Client Setup")
		fmt.Println()
		fmt.Println("  This will walk you through creating a GCP OAuth client.")
		fmt.Println("  You'll need a Google Cloud project (one will be detected")
		fmt.Println("  or created automatically).")
		fmt.Println()
		fmt.Println(warn("  When the setup offers to log you in at the end,"))
		fmt.Println(warn("  choose NO") + " — we'll handle login in step 2 with")
		fmt.Println("  the right permissions.")
		fmt.Println()

		var proceed bool
		huh.NewConfirm().
			Title("Ready to start setup?").
			Affirmative("Let's go").
			Negative("Cancel").
			Value(&proceed).
			Run()

		if !proceed {
			return fmt.Errorf("gws setup cancelled")
		}

		fmt.Println()
		c := exec.Command("sh", "-c", "gws auth setup")
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			fmt.Println()
			fmt.Println(warn("Setup had issues: ") + err.Error())
			fmt.Println()
			var cont bool
			huh.NewConfirm().
				Title("Continue to login anyway?").
				Affirmative("Yes, continue").
				Negative("Abort").
				Value(&cont).
				Run()
			if !cont {
				return fmt.Errorf("gws setup aborted")
			}
		}
		fmt.Println()
	} else {
		fmt.Println(success("  ✓") + " OAuth client already configured")
		fmt.Println()
	}

	// Step 2: Login with scoped access
	stepLabel := "Step 2 of 2"
	if !needsSetup {
		stepLabel = "Login"
	}
	fmt.Println(step(stepLabel) + "  Google Account Authorization")
	fmt.Println()
	fmt.Println("  Next, gws will confirm your scopes and give you a link")
	fmt.Println("  to authorize access to your Google account.")
	fmt.Println()

	boxContent := warn("Check ALL the permission boxes!") + "\n\n" +
		"For unverified apps, Google shows each\n" +
		"permission as an " + warn("unchecked checkbox") + ".\n\n" +
		"If you don't check them, only basic profile\n" +
		"access will be granted and " + warn("nothing will work") + "."

	fmt.Println(warningBox.Render(boxContent))
	fmt.Println()

	var proceed bool
	huh.NewConfirm().
		Title("Ready to continue?").
		Affirmative("Continue").
		Negative("Cancel").
		Value(&proceed).
		Run()

	if !proceed {
		return fmt.Errorf("gws login cancelled")
	}

	fmt.Println()
	loginCmd := "gws auth login -s gmail,drive,calendar,docs,sheets,chat,tasks,keep,forms"
	c := exec.Command("sh", "-c", loginCmd)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("gws login failed: %w", err)
	}

	fmt.Println()
	fmt.Println(success("  ✓ Google Workspace authenticated"))
	generateSkillsAfterAuth(tool)
	checkShellenvSetup()
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
