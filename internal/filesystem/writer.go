package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile writes content to destPath, creating parent directories as needed.
// If dryRun is true, it only prints what would be written without touching disk.
func WriteFile(destPath string, content []byte, dryRun bool) error {
	if dryRun {
		fmt.Printf("  [dry-run] would write: %s\n", destPath)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("creating directory for %s: %w", destPath, err)
	}
	if err := os.WriteFile(destPath, content, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", destPath, err)
	}
	return nil
}

// DirExists returns true if path exists and is a directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// DirIsEmpty returns true if path is an empty directory.
func DirIsEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

// CleanupDir removes a directory tree. Used to rollback partial generation.
// It only removes destDir if it was created by this generation (tracked via created flag).
func CleanupDir(destDir string, created bool) error {
	if !created {
		return nil // never remove pre-existing directories
	}
	return os.RemoveAll(destDir)
}
