package skills

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jwvictor/clinic/internal/config"
	"github.com/jwvictor/clinic/internal/registry"
)

// vendorSkillNames maps tool name → actual vendor skill directory name.
// e.g., "notion" → "notion-cli" when the vendor repo uses skills/notion-cli/.
var vendorSkillNames = map[string]string{}

// FetchVendorSkills clones a vendor repo's skills directory and installs
// all SKILL.md files to every target agent directory.
// Returns the count of skills installed and the name of the first skill directory.
func FetchVendorSkills(tool registry.ToolDef) (int, error) {
	if tool.SkillsSource == "" {
		return 0, fmt.Errorf("no skills source defined for %s", tool.Name)
	}

	// Clone to a temp directory (shallow, sparse if possible)
	cacheDir := filepath.Join(config.CacheDir(), "skills-repos", tool.Name)
	if err := os.RemoveAll(cacheDir); err != nil {
		return 0, err
	}

	repoURL := fmt.Sprintf("https://github.com/%s.git", tool.SkillsSource)
	cmd := exec.Command("git", "clone", "--depth=1", "--single-branch", repoURL, cacheDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("cloning %s: %w\n%s", repoURL, err, string(out))
	}

	// Find the skills subdirectory
	skillsSubdir := "skills"
	if tool.SkillsSubdir != "" {
		skillsSubdir = tool.SkillsSubdir
	}
	srcDir := filepath.Join(cacheDir, skillsSubdir)

	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		// Try root of repo if skills/ doesn't exist
		srcDir = cacheDir
	}

	// Find all SKILL.md files
	var skillDirsToCopy []string
	err = filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if d.Name() == "SKILL.md" && !d.IsDir() {
			skillDirsToCopy = append(skillDirsToCopy, filepath.Dir(path))
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("scanning skills: %w", err)
	}

	if len(skillDirsToCopy) == 0 {
		return 0, fmt.Errorf("no SKILL.md files found in %s/%s", tool.SkillsSource, skillsSubdir)
	}

	// Copy each skill to all target directories
	home, _ := os.UserHomeDir()
	targets := []string{
		filepath.Join(home, ".openclaw", "skills"),
		filepath.Join(home, ".claude", "skills"),
		filepath.Join(home, ".agents", "skills"),
	}

	installed := 0
	var installedNames []string
	for _, skillSrcDir := range skillDirsToCopy {
		skillName := filepath.Base(skillSrcDir)

		for _, targetRoot := range targets {
			targetDir := filepath.Join(targetRoot, skillName)
			if err := copySkillDir(skillSrcDir, targetDir); err != nil {
				// Log but don't fail — best effort for each target
				fmt.Fprintf(os.Stderr, "  ⚠ Failed to copy %s to %s: %s\n", skillName, targetDir, err)
				continue
			}
		}
		installedNames = append(installedNames, skillName)
		installed++
	}

	// Track the actual skill directory names so SkillPath can be accurate
	if len(installedNames) > 0 {
		vendorSkillNames[tool.Name] = installedNames[0]
	}

	// Clean up clone
	os.RemoveAll(cacheDir)

	return installed, nil
}

// copySkillDir copies all files from src to dst, creating dst if needed.
func copySkillDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Skip non-essential files
		name := strings.ToLower(d.Name())
		if name == ".git" || name == ".gitignore" || name == "license" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, 0644)
	})
}

// HasVendorSkills returns true if the tool has a vendor skills source.
func HasVendorSkills(tool registry.ToolDef) bool {
	return tool.SkillsSource != ""
}
