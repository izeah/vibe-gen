package generator

import (
	"embed"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	Language         Language
	PostInstallSteps []string
	ScanWarnings     string
	GitStatus        string
}

// Generate creates the project from cfg.
// It renders templates, writes files, then runs the secret scanner.
func Generate(cfg *ProjectConfig) (*Result, error) {
	// Populate derived fields.
	if cfg.Date == "" {
		cfg.Date = now().Format("2006-01-02")
		cfg.Year = now().Format("2006")
	}

	if cfg.Language == LanguageGo {
		if err := validation.ValidateModule(cfg.ModuleName); err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(cfg.Destination) == "" {
		return nil, fmt.Errorf("destination path cannot be empty")
	}
	dest, err := filepath.Abs(cfg.Destination)
	if err != nil {
		return nil, fmt.Errorf("resolving destination: %w", err)
	}
	cfg.Destination = dest

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
		Language:    cfg.Language,
	}

	// Files already written: shared and common templates never overwrite framework files.
	seen := map[string]bool{}

	writeFile := func(destPath string, content []byte) error {
		// gofmt generated Go so `make fmt-check` passes out of the box; keep raw output if it doesn't parse.
		if strings.HasSuffix(destPath, ".go") {
			if formatted, err := format.Source(content); err == nil {
				content = formatted
			}
		}
		if err := filesystem.WriteFile(destPath, content, cfg.DryRun); err != nil {
			return err
		}
		seen[destPath] = true
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
	if cfg.Variant != "" {
		frameworkRoot += "/" + cfg.Variant
	}
	if err := renderDir(renderer, Templates, frameworkRoot, cfg.Destination, vars, writeFile); err != nil {
		return rollback(err)
	}

	// 1b. Render shared templates (e.g. the Node API core); framework files win.
	for _, shared := range cfg.Shared {
		writeNew := func(destPath string, content []byte) error {
			if seen[destPath] {
				return nil
			}
			return writeFile(destPath, content)
		}
		if err := renderDir(renderer, Templates, "templates/"+string(cfg.Language)+"/"+shared, cfg.Destination, vars, writeNew); err != nil {
			return rollback(err)
		}
	}

	// 2. Render common doc templates.
	if err := renderCommonDocs(renderer, cfg, seen, writeFile); err != nil {
		return rollback(err)
	}

	// 3. Post-generation secret scan (skip on dry-run).
	if !cfg.DryRun {
		findings, scanErr := validation.ScanDirectory(cfg.Destination)
		if scanErr == nil && len(findings) > 0 {
			result.ScanWarnings = validation.FormatFindings(findings)
		}
	}

	// 4. Make sure the project is a git repository (vibe workflow relies on git diff/status).
	if !cfg.DryRun {
		result.GitStatus = ensureGitRepo(cfg.Destination)
	}

	// 5. Build post-install steps with rendered variables.
	steps, err := renderPostInstallSteps(cfg.PostInstallSteps, vars)
	if err != nil {
		return rollback(err)
	}
	result.PostInstallSteps = steps

	return result, nil
}

// renderDir walks all .tmpl files under srcRoot in the embedded FS and renders them.
func renderDir(renderer *tmpl.Renderer, fsys fs.FS, srcRoot, destRoot string, vars map[string]string, writeFile func(string, []byte) error) error {
	return tmpl.WalkTemplates(fsys, srcRoot, func(path string) error {
		content, err := renderer.Render(path)
		if err != nil {
			return err
		}
		// Output path is relative to srcRoot, so variant roots (templates/<lang>/<fw>/<variant>) work too.
		rel := strings.TrimSuffix(strings.TrimPrefix(path, srcRoot+"/"), ".tmpl")
		destPath := filepath.Join(destRoot, filepath.FromSlash(rel))
		return writeFile(destPath, content)
	})
}

// renderCommonDocs renders the common doc templates that apply to this project type.
// Common template output paths are relative to the project root (e.g. "docs/api.md", ".gitignore").
func renderCommonDocs(renderer *tmpl.Renderer, cfg *ProjectConfig, seen map[string]bool, writeFile func(string, []byte) error) error {
	type docFile struct {
		tmplPath string
		enabled  bool
	}
	docs := []docFile{
		{"templates/common/CLAUDE.md.tmpl", true},
		{"templates/common/README.md.tmpl", true},
		{"templates/common/chat_history.md.tmpl", true},
		{"templates/common/gitignore.tmpl", true},
		{"templates/common/npmrc.tmpl", cfg.Language != LanguageGo},
		{"templates/common/.claude/settings.json.tmpl", true},
		{"templates/common/docs/architecture.md.tmpl", cfg.DocProfile.Architecture},
		{"templates/common/docs/coding-standards.md.tmpl", cfg.DocProfile.CodingStandards},
		{"templates/common/docs/security.md.tmpl", cfg.DocProfile.Security},
		{"templates/common/docs/testing.md.tmpl", cfg.DocProfile.Testing},
		{"templates/common/docs/decisions.md.tmpl", cfg.DocProfile.Decisions},
		{"templates/common/docs/roadmap.md.tmpl", cfg.DocProfile.Roadmap},
		{"templates/common/docs/progress.md.tmpl", cfg.DocProfile.Progress},
		{"templates/common/docs/task.md.tmpl", true},
		{"templates/common/docs/vibe-workflow.md.tmpl", cfg.DocProfile.VibeWorkflow},
		{"templates/common/docs/git.md.tmpl", cfg.DocProfile.Git},
		{"templates/common/docs/observability.md.tmpl", cfg.DocProfile.Observability},
		{"templates/common/docs/performance.md.tmpl", cfg.DocProfile.Performance},
		{"templates/common/docs/api.md.tmpl", cfg.DocProfile.API},
		{"templates/common/docs/database.md.tmpl", cfg.DocProfile.Database},
		{"templates/common/DESIGN.md.tmpl", cfg.ProjectType == ProjectTypeFrontend},
		{"templates/common/.semgrep/antislop.yml.tmpl", cfg.ProjectType == ProjectTypeFrontend},
		{"templates/common/docs/design_system.md.tmpl", cfg.ProjectType == ProjectTypeFrontend},
	}

	for _, doc := range docs {
		if !doc.enabled {
			continue
		}
		destPath := filepath.Join(cfg.Destination, filepath.FromSlash(outNameFromTmpl(doc.tmplPath)))
		if seen[destPath] { // the framework template provides this file (framework wins)
			continue
		}
		content, err := renderer.Render(doc.tmplPath)
		if err != nil {
			return err
		}
		if err := writeFile(destPath, content); err != nil {
			return err
		}
	}
	return nil
}

// ensureGitRepo runs `git init` unless dir is already inside a git work tree
// (e.g. generated into a monorepo). Never fails generation; returns a status line.
func ensureGitRepo(dir string) string {
	if _, err := exec.LookPath("git"); err != nil {
		return "⚠ git not found — run `git init` manually"
	}
	// rev-parse reads no work-tree state, so a parent repo's core.fsmonitor never runs.
	if exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree").Run() == nil {
		return "git: already inside a repository, skipped git init"
	}
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		return fmt.Sprintf("⚠ git init failed: %s", strings.TrimSpace(string(out)))
	}
	return "✓ git repository initialized"
}

// outNameFromTmpl derives the project-relative output path from a common template path
// (e.g. "templates/common/docs/api.md.tmpl" → "docs/api.md"), so a framework file at the
// same path wins instead of being overwritten.
func outNameFromTmpl(path string) string {
	rel := strings.TrimSuffix(strings.TrimPrefix(path, "templates/common/"), ".tmpl")
	if rel == "gitignore" || rel == "npmrc" { // dotfiles are stored without the dot
		return "." + rel
	}
	return rel
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
		base := validation.ToSlug(filepath.Base(wd))
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
	if r.GitStatus != "" {
		sb.WriteString("  " + r.GitStatus + "\n")
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
			if r.Language == LanguageGo {
				sb.WriteString(" Swagger UI:    http://localhost:3000/swagger/index.html (run `make swag`)\n")
			} else {
				sb.WriteString(" Swagger UI:    http://localhost:3000/swagger/ (OpenAPI: /openapi.json)\n")
			}
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
