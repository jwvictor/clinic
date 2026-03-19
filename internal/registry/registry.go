package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/togglemedia/clinic/internal/config"
	registrydata "github.com/togglemedia/clinic/registry"
)

const (
	// remoteBaseURL is the raw GitHub URL for the registry directory.
	remoteBaseURL = "https://raw.githubusercontent.com/jwvictor/clinic/main/registry"

	// cacheTTL controls how often we check for remote updates.
	cacheTTL = 24 * time.Hour

	// fetchTimeout is the HTTP timeout for remote fetches.
	fetchTimeout = 5 * time.Second

	// currentSchemaVersion is the max schema version this binary understands.
	currentSchemaVersion = 1
)

// RegistryIndex is the parsed index.json manifest.
type RegistryIndex struct {
	SchemaVersion int      `json:"schema_version"`
	Tools         []string `json:"tools"`
	Stacks        []string `json:"stacks"`
}

// Registry holds all known tool and stack definitions.
type Registry struct {
	Tools  map[string]ToolDef
	Stacks map[string]StackDef
}

// Load returns the registry, trying cache → remote → embedded fallback.
func Load() *Registry {
	// 1. Try fresh cache
	if !isCacheStale() {
		if reg, err := loadFromCache(); err == nil {
			return reg
		}
	}

	// 2. Try remote fetch (writes to cache on success)
	if reg, err := fetchAndCache(); err == nil {
		return reg
	}

	// 3. Fall back to compiled-in embedded registry
	reg, _ := loadFromEmbedded()
	return reg
}

// ForceRefresh fetches the remote registry regardless of cache freshness.
func ForceRefresh() (*Registry, error) {
	return fetchAndCache()
}

// GetTool returns a tool definition by name.
func (r *Registry) GetTool(name string) (ToolDef, bool) {
	t, ok := r.Tools[name]
	return t, ok
}

// GetStack returns a stack definition by name.
func (r *Registry) GetStack(name string) (StackDef, bool) {
	s, ok := r.Stacks[name]
	return s, ok
}

// ToolNames returns a sorted list of all available tool names.
func (r *Registry) ToolNames() []string {
	names := make([]string, 0, len(r.Tools))
	for name := range r.Tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// StackNames returns a sorted list of all available stack names.
func (r *Registry) StackNames() []string {
	names := make([]string, 0, len(r.Stacks))
	for name := range r.Stacks {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// --- Cache ---

func cacheDir() string {
	return filepath.Join(config.CacheDir(), "registry")
}

func isCacheStale() bool {
	info, err := os.Stat(filepath.Join(cacheDir(), "index.json"))
	if err != nil {
		return true
	}
	return time.Since(info.ModTime()) > cacheTTL
}

func loadFromCache() (*Registry, error) {
	dir := cacheDir()
	indexBytes, err := os.ReadFile(filepath.Join(dir, "index.json"))
	if err != nil {
		return nil, err
	}
	toolsDir := os.DirFS(filepath.Join(dir, "tools"))
	stacksDir := os.DirFS(filepath.Join(dir, "stacks"))
	return parseRegistry(indexBytes, toolsDir, stacksDir)
}

// --- Remote fetch ---

func fetchAndCache() (*Registry, error) {
	client := &http.Client{Timeout: fetchTimeout}

	// Fetch and validate index
	indexBytes, err := httpGet(client, remoteBaseURL+"/index.json")
	if err != nil {
		return nil, fmt.Errorf("fetch index: %w", err)
	}

	var idx RegistryIndex
	if err := json.Unmarshal(indexBytes, &idx); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}
	if idx.SchemaVersion > currentSchemaVersion {
		return nil, fmt.Errorf("registry schema version %d is newer than this binary supports (%d) — please update clinic", idx.SchemaVersion, currentSchemaVersion)
	}

	// Prepare cache directories
	dir := cacheDir()
	os.MkdirAll(filepath.Join(dir, "tools"), 0755)
	os.MkdirAll(filepath.Join(dir, "stacks"), 0755)

	if err := os.WriteFile(filepath.Join(dir, "index.json"), indexBytes, 0644); err != nil {
		return nil, fmt.Errorf("cache index: %w", err)
	}

	// Fetch and cache all tools
	for _, name := range idx.Tools {
		data, err := httpGet(client, fmt.Sprintf("%s/tools/%s.json", remoteBaseURL, name))
		if err != nil {
			return nil, fmt.Errorf("fetch tool %s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "tools", name+".json"), data, 0644); err != nil {
			return nil, fmt.Errorf("cache tool %s: %w", name, err)
		}
	}

	// Fetch and cache all stacks
	for _, name := range idx.Stacks {
		data, err := httpGet(client, fmt.Sprintf("%s/stacks/%s.json", remoteBaseURL, name))
		if err != nil {
			return nil, fmt.Errorf("fetch stack %s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "stacks", name+".json"), data, 0644); err != nil {
			return nil, fmt.Errorf("cache stack %s: %w", name, err)
		}
	}

	return loadFromCache()
}

func httpGet(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

// --- Embedded fallback ---

func loadFromEmbedded() (*Registry, error) {
	// embed.FS uses paths like "tools/gh.json" relative to the embed root
	toolsSub, err := fs.Sub(registrydata.Tools, "tools")
	if err != nil {
		return nil, err
	}
	stacksSub, err := fs.Sub(registrydata.Stacks, "stacks")
	if err != nil {
		return nil, err
	}
	return parseRegistry(registrydata.Index, toolsSub, stacksSub)
}

// --- Shared loading logic ---

// parseRegistry builds a Registry from an index and two FS dirs containing tool/stack JSON files.
// Both toolsFS and stacksFS should be rooted so that files are at "name.json" (no subdirectory prefix).
func parseRegistry(indexBytes []byte, toolsFS fs.FS, stacksFS fs.FS) (*Registry, error) {
	var idx RegistryIndex
	if err := json.Unmarshal(indexBytes, &idx); err != nil {
		return nil, err
	}

	reg := &Registry{
		Tools:  make(map[string]ToolDef, len(idx.Tools)),
		Stacks: make(map[string]StackDef, len(idx.Stacks)),
	}

	for _, name := range idx.Tools {
		data, err := fs.ReadFile(toolsFS, name+".json")
		if err != nil {
			continue
		}
		var t ToolDef
		if err := json.Unmarshal(data, &t); err != nil {
			continue
		}
		reg.Tools[t.Name] = t
	}

	for _, name := range idx.Stacks {
		data, err := fs.ReadFile(stacksFS, name+".json")
		if err != nil {
			continue
		}
		var s StackDef
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		reg.Stacks[s.Name] = s
	}

	return reg, nil
}
