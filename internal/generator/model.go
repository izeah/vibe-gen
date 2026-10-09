package generator

import "time"

// Language represents a supported programming language.
type Language string

const (
	LanguageGo         Language = "go"
	LanguageJavaScript Language = "javascript"
	LanguageTypeScript Language = "typescript"
)

// ProjectType categorizes what kind of project is generated.
type ProjectType string

const (
	ProjectTypeAPI      ProjectType = "api"
	ProjectTypeWorker   ProjectType = "worker"
	ProjectTypeFrontend ProjectType = "frontend"
)

// DocProfile defines which AI documentation files to generate.
type DocProfile struct {
	Architecture    bool
	API             bool
	Database        bool
	Observability   bool
	Performance     bool
	Security        bool
	Testing         bool
	Git             bool
	Decisions       bool
	Roadmap         bool
	Progress        bool
	VibeWorkflow    bool
	CodingStandards bool
}

// BackendDocProfile returns the full doc set for backend/API projects.
func BackendDocProfile() DocProfile {
	return DocProfile{
		Architecture:    true,
		API:             true,
		Database:        true,
		Observability:   true,
		Performance:     true,
		Security:        true,
		Testing:         true,
		Git:             true,
		Decisions:       true,
		Roadmap:         true,
		Progress:        true,
		VibeWorkflow:    true,
		CodingStandards: true,
	}
}

// WorkerDocProfile returns docs for worker/cron projects.
func WorkerDocProfile() DocProfile {
	return DocProfile{
		Architecture:    true,
		Database:        true,
		Observability:   true,
		Performance:     true,
		Security:        true,
		Testing:         true,
		Git:             true,
		Decisions:       true,
		Roadmap:         true,
		Progress:        true,
		VibeWorkflow:    true,
		CodingStandards: true,
	}
}

// FrontendDocProfile returns docs for frontend projects.
func FrontendDocProfile() DocProfile {
	return DocProfile{
		Architecture:    true,
		Performance:     true,
		Security:        true,
		Testing:         true,
		Git:             true,
		Decisions:       true,
		Roadmap:         true,
		Progress:        true,
		VibeWorkflow:    true,
		CodingStandards: true,
	}
}

// ProjectConfig holds all configuration needed to generate a project.
type ProjectConfig struct {
	// Name is the raw project name input by the user.
	Name string
	// Slug is the kebab-case version of Name.
	Slug string
	// Title is the Title Cased version of Name.
	Title string
	// Language is the programming language.
	Language Language
	// Framework is the display name of the framework.
	Framework string
	// FrameworkID is the internal identifier used to locate templates.
	FrameworkID string
	// Shared lists template dirs under templates/<lang>/ rendered after the framework (framework files win).
	Shared []string
	// Variant is the framework sub-option ID (e.g. "app" for Next.js App Router); empty if none.
	Variant string
	// ProjectType categorizes the project.
	ProjectType ProjectType
	// Destination is the absolute path where the project is generated.
	Destination string
	// ModuleName is the Go module path (Go projects only).
	ModuleName string
	// DryRun skips file writing and only shows what would be generated.
	DryRun bool
	// Date is the generation date (YYYY-MM-DD).
	Date string
	// Year is the generation year.
	Year string
	// DocProfile defines which doc files to generate.
	DocProfile DocProfile
	// PostInstallSteps are instructions shown after generation.
	PostInstallSteps []string
}

// TemplateVars returns a map of template variables for rendering.
// IsTypeScript is "true" or "" so templates can branch with {{if .IsTypeScript}}.
func (c *ProjectConfig) TemplateVars() map[string]string {
	isTS := ""
	if c.Language == LanguageTypeScript {
		isTS = "true"
	}
	return map[string]string{
		"Variant":      c.Variant,
		"IsTypeScript": isTS,
		"ProjectName":  c.Name,
		"ProjectSlug":  c.Slug,
		"ProjectTitle": c.Title,
		"Language":     string(c.Language),
		"Framework":    c.Framework,
		"FrameworkID":  c.FrameworkID,
		"ProjectType":  string(c.ProjectType),
		"ModuleName":   c.ModuleName,
		"Destination":  c.Destination,
		"GoVersion":    "1.27",
		"Date":         c.Date,
		"Year":         c.Year,
	}
}

// now returns the current time for use in project generation.
func now() time.Time {
	return time.Now()
}
