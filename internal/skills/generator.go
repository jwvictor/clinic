package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/togglemedia/clinic/internal/installer"
	"github.com/togglemedia/clinic/internal/registry"
)

// GenerateData holds the template variables for skill generation.
type GenerateData struct {
	Name        string
	Command     string
	Description string
	Version     string
	AuthUser    string
	AuthEnvVar  string
	AuthCmd     string
	Category    string
	NeedsAuth   bool
}

// targetDirs returns all directories where skills should be written.
func targetDirs() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".openclaw", "skills"), // OpenClaw (primary target)
		filepath.Join(home, ".claude", "skills"),   // Claude Code
		filepath.Join(home, ".agents", "skills"),   // Agent Skills open standard
	}
}

// skillDirsForTool returns all target directories for a specific tool.
func skillDirsForTool(toolName string) []string {
	var dirs []string
	for _, root := range targetDirs() {
		dirs = append(dirs, filepath.Join(root, toolName))
	}
	return dirs
}

// Generate creates skill files for the given tool.
// It uses a three-tier strategy:
//  1. Vendor skills — if the tool ships its own skills (e.g., gws has 93), fetch them
//  2. Curated skills — hand-written templates for popular tools (gh, aws, stripe, etc.)
//  3. Generic fallback — basic template for everything else
//
// If the tool requires auth and authOK is false, skills are NOT generated
// (and any existing skills are removed) to prevent agents from trying to
// use unauthenticated tools.
func Generate(tool registry.ToolDef, status installer.Status, authUser string, authOK bool) (string, error) {
	// Don't generate skills for unauthenticated tools that need auth
	needsAuth := tool.Auth.InjectType != "" && tool.Auth.InjectType != "none"
	if needsAuth && !authOK {
		// Remove any stale skills from a previous install
		Remove(tool.Name)
		return "", fmt.Errorf("skipped — authenticate first")
	}
	// Tier 1: Vendor-shipped skills
	if HasVendorSkills(tool) {
		count, err := FetchVendorSkills(tool)
		if err != nil {
			// Fall through to curated/generic on vendor failure
			fmt.Fprintf(os.Stderr, "  ⚠ Vendor skills fetch failed (%s), falling back to curated\n", err)
		} else {
			return fmt.Sprintf("%d vendor skills installed", count), nil
		}
	}

	// Build template data
	cleanAuthUser := authUser
	if !needsAuth || authUser == "n/a" {
		cleanAuthUser = ""
	}

	data := GenerateData{
		Name:        tool.Name,
		Command:     tool.Command,
		Description: tool.Description,
		Version:     status.Version,
		AuthUser:    cleanAuthUser,
		AuthEnvVar:  tool.Auth.EnvVar,
		AuthCmd:     tool.Auth.AuthCmd,
		Category:    tool.Category,
		NeedsAuth:   needsAuth,
	}

	// Tier 2: Curated skill template
	tmplStr, hasCurated := curatedSkills[tool.Name]
	if !hasCurated {
		// Tier 3: Generic fallback
		tmplStr = genericTemplate
	}

	content, err := renderTemplate(tmplStr, data)
	if err != nil {
		return "", fmt.Errorf("rendering skill template: %w", err)
	}

	// Write to all target directories
	for _, dir := range skillDirsForTool(tool.Name) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("creating skill directory %s: %w", dir, err)
		}
		skillPath := filepath.Join(dir, "SKILL.md")
		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("writing skill file %s: %w", skillPath, err)
		}
	}

	if hasCurated {
		return "curated skill", nil
	}
	return "generic skill", nil
}

// Remove deletes the skill directories for a tool across all platforms.
func Remove(toolName string) error {
	var lastErr error
	for _, dir := range skillDirsForTool(toolName) {
		if err := os.RemoveAll(dir); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// SkillPath returns the primary skill path (OpenClaw) for display purposes.
func SkillPath(toolName string) string {
	home, _ := os.UserHomeDir()
	// Vendor skills may use a different directory name than the tool name
	dirName := toolName
	if vendorName, ok := vendorSkillNames[toolName]; ok {
		dirName = vendorName
	}
	return filepath.Join(home, ".openclaw", "skills", dirName, "SKILL.md")
}

const genericTemplate = `---
name: {{.Name}}
description: >
  Use when the user needs {{.Description}}.
  The {{.Command}} CLI is installed{{if .NeedsAuth}} and authenticated{{end}}.
allowed-tools: Bash({{.Command}}:*)
---

You have the ` + "`{{.Command}}`" + ` CLI (v{{.Version}}) installed{{if .AuthUser}} and authenticated as **{{.AuthUser}}**{{end}}.

## Usage
Run ` + "`{{.Command}} --help`" + ` to see available commands. Use ` + "`{{.Command}} <subcommand> --help`" + ` for details on any subcommand.
Prefer ` + "`--json`" + ` or structured output flags when parsing results programmatically.
{{- if .NeedsAuth}}

## Auth
{{- if .AuthEnvVar}}
Managed by Clinic. Token injected via ` + "`{{.AuthEnvVar}}`" + ` env var.
{{- end}}
If auth fails, run ` + "`clinic auth {{.Name}}`" + ` or ` + "`clinic doctor`" + `.
{{- if .AuthCmd}}
To re-authenticate manually: ` + "`{{.AuthCmd}}`" + `
{{- end}}
{{- end}}
`

func renderTemplate(tmplStr string, data GenerateData) (string, error) {
	tmpl, err := template.New("skill").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
