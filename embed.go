// Package vibegen is the root package for the vibe-gen tool.
// It provides the embedded templates filesystem.
package vibegen

import "embed"

// Templates contains all embedded project templates.
// The embed directive is here at module root so "templates" path is valid.
//
//go:embed all:templates
var Templates embed.FS
