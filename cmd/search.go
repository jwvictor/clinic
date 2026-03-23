package cmd

import (
	"fmt"
	"strings"

	"github.com/jwvictor/clinic/internal/registry"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search available tools by name, description, or category",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.ToLower(args[0])
		reg := registry.Load()

		var matches []registry.ToolDef
		for _, name := range reg.ToolNames() {
			t := reg.Tools[name]
			if strings.Contains(strings.ToLower(t.Name), query) ||
				strings.Contains(strings.ToLower(t.Description), query) ||
				strings.Contains(strings.ToLower(t.Category), query) {
				matches = append(matches, t)
			}
		}

		if len(matches) == 0 {
			fmt.Printf("No tools matching %q. Try 'clinic list --all' to see everything available.\n", args[0])
			return nil
		}

		for _, t := range matches {
			fmt.Printf("%-16s %-13s %s\n", t.Name, t.Category, t.Description)
		}

		return nil
	},
}
