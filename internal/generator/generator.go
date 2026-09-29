package generator

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/izeah01/vibe-gen/internal/filesystem"
	tmpl "github.com/izeah01/vibe-gen/internal/template"
	"github.com/izeah01/vibe-gen/internal/validation"
)

// Templates is the embedded filesystem containing all project templates.
// It is set by main via SetTemplates after the embed directive in the binary.
var Templates embed.FS

// SetTemplates injects the embedded FS. Called once at startup from cmd/vibe-gen/main.go.
func SetTemplates(f embed.FS) {
	Templates = f
}

// Result contains the outcome of a generation run.
type Result struct {
	Destination      string
	FilesWritten     int
	DryRun           bool
	ProjectType      ProjectType
	PostInstallSteps []string
	ScanWarnings     string
}

// Generate creates the project from cfg.
// It renders templates, writes files, then runs the secret scanner.
func Generate(cfg *ProjectConfig) (*Result, error) {
	// Populate derived fields.
	if cfg.Date == "" {
		cfg.Date = now().Format("2006-01-02")
		cfg.Year = now().Format("2006")
	}

	dest, err := filepath.Abs(cfg.Destination)
	if err != nil {
		return nil, fmt.Errorf("resolving destination: %w", err)
	}
	cfg.Destination = dest

	if err := filesystem.ValidateDest(cfg.Destination); err != nil {
		return nil, err
	}

	// Detect if dest already exists and is non-empty.
	dirCreated := false
	if filesystem.DirExists(cfg.Destination) {
		empty, err := filesystem.DirIsEmpty(cfg.Destination)
		if err != nil {
			return nil, fmt.Errorf("checking destination: %w", err)
		}
		if !empty {
			return nil, fmt.Errorf("destination %q already exists and is not empty; use a different path or remove it first", cfg.Destination)
		}
	} else {
		dirCreated = true
	}

	vars := cfg.TemplateVars()
	renderer := tmpl.New(Templates, vars)

	result := &Result{
		Destination: cfg.Destination,
		DryRun:      cfg.DryRun,
		ProjectType: cfg.ProjectType,
	}

	// Track written files for rollback on error.
	var written []string

	writeFile := func(destPath string, content []byte) error {
		if err := filesystem.WriteFile(destPath, content, cfg.DryRun); err != nil {
			return err
		}
		written = append(written, destPath)
		result.FilesWritten++
		return nil
	}

	// Helper to rollback partial writes.
	rollback := func(genErr error) (*Result, error) {
		if !cfg.DryRun {
			filesystem.CleanupDir(cfg.Destination, dirCreated)
		}
		return nil, fmt.Errorf("generation failed: %w", genErr)
	}

	// 1. Render framework-specific templates.
	// Embedded FS always uses forward slashes — do not use filepath.Join here.
	frameworkRoot := "templates/" + cfg.FrameworkID
	if err := renderDir(renderer, Templates, frameworkRoot, cfg.Destination, vars, writeFile); err != nil {
		return rollback(err)
	}

	// 2. Render common doc templates.
	if err := renderCommonDocs(renderer, Templates, cfg, writeFile); err != nil {
		return rollback(err)
	}

	// 3. Post-generation secret scan (skip on dry-run).
	if !cfg.DryRun {
		findings, scanErr := validation.ScanDirectory(cfg.Destination)
		if scanErr == nil && len(findings) > 0 {
			result.ScanWarnings = validation.FormatFindings(findings)
		}
	}

	// 4. Build post-install steps with rendered variables.
	steps, err := renderPostInstallSteps(cfg.PostInstallSteps, vars)
	if err != nil {
		return rollback(err)
	}
	result.PostInstallSteps = steps

	_ = written // used only for rollback tracking
	return result, nil
}

// renderDir walks all .tmpl files under srcRoot in the embedded FS and renders them.
func renderDir(renderer *tmpl.Renderer, fsys fs.FS, srcRoot, destRoot string, vars map[string]string, writeFile func(string, []byte) error) error {
	return tmpl.WalkTemplates(fsys, srcRoot, func(path string) error {
		content, relOut, err := renderer.Render(path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(destRoot, relOut)
		return writeFile(destPath, content)
	})
}

// renderCommonDocs renders the common doc templates that apply to this project type.
// Common template output paths are relative to the project root (e.g. "docs/api.md", ".gitignore").
func renderCommonDocs(renderer *tmpl.Renderer, fsys fs.FS, cfg *ProjectConfig, writeFile func(string, []byte) error) error {
	type docFile struct {
		tmplPath string
		enabled  bool
	}
	docs := []docFile{
		{"templates/common/CLAUDE.md.tmpl", true},
		{"templates/common/README.md.tmpl", true},
		{"templates/common/chat_history.md.tmpl", true},
		{"templates/common/gitignore.tmpl", true},
		{"templates/common/docs/architecture.md.tmpl", cfg.DocProfile.Architecture},
		{"templates/common/docs/coding-standards.md.tmpl", cfg.DocProfile.CodingStandards},
		{"templates/common/docs/security.md.tmpl", cfg.DocProfile.Security},
		{"templates/common/docs/testing.md.tmpl", cfg.DocProfile.Testing},
		{"templates/common/docs/decisions.md.tmpl", cfg.DocProfile.Decisions},
		{"templates/common/docs/roadmap.md.tmpl", cfg.DocProfile.Roadmap},
		{"templates/common/docs/progress.md.tmpl", cfg.DocProfile.Progress},
		{"templates/common/docs/vibe-workflow.md.tmpl", cfg.DocProfile.VibeWorkflow},
		{"templates/common/docs/git.md.tmpl", cfg.DocProfile.Git},
		{"templates/common/docs/observability.md.tmpl", cfg.DocProfile.Observability},
		{"templates/common/docs/performance.md.tmpl", cfg.DocProfile.Performance},
		{"templates/common/docs/api.md.tmpl", cfg.DocProfile.API},
		{"templates/common/docs/database.md.tmpl", cfg.DocProfile.Database},
	}

	for _, doc := range docs {
		if !doc.enabled {
			continue
		}
		// Check if the framework template already provides this file (framework wins).
		if fileExistsInDest(cfg.Destination, outNameFromTmpl(doc.tmplPath)) {
			continue
		}
		content, relOut, err := renderer.Render(doc.tmplPath)
		if err != nil {
			return err
		}
		// gitignore.tmpl → .gitignore (strip prefix already handled by renderer)
		if relOut == "gitignore" {
			relOut = ".gitignore"
		}
		destPath := filepath.Join(cfg.Destination, relOut)
		if err := writeFile(destPath, content); err != nil {
			return err
		}
	}
	return nil
}

// outNameFromTmpl derives the output filename from a template path.
func outNameFromTmpl(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ".tmpl")
	if base == "gitignore" {
		return ".gitignore"
	}
	return base
}

// fileExistsInDest checks if a file already exists in the destination.
func fileExistsInDest(dest, filename string) bool {
	_, err := os.Stat(filepath.Join(dest, filename))
	return err == nil
}

// renderPostInstallSteps applies template vars to each step string.
func renderPostInstallSteps(steps []string, vars map[string]string) ([]string, error) {
	rendered := make([]string, 0, len(steps))
	for _, s := range steps {
		out, err := tmpl.RenderString(s, vars)
		if err != nil {
			return nil, err
		}
		rendered = append(rendered, out)
	}
	return rendered, nil
}

// SuggestModuleName derives a Go module name suggestion from the working directory and project name.
// Strategy: if pwd is under GOPATH/src, use the relative path. Otherwise use project name.
func SuggestModuleName(projectName string) string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, _ := os.UserHomeDir()
		gopath = filepath.Join(home, "go")
	}
	srcRoot := filepath.Join(gopath, "src")

	wd, err := os.Getwd()
	if err != nil {
		return projectName
	}

	rel, err := filepath.Rel(srcRoot, wd)
	if err != nil || strings.HasPrefix(rel, "..") {
		// Not under GOPATH/src; use the last folder of pwd as module suggestion.
		base := filepath.Base(wd)
		return base + "/" + projectName
	}
	return rel + "/" + projectName
}

// FormatPostInstall returns the formatted post-generation message.
func FormatPostInstall(r *Result) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n✓ Project generated: %s\n", r.Destination))
	if r.DryRun {
		sb.WriteString("  (dry-run: no files were written)\n")
	} else {
		sb.WriteString(fmt.Sprintf("  %d files written\n", r.FilesWritten))
	}
	if len(r.PostInstallSteps) > 0 {
		sb.WriteString("\n──────────────────────────────────\n")
		sb.WriteString(" Next steps\n")
		sb.WriteString("──────────────────────────────────\n")
		for _, step := range r.PostInstallSteps {
			sb.WriteString(fmt.Sprintf(" $ %s\n", step))
		}
		sb.WriteString("──────────────────────────────────\n")
		sb.WriteString(" Documentation: docs/\n")
		if r.ProjectType == ProjectTypeAPI {
			sb.WriteString(" Swagger UI:    http://localhost:8080/swagger/index.html (run `make swag`)\n")
		}
		sb.WriteString(" AI Context:    CLAUDE.md\n")
		sb.WriteString("──────────────────────────────────\n")
	}
	if r.ScanWarnings != "" {
		sb.WriteString("\n")
		sb.WriteString(r.ScanWarnings)
	}
	return sb.String()
}

// ensure time package is referenced (used by now() in model.go).
var _ = time.RFC3339
