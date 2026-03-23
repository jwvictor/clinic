package versioncheck

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	versionURL   = "https://getclinic.sh/api/version"
	cacheTTL     = 24 * time.Hour
	fetchTimeout = 2 * time.Second
)

type versionManifest struct {
	Latest       string `json:"latest"`
	MinSupported string `json:"min_supported"`
}

type cachedCheck struct {
	Latest    string `json:"latest"`
	CheckedAt int64  `json:"checked_at"`
}

func cachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".clinic", "version-check.json")
}

// CheckAndNotify prints a message to stderr if a newer version of clinic is
// available. It fetches the remote manifest at most once per 24 hours and
// caches the result locally. Network failures are silently ignored — this
// must never block or break normal command execution.
func CheckAndNotify(currentVersion string) {
	// Don't nag on dev builds
	if currentVersion == "" || strings.HasSuffix(currentVersion, "-dev") {
		return
	}

	latest, ok := getLatestVersion()
	if !ok || latest == "" {
		return
	}

	if !isNewer(latest, currentVersion) {
		return
	}

	fmt.Fprintf(os.Stderr, "\nA newer version of clinic is available (v%s → v%s). Run \"clinic self-update\" to upgrade.\n", currentVersion, latest)
}

// getLatestVersion returns the latest version string, using cache if fresh.
func getLatestVersion() (string, bool) {
	// Try cache first
	if cached, ok := loadCache(); ok {
		if time.Since(time.Unix(cached.CheckedAt, 0)) < cacheTTL {
			return cached.Latest, true
		}
	}

	// Fetch with a short timeout — never block the user
	client := &http.Client{Timeout: fetchTimeout}
	resp, err := client.Get(versionURL)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", false
	}

	var manifest versionManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return "", false
	}

	// Cache the result
	saveCache(cachedCheck{
		Latest:    manifest.Latest,
		CheckedAt: time.Now().Unix(),
	})

	return manifest.Latest, true
}

func loadCache() (cachedCheck, bool) {
	data, err := os.ReadFile(cachePath())
	if err != nil {
		return cachedCheck{}, false
	}
	var c cachedCheck
	if err := json.Unmarshal(data, &c); err != nil {
		return cachedCheck{}, false
	}
	return c, true
}

func saveCache(c cachedCheck) {
	data, err := json.Marshal(c)
	if err != nil {
		return
	}
	os.MkdirAll(filepath.Dir(cachePath()), 0755)
	os.WriteFile(cachePath(), data, 0644)
}

// isNewer returns true if latest is a higher version than current.
// Both should be bare version strings like "0.1.6" (no "v" prefix).
func isNewer(latest, current string) bool {
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	lp := parseVersion(latest)
	cp := parseVersion(current)

	for i := 0; i < 3; i++ {
		if lp[i] > cp[i] {
			return true
		}
		if lp[i] < cp[i] {
			return false
		}
	}
	return false
}

func parseVersion(v string) [3]int {
	var parts [3]int
	segments := strings.SplitN(v, ".", 3)
	for i, s := range segments {
		if i >= 3 {
			break
		}
		// Strip any suffix like "-dev", "-rc1"
		s = strings.SplitN(s, "-", 2)[0]
		fmt.Sscanf(s, "%d", &parts[i])
	}
	return parts
}
