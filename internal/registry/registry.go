package registry

import "github.com/izeah01/vibe-gen/internal/generator"

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
	// TemplatePath is relative to the templates/ directory.
	TemplatePath string
	// DocProfile defines which documentation files to generate.
	DocProfile generator.DocProfile
	// PostInstallSteps lists instructions shown after generation.
	PostInstallSteps []string
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

// all registered languages and their frameworks.
var languages = []Language{
	{
		ID:          generator.LanguageGo,
		DisplayName: "Golang",
		Frameworks: []Framework{
			{
				ID:           "echo",
				Name:         "Echo (API)",
				DisplayName:  "Echo    (API)",
				ProjectType:  generator.ProjectTypeAPI,
				TemplatePath: "go/echo",
				DocProfile:   generator.BackendDocProfile(),
				PostInstallSteps: []string{
					"cd {{.ProjectSlug}}",
					"cp .env.example .env",
					"cp config.example.yml config.yml",
					"go mod tidy",
					"make gen-keys    # generate RSA keys for JWT (./secret-auth/)",
					"make swag        # generate Swagger docs (swag init -g main.go -o docs/swagger)",
					"make run",
				},
			},
			{
				ID:           "gin",
				Name:         "Gin (API)",
				DisplayName:  "Gin     (API)",
				ProjectType:  generator.ProjectTypeAPI,
				TemplatePath: "go/gin",
				DocProfile:   generator.BackendDocProfile(),
				PostInstallSteps: []string{
					"cd {{.ProjectSlug}}",
					"cp .env.example .env",
					"cp config.example.yml config.yml",
					"go mod tidy",
					"make gen-keys    # generate RSA keys for JWT (./secret-auth/)",
					"make swag        # generate Swagger docs (swag init -g main.go -o docs/swagger)",
					"make run",
				},
			},
			{
				ID:           "fiber",
				Name:         "Fiber (API)",
				DisplayName:  "Fiber   (API)",
				ProjectType:  generator.ProjectTypeAPI,
				TemplatePath: "go/fiber",
				DocProfile:   generator.BackendDocProfile(),
				PostInstallSteps: []string{
					"cd {{.ProjectSlug}}",
					"cp .env.example .env",
					"cp config.example.yml config.yml",
					"go mod tidy",
					"make gen-keys    # generate RSA keys for JWT (./secret-auth/)",
					"make swag        # generate Swagger docs (swag init -g main.go -o docs/swagger)",
					"make run",
				},
			},
			{
				ID:           "cron-job",
				Name:         "Cron Job (Worker)",
				DisplayName:  "Cron Job (Worker)",
				ProjectType:  generator.ProjectTypeWorker,
				TemplatePath: "go/cron-job",
				DocProfile:   generator.WorkerDocProfile(),
				PostInstallSteps: []string{
					"cd {{.ProjectSlug}}",
					"cp .env.example .env",
					"cp config.example.yaml config.yaml",
					"go mod tidy",
					"go run main.go",
				},
			},
		},
	},
	{
		ID:          generator.LanguageJavaScript,
		DisplayName: "JavaScript",
		Frameworks: []Framework{
			{
				ID:           "express",
				Name:         "ExpressJS (Backend API)",
				DisplayName:  "ExpressJS       (Backend API)",
				ProjectType:  generator.ProjectTypeAPI,
				TemplatePath: "javascript/express",
				DocProfile:   generator.BackendDocProfile(),
				PostInstallSteps: []string{
					"cd {{.ProjectSlug}}",
					"cp .env.example .env",
					"npm install",
					"npm run dev",
				},
			},
			{
				ID:           "nestjs",
				Name:         "NestJS (Backend API)",
				DisplayName:  "NestJS          (Backend API)",
				ProjectType:  generator.ProjectTypeAPI,
				TemplatePath: "javascript/nestjs",
				DocProfile:   generator.BackendDocProfile(),
				PostInstallSteps: []string{
					"cd {{.ProjectSlug}}",
					"cp .env.example .env",
					"npm install",
					"npm run start:dev",
				},
			},
			{
				ID:           "cron",
				Name:         "Cron Scheduler (Worker)",
				DisplayName:  "Cron Scheduler  (Worker)",
				ProjectType:  generator.ProjectTypeWorker,
				TemplatePath: "javascript/cron",
				DocProfile:   generator.WorkerDocProfile(),
				PostInstallSteps: []string{
					"cd {{.ProjectSlug}}",
					"cp .env.example .env",
					"npm install",
					"npm run dev",
				},
			},
			{
				ID:           "react",
				Name:         "React/Vite (Frontend)",
				DisplayName:  "React/Vite      (Frontend)",
				ProjectType:  generator.ProjectTypeFrontend,
				TemplatePath: "javascript/react",
				DocProfile:   generator.FrontendDocProfile(),
				PostInstallSteps: []string{
					"cd {{.ProjectSlug}}",
					"cp .env.example .env.local",
					"npm install",
					"npm run dev",
				},
			},
			{
				ID:           "next",
				Name:         "Next.js (Frontend)",
				DisplayName:  "Next.js         (Frontend)",
				ProjectType:  generator.ProjectTypeFrontend,
				TemplatePath: "javascript/next",
				DocProfile:   generator.FrontendDocProfile(),
				PostInstallSteps: []string{
					"cd {{.ProjectSlug}}",
					"cp .env.example .env.local",
					"npm install",
					"npm run dev",
				},
			},
			{
				ID:           "vue",
				Name:         "Vue.js/Vite (Frontend)",
				DisplayName:  "Vue.js/Vite     (Frontend)",
				ProjectType:  generator.ProjectTypeFrontend,
				TemplatePath: "javascript/vue",
				DocProfile:   generator.FrontendDocProfile(),
				PostInstallSteps: []string{
					"cd {{.ProjectSlug}}",
					"cp .env.example .env.local",
					"npm install",
					"npm run dev",
				},
			},
			{
				ID:           "svelte",
				Name:         "Svelte/Vite (Frontend)",
				DisplayName:  "Svelte/Vite     (Frontend)",
				ProjectType:  generator.ProjectTypeFrontend,
				TemplatePath: "javascript/svelte",
				DocProfile:   generator.FrontendDocProfile(),
				PostInstallSteps: []string{
					"cd {{.ProjectSlug}}",
					"cp .env.example .env.local",
					"npm install",
					"npm run dev",
				},
			},
			{
				ID:           "nuxt",
				Name:         "NuxtJS (Frontend)",
				DisplayName:  "NuxtJS          (Frontend)",
				ProjectType:  generator.ProjectTypeFrontend,
				TemplatePath: "javascript/nuxt",
				DocProfile:   generator.FrontendDocProfile(),
				PostInstallSteps: []string{
					"cd {{.ProjectSlug}}",
					"cp .env.example .env",
					"npm install",
					"npm run dev",
				},
			},
		},
	},
}

// Languages returns all registered languages.
func Languages() []Language {
	return languages
}

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
