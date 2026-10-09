//go:build smoke

// Package smoke generates every ready template and proves the output builds, lints and tests.
// Run with `make smoke` (all) or `make smoke FRAMEWORK=<id>` (one).
package smoke

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	vibegen "github.com/izeah01/vibe-gen"
	"github.com/izeah01/vibe-gen/internal/generator"
	"github.com/izeah01/vibe-gen/internal/registry"
)

// steps lists the commands each language must pass, run in order inside the generated project.
// Requires semgrep on PATH (brew install semgrep).
var nodeSteps = [][]string{
	{"npm", "install", "--no-audit", "--no-fund", "--ignore-scripts"},
	{"npm", "run", "lint"},
	{"npm", "test"},
	{"npm", "run", "build", "--if-present"},
	{"npm", "audit", "--omit=dev", "--audit-level=high"},
	{"make", "sast"},
}

var steps = map[generator.Language][][]string{
	generator.LanguageGo: {
		{"go", "mod", "tidy"},
		{"go", "build", "./..."},
		{"go", "vet", "./..."},
		{"sh", "-c", `out=$(gofmt -l .); [ -z "$out" ] || { echo "$out"; exit 1; }`},
		{"go", "test", "./..."},
		{"go", "run", "golang.org/x/vuln/cmd/govulncheck@latest", "./..."},
		// Raw command, not make sast: the cron-job template has no Makefile.
		{"semgrep", "scan", "--config", "p/default", "--error", "--metrics=off", "--quiet"},
	},
	generator.LanguageJavaScript: nodeSteps,
	// TypeScript adds a typecheck after lint.
	generator.LanguageTypeScript: slices.Insert(slices.Clone(nodeSteps), 2, []string{"npm", "run", "typecheck"}),
}

func TestSmoke(t *testing.T) {
	generator.SetTemplates(vibegen.Templates)
	only := os.Getenv("FRAMEWORK")

	ran := 0
	for _, lang := range registry.Languages() {
		for _, fw := range lang.Frameworks {
			if only != "" && fw.ID != only {
				continue
			}
			ran++
			variants := fw.Variants
			if len(variants) == 0 {
				variants = []registry.Variant{{}}
			}
			for _, v := range variants {
				name := string(lang.ID) + "/" + fw.ID
				if v.ID != "" {
					name += "/" + v.ID
				}
				t.Run(name, func(t *testing.T) {
					smoke(t, lang.ID, fw, v)
				})
			}
		}
	}
	if ran == 0 {
		t.Fatalf("no ready template matches FRAMEWORK=%q", only)
	}
}

// smoke generates one template (language × framework × variant) and runs its language steps.
func smoke(t *testing.T, lang generator.Language, fw registry.Framework, v registry.Variant) {
	cmds, ok := steps[lang]
	if !ok {
		t.Fatalf("no smoke steps for language %q; add them before marking a template ready", lang)
	}
	dir := filepath.Join(t.TempDir(), "smoke-"+fw.ID)
	cfg := fw.ProjectConfig(lang, v, "smoke-"+fw.ID, dir, "example.com/smoke-"+fw.ID)
	if _, err := generator.Generate(cfg); err != nil {
		t.Fatalf("generate: %v", err)
	}
	if fw.ProjectType == generator.ProjectTypeFrontend {
		for _, f := range []string{"DESIGN.md", ".semgrep/antislop.yml"} {
			if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
				t.Fatalf("frontend project is missing %s: %v", f, err)
			}
		}
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v\n%s", strings.Join(c, " "), err, out)
		}
	}
}
