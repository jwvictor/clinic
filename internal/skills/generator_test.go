package skills

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestMissingFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	m := LoadManifest()
	if m == nil {
		t.Fatal("LoadManifest() returned nil, want empty map")
	}
	if len(m) != 0 {
		t.Errorf("len(manifest) = %d, want 0", len(m))
	}
}

func TestSaveAndLoadManifestRoundtrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	original := map[string][]string{
		"gws":    {"gws-gmail", "gws-drive", "gws-calendar"},
		"docker": {"docker"},
	}

	SaveManifest(original)

	// Verify the file was written
	path := filepath.Join(tmp, ".clinic", "skills-manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("manifest file not written: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string][]string
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("manifest is not valid JSON: %v", err)
	}

	// Load it back
	loaded := LoadManifest()
	if len(loaded) != 2 {
		t.Fatalf("len(loaded) = %d, want 2", len(loaded))
	}
	if len(loaded["gws"]) != 3 {
		t.Errorf("len(loaded[\"gws\"]) = %d, want 3", len(loaded["gws"]))
	}
	if len(loaded["docker"]) != 1 {
		t.Errorf("len(loaded[\"docker\"]) = %d, want 1", len(loaded["docker"]))
	}
}

func TestTrackSkillDirsUpdatesManifest(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Track some dirs for tool A
	TrackSkillDirs("toolA", []string{"dir1", "dir2"})
	m := LoadManifest()
	if len(m["toolA"]) != 2 {
		t.Fatalf("toolA dirs = %d, want 2", len(m["toolA"]))
	}

	// Track dirs for tool B — should not affect tool A
	TrackSkillDirs("toolB", []string{"dir3"})
	m = LoadManifest()
	if len(m) != 2 {
		t.Fatalf("manifest has %d tools, want 2", len(m))
	}
	if len(m["toolA"]) != 2 {
		t.Errorf("toolA dirs after adding toolB = %d, want 2", len(m["toolA"]))
	}

	// Update tool A — should replace its entry
	TrackSkillDirs("toolA", []string{"dir4"})
	m = LoadManifest()
	if len(m["toolA"]) != 1 {
		t.Errorf("toolA dirs after update = %d, want 1", len(m["toolA"]))
	}
	if m["toolA"][0] != "dir4" {
		t.Errorf("toolA[0] = %q, want %q", m["toolA"][0], "dir4")
	}
}

func TestRenderTemplate(t *testing.T) {
	data := GenerateData{
		Name:        "gh",
		Command:     "gh",
		Description: "interact with GitHub",
		Version:     "2.30.0",
		AuthUser:    "octocat",
		AuthEnvVar:  "GH_TOKEN",
		AuthCmd:     "gh auth login",
		Category:    "devtools",
		NeedsAuth:   true,
	}

	result, err := renderTemplate(genericTemplate, data)
	if err != nil {
		t.Fatalf("renderTemplate() error: %v", err)
	}

	// Check key parts are present
	checks := []string{
		"name: gh",
		"gh",
		"interact with GitHub",
		"v2.30.0",
		"octocat",
		"GH_TOKEN",
		"gh auth login",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("rendered template missing %q", check)
		}
	}
}

func TestRenderTemplateNoAuth(t *testing.T) {
	data := GenerateData{
		Name:        "jq",
		Command:     "jq",
		Description: "process JSON",
		Version:     "1.7",
		NeedsAuth:   false,
	}

	result, err := renderTemplate(genericTemplate, data)
	if err != nil {
		t.Fatalf("renderTemplate() error: %v", err)
	}

	if strings.Contains(result, "## Auth") {
		t.Error("template should NOT contain Auth section for non-auth tool")
	}
	if !strings.Contains(result, "jq") {
		t.Error("template should contain tool name")
	}
}

func TestTargetDirs(t *testing.T) {
	dirs := targetDirs()
	if len(dirs) != 3 {
		t.Fatalf("targetDirs() returned %d dirs, want 3", len(dirs))
	}

	// Each should contain the expected skill directory names
	expectations := []string{".openclaw/skills", ".claude/skills", ".agents/skills"}
	for i, exp := range expectations {
		if !strings.Contains(dirs[i], exp) {
			t.Errorf("targetDirs()[%d] = %q, want to contain %q", i, dirs[i], exp)
		}
	}
}

func TestSkillDirsForTool(t *testing.T) {
	dirs := skillDirsForTool("gh")
	if len(dirs) != 3 {
		t.Fatalf("skillDirsForTool() returned %d dirs, want 3", len(dirs))
	}
	for _, d := range dirs {
		if !strings.HasSuffix(d, "/gh") {
			t.Errorf("skillDirsForTool dir %q should end with /gh", d)
		}
	}
}
