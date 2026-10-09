package registry

import (
	"slices"

	"github.com/izeah01/vibe-gen/internal/generator"
	"github.com/izeah01/vibe-gen/internal/validation"
)

// Variant is a sub-option of a framework (e.g. Next.js App Router vs Pages Router).
// Its templates live in templates/<lang>/<framework>/<variant>/.
type Variant struct {
	ID   string
	Name string
}

// Framework describes a selectable framework option in the TUI.
type Framework struct {
	// ID is the internal identifier used to locate templates.
	ID string
	// Name is the clean framework name used in generated files and templates (no TUI padding).
	Name string
	// DisplayName is shown in the TUI selection list (may include spacing for alignment).
	DisplayName string
	// ProjectType categorizes the project.
	ProjectType generator.ProjectType
	// DocProfile defines which documentation files to generate.
	DocProfile generator.DocProfile
	// PostInstallSteps lists instructions shown after generation.
	PostInstallSteps []string
	// Shared lists template dirs under templates/<lang>/ rendered after the framework's own
	// templates (e.g. "_shared/api"); a file the framework provides wins.
	Shared []string
	// Variants lists sub-options; the first one is the default. Empty for most frameworks.
	Variants []Variant
}

// Language describes a selectable programming language.
type Language struct {
	// ID is the internal normalized identifier.
	ID generator.Language
	// DisplayName is shown in the TUI.
	DisplayName string
	// Frameworks is the ordered list of framework options.
	Frameworks []Framework
}

// Post-install steps shared by several frameworks. npm install is safe: generated Node
// projects ship an .npmrc with ignore-scripts=true (no dependency lifecycle scripts).
var (
	goAPISteps = []string{
		"cd {{.ProjectSlug}}",
		"cp .env.example .env",
		"cp config.example.yml config.yml",
		"go mod tidy",
		"make pg-start redis-start   # local PostgreSQL 18 + Redis (Docker)",
		"make migrate-up             # users + samples tables",
		"make gen-keys               # RSA keys for JWT (./secret-auth/)",
		"make create-admin EMAIL=admin@example.com   # first admin (prompts for password)",
		"make swag                   # Swagger docs (docs/swagger)",
		"make run",
	}
	nodeAPISteps = []string{
		"cd {{.ProjectSlug}}",
		"cp .env.example .env",
		"npm install",
		"make pg-start redis-start   # local PostgreSQL 18 + Redis (Docker)",
		"make migrate-up             # users + samples tables",
		"make gen-keys               # RSA keys for JWT (./secret-auth/)",
		"make create-admin EMAIL=admin@example.com   # first admin (prompts for password)",
		"make run",
	}
	spaSteps = []string{
		"cd {{.ProjectSlug}}",
		"cp .env.example .env",
		"npm install",
		"# API: AUTH_COOKIE_MODE=true, APP_CORS_ALLOW_ORIGIN=http://localhost:5173 (same site, see docs/security.md)",
		"make run                     # http://localhost:5173 (proxies /v1 to the API)",
		"make test-e2e                # Playwright against a mock API (no backend needed)",
	}
)

// bffSteps are the post-install steps of a BFF framework; envFile is padded for alignment.
func bffSteps(envFile string) []string {
	return []string{
		"cd {{.ProjectSlug}}",
		"cp .env.example " + envFile + " # set BFF_COOKIE_SECRET: openssl rand -base64 32",
		"npm install",
		"# start a backend API on :3000 (Go or Node template, AUTH_COOKIE_MODE=false)",
		"make run                     # http://localhost:3001",
		"make test-e2e                # Playwright against a mock API (no backend needed)",
	}
}

// all registered languages and their frameworks.
var languages = []Language{
	{
		ID:          generator.LanguageGo,
		DisplayName: "Golang",
		Frameworks: []Framework{
			{
				ID:               "echo",
				Name:             "Echo (API)",
				DisplayName:      "Echo    (API)",
				ProjectType:      generator.ProjectTypeAPI,
				DocProfile:       generator.BackendDocProfile(),
				PostInstallSteps: goAPISteps,
			},
			{
				ID:               "gin",
				Name:             "Gin (API)",
				DisplayName:      "Gin     (API)",
				ProjectType:      generator.ProjectTypeAPI,
				DocProfile:       generator.BackendDocProfile(),
				PostInstallSteps: goAPISteps,
			},
			{
				ID:               "fiber",
				Name:             "Fiber (API)",
				DisplayName:      "Fiber   (API)",
				ProjectType:      generator.ProjectTypeAPI,
				DocProfile:       generator.BackendDocProfile(),
				PostInstallSteps: goAPISteps,
			},
			{
				ID:          "cron-job",
				Name:        "Cron Job (Worker)",
				DisplayName: "Cron Job (Worker)",
				ProjectType: generator.ProjectTypeWorker,
				DocProfile:  generator.WorkerDocProfile(),
				PostInstallSteps: []string{
					"cd {{.ProjectSlug}}",
					"cp .env.example .env",
					"go mod tidy",
					"docker run -d --name redis -p 127.0.0.1:6379:6379 redis:7   # jobs run under a Redis lock",
					"go run .",
				},
			},
		},
	},
	{
		ID:          generator.LanguageJavaScript,
		DisplayName: "JavaScript",
		Frameworks:  nodeFrameworks(generator.LanguageJavaScript),
	},
	{
		ID:          generator.LanguageTypeScript,
		DisplayName: "TypeScript",
		Frameworks:  nodeFrameworks(generator.LanguageTypeScript),
	},
}

// nodeFrameworks returns the JavaScript/TypeScript framework list. Both languages share it:
// TypeScript templates are the source and JavaScript ones are derived by `make gen-js`.
// NestJS is TypeScript-only (decorators do not survive type stripping).
func nodeFrameworks(lang generator.Language) []Framework {
	fws := []Framework{
		{
			ID:               "express",
			Name:             "Express 5 (Backend API)",
			DisplayName:      "Express 5       (Backend API)",
			ProjectType:      generator.ProjectTypeAPI,
			DocProfile:       generator.BackendDocProfile(),
			Shared:           []string{"_shared/api", "_shared/express-http"},
			PostInstallSteps: nodeAPISteps,
		},
		{
			ID:               "hono",
			Name:             "Hono (Backend API)",
			DisplayName:      "Hono            (Backend API)",
			ProjectType:      generator.ProjectTypeAPI,
			DocProfile:       generator.BackendDocProfile(),
			Shared:           []string{"_shared/api"},
			PostInstallSteps: nodeAPISteps,
		},
		{
			ID:               "nestjs",
			Name:             "NestJS (Backend API)",
			DisplayName:      "NestJS          (Backend API)",
			ProjectType:      generator.ProjectTypeAPI,
			DocProfile:       generator.BackendDocProfile(),
			Shared:           []string{"_shared/api", "_shared/express-http"},
			PostInstallSteps: nodeAPISteps,
		},
		{
			ID:          "cron",
			Name:        "Cron Scheduler (Worker)",
			DisplayName: "Cron Scheduler  (Worker)",
			ProjectType: generator.ProjectTypeWorker,
			DocProfile:  generator.WorkerDocProfile(),
			PostInstallSteps: []string{
				"cd {{.ProjectSlug}}",
				"cp .env.example .env",
				"npm install",
				"make pg-start redis-start   # local PostgreSQL 18 + Redis (Docker); schema comes from the API project",
				"make run-job JOB=sampleReport   # run one job now",
				"make run",
			},
		},
		{
			ID:               "react",
			Name:             "React/Vite (Frontend SPA)",
			DisplayName:      "React/Vite      (Frontend SPA)",
			ProjectType:      generator.ProjectTypeFrontend,
			DocProfile:       generator.FrontendDocProfile(),
			Shared:           []string{"_shared/spa"},
			PostInstallSteps: spaSteps,
		},
		{
			ID:          "next",
			Name:        "Next.js (Frontend)",
			DisplayName: "Next.js         (Frontend)",
			ProjectType: generator.ProjectTypeFrontend,
			DocProfile:  generator.FrontendDocProfile(),
			Variants: []Variant{
				{ID: "app", Name: "App Router"},
				{ID: "pages", Name: "Pages Router"},
			},
			Shared:           []string{"_shared/bff", "_shared/next"},
			PostInstallSteps: bffSteps(".env.local  "),
		},
		{
			ID:               "vue",
			Name:             "Vue/Vite (Frontend SPA)",
			DisplayName:      "Vue/Vite        (Frontend SPA)",
			ProjectType:      generator.ProjectTypeFrontend,
			DocProfile:       generator.FrontendDocProfile(),
			Shared:           []string{"_shared/spa"},
			PostInstallSteps: spaSteps,
		},
		{
			ID:               "sveltekit",
			Name:             "SvelteKit (Frontend)",
			DisplayName:      "SvelteKit       (Frontend)",
			ProjectType:      generator.ProjectTypeFrontend,
			DocProfile:       generator.FrontendDocProfile(),
			Shared:           []string{"_shared/bff"},
			PostInstallSteps: bffSteps(".env        "),
		},
		{
			ID:               "nuxt",
			Name:             "Nuxt 4 (Frontend)",
			DisplayName:      "Nuxt 4          (Frontend)",
			ProjectType:      generator.ProjectTypeFrontend,
			DocProfile:       generator.FrontendDocProfile(),
			Shared:           []string{"_shared/bff"},
			PostInstallSteps: bffSteps(".env        "),
		},
	}
	if lang == generator.LanguageJavaScript {
		fws = slices.DeleteFunc(fws, func(f Framework) bool { return f.ID == "nestjs" })
	}
	return fws
}

// Languages returns every registered language with its frameworks.
func Languages() []Language { return languages }

// FindLanguage returns the Language with the given ID, and a bool indicating if found.
func FindLanguage(id generator.Language) (Language, bool) {
	for _, l := range languages {
		if l.ID == id {
			return l, true
		}
	}
	return Language{}, false
}

// FindFramework returns the Framework within a language, and a bool indicating if found.
func FindFramework(langID generator.Language, frameworkID string) (Framework, bool) {
	lang, ok := FindLanguage(langID)
	if !ok {
		return Framework{}, false
	}
	for _, f := range lang.Frameworks {
		if f.ID == frameworkID {
			return f, true
		}
	}
	return Framework{}, false
}

// FindVariant returns the variant with the given ID; an empty ID selects the default (first) variant.
// Frameworks without variants only accept an empty ID.
func (f Framework) FindVariant(id string) (Variant, bool) {
	if len(f.Variants) == 0 {
		return Variant{}, id == ""
	}
	if id == "" {
		return f.Variants[0], true
	}
	for _, v := range f.Variants {
		if v.ID == id {
			return v, true
		}
	}
	return Variant{}, false
}

// ProjectConfig builds the generator config for this framework.
func (f Framework) ProjectConfig(lang generator.Language, variant Variant, name, dest, moduleName string) *generator.ProjectConfig {
	if lang != generator.LanguageGo {
		moduleName = ""
	}
	return &generator.ProjectConfig{
		Name:             name,
		Slug:             validation.ToSlug(name),
		Title:            validation.ToTitle(name),
		Language:         lang,
		Framework:        f.Name,
		FrameworkID:      string(lang) + "/" + f.ID,
		Variant:          variant.ID,
		Shared:           f.Shared,
		ProjectType:      f.ProjectType,
		Destination:      dest,
		ModuleName:       moduleName,
		DocProfile:       f.DocProfile,
		PostInstallSteps: f.PostInstallSteps,
	}
}
