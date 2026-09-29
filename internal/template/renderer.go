package tmpl

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"
)

// TemplateData is a dynamic struct that allows {{.Key}} access on a map[string]string
// via a custom FuncMap approach. We convert the map to a map[string]interface{}
// so that text/template's field lookup works with {{.Key}} syntax.
type TemplateData map[string]interface{}

// Renderer renders Go text/template files with a variable map.
type Renderer struct {
	fs   fs.FS
	data TemplateData
}

// New creates a Renderer with the given embedded FS and template variables.
func New(fsys fs.FS, vars map[string]string) *Renderer {
	data := make(TemplateData, len(vars))
	for k, v := range vars {
		data[k] = v
	}
	return &Renderer{fs: fsys, data: data}
}

// Render renders the template at srcPath within the embedded FS.
// The output filename is derived by stripping the ".tmpl" suffix from srcPath.
// Returns rendered content and the output file's relative path.
func (r *Renderer) Render(srcPath string) ([]byte, string, error) {
	content, err := fs.ReadFile(r.fs, srcPath)
	if err != nil {
		return nil, "", fmt.Errorf("reading template %s: %w", srcPath, err)
	}

	tmplName := filepath.Base(srcPath)
	t, err := template.New(tmplName).Option("missingkey=error").Parse(string(content))
	if err != nil {
		return nil, "", fmt.Errorf("parsing template %s: %w", srcPath, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, r.data); err != nil {
		return nil, "", fmt.Errorf("executing template %s: %w", srcPath, err)
	}

	// Derive output path by stripping .tmpl suffix.
	outPath := strings.TrimSuffix(srcPath, ".tmpl")
	// Strip the "templates/<lang>/<framework>/" prefix to get project-relative path.
	outPath = stripTemplatePrefix(outPath)

	return buf.Bytes(), outPath, nil
}

// RenderString renders a template string directly (used for PostInstallSteps).
func RenderString(tmplStr string, vars map[string]string) (string, error) {
	data := make(TemplateData, len(vars))
	for k, v := range vars {
		data[k] = v
	}
	t, err := template.New("inline").Option("missingkey=error").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// stripTemplatePrefix removes the leading "templates/<lang>/<framework>/" from a path.
// For common templates it removes "templates/common/".
// Examples:
//
//	"templates/go/echo/internal/config/type.go"  → "internal/config/type.go"
//	"templates/go/cron-job/main.go"              → "main.go"
//	"templates/common/docs/api.md"               → "docs/api.md"
func stripTemplatePrefix(path string) string {
	// Use forward slash splitting always (embedded FS uses /).
	parts := strings.SplitN(path, "/", 5)
	// parts[0] = "templates"
	// parts[1] = "<lang>"  (e.g. "go", "javascript", "common")
	// parts[2] = "<framework>" (e.g. "echo", "gin", "react")  OR for common: "docs" etc.
	// parts[3+] = rest of path
	if len(parts) < 2 {
		return path
	}
	if parts[1] == "common" {
		// templates/common/<rest>
		if len(parts) >= 3 {
			return strings.Join(parts[2:], "/")
		}
	}
	// templates/<lang>/<framework>/<rest>
	if len(parts) >= 4 {
		return strings.Join(parts[3:], "/")
	}
	if len(parts) == 3 {
		return parts[2]
	}
	return path
}

// WalkTemplates calls fn for each .tmpl file under rootPath in the embedded FS.
func WalkTemplates(fsys fs.FS, rootPath string, fn func(path string) error) error {
	return fs.WalkDir(fsys, rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".tmpl") {
			return nil
		}
		return fn(path)
	})
}
