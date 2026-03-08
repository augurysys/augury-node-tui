package styles

import "github.com/charmbracelet/lipgloss"

// Color palette inspired by modern TUI tools (k9s, lazygit)
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#7aa2f7") // Soft blue
	ColorSecondary = lipgloss.Color("#bb9af7") // Soft purple
	ColorAccent    = lipgloss.Color("#7dcfff") // Cyan

	// Status colors
	ColorSuccess = lipgloss.Color("#9ece6a") // Green
	ColorWarning = lipgloss.Color("#e0af68") // Yellow
	ColorError   = lipgloss.Color("#f7768e") // Red
	ColorInfo    = lipgloss.Color("#7aa2f7") // Blue

	// UI colors
	ColorFg       = lipgloss.Color("#c0caf5") // Light text
	ColorFgDim    = lipgloss.Color("#565f89") // Dimmed text
	ColorBg       = lipgloss.Color("#1a1b26") // Dark background
	ColorBgLight  = lipgloss.Color("#24283b") // Lighter background
	ColorBorder   = lipgloss.Color("#414868") // Border color
	ColorSelected = lipgloss.Color("#283457") // Selected background
)

// Base styles
var (
	// Text styles
	Bold      = lipgloss.NewStyle().Bold(true)
	Dim       = lipgloss.NewStyle().Foreground(ColorFgDim)
	Highlight = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)

	// Status styles
	Success = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	Warning = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
	Error   = lipgloss.NewStyle().Foreground(ColorError).Bold(true)
	Info    = lipgloss.NewStyle().Foreground(ColorInfo)

	// Component styles
	Border = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1)

	Section = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColorBorder).
		BorderTop(true).
		Padding(0, 1).
		MarginTop(1)

	Header = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		Padding(0, 1)

	Item = lipgloss.NewStyle().
		Padding(0, 1)

	ItemSelected = lipgloss.NewStyle().
		Background(ColorSelected).
		Foreground(ColorPrimary).
		Bold(true).
		Padding(0, 1)

	KeyHelp = lipgloss.NewStyle().
		Foreground(ColorFgDim).
		Padding(1, 0)

	Title = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)
)

// Helper functions
func StatusStyle(status string) lipgloss.Style {
	switch status {
	case "ready", "clean", "success", "built":
		return Success
	case "not ready", "dirty", "warning", "missing":
		return Warning
	case "error", "failed":
		return Error
	default:
		return Dim
	}
}

func KeyBinding(key, desc string) string {
	keyStyle := lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(ColorFg)
	return keyStyle.Render(key) + " " + descStyle.Render(desc)
}
