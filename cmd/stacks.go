package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/togglemedia/clinic/internal/registry"
)

var stacksCmd = &cobra.Command{
	Use:   "stacks",
	Short: "Browse available stacks",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := registry.Load()

		fmt.Println("Available stacks:")
		fmt.Println()

		for _, s := range reg.Stacks {
			fmt.Printf("  %s\n", s.Name)
			fmt.Printf("  %s\n", s.Description)
			fmt.Printf("  Tools: %s\n", strings.Join(s.Tools, ", "))
			fmt.Println()
		}

		fmt.Println("Run: clinic init --stack <name>")
		return nil
	},
}
