package prompt

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	colorPrimary   = lipgloss.Color("#7C3AED") // vibrant purple
	colorAccent    = lipgloss.Color("#06B6D4") // cyan
	colorSuccess   = lipgloss.Color("#10B981") // green
	colorWarning   = lipgloss.Color("#F59E0B") // amber
	colorError     = lipgloss.Color("#EF4444") // red
	colorMuted     = lipgloss.Color("#6B7280") // gray
	colorHighlight = lipgloss.Color("#F3F4F6") // near white
)

// Styles
var (
	bannerStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Italic(true)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(12)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorHighlight).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	sectionStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Bold(true)

	dividerStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	summaryBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 2)
)

// Banner returns the styled ASCII banner for the tool.
func Banner() string {
	title := bannerStyle.Render(`
 ██╗   ██╗██╗██████╗ ███████╗      ██████╗ ███████╗███╗   ██╗
 ██║   ██║██║██╔══██╗██╔════╝     ██╔════╝ ██╔════╝████╗  ██║
 ██║   ██║██║██████╔╝█████╗       ██║  ███╗█████╗  ██╔██╗ ██║
 ╚██╗ ██╔╝██║██╔══██╗██╔══╝       ██║   ██║██╔══╝  ██║╚██╗██║
  ╚████╔╝ ██║██████╔╝███████╗     ╚██████╔╝███████╗██║ ╚████║
   ╚═══╝  ╚═╝╚═════╝ ╚══════╝      ╚═════╝ ╚══════╝╚═╝  ╚═══╝`)

	subtitle := subtitleStyle.Render("    Boilerplate Generator Tool  •  By Izeah01")
	divider := dividerStyle.Render("  ─────────────────────────────────────────────────────────")

	return title + "\n" + subtitle + "\n" + divider + "\n"
}
