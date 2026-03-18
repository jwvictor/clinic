package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.1.0-dev"

var rootCmd = &cobra.Command{
	Use:   "cliq",
	Short: "Your clique of CLI tools — managed, authenticated, agent-ready",
	Long: `Cliq manages collections of agent-friendly CLI tools as unified, opinionated stacks.
It handles discovery, installation, authentication, skill generation, and lifecycle
management — turning a bare terminal into a fully agent-capable workspace in one command.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(stacksCmd)
	rootCmd.AddCommand(shellenvCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Cliq version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cliq version %s\n", version)
	},
}
