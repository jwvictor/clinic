package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLockfileSaveAndLoad(t *testing.T) {
	// Use a temp dir as the clinic home so we don't touch real config
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	// Ensure the .clinic dir exists
	if err := os.MkdirAll(filepath.Join(tmp, ".clinic"), 0755); err != nil {
		t.Fatal(err)
	}

	lf := &Lockfile{
		Project: ProjectConfig{
			Stack:         "devops",
			ClinicVersion: "1.2.3",
		},
		Tools: map[string]ToolLock{
			"gh": {
				Version:      "2.30.0",
				InstalledVia: "brew",
				PreExisting:  true,
			},
			"jq": {
				Version:      "1.7",
				InstalledVia: "brew",
			},
		},
	}

	// Save
	if err := lf.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify the file was written and is valid JSON
	data, err := os.ReadFile(LockfilePath())
	if err != nil {
		t.Fatalf("reading lockfile after save: %v", err)
	}
	var parsed Lockfile
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}

	// Load it back
	loaded, err := LoadLockfile()
	if err != nil {
		t.Fatalf("LoadLockfile() error: %v", err)
	}

	if loaded.Project.Stack != "devops" {
		t.Errorf("Project.Stack = %q, want %q", loaded.Project.Stack, "devops")
	}
	if loaded.Project.ClinicVersion != "1.2.3" {
		t.Errorf("Project.ClinicVersion = %q, want %q", loaded.Project.ClinicVersion, "1.2.3")
	}
	if len(loaded.Tools) != 2 {
		t.Fatalf("len(Tools) = %d, want 2", len(loaded.Tools))
	}
	gh, ok := loaded.Tools["gh"]
	if !ok {
		t.Fatal("Tools[\"gh\"] not found")
	}
	if gh.Version != "2.30.0" {
		t.Errorf("gh.Version = %q, want %q", gh.Version, "2.30.0")
	}
	if !gh.PreExisting {
		t.Error("gh.PreExisting = false, want true")
	}
}

func TestLoadLockfileMissingFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// No .clinic dir at all — should return empty lockfile, not error
	lf, err := LoadLockfile()
	if err != nil {
		t.Fatalf("LoadLockfile() on missing file: %v", err)
	}
	if lf.Project.ClinicVersion != "0.1.0" {
		t.Errorf("default ClinicVersion = %q, want %q", lf.Project.ClinicVersion, "0.1.0")
	}
	if lf.Tools == nil {
		t.Error("Tools map should be initialized, got nil")
	}
	if len(lf.Tools) != 0 {
		t.Errorf("len(Tools) = %d, want 0", len(lf.Tools))
	}
}

func TestLockfileToolsMapManipulation(t *testing.T) {
	lf := &Lockfile{
		Project: ProjectConfig{ClinicVersion: "0.1.0"},
		Tools:   make(map[string]ToolLock),
	}

	// Add a tool
	lf.Tools["docker"] = ToolLock{
		Version:      "24.0.0",
		InstalledVia: "brew",
	}

	if _, ok := lf.Tools["docker"]; !ok {
		t.Fatal("docker not found after adding")
	}

	// Remove a tool
	delete(lf.Tools, "docker")
	if _, ok := lf.Tools["docker"]; ok {
		t.Fatal("docker still found after removing")
	}
}

func TestSaveToolEnvAndLoadAllEnvFiles(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := os.MkdirAll(filepath.Join(tmp, ".clinic", "env"), 0755); err != nil {
		t.Fatal(err)
	}

	vars := map[string]string{
		"GH_TOKEN": "test-token-123",
	}
	if err := SaveToolEnv("gh", vars); err != nil {
		t.Fatalf("SaveToolEnv() error: %v", err)
	}

	// Verify file was written
	envPath := filepath.Join(tmp, ".clinic", "env", "gh.env")
	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("reading env file: %v", err)
	}
	content := string(data)
	if content == "" {
		t.Fatal("env file is empty")
	}

	// Load all env files
	allEnv, err := LoadAllEnvFiles()
	if err != nil {
		t.Fatalf("LoadAllEnvFiles() error: %v", err)
	}
	if allEnv == "" {
		t.Fatal("LoadAllEnvFiles() returned empty string")
	}
}

func TestLoadAllEnvFilesMissingDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// No env dir — should return empty string, no error
	result, err := LoadAllEnvFiles()
	if err != nil {
		t.Fatalf("LoadAllEnvFiles() error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestEnsureDirs(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs() error: %v", err)
	}

	// Verify key directories exist
	expected := []string{
		filepath.Join(tmp, ".clinic"),
		filepath.Join(tmp, ".clinic", "bin"),
		filepath.Join(tmp, ".clinic", "versions"),
		filepath.Join(tmp, ".clinic", "cache"),
		filepath.Join(tmp, ".clinic", "env"),
		filepath.Join(tmp, ".clinic", "cache", "registry"),
		filepath.Join(tmp, ".clinic", "cache", "downloads"),
	}
	for _, dir := range expected {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("directory %s does not exist: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func TestPlatformReturnsNonEmpty(t *testing.T) {
	os_, arch := Platform()
	if os_ == "" {
		t.Error("Platform() returned empty OS")
	}
	if arch == "" {
		t.Error("Platform() returned empty arch")
	}
}
