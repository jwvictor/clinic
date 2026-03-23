package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var selfUpdateCmd = &cobra.Command{
	Use:     "self-update",
	Aliases: []string{"selfupdate"},
	Short:   "Update clinic itself to the latest release",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Fetch latest release info from GitHub
		fmt.Println("Checking for updates...")

		latestVersion, err := fetchLatestVersion()
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}

		// Strip leading "v" for comparison with our version var
		latestClean := strings.TrimPrefix(latestVersion, "v")
		currentClean := strings.TrimPrefix(version, "v")

		// 2. Compare versions
		if latestClean == currentClean {
			fmt.Printf("clinic is already up to date (v%s)\n", currentClean)
			return nil
		}

		fmt.Printf("Updating clinic: v%s → v%s\n", currentClean, latestClean)

		// 3. Detect OS and arch
		osName := runtime.GOOS
		archName := runtime.GOARCH
		if osName != "darwin" && osName != "linux" {
			return fmt.Errorf("unsupported OS: %s", osName)
		}
		if archName != "amd64" && archName != "arm64" {
			return fmt.Errorf("unsupported architecture: %s", archName)
		}

		// 4. Download tarball
		tarball := fmt.Sprintf("clinic_%s_%s_%s.tar.gz", latestClean, osName, archName)
		downloadURL := fmt.Sprintf("https://github.com/jwvictor/clinic/releases/download/%s/%s", latestVersion, tarball)

		fmt.Printf("Downloading %s...\n", downloadURL)

		tmpDir, err := os.MkdirTemp("", "clinic-update-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		tarballPath := filepath.Join(tmpDir, tarball)
		if err := downloadFile(downloadURL, tarballPath); err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		// 5. Extract binary from tarball
		binaryPath, err := extractBinary(tarballPath, tmpDir, "clinic")
		if err != nil {
			return fmt.Errorf("failed to extract archive: %w", err)
		}

		// 6. Replace current binary
		currentBinary, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to determine current binary path: %w", err)
		}
		currentBinary, err = filepath.EvalSymlinks(currentBinary)
		if err != nil {
			return fmt.Errorf("failed to resolve binary path: %w", err)
		}

		if err := replaceBinary(binaryPath, currentBinary); err != nil {
			return fmt.Errorf("failed to replace binary: %w", err)
		}

		fmt.Printf("✓ clinic updated to v%s\n", latestClean)
		return nil
	},
}

// githubRelease is the subset of the GitHub release JSON we care about.
type githubRelease struct {
	TagName string `json:"tag_name"`
}

func fetchLatestVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/jwvictor/clinic/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release JSON: %w", err)
	}

	if release.TagName == "" {
		return "", fmt.Errorf("could not determine latest version from GitHub releases")
	}

	return release.TagName, nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func extractBinary(tarballPath, destDir, binaryName string) (string, error) {
	f, err := os.Open(tarballPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Match the binary by base name
		if filepath.Base(hdr.Name) == binaryName && hdr.Typeflag == tar.TypeReg {
			outPath := filepath.Join(destDir, binaryName)
			out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
			return outPath, nil
		}
	}

	return "", fmt.Errorf("binary %q not found in archive", binaryName)
}

func replaceBinary(newBinary, currentBinary string) error {
	// Read the new binary into memory
	data, err := os.ReadFile(newBinary)
	if err != nil {
		return err
	}

	// Try direct write first
	if err := os.WriteFile(currentBinary, data, 0755); err == nil {
		return nil
	}

	// If direct write fails (e.g. permission denied), try rename approach:
	// write next to the target, then atomic rename
	tmpPath := currentBinary + ".new"
	if err := os.WriteFile(tmpPath, data, 0755); err != nil {
		return fmt.Errorf("cannot write to %s: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, currentBinary); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cannot replace binary (you may need to run with sudo): %w", err)
	}

	return nil
}
