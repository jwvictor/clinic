package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/togglemedia/cliq/internal/installer"
	"github.com/togglemedia/cliq/internal/registry"
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

// Generate creates a SKILL.md file for the given tool.
func Generate(tool registry.ToolDef, status installer.Status, authUser string) error {
	home, _ := os.UserHomeDir()
	skillDir := filepath.Join(home, ".claude", "skills", tool.Name)

	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("creating skill directory: %w", err)
	}

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

	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing skill file: %w", err)
	}

	return nil
}

// Remove deletes the skill directory for a tool.
func Remove(toolName string) error {
	home, _ := os.UserHomeDir()
	skillDir := filepath.Join(home, ".claude", "skills", toolName)
	return os.RemoveAll(skillDir)
}

// SkillPath returns the path where a tool's skill would be generated.
func SkillPath(toolName string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "skills", toolName, "SKILL.md")
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
Managed by Cliq. Token injected via ` + "`{{.AuthEnvVar}}`" + ` env var.
{{- end}}
If auth fails, run ` + "`cliq auth {{.Name}}`" + ` or ` + "`cliq doctor`" + `.
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
