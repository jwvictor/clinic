package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	method := detectMethod(tool)

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

func detectMethod(tool registry.ToolDef) string {
	// Try each install method's actual formula/package name — the command name
	// often differs (e.g., command "clx" is brew formula "circumflex").
	for _, m := range tool.InstallMethods {
		switch m.Type {
		case "brew":
			name := m.Formula
			if name == "" {
				name = tool.Command
			}
			// Check both formula and cask
			out, err := exec.Command("brew", "list", "--formula", name).CombinedOutput()
			if err == nil && len(out) > 0 {
				return "brew"
			}
			out, err = exec.Command("brew", "list", "--cask", name).CombinedOutput()
			if err == nil && len(out) > 0 {
				return "brew"
			}
		case "npm":
			name := m.Package
			if name == "" {
				name = tool.Command
			}
			out, err := exec.Command("npm", "list", "-g", "--depth=0", name).CombinedOutput()
			if err == nil && strings.Contains(string(out), name) {
				return "npm"
			}
		}
	}
	// Fallback: check by command name (formula and cask)
	out, err := exec.Command("brew", "list", "--formula", tool.Command).CombinedOutput()
	if err == nil && len(out) > 0 {
		return "brew"
	}
	out, err = exec.Command("brew", "list", "--cask", tool.Command).CombinedOutput()
	if err == nil && len(out) > 0 {
		return "brew"
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
					// go install puts binaries in ~/go/bin which may not be in PATH —
					// symlink into /usr/local/bin so it's always available
					symlinkToPath(tool.Command, gobin())
					return "go_install", nil
				}
			}
		case "cargo_install":
			if hasCargo() && method.Package != "" {
				args := []string{"install", "--git", method.Package}
				err := runInstall("cargo", args...)
				if err == nil {
					// cargo install puts binaries in ~/.cargo/bin which may not be in PATH
					symlinkToPath(tool.Command, cargobin())
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
	case "go_install":
		binPath := filepath.Join(gobin(), tool.Command)
		os.Remove(binPath)
		// Remove symlink if we created one
		os.Remove(filepath.Join("/usr/local/bin", tool.Command))
		return nil
	case "cargo_install":
		return runInstall("cargo", "uninstall", tool.Command)
	default:
		return fmt.Errorf("don't know how to uninstall %s (installed via %s)", tool.Name, installedVia)
	}
}

func gobin() string {
	if gb := os.Getenv("GOBIN"); gb != "" {
		return gb
	}
	gp := os.Getenv("GOPATH")
	if gp == "" {
		home, _ := os.UserHomeDir()
		gp = filepath.Join(home, "go")
	}
	return filepath.Join(gp, "bin")
}

func cargobin() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cargo", "bin")
}

// symlinkToPath creates a symlink in /usr/local/bin pointing to a binary
// in a non-PATH directory (e.g. ~/go/bin). Fails silently if it can't.
func symlinkToPath(command string, srcDir string) {
	src := filepath.Join(srcDir, command)
	dst := filepath.Join("/usr/local/bin", command)
	// Don't overwrite an existing file
	if _, err := os.Lstat(dst); err == nil {
		return
	}
	os.Symlink(src, dst)
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
