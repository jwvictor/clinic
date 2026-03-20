package registry

// ToolDef defines a CLI tool that Clinic can manage.
type ToolDef struct {
	Name           string         `json:"name"`
	Command        string         `json:"command"`
	Description    string         `json:"description"`
	Language       string         `json:"language"`
	Category       string         `json:"category"`
	VersionCommand string         `json:"version_command"`
	VersionPattern string         `json:"version_pattern"`
	InstallMethods []InstallMethod `json:"install_methods"`
	Auth           AuthDef        `json:"auth"`
	SkillTemplate  string         `json:"skill_template"`
	// SkillsSource is a GitHub repo containing vendor-shipped skills.
	// Format: "owner/repo" (e.g., "googleworkspace/cli").
	// If set, Clinic fetches skills from this repo instead of generating them.
	// The repo is expected to have a skills/ directory with SKILL.md files.
	SkillsSource   string         `json:"skills_source,omitempty"`
	// SkillsSubdir overrides the default "skills" subdirectory in the source repo.
	SkillsSubdir   string         `json:"skills_subdir,omitempty"`
}

// InstallMethod defines one way to install a tool.
type InstallMethod struct {
	Type        string   `json:"type"`     // brew, npm, curl_script, shell, go_install, cargo_install
	Platforms   []string `json:"platforms"` // macos, linux
	Formula     string   `json:"formula,omitempty"`
	Package     string   `json:"package,omitempty"`
	Global      bool     `json:"global,omitempty"`
	URLTemplate string   `json:"url_template,omitempty"`
	BinaryName  string   `json:"binary_name,omitempty"`
	ScriptURL    string   `json:"script_url,omitempty"`
	ShellCommand string   `json:"shell_command,omitempty"` // raw shell command to run (alternative to script_url)
	Requires    string   `json:"requires,omitempty"`
}

// AuthDef defines how to authenticate a tool.
type AuthDef struct {
	InjectType      string          `json:"inject_type"`                        // env, file, none
	EnvVar          string          `json:"env_var,omitempty"`
	AuthCheck       string          `json:"auth_check,omitempty"`               // command to check if authed
	AuthCheckPattern string         `json:"auth_check_pattern,omitempty"`      // regex that must match auth_check output for auth to be OK
	AuthCmd         string          `json:"auth_command,omitempty"`             // command to authenticate (interactive/browser)
	AuthCmdHeadless string          `json:"auth_command_headless,omitempty"`    // command to authenticate (no browser, device-code flow)
	AuthHint        string          `json:"auth_hint,omitempty"`                // message shown before running auth command (e.g., where to get a token)
	AuthEnvPrompts  []AuthEnvPrompt `json:"auth_env_prompts,omitempty"`         // for env-var-only tools: prompt user for these values
}

// AuthEnvPrompt defines a single env var to prompt the user for.
type AuthEnvPrompt struct {
	EnvVar   string `json:"env_var"`
	Label    string `json:"label"`
	Required bool   `json:"required"`
	Secret   bool   `json:"secret"` // mask input
}

// StackDef defines a curated bundle of tools.
type StackDef struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tools       []string `json:"tools"`
}
