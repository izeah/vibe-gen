package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ScanResult holds findings from secret scanning.
type ScanResult struct {
	File    string
	Line    int
	Pattern string
	Excerpt string
}

// secretPatterns lists regex patterns that indicate potential hardcoded secrets.
var secretPatterns = []*secretRule{
	{name: "Password assignment", pattern: regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*["'][^"'${\s]{4,}["']`)},
	{name: "API key assignment", pattern: regexp.MustCompile(`(?i)(api[_\-]?key|apikey)\s*[:=]\s*["'][^"'${\s]{8,}["']`)},
	{name: "Secret/token assignment", pattern: regexp.MustCompile(`(?i)(secret|token|access[_\-]?token)\s*[:=]\s*["'][^"'${\s]{8,}["']`)},
	{name: "Private key block", pattern: regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`)},
	{name: "JWT-like string", pattern: regexp.MustCompile(`eyJ[A-Za-z0-9_\-]{10,}\.[A-Za-z0-9_\-]{10,}\.[A-Za-z0-9_\-]{10,}`)},
	{name: "AWS key pattern", pattern: regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{name: "Stripe key pattern", pattern: regexp.MustCompile(`sk_(live|test)_[0-9a-zA-Z]{24,}`)},
}

// safeValuePatterns matches values that are clearly placeholders — safe to ignore.
var safeValuePatterns = []*regexp.Regexp{
	regexp.MustCompile(`^["']?<[^>]+>["']?$`),            // <placeholder>
	regexp.MustCompile(`^["']?\{\{[^}]+\}\}["']?$`),       // {{VARIABLE}}
	regexp.MustCompile(`(?i)^["']?(your[_\-]?|change[_\-]?me|todo|fixme|xxx|placeholder|example|sample|test|local|demo|dummy)`),
	regexp.MustCompile(`^["']?$`),                          // empty string
}

type secretRule struct {
	name    string
	pattern *regexp.Regexp
}

// ScanDirectory walks destDir and scans all files for secret patterns.
// Returns a list of findings and an error if scanning itself failed.
func ScanDirectory(destDir string) ([]ScanResult, error) {
	var findings []ScanResult

	err := filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable files
		}
		if info.IsDir() {
			// Skip .git and node_modules directories.
			base := info.Name()
			if base == ".git" || base == "node_modules" || base == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		// Only scan text-like files.
		if !isTextFile(path) {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			for _, rule := range secretPatterns {
				match := rule.pattern.FindString(line)
				if match == "" {
					continue
				}
				// Check if the matched value looks like a safe placeholder.
				if isSafeValue(match) {
					continue
				}
				rel, _ := filepath.Rel(destDir, path)
				findings = append(findings, ScanResult{
					File:    rel,
					Line:    lineNum + 1,
					Pattern: rule.name,
					Excerpt: strings.TrimSpace(line),
				})
			}
		}
		return nil
	})

	return findings, err
}

// FormatFindings returns a human-readable summary of scan results.
func FormatFindings(findings []ScanResult) string {
	if len(findings) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("⚠  Secret scan warnings:\n")
	for _, f := range findings {
		sb.WriteString(fmt.Sprintf("   %s:%d  [%s]\n   %s\n\n", f.File, f.Line, f.Pattern, f.Excerpt))
	}
	return sb.String()
}

// isTextFile returns true for file extensions that should be scanned.
func isTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	textExts := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".json": true, ".yaml": true, ".yml": true, ".toml": true,
		".env": true, ".example": true, ".md": true, ".txt": true,
		".sh": true, ".makefile": true, "": true,
	}
	// Also match files with no extension by name.
	base := strings.ToLower(filepath.Base(path))
	noExtFiles := map[string]bool{
		"makefile": true, "dockerfile": true, ".env": true,
		".env.example": true, ".gitignore": true,
	}
	return textExts[ext] || noExtFiles[base]
}

// isSafeValue returns true if the match looks like a known safe placeholder.
func isSafeValue(match string) bool {
	val := strings.TrimSpace(match)
	if idx := strings.IndexAny(match, ":="); idx != -1 {
		val = strings.TrimSpace(match[idx+1:])
	}
	for _, p := range safeValuePatterns {
		if p.MatchString(val) || p.MatchString(match) {
			return true
		}
	}
	return false
}
