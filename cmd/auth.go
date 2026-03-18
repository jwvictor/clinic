package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/togglemedia/cliq/internal/config"
	"github.com/togglemedia/cliq/internal/doctor"
	"github.com/togglemedia/cliq/internal/registry"
)

var authStatus bool

var authCmd = &cobra.Command{
	Use:   "auth [tool]",
	Short: "Authenticate a CLI tool or check auth status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if authStatus || len(args) == 0 {
			return showAuthStatus()
		}
		return runAuth(args[0])
	},
}

func init() {
	authCmd.Flags().BoolVar(&authStatus, "status", false, "Show auth status for all tools")
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
		if tool.Auth.InjectType == "none" {
			fmt.Printf("%-16s %-10s %s\n", toolName, "n/a", "no auth needed")
			continue
		}

		health := doctor.Check(tool)
		if health.AuthOK {
			fmt.Printf("%-16s %-10s %s\n", toolName, "✓ ok", health.AuthUser)
		} else {
			fmt.Printf("%-16s %-10s run: %s\n", toolName, "✗ no", tool.Auth.AuthCmd)
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

	if tool.Auth.InjectType == "none" {
		fmt.Printf("%s does not require authentication.\n", toolName)
		return nil
	}

	fmt.Printf("Authenticating %s...\n\n", toolName)

	// Run the tool's native auth command interactively
	parts := strings.Fields(tool.Auth.AuthCmd)
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
