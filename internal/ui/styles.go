package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary  = lipgloss.Color("#7C3AED") // violet
	colorMuted    = lipgloss.Color("#6B7280")
	colorSuccess  = lipgloss.Color("#10B981")
	colorWarning  = lipgloss.Color("#F59E0B")
	colorDanger   = lipgloss.Color("#EF4444")
	colorSelected = lipgloss.Color("#7C3AED")
	colorHeader   = lipgloss.Color("#374151")
	colorBg       = lipgloss.Color("#1F2937")

	styleBold    = lipgloss.NewStyle().Bold(true)
	styleMuted   = lipgloss.NewStyle().Foreground(colorMuted)
	styleSuccess = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	styleDanger  = lipgloss.NewStyle().Foreground(colorDanger).Bold(true)
	styleWarning = lipgloss.NewStyle().Foreground(colorWarning)
	stylePrimary = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)

	styleLogo = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			Padding(0, 1)

	styleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)

	styleCategoryHeader = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true).
				MarginTop(1)

	styleSelected = lipgloss.NewStyle().
			Foreground(colorSelected).
			Bold(true)

	styleCursor = lipgloss.NewStyle().
			Background(colorPrimary).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	styleSize = lipgloss.NewStyle().
			Foreground(colorWarning).
			Width(9).
			Align(lipgloss.Right)

	styleAge = lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(10).
			Align(lipgloss.Right)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)

	styleStatusBar = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)
)

const logo = "🧹 broom"
const tagline = "Sweep your dev machine clean"
