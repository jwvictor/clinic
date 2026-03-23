package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ClinicHome returns the path to ~/.clinic/
func ClinicHome() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".clinic")
}

// BinDir returns ~/.clinic/bin/
func BinDir() string {
	return filepath.Join(ClinicHome(), "bin")
}

// VersionsDir returns ~/.clinic/versions/
func VersionsDir() string {
	return filepath.Join(ClinicHome(), "versions")
}

// CacheDir returns ~/.clinic/cache/
func CacheDir() string {
	return filepath.Join(ClinicHome(), "cache")
}

// EnvDir returns ~/.clinic/env/
func EnvDir() string {
	return filepath.Join(ClinicHome(), "env")
}

// EnsureDirs creates the clinic directory structure if it doesn't exist.
func EnsureDirs() error {
	dirs := []string{
		ClinicHome(),
		BinDir(),
		VersionsDir(),
		CacheDir(),
		EnvDir(),
		filepath.Join(CacheDir(), "registry"),
		filepath.Join(CacheDir(), "downloads"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}

// Platform returns the current OS and architecture in normalized form.
func Platform() (os_ string, arch string) {
	os_ = runtime.GOOS
	arch = runtime.GOARCH
	return
}

// currentSchemaVersion is the latest lockfile schema version that this build of clinic understands.
const currentSchemaVersion = 1

// Lockfile represents clinic.toml (simplified as JSON for now, will migrate to TOML).
type Lockfile struct {
	SchemaVersion int                    `json:"schema_version"`
	Project       ProjectConfig          `json:"project"`
	Tools         map[string]ToolLock    `json:"tools"`
}

type ProjectConfig struct {
	Stack        string `json:"stack,omitempty"`
	ClinicVersion  string `json:"clinic_version"`
}

type ToolLock struct {
	Version      string `json:"version"`
	InstalledVia string `json:"installed_via"`
	PreExisting  bool   `json:"pre_existing,omitempty"`
}

// LockfilePath returns the path to ~/.clinic/clinic.json.
func LockfilePath() string {
	return filepath.Join(ClinicHome(), "clinic.json")
}

// LoadLockfile reads the lockfile from the current directory.
func LoadLockfile() (*Lockfile, error) {
	data, err := os.ReadFile(LockfilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return &Lockfile{
				SchemaVersion: currentSchemaVersion,
				Project:       ProjectConfig{ClinicVersion: "0.1.0"},
				Tools:         make(map[string]ToolLock),
			}, nil
		}
		return nil, err
	}
	var lf Lockfile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, err
	}
	if lf.Tools == nil {
		lf.Tools = make(map[string]ToolLock)
	}
	// Backfill schema version for lockfiles written before versioning was added.
	if lf.SchemaVersion == 0 {
		lf.SchemaVersion = 1
	}
	if lf.SchemaVersion > currentSchemaVersion {
		return nil, fmt.Errorf(
			"lockfile has schema_version %d but this build of clinic only supports up to %d — please update clinic",
			lf.SchemaVersion, currentSchemaVersion,
		)
	}
	return &lf, nil
}

// SaveToolEnv writes env vars for a tool to ~/.clinic/env/<tool>.env
func SaveToolEnv(toolName string, vars map[string]string) error {
	if err := os.MkdirAll(EnvDir(), 0700); err != nil {
		return err
	}
	var lines []string
	for k, v := range vars {
		lines = append(lines, fmt.Sprintf("export %s=%q", k, v))
	}
	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(filepath.Join(EnvDir(), toolName+".env"), []byte(content), 0600)
}

// LoadAllEnvFiles reads all .env files from ~/.clinic/env/ and returns
// shell export lines suitable for eval.
func LoadAllEnvFiles() (string, error) {
	entries, err := os.ReadDir(EnvDir())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	var lines []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".env") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(EnvDir(), e.Name()))
		if err != nil {
			continue
		}
		lines = append(lines, strings.TrimSpace(string(data)))
	}
	return strings.Join(lines, "\n"), nil
}

// ShellRCFile returns the path to the user's shell RC file based on $SHELL.
func ShellRCFile() string {
	home, _ := os.UserHomeDir()
	shell := os.Getenv("SHELL")
	switch {
	case strings.Contains(shell, "zsh"):
		return filepath.Join(home, ".zshrc")
	case strings.Contains(shell, "bash"):
		// Prefer .bash_profile on macOS, .bashrc on Linux
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, ".bash_profile")
		}
		return filepath.Join(home, ".bashrc")
	case strings.Contains(shell, "fish"):
		return filepath.Join(home, ".config", "fish", "config.fish")
	default:
		return filepath.Join(home, ".profile")
	}
}

// HasShellenvInRC checks if eval "$(clinic shellenv)" is already in the shell RC file.
func HasShellenvInRC() bool {
	data, err := os.ReadFile(ShellRCFile())
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "clinic shellenv")
}

// Save writes the lockfile to disk.
func (lf *Lockfile) Save() error {
	lf.SchemaVersion = currentSchemaVersion
	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(LockfilePath(), data, 0644)
}
