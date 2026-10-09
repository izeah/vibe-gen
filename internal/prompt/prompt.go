package prompt

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/izeah01/vibe-gen/internal/generator"
	"github.com/izeah01/vibe-gen/internal/registry"
	"github.com/izeah01/vibe-gen/internal/validation"
)

// step tracks which screen the TUI is on.
type step int

const (
	stepLanguage step = iota
	stepFramework
	stepVariant // only for frameworks with variants
	stepName
	stepDest
	stepModuleName // Go only
	stepSummary
	stepDone
)

// Model is the Bubble Tea model for the interactive flow.
type Model struct {
	step step

	// language selection
	languages  []registry.Language
	langCursor int

	// framework selection
	frameworks  []registry.Framework
	frameCursor int

	// variant selection
	variantCursor int

	// text inputs
	nameInput   textinput.Model
	destInput   textinput.Model
	moduleInput textinput.Model // Go only

	// collected values
	selectedLang      registry.Language
	selectedFramework registry.Framework
	selectedVariant   registry.Variant

	// confirmation
	confirmed bool

	// result
	Result *generator.ProjectConfig
	Err    error

	// display
	width int
}

// New creates a fresh TUI model.
func New() Model {
	ni := textinput.New()
	ni.Placeholder = "my-awesome-api"
	ni.CharLimit = 64
	ni.Width = 40

	di := textinput.New()
	di.CharLimit = 256
	di.Width = 40

	mi := textinput.New()
	mi.CharLimit = 256
	mi.Width = 40

	return Model{
		step:        stepLanguage,
		languages:   registry.Languages(),
		nameInput:   ni,
		destInput:   di,
		moduleInput: mi,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case tea.KeyMsg:
		switch m.step {
		case stepLanguage:
			return m.updateLanguage(msg)
		case stepFramework:
			return m.updateFramework(msg)
		case stepVariant:
			return m.updateVariant(msg)
		case stepName:
			return m.updateName(msg)
		case stepDest:
			return m.updateDest(msg)
		case stepModuleName:
			return m.updateModule(msg)
		case stepSummary:
			return m.updateSummary(msg)
		}
	}
	return m, nil
}

// — Language selection —

func (m Model) updateLanguage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.langCursor > 0 {
			m.langCursor--
		}
	case "down", "j":
		if m.langCursor < len(m.languages)-1 {
			m.langCursor++
		}
	case "enter", " ":
		m.selectedLang = m.languages[m.langCursor]
		m.frameworks = m.selectedLang.Frameworks
		m.frameCursor = 0
		m.step = stepFramework
	}
	return m, nil
}

// — Framework selection —

func (m Model) updateFramework(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.step = stepLanguage
	case "up", "k":
		if m.frameCursor > 0 {
			m.frameCursor--
		}
	case "down", "j":
		if m.frameCursor < len(m.frameworks)-1 {
			m.frameCursor++
		}
	case "enter", " ":
		m.selectedFramework = m.frameworks[m.frameCursor]
		m.selectedVariant = registry.Variant{}
		if len(m.selectedFramework.Variants) > 0 {
			m.variantCursor = 0
			m.step = stepVariant
			return m, nil
		}
		m.nameInput.Focus()
		m.step = stepName
	}
	return m, nil
}

// — Variant selection —

func (m Model) updateVariant(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	variants := m.selectedFramework.Variants
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.step = stepFramework
	case "up", "k":
		if m.variantCursor > 0 {
			m.variantCursor--
		}
	case "down", "j":
		if m.variantCursor < len(variants)-1 {
			m.variantCursor++
		}
	case "enter", " ":
		m.selectedVariant = variants[m.variantCursor]
		m.nameInput.Focus()
		m.step = stepName
	}
	return m, nil
}

// — Project name input —

func (m Model) updateName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.nameInput.Blur()
		m.step = stepFramework
		if len(m.selectedFramework.Variants) > 0 {
			m.step = stepVariant
		}
	case "enter":
		name := strings.TrimSpace(m.nameInput.Value())
		if err := validation.ValidateName(name); err != nil {
			m.Err = err
			return m, nil
		}
		m.Err = nil
		slug := validation.ToSlug(name)
		// Pre-fill destination.
		m.destInput.SetValue("./" + slug)
		m.destInput.Focus()
		m.nameInput.Blur()
		m.step = stepDest
		return m, nil
	}
	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

// — Destination input —

func (m Model) updateDest(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.destInput.Blur()
		m.nameInput.Focus()
		m.step = stepName
	case "enter":
		dest := strings.TrimSpace(m.destInput.Value())
		if dest == "" {
			dest = "./" + validation.ToSlug(m.nameInput.Value())
			m.destInput.SetValue(dest)
		}
		m.destInput.Blur()
		// For Go projects, go to module name step.
		if m.selectedLang.ID == generator.LanguageGo {
			suggestion := generator.SuggestModuleName(m.nameInput.Value())
			m.moduleInput.SetValue(suggestion)
			m.moduleInput.Focus()
			m.step = stepModuleName
		} else {
			m.step = stepSummary
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.destInput, cmd = m.destInput.Update(msg)
	return m, cmd
}

// — Go module name input —

func (m Model) updateModule(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.moduleInput.Blur()
		m.destInput.Focus()
		m.step = stepDest
	case "enter":
		m.moduleInput.Blur()
		m.step = stepSummary
		return m, nil
	}
	var cmd tea.Cmd
	m.moduleInput, cmd = m.moduleInput.Update(msg)
	return m, cmd
}

// — Summary + confirmation —

func (m Model) updateSummary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "n", "N":
		return m, tea.Quit
	case "y", "Y", "enter":
		m.confirmed = true
		m.Result = m.buildConfig()
		m.step = stepDone
		return m, tea.Quit
	}
	return m, nil
}

// — View —

func (m Model) View() string {
	var sb strings.Builder
	sb.WriteString(Banner())
	sb.WriteString("\n")

	switch m.step {
	case stepLanguage:
		sb.WriteString(sectionStyle.Render("  Select Language") + "\n\n")
		for i, lang := range m.languages {
			cursor := "  "
			line := "  " + lang.DisplayName
			if i == m.langCursor {
				cursor = cursorStyle.Render("▶ ")
				line = selectedStyle.Render("  " + lang.DisplayName)
			}
			sb.WriteString(cursor + line + "\n")
		}
		sb.WriteString("\n" + helpStyle.Render("  ↑/↓ navigate  •  enter select  •  q quit"))

	case stepFramework:
		sb.WriteString(sectionStyle.Render("  Select Framework / Project Type") + "\n\n")
		for i, fw := range m.frameworks {
			cursor := "  "
			line := "  " + fw.DisplayName
			if i == m.frameCursor {
				cursor = cursorStyle.Render("▶ ")
				line = selectedStyle.Render("  " + fw.DisplayName)
			}
			sb.WriteString(cursor + line + "\n")
		}
		sb.WriteString("\n" + helpStyle.Render("  ↑/↓ navigate  •  enter select  •  esc back  •  q quit"))

	case stepVariant:
		sb.WriteString(sectionStyle.Render("  Select "+m.selectedFramework.Name+" Variant") + "\n\n")
		for i, v := range m.selectedFramework.Variants {
			cursor := "  "
			line := "  " + v.Name
			if i == m.variantCursor {
				cursor = cursorStyle.Render("▶ ")
				line = selectedStyle.Render("  " + v.Name)
			}
			sb.WriteString(cursor + line + "\n")
		}
		sb.WriteString("\n" + helpStyle.Render("  ↑/↓ navigate  •  enter select  •  esc back  •  q quit"))

	case stepName:
		sb.WriteString(sectionStyle.Render("  Project Name") + "\n\n")
		sb.WriteString("  " + m.nameInput.View() + "\n")
		if m.Err != nil {
			sb.WriteString("\n  " + errorStyle.Render("✗ "+m.Err.Error()) + "\n")
		}
		sb.WriteString("\n" + helpStyle.Render("  enter confirm  •  esc back"))

	case stepDest:
		sb.WriteString(sectionStyle.Render("  Destination Directory") + "\n\n")
		sb.WriteString("  " + m.destInput.View() + "\n")
		sb.WriteString("\n" + helpStyle.Render("  enter confirm  •  esc back"))

	case stepModuleName:
		sb.WriteString(sectionStyle.Render("  Go Module Name") + "\n\n")
		sb.WriteString("  " + m.moduleInput.View() + "\n")
		sb.WriteString("\n" + helpStyle.Render("  enter confirm  •  esc back"))

	case stepSummary:
		sb.WriteString(m.renderSummary())
		sb.WriteString("\n" + helpStyle.Render("  y/enter — generate   •   n/q — cancel"))

	case stepDone:
		sb.WriteString(successStyle.Render("  ✓ Generating project...") + "\n")
	}

	return sb.String()
}

func (m Model) renderSummary() string {
	name := m.nameInput.Value()
	dest := m.destInput.Value()
	if dest == "" {
		dest = "./" + validation.ToSlug(name)
	}

	rows := []string{
		labelStyle.Render("Language") + " : " + valueStyle.Render(m.selectedLang.DisplayName),
		labelStyle.Render("Framework") + " : " + valueStyle.Render(m.selectedFramework.DisplayName),
	}
	if m.selectedVariant.ID != "" {
		rows = append(rows, labelStyle.Render("Variant")+" : "+valueStyle.Render(m.selectedVariant.Name))
	}
	rows = append(rows,
		labelStyle.Render("Project")+" : "+valueStyle.Render(name),
		labelStyle.Render("Location")+" : "+valueStyle.Render(dest),
	)
	if m.selectedLang.ID == generator.LanguageGo {
		rows = append(rows, labelStyle.Render("Module")+" : "+valueStyle.Render(m.moduleInput.Value()))
	}

	content := strings.Join(rows, "\n")
	box := summaryBoxStyle.Render(content)

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(box)
	sb.WriteString("\n\n  Generate? " + successStyle.Render("[Y]") + " / " + errorStyle.Render("[n]") + "\n\n")
	return sb.String()
}

// buildConfig constructs a ProjectConfig from the current model state.
func (m Model) buildConfig() *generator.ProjectConfig {
	name := strings.TrimSpace(m.nameInput.Value())
	dest := strings.TrimSpace(m.destInput.Value())
	if dest == "" {
		dest = "./" + validation.ToSlug(name)
	}
	// Make dest absolute relative to cwd.
	if !filepath.IsAbs(dest) {
		wd, _ := os.Getwd()
		dest = filepath.Join(wd, dest)
	}

	return m.selectedFramework.ProjectConfig(m.selectedLang.ID, m.selectedVariant, name, dest, strings.TrimSpace(m.moduleInput.Value()))
}
