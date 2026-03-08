package styles

import "github.com/charmbracelet/lipgloss"

// Catppuccin Mocha palette
type PaletteColors struct {
	// Base colors
	Base     string
	Mantle   string
	Surface0 string
	Surface1 string
	Overlay0 string
	Text     string
	Subtext0 string

	// Semantic status colors
	Success string
	Warning string
	Error   string
	Info    string

	// Categorical accent colors
	AccentPink     string
	AccentMauve    string
	AccentPeach    string
	AccentTeal     string
	AccentSapphire string
}

var Palette = PaletteColors{
	// Base
	Base:     "#1E1E2E",
	Mantle:   "#181825",
	Surface0: "#313244",
	Surface1: "#45475A",
	Overlay0: "#6C7086",
	Text:     "#CDD6F4",
	Subtext0: "#A6ADC8",

	// Status
	Success: "#A6E3A1",
	Warning: "#F9E2AF",
	Error:   "#F38BA8",
	Info:    "#89B4FA",

	// Accents
	AccentPink:     "#F5C2E7",
	AccentMauve:    "#CBA6F7",
	AccentPeach:    "#FAB387",
	AccentTeal:     "#94E2D5",
	AccentSapphire: "#74C7EC",
}

// Predefined typography styles
var (
	TitleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(Palette.AccentPink))
	SectionStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(Palette.Text))
	BodyStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Text))
	DimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.Overlay0))
	HighlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(Palette.AccentMauve)).Background(lipgloss.Color(Palette.Surface1))

	BorderThick = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(Palette.Surface1))
	BorderThin  = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(Palette.Overlay0))
)

// Spacing constants
const (
	SpacingCompact  = 0
	SpacingNormal   = 1
	SpacingSpacious = 2
)
