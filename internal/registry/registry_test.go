package registry

import (
	"encoding/json"
	"io/fs"
	"testing"
	"testing/fstest"

	registrydata "github.com/jwvictor/clinic/registry"
)

func TestLoadReturnsNonEmptyRegistry(t *testing.T) {
	// Load uses embedded fallback if cache/remote are unavailable
	reg := Load()
	if reg == nil {
		t.Fatal("Load() returned nil")
	}
	if len(reg.Tools) == 0 {
		t.Error("Load() returned registry with no tools")
	}
	if len(reg.Stacks) == 0 {
		t.Error("Load() returned registry with no stacks")
	}
}

func TestGetToolValid(t *testing.T) {
	reg := Load()

	tool, ok := reg.GetTool("gh")
	if !ok {
		t.Fatal("GetTool(\"gh\") not found")
	}
	if tool.Name != "gh" {
		t.Errorf("tool.Name = %q, want %q", tool.Name, "gh")
	}
	if tool.Command == "" {
		t.Error("tool.Command is empty")
	}
	if tool.Description == "" {
		t.Error("tool.Description is empty")
	}
}

func TestGetToolInvalid(t *testing.T) {
	reg := Load()

	_, ok := reg.GetTool("_nonexistent_tool_xyz")
	if ok {
		t.Error("GetTool() returned ok=true for nonexistent tool")
	}
}

func TestGetStackValid(t *testing.T) {
	reg := Load()

	stack, ok := reg.GetStack("basics")
	if !ok {
		t.Fatal("GetStack(\"basics\") not found")
	}
	if stack.Name != "basics" {
		t.Errorf("stack.Name = %q, want %q", stack.Name, "basics")
	}
	if len(stack.Tools) == 0 {
		t.Error("basics stack has no tools")
	}
}

func TestGetStackInvalid(t *testing.T) {
	reg := Load()

	_, ok := reg.GetStack("_nonexistent_stack_xyz")
	if ok {
		t.Error("GetStack() returned ok=true for nonexistent stack")
	}
}

func TestToolNamesAreSorted(t *testing.T) {
	reg := Load()
	names := reg.ToolNames()
	if len(names) == 0 {
		t.Fatal("ToolNames() returned empty slice")
	}
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("ToolNames() not sorted: %q before %q", names[i-1], names[i])
		}
	}
}

func TestStackNamesAreSorted(t *testing.T) {
	reg := Load()
	names := reg.StackNames()
	if len(names) == 0 {
		t.Fatal("StackNames() returned empty slice")
	}
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("StackNames() not sorted: %q before %q", names[i-1], names[i])
		}
	}
}

func TestAllToolsInIndexCanBeLoaded(t *testing.T) {
	// Parse the embedded index to get tool names
	var idx RegistryIndex
	if err := json.Unmarshal(registrydata.Index, &idx); err != nil {
		t.Fatalf("parsing embedded index.json: %v", err)
	}

	reg := Load()

	for _, toolName := range idx.Tools {
		tool, ok := reg.GetTool(toolName)
		if !ok {
			t.Errorf("tool %q listed in index.json but not loadable", toolName)
			continue
		}
		if tool.Name == "" {
			t.Errorf("tool %q has empty Name field", toolName)
		}
	}
}

func TestAllStacksInIndexCanBeLoaded(t *testing.T) {
	var idx RegistryIndex
	if err := json.Unmarshal(registrydata.Index, &idx); err != nil {
		t.Fatalf("parsing embedded index.json: %v", err)
	}

	reg := Load()

	for _, stackName := range idx.Stacks {
		stack, ok := reg.GetStack(stackName)
		if !ok {
			t.Errorf("stack %q listed in index.json but not loadable", stackName)
			continue
		}
		if stack.Name == "" {
			t.Errorf("stack %q has empty Name field", stackName)
		}
	}
}

func TestParseRegistryWithSyntheticData(t *testing.T) {
	indexJSON := []byte(`{"schema_version":1,"tools":["mytool"],"stacks":["mystack"]}`)

	toolJSON := []byte(`{
		"name":"mytool",
		"command":"mytool",
		"description":"A test tool",
		"category":"testing",
		"version_command":"mytool --version",
		"install_methods":[],
		"auth":{"inject_type":"none"}
	}`)
	stackJSON := []byte(`{
		"name":"mystack",
		"description":"A test stack",
		"tools":["mytool"]
	}`)

	toolsFS := fstest.MapFS{
		"mytool.json": &fstest.MapFile{Data: toolJSON},
	}
	stacksFS := fstest.MapFS{
		"mystack.json": &fstest.MapFile{Data: stackJSON},
	}

	reg, err := parseRegistry(indexJSON, fs.FS(toolsFS), fs.FS(stacksFS))
	if err != nil {
		t.Fatalf("parseRegistry() error: %v", err)
	}

	if len(reg.Tools) != 1 {
		t.Fatalf("len(Tools) = %d, want 1", len(reg.Tools))
	}
	tool, ok := reg.GetTool("mytool")
	if !ok {
		t.Fatal("mytool not found")
	}
	if tool.Description != "A test tool" {
		t.Errorf("tool.Description = %q, want %q", tool.Description, "A test tool")
	}

	if len(reg.Stacks) != 1 {
		t.Fatalf("len(Stacks) = %d, want 1", len(reg.Stacks))
	}
	stack, ok := reg.GetStack("mystack")
	if !ok {
		t.Fatal("mystack not found")
	}
	if len(stack.Tools) != 1 || stack.Tools[0] != "mytool" {
		t.Errorf("stack.Tools = %v, want [mytool]", stack.Tools)
	}
}

func TestParseRegistrySkipsMissingFiles(t *testing.T) {
	indexJSON := []byte(`{"schema_version":1,"tools":["exists","missing"],"stacks":[]}`)

	toolJSON := []byte(`{"name":"exists","command":"exists","description":"exists"}`)
	toolsFS := fstest.MapFS{
		"exists.json": &fstest.MapFile{Data: toolJSON},
		// "missing.json" intentionally absent
	}
	stacksFS := fstest.MapFS{}

	reg, err := parseRegistry(indexJSON, fs.FS(toolsFS), fs.FS(stacksFS))
	if err != nil {
		t.Fatalf("parseRegistry() error: %v", err)
	}

	if len(reg.Tools) != 1 {
		t.Errorf("len(Tools) = %d, want 1 (should skip missing)", len(reg.Tools))
	}
}
