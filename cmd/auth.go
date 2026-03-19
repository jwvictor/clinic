package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/togglemedia/clinic/internal/config"
	"github.com/togglemedia/clinic/internal/doctor"
	"github.com/togglemedia/clinic/internal/registry"
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

	// Determine which auth command to use
	headless := authHeadless || detectHeadless()
	authCommand := tool.Auth.AuthCmd

	if headless && tool.Auth.AuthCmdHeadless != "" {
		authCommand = tool.Auth.AuthCmdHeadless
		fmt.Printf("Authenticating %s (headless mode)...\n", toolName)
		fmt.Println("A URL will be displayed — open it on any device with a browser.\n")
	} else if headless {
		// Headless requested but no headless auth flow available — don't fall
		// through to the browser-based command.
		fmt.Printf("⚠ No headless auth flow available for %s.\n", toolName)
		if tool.Auth.EnvVar != "" {
			fmt.Printf("  Set the %s environment variable directly instead:\n\n", tool.Auth.EnvVar)
			fmt.Printf("    export %s=\"your-token-here\"\n\n", tool.Auth.EnvVar)
		}
		return nil
	} else {
		fmt.Printf("Authenticating %s...\n\n", toolName)
	}

	// Run the auth command interactively
	parts := strings.Fields(authCommand)
	c := exec.Command(parts[0], parts[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	fmt.Printf("\n✓ %s authenticated\n", toolName)
	return nil
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
