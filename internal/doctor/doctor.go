package doctor

import (
	"os/exec"
	"strings"

	"github.com/togglemedia/clinic/internal/installer"
	"github.com/togglemedia/clinic/internal/registry"
)

// ToolHealth represents the full health status of a managed tool.
type ToolHealth struct {
	Name         string
	Installed    bool
	Version      string
	InstalledVia string
	AuthOK       bool
	AuthUser     string
	HasSkill     bool
}

// Check runs health checks on a single tool.
func Check(tool registry.ToolDef) ToolHealth {
	status := installer.Detect(tool)

	health := ToolHealth{
		Name:         tool.Name,
		Installed:    status.Installed,
		Version:      status.Version,
		InstalledVia: status.InstalledVia,
	}

	if !status.Installed {
		return health
	}

	// Check auth
	if tool.Auth.AuthCheck != "" {
		health.AuthOK, health.AuthUser = checkAuth(tool.Auth.AuthCheck)
	} else if tool.Auth.InjectType == "none" {
		health.AuthOK = true
		health.AuthUser = "n/a"
	}

	// Check skill file existence
	health.HasSkill = checkSkillExists(tool.Name)

	return health
}

func checkAuth(command string) (bool, string) {
	parts := strings.Fields(command)
	out, err := exec.Command(parts[0], parts[1:]...).CombinedOutput()
	if err != nil {
		return false, ""
	}
	// Extract first line as a rough "user" indicator
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 0 {
		return true, truncate(lines[0], 50)
	}
	return true, ""
}

func checkSkillExists(toolName string) bool {
	// Check common skill locations
	paths := skillPaths(toolName)
	for _, p := range paths {
		if fileExists(p) {
			return true
		}
	}
	return false
}

func skillPaths(toolName string) []string {
	home := homeDir()
	return []string{
		home + "/.openclaw/skills/" + toolName + "/SKILL.md",
		home + "/.claude/skills/" + toolName + "/SKILL.md",
		home + "/.agents/skills/" + toolName + "/SKILL.md",
	}
}

func homeDir() string {
	home, _ := exec.Command("sh", "-c", "echo $HOME").Output()
	return strings.TrimSpace(string(home))
}

func fileExists(path string) bool {
	_, err := exec.Command("test", "-f", path).CombinedOutput()
	return err == nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
