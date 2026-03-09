package styles

import "github.com/charmbracelet/lipgloss"

// Palette defines the application color scheme (Catppuccin Mocha)
type Palette struct {
	// Base colors
	Base     string // #1E1E2E (background)
	Surface0 string // #313244 (elevated surfaces)
	Overlay0 string // #6C7086 (dimmed text)
	Text     string // #CDD6F4 (primary text)

	// Semantic status colors
	Success string // #A6E3A1 (green)
	Warning string // #F9E2AF (yellow)
	Error   string // #F38BA8 (red)
	Info    string // #89B4FA (blue)

	// Accent colors (categorical mapping)
	AccentPink   string // #F5C2E7 (platforms)
	AccentMauve  string // #CBA6F7 (builds)
	AccentPeach  string // #FAB387 (caches)
	AccentTeal   string // #94E2D5 (validations)
	AccentBlue   string // #89B4FA (selection, links)
	AccentLavender string // #B4BEFE (table headers)
}

// DefaultPalette returns Catppuccin Mocha palette
func DefaultPalette() Palette {
	return Palette{
		Base:     "#1E1E2E",
		Surface0: "#313244",
		Overlay0: "#6C7086",
		Text:     "#CDD6F4",

		Success: "#A6E3A1",
		Warning: "#F9E2AF",
		Error:   "#F38BA8",
		Info:    "#89B4FA",

		AccentPink:       "#F5C2E7",
		AccentMauve:      "#CBA6F7",
		AccentPeach:      "#FAB387",
		AccentTeal:       "#94E2D5",
		AccentBlue:       "#89B4FA",
		AccentLavender:   "#B4BEFE",
	}
}

// Typography defines text styles
type Typography struct {
	Title     lipgloss.Style
	Section   lipgloss.Style
	Body      lipgloss.Style
	Dim       lipgloss.Style
	Highlight lipgloss.Style
}

// DefaultTypography returns standard text styles
func DefaultTypography() Typography {
	p := DefaultPalette()
	return Typography{
		Title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.AccentPink)),
		Section:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.Text)),
		Body:      lipgloss.NewStyle().Foreground(lipgloss.Color(p.Text)),
		Dim:       lipgloss.NewStyle().Foreground(lipgloss.Color(p.Overlay0)),
		Highlight: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.AccentMauve)),
	}
}

// Borders defines border styles
type Borders struct {
	Thick  lipgloss.Border
	Thin   lipgloss.Border
	Double lipgloss.Border
	None   lipgloss.Border
}

// DefaultBorders returns standard border styles
func DefaultBorders() Borders {
	return Borders{
		Thick:  lipgloss.ThickBorder(),
		Thin:   lipgloss.NormalBorder(),
		Double: lipgloss.DoubleBorder(),
		None:   lipgloss.Border{},
	}
}

// Spacing constants
const (
	SpacingCompact  = 0
	SpacingNormal   = 1
	SpacingSpacious = 2
)
