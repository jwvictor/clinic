package installer

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/togglemedia/clinic/internal/registry"
)

// Status represents the install state of a tool.
type Status struct {
	Installed    bool
	Version      string
	InstalledVia string // brew, npm, binary, curl_script, go_install, cargo_install, unknown
}

// Detect checks if a tool is already installed and returns its status.
func Detect(tool registry.ToolDef) Status {
	path, err := exec.LookPath(tool.Command)
	if err != nil || path == "" {
		return Status{Installed: false}
	}

	version := detectVersion(tool)
	method := detectMethod(tool.Command)

	return Status{
		Installed:    true,
		Version:      version,
		InstalledVia: method,
	}
}

func detectVersion(tool registry.ToolDef) string {
	if tool.VersionCommand == "" {
		return "unknown"
	}
	parts := strings.Fields(tool.VersionCommand)
	out, err := exec.Command(parts[0], parts[1:]...).CombinedOutput()
	if err != nil {
		return "unknown"
	}
	if tool.VersionPattern == "" {
		return strings.TrimSpace(string(out))
	}
	re, err := regexp.Compile(tool.VersionPattern)
	if err != nil {
		return strings.TrimSpace(string(out))
	}
	matches := re.FindStringSubmatch(string(out))
	if len(matches) > 1 {
		return matches[1]
	}
	return strings.TrimSpace(string(out))
}

func detectMethod(command string) string {
	// Check if it's a brew-managed binary
	out, err := exec.Command("brew", "list", "--formula", command).CombinedOutput()
	if err == nil && len(out) > 0 {
		return "brew"
	}
	// Check if it's an npm global
	out, err = exec.Command("npm", "list", "-g", "--depth=0", command).CombinedOutput()
	if err == nil && strings.Contains(string(out), command) {
		return "npm"
	}
	return "unknown"
}

// Install attempts to install a tool using the first available method.
func Install(tool registry.ToolDef) (string, error) {
	os_ := runtime.GOOS
	platform := "linux"
	if os_ == "darwin" {
		platform = "macos"
	}

	for _, method := range tool.InstallMethods {
		if !supportsplatform(method, platform) {
			continue
		}
		switch method.Type {
		case "brew":
			if hasBrew() {
				err := runInstall("brew", "install", method.Formula)
				if err == nil {
					return "brew", nil
				}
			}
		case "npm":
			if hasNpm() {
				if method.Requires != "" && !checkNodeVersion(method.Requires) {
					continue
				}
				args := []string{"install"}
				if method.Global {
					args = append(args, "-g")
				}
				args = append(args, method.Package)
				err := runInstall("npm", args...)
				if err != nil && method.Global && runtime.GOOS == "linux" {
					// Retry with sudo on Linux — apt-installed Node needs it for global installs
					sudoArgs := append([]string{"npm"}, args...)
					err = runInstall("sudo", sudoArgs...)
				}
				if err == nil {
					return "npm", nil
				}
			}
		case "curl_script":
			if method.ScriptURL != "" {
				var err error
				if strings.HasSuffix(method.ScriptURL, ".tgz") || strings.HasSuffix(method.ScriptURL, ".tar.gz") {
					// Binary tarball — download, extract, move to /usr/local/bin
					binaryName := tool.Command
					if method.BinaryName != "" {
						binaryName = method.BinaryName
					}
					script := fmt.Sprintf(
						"cd /tmp && curl -sSL %s -o _clinic_dl.tgz && tar -xzf _clinic_dl.tgz %s && sudo mv %s /usr/local/bin/ && rm -f _clinic_dl.tgz",
						method.ScriptURL, binaryName, binaryName,
					)
					err = runShell(script)
				} else {
					err = runShell(fmt.Sprintf("curl -sSL %s | bash", method.ScriptURL))
				}
				if err == nil {
					return "curl_script", nil
				}
			}
		case "shell":
			if method.ShellCommand != "" {
				err := runShell(method.ShellCommand)
				if err == nil {
					return "shell", nil
				}
			}
		case "go_install":
			if hasGo() && method.Package != "" {
				err := runInstall("go", "install", method.Package)
				if err == nil {
					return "go_install", nil
				}
			}
		case "cargo_install":
			if hasCargo() && method.Package != "" {
				args := []string{"install", "--git", method.Package}
				err := runInstall("cargo", args...)
				if err == nil {
					return "cargo_install", nil
				}
			}
		}
	}

	return "", fmt.Errorf("no available install method for %s on %s", tool.Name, platform)
}

func supportsplatform(method registry.InstallMethod, platform string) bool {
	for _, p := range method.Platforms {
		if p == platform {
			return true
		}
	}
	return false
}

func hasBrew() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

func hasNpm() bool {
	_, err := exec.LookPath("npm")
	return err == nil
}

func hasGo() bool {
	_, err := exec.LookPath("go")
	return err == nil
}

func hasCargo() bool {
	_, err := exec.LookPath("cargo")
	return err == nil
}

func checkNodeVersion(requires string) bool {
	out, err := exec.Command("node", "--version").Output()
	if err != nil {
		return false
	}
	// Simple check — just verify node exists. Full semver parsing can come later.
	_ = out
	return true
}

// Uninstall removes a tool using the package manager that installed it.
func Uninstall(tool registry.ToolDef, installedVia string) error {
	switch installedVia {
	case "brew":
		formula := tool.Command
		for _, m := range tool.InstallMethods {
			if m.Type == "brew" && m.Formula != "" {
				formula = m.Formula
				break
			}
		}
		return runInstall("brew", "uninstall", formula)
	case "npm":
		pkg := tool.Command
		for _, m := range tool.InstallMethods {
			if m.Type == "npm" && m.Package != "" {
				pkg = m.Package
				break
			}
		}
		args := []string{"uninstall", "-g", pkg}
		err := runInstall("npm", args...)
		if err != nil && runtime.GOOS == "linux" {
			return runInstall("sudo", append([]string{"npm"}, args...)...)
		}
		return err
	default:
		return fmt.Errorf("don't know how to uninstall %s (installed via %s)", tool.Name, installedVia)
	}
}

func runInstall(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}

func runShell(command string) error {
	cmd := exec.Command("sh", "-c", command)
	return cmd.Run()
}
