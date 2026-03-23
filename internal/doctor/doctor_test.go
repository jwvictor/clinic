package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTruncateShortString(t *testing.T) {
	result := truncate("hello", 10)
	if result != "hello" {
		t.Errorf("truncate(%q, 10) = %q, want %q", "hello", result, "hello")
	}
}

func TestTruncateExactLength(t *testing.T) {
	result := truncate("hello", 5)
	if result != "hello" {
		t.Errorf("truncate(%q, 5) = %q, want %q", "hello", result, "hello")
	}
}

func TestTruncateLongString(t *testing.T) {
	input := "this is a very long string that should be truncated"
	result := truncate(input, 10)
	if result != "this is a ..." {
		t.Errorf("truncate(%q, 10) = %q, want %q", input, result, "this is a ...")
	}
	if len(result) != 13 { // 10 chars + "..."
		t.Errorf("truncated length = %d, want 13", len(result))
	}
}

func TestTruncateEmptyString(t *testing.T) {
	result := truncate("", 10)
	if result != "" {
		t.Errorf("truncate(%q, 10) = %q, want %q", "", result, "")
	}
}

func TestFileExistsWithExistingFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if !fileExists(path) {
		t.Errorf("fileExists(%q) = false, want true", path)
	}
}

func TestFileExistsWithNonExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.txt")
	if fileExists(path) {
		t.Errorf("fileExists(%q) = true, want false", path)
	}
}

func TestFileExistsWithDirectory(t *testing.T) {
	// test -f should return false for directories
	tmp := t.TempDir()
	if fileExists(tmp) {
		t.Errorf("fileExists(%q) = true for directory, want false", tmp)
	}
}

func TestSkillPathsReturnsExpected(t *testing.T) {
	paths := skillPaths("gh")
	if len(paths) != 3 {
		t.Fatalf("skillPaths() returned %d paths, want 3", len(paths))
	}

	expectations := []string{
		".openclaw/skills/gh/SKILL.md",
		".claude/skills/gh/SKILL.md",
		".agents/skills/gh/SKILL.md",
	}
	for i, exp := range expectations {
		if !strings.HasSuffix(paths[i], exp) {
			t.Errorf("skillPaths()[%d] = %q, want suffix %q", i, paths[i], exp)
		}
	}
}

func TestCheckSkillExistsWithTempSkillFile(t *testing.T) {
	home, _ := os.UserHomeDir()

	// Create a temp SKILL.md in one of the expected locations
	skillDir := filepath.Join(home, ".openclaw", "skills", "_test_clinic_doctor_tool")
	skillPath := filepath.Join(skillDir, "SKILL.md")

	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillPath, []byte("test skill"), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(skillDir)

	if !checkSkillExists("_test_clinic_doctor_tool") {
		t.Error("checkSkillExists() = false after creating SKILL.md, want true")
	}
}

func TestCheckSkillExistsMissingTool(t *testing.T) {
	if checkSkillExists("_nonexistent_tool_that_should_never_exist_xyz123") {
		t.Error("checkSkillExists() = true for nonexistent tool, want false")
	}
}

func TestToolHealthStruct(t *testing.T) {
	// Basic struct construction test
	h := ToolHealth{
		Name:         "gh",
		Installed:    true,
		Version:      "2.30.0",
		InstalledVia: "brew",
		AuthOK:       true,
		AuthUser:     "octocat",
		HasSkill:     true,
	}

	if h.Name != "gh" {
		t.Errorf("Name = %q, want %q", h.Name, "gh")
	}
	if !h.Installed {
		t.Error("Installed = false, want true")
	}
	if !h.AuthOK {
		t.Error("AuthOK = false, want true")
	}
}
