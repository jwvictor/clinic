package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/jwvictor/clinic/internal/config"
)

var shellenvCmd = &cobra.Command{
	Use:   "shellenv",
	Short: "Print shell environment setup (add to .zshrc/.bashrc)",
	Long: `Outputs shell commands that add Clinic's bin directory to PATH
and export any saved API keys/tokens.

Add this to your shell profile:
  eval "$(clinic shellenv)"`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("export PATH=\"%s:$PATH\"\n", config.BinDir())

		// Export saved env vars from ~/.clinic/env/*.env
		envLines, err := config.LoadAllEnvFiles()
		if err == nil && envLines != "" {
			fmt.Println(envLines)
		}
	},
}
