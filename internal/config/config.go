package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
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

// EnsureDirs creates the clinic directory structure if it doesn't exist.
func EnsureDirs() error {
	dirs := []string{
		ClinicHome(),
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

// Lockfile represents clinic.toml (simplified as JSON for now, will migrate to TOML).
type Lockfile struct {
	Project ProjectConfig          `json:"project"`
	Tools   map[string]ToolLock    `json:"tools"`
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
				Project: ProjectConfig{ClinicVersion: "0.1.0"},
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
