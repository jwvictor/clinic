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

// skillDirs returns all directories where skills should be written.
// The same SKILL.md works for all targets (Agent Skills standard).
func skillDirs(toolName string) []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".openclaw", "skills", toolName),  // OpenClaw (primary target)
		filepath.Join(home, ".claude", "skills", toolName),    // Claude Code
		filepath.Join(home, ".agents", "skills", toolName),    // Agent Skills open standard
	}
}

// Generate creates SKILL.md files for the given tool across all agent platforms.
func Generate(tool registry.ToolDef, status installer.Status, authUser string) error {
	needsAuth := tool.Auth.InjectType != "" && tool.Auth.InjectType != "none"
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

	content, err := renderSkill(data)
	if err != nil {
		return fmt.Errorf("rendering skill template: %w", err)
	}

	for _, dir := range skillDirs(tool.Name) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating skill directory %s: %w", dir, err)
		}
		skillPath := filepath.Join(dir, "SKILL.md")
		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing skill file %s: %w", skillPath, err)
		}
	}

	return nil
}

// Remove deletes the skill directories for a tool across all platforms.
func Remove(toolName string) error {
	var lastErr error
	for _, dir := range skillDirs(toolName) {
		if err := os.RemoveAll(dir); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// SkillPaths returns all paths where a tool's skill is generated.
func SkillPaths(toolName string) []string {
	var paths []string
	for _, dir := range skillDirs(toolName) {
		paths = append(paths, filepath.Join(dir, "SKILL.md"))
	}
	return paths
}

// SkillPath returns the primary skill path (OpenClaw) for display purposes.
func SkillPath(toolName string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".openclaw", "skills", toolName, "SKILL.md")
}

const skillTemplate = `---
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

func renderSkill(data GenerateData) (string, error) {
	tmpl, err := template.New("skill").Parse(skillTemplate)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
