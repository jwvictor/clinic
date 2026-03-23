package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/jwvictor/clinic/internal/config"
	"github.com/jwvictor/clinic/internal/installer"
	"github.com/jwvictor/clinic/internal/registry"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade [tool]",
	Short: "Upgrade one or all installed tools via their package managers",
	Long: `Upgrade runs the appropriate package-manager upgrade command for each
installed tool (e.g. brew upgrade, npm update -g, go install, cargo install).

If a tool name is given, only that tool is upgraded. Otherwise all installed
tools in the lockfile are upgraded.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		lf, err := config.LoadLockfile()
		if err != nil {
			return err
		}
		reg := registry.Load()

		var toolNames []string
		if len(args) > 0 {
			toolNames = args
		} else {
			for name := range lf.Tools {
				toolNames = append(toolNames, name)
			}
		}

		if len(toolNames) == 0 {
			fmt.Println("No tools installed. Run 'clinic add <tool>' first.")
			return nil
		}

		for _, toolName := range toolNames {
			lock, inLockfile := lf.Tools[toolName]
			if !inLockfile {
				fmt.Printf("%-16s skipped (not in lockfile)\n", toolName)
				continue
			}

			tool, ok := reg.GetTool(toolName)
			if !ok {
				fmt.Printf("%-16s skipped (not in registry)\n", toolName)
				continue
			}

			oldVersion := lock.Version
			installedVia := lock.InstalledVia

			// Run the appropriate upgrade command
			upgradeErr := runUpgrade(tool, installedVia)
			if upgradeErr != nil {
				fmt.Printf("%-16s v%-10s  ✗ upgrade failed: %v\n", toolName, oldVersion, upgradeErr)
				continue
			}

			// Detect the new version and update the lockfile
			status := installer.Detect(tool)
			if !status.Installed {
				fmt.Printf("%-16s v%-10s  ✗ not found after upgrade\n", toolName, oldVersion)
				continue
			}

			newVersion := status.Version
			lf.Tools[toolName] = config.ToolLock{
				Version:      newVersion,
				InstalledVia: status.InstalledVia,
			}

			if oldVersion != newVersion {
				fmt.Printf("%-16s v%s → v%s  ✓ upgraded\n", toolName, oldVersion, newVersion)
			} else {
				fmt.Printf("%-16s v%-10s  ✓ already latest\n", toolName, newVersion)
			}
		}

		return lf.Save()
	},
}

// runUpgrade executes the package-manager upgrade command for a tool.
func runUpgrade(tool registry.ToolDef, installedVia string) error {
	switch installedVia {
	case "brew":
		formula := tool.Command
		for _, m := range tool.InstallMethods {
			if m.Type == "brew" && m.Formula != "" {
				formula = m.Formula
				break
			}
		}
		return installer.RunInstallCmd("brew", "upgrade", formula)

	case "npm":
		pkg := tool.Command
		for _, m := range tool.InstallMethods {
			if m.Type == "npm" && m.Package != "" {
				pkg = m.Package
				break
			}
		}
		args := []string{"update", "-g", pkg}
		err := installer.RunInstallCmd("npm", args...)
		if err != nil && runtime.GOOS == "linux" {
			// Retry with sudo on Linux — apt-installed Node needs it for global installs
			sudoArgs := append([]string{"npm"}, args...)
			err = installer.RunInstallCmd("sudo", sudoArgs...)
		}
		return err

	case "go_install":
		for _, m := range tool.InstallMethods {
			if m.Type == "go_install" && m.Package != "" {
				return installer.RunInstallCmd("go", "install", m.Package)
			}
		}
		return fmt.Errorf("no go_install method found in registry for %s", tool.Name)

	case "cargo_install":
		for _, m := range tool.InstallMethods {
			if m.Type == "cargo_install" && m.Package != "" {
				return installer.RunInstallCmd("cargo", "install", "--git", m.Package)
			}
		}
		return fmt.Errorf("no cargo_install method found in registry for %s", tool.Name)

	case "curl_script", "shell", "unknown":
		return fmt.Errorf("manual upgrade needed (installed via %s)", installedVia)

	default:
		return fmt.Errorf("manual upgrade needed (installed via %s)", installedVia)
	}
}
