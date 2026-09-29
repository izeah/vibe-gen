package filesystem

import (
	"fmt"
	"io/fs"
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

// CopyFS copies all files from src (an fs.FS) under srcRoot to destDir.
// It respects SafeJoin to prevent path traversal.
func CopyFS(src fs.FS, srcRoot, destDir string, dryRun bool) error {
	return fs.WalkDir(src, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Compute relative path from srcRoot.
		rel, relErr := filepath.Rel(srcRoot, path)
		if relErr != nil {
			return relErr
		}
		dest, joinErr := SafeJoin(destDir, rel)
		if joinErr != nil {
			return joinErr
		}
		if d.IsDir() {
			if dryRun {
				fmt.Printf("  [dry-run] would mkdir: %s\n", dest)
				return nil
			}
			return os.MkdirAll(dest, 0o755)
		}
		data, readErr := fs.ReadFile(src, path)
		if readErr != nil {
			return fmt.Errorf("reading embedded file %s: %w", path, readErr)
		}
		return WriteFile(dest, data, dryRun)
	})
}
