package validation

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var (
	// validNamePattern matches safe project names: lowercase letters, digits, hyphens, underscores.
	validNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\-]{0,63}$`)
	// traversalPattern catches path traversal attempts.
	traversalPattern = regexp.MustCompile(`(\.\.|/|\\|:)`)

	errEmptyName     = errors.New("project name cannot be empty")
	errNameTooLong   = errors.New("project name must be 64 characters or fewer")
	errInvalidChars  = errors.New("project name must start with a letter and contain only letters, digits, hyphens, or underscores")
	errPathTraversal = errors.New("project name contains unsafe path characters")
	errReservedName  = errors.New("project name is a reserved word")

	// validModulePattern: slash-separated path elements of letters, digits and ._~- (Go module path rules,
	// minus the rarely-used corners); keeps newlines/spaces out of go.mod and import paths.
	validModulePattern = regexp.MustCompile(`^[A-Za-z0-9_~-][A-Za-z0-9._~-]*(/[A-Za-z0-9_~-][A-Za-z0-9._~-]*)*$`)
	errInvalidModule   = errors.New("module name must be a Go module path like github.com/acme/app (letters, digits, ._~- separated by /)")
)

// reserved names that should not be used as project names.
var reservedNames = map[string]bool{
	"con": true, "prn": true, "aux": true, "nul": true,
	"com1": true, "com2": true, "com3": true, "com4": true,
	"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true,
	".": true, "..": true,
}

// ValidateName checks that the project name is safe and valid.
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errEmptyName
	}
	if len(name) > 64 {
		return errNameTooLong
	}
	if traversalPattern.MatchString(name) {
		return errPathTraversal
	}
	if reservedNames[strings.ToLower(name)] {
		return errReservedName
	}
	if !validNamePattern.MatchString(name) {
		return errInvalidChars
	}
	return nil
}

// ValidateModule checks that a Go module path is safe to write into go.mod and imports.
func ValidateModule(path string) error {
	if len(path) > 255 || !validModulePattern.MatchString(path) {
		return errInvalidModule
	}
	return nil
}

// ToSlug converts a project name to kebab-case slug.
// Example: "MyPaymentAPI" → "my-payment-api"
func ToSlug(name string) string {
	var sb strings.Builder
	runes := []rune(name)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 && !unicode.IsUpper(runes[i-1]) {
				sb.WriteRune('-')
			} else if i > 0 && i+1 < len(runes) && unicode.IsLower(runes[i+1]) && unicode.IsUpper(runes[i-1]) {
				sb.WriteRune('-')
			}
			sb.WriteRune(unicode.ToLower(r))
		} else if r == '_' || r == ' ' {
			sb.WriteRune('-')
		} else {
			sb.WriteRune(r)
		}
	}
	slug := sb.String()
	// Collapse multiple hyphens.
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return strings.Trim(slug, "-")
}

// ToTitle converts a project name to Title Case.
// Example: "my-payment-api" → "My Payment Api"
func ToTitle(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(string(w[0])) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}
