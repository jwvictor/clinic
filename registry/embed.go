// Package registrydata provides embedded access to the JSON registry files.
// This exists at the repo root level so go:embed can reference the JSON files
// in the same directory tree.
package registrydata

import "embed"

// Index is the raw bytes of index.json, compiled into the binary.
//
//go:embed index.json
var Index []byte

// Tools contains all tool JSON files, compiled into the binary.
//
//go:embed tools/*.json
var Tools embed.FS

// Stacks contains all stack JSON files, compiled into the binary.
//
//go:embed stacks/*.json
var Stacks embed.FS
