package doctor

import (
	"os"
	"os/exec"
	"regexp"
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
		health.AuthOK, health.AuthUser = checkAuth(tool.Auth)
	} else if tool.Auth.InjectType == "none" {
		health.AuthOK = true
		health.AuthUser = "n/a"
	} else if tool.Auth.EnvVar != "" {
		// No auth_check command, but there's an env var — check if it's set
		if os.Getenv(tool.Auth.EnvVar) != "" {
			health.AuthOK = true
			health.AuthUser = "via " + tool.Auth.EnvVar
		}
	}

	// Check skill file existence
	health.HasSkill = checkSkillExists(tool.Name)

	return health
}

func checkAuth(auth registry.AuthDef) (bool, string) {
	parts := strings.Fields(auth.AuthCheck)
	out, err := exec.Command(parts[0], parts[1:]...).CombinedOutput()
	if err != nil {
		return false, ""
	}
	output := strings.TrimSpace(string(out))

	// If the tool defines an explicit success pattern, use it.
	// Output must match for auth to be considered OK.
	if auth.AuthCheckPattern != "" {
		re, err := regexp.Compile(auth.AuthCheckPattern)
		if err != nil {
			return false, ""
		}
		matches := re.FindStringSubmatch(output)
		if matches == nil {
			return false, ""
		}
		// If the pattern has a capture group, use it as the user identifier
		user := ""
		if len(matches) > 1 {
			user = truncate(matches[1], 50)
		}
		return true, user
	}

	// Fallback heuristic: exit code 0 + no obvious failure phrases = authenticated
	lower := strings.ToLower(output)
	failPhrases := []string{
		"not logged in",
		"not authenticated",
		"no account",
		"no auth",
		"login required",
		"unauthenticated",
		"unauthorized",
		"please log in",
		"please login",
		"to log in",
		"to login",
	}
	for _, phrase := range failPhrases {
		if strings.Contains(lower, phrase) {
			return false, ""
		}
	}

	// Extract first line as a rough "user" indicator
	lines := strings.Split(output, "\n")
	if len(lines) > 0 && lines[0] != "" {
		return true, truncate(lines[0], 50)
	}
	return true, ""
}

func checkSkillExists(toolName string) bool {
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
