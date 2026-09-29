package filesystem

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	errEmptyDest   = errors.New("destination path cannot be empty")
	errPathTraversal = errors.New("destination path contains unsafe traversal sequence")
)

// ValidateDest checks that the destination path is safe to write to.
func ValidateDest(dest string) error {
	if strings.TrimSpace(dest) == "" {
		return errEmptyDest
	}
	cleaned := filepath.Clean(dest)
	// Reject paths that try to escape via traversal after cleaning.
	if strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || cleaned == ".." {
		return errPathTraversal
	}
	return nil
}

// SafeJoin joins base and rel, returning an error if the result escapes base.
func SafeJoin(base, rel string) (string, error) {
	joined := filepath.Join(base, rel)
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", fmt.Errorf("resolving base path: %w", err)
	}
	absJoined, err := filepath.Abs(joined)
	if err != nil {
		return "", fmt.Errorf("resolving joined path: %w", err)
	}
	if !strings.HasPrefix(absJoined, absBase+string(filepath.Separator)) && absJoined != absBase {
		return "", fmt.Errorf("path %q escapes destination root %q", rel, base)
	}
	return absJoined, nil
}
