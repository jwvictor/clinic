package cmd

import (
	"fmt"

	"github.com/jwvictor/clinic/internal/versioncheck"
	"github.com/spf13/cobra"
)

var version = "0.1.0-dev"

var rootCmd = &cobra.Command{
	Use:   "clinic",
	Short: "Your CLI tools, managed — installed, authenticated, agent-ready",
	Long: `Clinic manages collections of agent-friendly CLI tools as unified, opinionated stacks.
It handles discovery, installation, authentication, skill generation, and lifecycle
management — turning a bare terminal into a fully agent-capable workspace in one command.`,
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		versioncheck.CheckAndNotify(version)
	},
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
	rootCmd.AddCommand(nukeCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(selfUpdateCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Clinic version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("clinic version %s\n", version)
	},
}
