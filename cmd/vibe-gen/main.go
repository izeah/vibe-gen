package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	vibegen "github.com/izeah01/vibe-gen"
	"github.com/izeah01/vibe-gen/internal/generator"
	"github.com/izeah01/vibe-gen/internal/prompt"
	"github.com/izeah01/vibe-gen/internal/registry"
	"github.com/izeah01/vibe-gen/internal/validation"
)

func main() {
	generator.SetTemplates(vibegen.Templates)

	rootCmd := &cobra.Command{
		Use:   "vibe-gen",
		Short: "Vibe Coding Boilerplate Generator — By Izeah01",
		Long:  "Generate AI-ready project boilerplates interactively or via flags.",
		// Default: run interactive TUI.
		RunE: runInteractive,
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a project non-interactively",
		Long:  "Generate a project boilerplate using flags (non-interactive mode).",
		RunE:  runNonInteractive,
	}

	initCmd.Flags().String("lang", "", "Programming language (go, javascript)")
	initCmd.Flags().String("framework", "", "Framework ID (echo, gin, fiber, cron-job, express, nestjs, cron, react, next, vue, svelte, nuxt)")
	initCmd.Flags().String("name", "", "Project name")
	initCmd.Flags().String("dest", "", "Destination directory (default: ./<name>)")
	initCmd.Flags().String("module", "", "Go module name (Go projects only; auto-detected if empty)")
	initCmd.Flags().Bool("dry-run", false, "Show what would be generated without writing files")

	rootCmd.Flags().Bool("dry-run", false, "Show what would be generated without writing files")
	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runInteractive starts the Bubble Tea TUI.
func runInteractive(cmd *cobra.Command, _ []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	m := prompt.New()
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	result, ok := finalModel.(prompt.Model)
	if !ok || result.Result == nil {
		fmt.Println("\n  Cancelled.")
		return nil
	}

	cfg := result.Result
	cfg.DryRun = dryRun

	return runGeneration(cfg)
}

// runNonInteractive generates a project from flags without TUI.
func runNonInteractive(cmd *cobra.Command, _ []string) error {
	langStr, _ := cmd.Flags().GetString("lang")
	frameworkID, _ := cmd.Flags().GetString("framework")
	name, _ := cmd.Flags().GetString("name")
	dest, _ := cmd.Flags().GetString("dest")
	moduleName, _ := cmd.Flags().GetString("module")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Validate required flags.
	missing := []string{}
	if langStr == "" {
		missing = append(missing, "--lang")
	}
	if frameworkID == "" {
		missing = append(missing, "--framework")
	}
	if name == "" {
		missing = append(missing, "--name")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required flags: %s", strings.Join(missing, ", "))
	}

	// Validate project name.
	if err := validation.ValidateName(name); err != nil {
		return fmt.Errorf("invalid project name: %w", err)
	}

	slug := validation.ToSlug(name)
	if dest == "" {
		dest = "./" + slug
	}

	// Lookup language.
	lang := generator.Language(langStr)
	langEntry, ok := registry.FindLanguage(lang)
	if !ok {
		return fmt.Errorf("unsupported language %q; supported: go, javascript", langStr)
	}

	// Lookup framework.
	fw, ok := registry.FindFramework(lang, frameworkID)
	if !ok {
		return fmt.Errorf("unsupported framework %q for language %q", frameworkID, langStr)
	}

	// Auto-detect module name for Go.
	if lang == generator.LanguageGo && moduleName == "" {
		moduleName = generator.SuggestModuleName(name)
	}

	cfg := &generator.ProjectConfig{
		Name:             name,
		Slug:             slug,
		Title:            validation.ToTitle(name),
		Language:         lang,
		Framework:        fw.Name,
		FrameworkID:      string(langEntry.ID) + "/" + fw.ID,
		ProjectType:      fw.ProjectType,
		Destination:      dest,
		ModuleName:       moduleName,
		DocProfile:       fw.DocProfile,
		PostInstallSteps: fw.PostInstallSteps,
		DryRun:           dryRun,
	}

	return runGeneration(cfg)
}

// runGeneration executes the generation and prints results.
func runGeneration(cfg *generator.ProjectConfig) error {
	if cfg.DryRun {
		fmt.Print("\n  [dry-run mode] No files will be written.\n\n")
	}

	result, err := generator.Generate(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "\n"+fmt.Sprintf("  ✗ Generation failed: %s", err.Error()))
		os.Exit(1)
	}

	fmt.Print(generator.FormatPostInstall(result))
	return nil
}
