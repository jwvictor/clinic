package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// CliqHome returns the path to ~/.cliq/
func CliqHome() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cliq")
}

// BinDir returns ~/.cliq/bin/
func BinDir() string {
	return filepath.Join(CliqHome(), "bin")
}

// VersionsDir returns ~/.cliq/versions/
func VersionsDir() string {
	return filepath.Join(CliqHome(), "versions")
}

// CacheDir returns ~/.cliq/cache/
func CacheDir() string {
	return filepath.Join(CliqHome(), "cache")
}

// EnsureDirs creates the cliq directory structure if it doesn't exist.
func EnsureDirs() error {
	dirs := []string{
		CliqHome(),
		BinDir(),
		VersionsDir(),
		CacheDir(),
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

// Lockfile represents cliq.toml (simplified as JSON for now, will migrate to TOML).
type Lockfile struct {
	Project ProjectConfig          `json:"project"`
	Tools   map[string]ToolLock    `json:"tools"`
}

type ProjectConfig struct {
	Stack        string `json:"stack,omitempty"`
	CliqVersion  string `json:"cliq_version"`
}

type ToolLock struct {
	Version      string `json:"version"`
	InstalledVia string `json:"installed_via"`
}

// LockfilePath returns the path to cliq.json in the current directory.
func LockfilePath() string {
	return "cliq.json"
}

// LoadLockfile reads the lockfile from the current directory.
func LoadLockfile() (*Lockfile, error) {
	data, err := os.ReadFile(LockfilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return &Lockfile{
				Project: ProjectConfig{CliqVersion: "0.1.0"},
				Tools:   make(map[string]ToolLock),
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
	return &lf, nil
}

// Save writes the lockfile to disk.
func (lf *Lockfile) Save() error {
	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(LockfilePath(), data, 0644)
}
