package primitives

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/augurysys/augury-node-tui/internal/styles"
)

// CardStyle defines card appearance variants
type CardStyle int

const (
	CardCompact    CardStyle = iota // No padding
	CardNormal                      // Normal padding
	CardEmphasized                  // Thick border, accent color
)

// Card is a bordered container with optional title
type Card struct {
	Title   string
	Content string
	Style   CardStyle
}

// Render produces the styled card within given width
func (c Card) Render(width int) string {
	palette := styles.DefaultPalette()
	borders := styles.DefaultBorders()
	typo := styles.DefaultTypography()

	// Choose border and style based on variant
	var border lipgloss.Border
	var borderColor string
	var padding int
	var bgColor string

	switch c.Style {
	case CardCompact:
		border = borders.Thin
		borderColor = palette.Overlay0
		padding = 0
		bgColor = ""
	case CardEmphasized:
		border = borders.Thick
		borderColor = palette.AccentMauve
		padding = 2
		bgColor = palette.Surface0
	default: // CardNormal
		border = borders.Double
		borderColor = palette.Overlay0
		padding = 2
		bgColor = palette.Surface0
	}

	style := lipgloss.NewStyle().
		Border(border).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(width - 2). // Account for borders
		Padding(1, padding) // Vertical padding for breathing room

	if bgColor != "" {
		style = style.Background(lipgloss.Color(bgColor))
	}

	// Word-wrap content
	contentWidth := width - 2 - padding*2
	wrapped := wordWrap(c.Content, contentWidth)

	// Combine title and content - title centered and bold
	var content string
	if c.Title != "" {
		titleStyle := typo.Section.Copy().Bold(true)
		titleLine := titleStyle.Render(c.Title)
		content = lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, titleLine) + "\n" + wrapped
	} else {
		content = wrapped
	}

	return style.Render(content)
}

// wordWrap breaks text at word boundaries to fit width (display width, not bytes)
func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	words := strings.Fields(text)
	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		wordWidth := runewidth.StringWidth(word)
		lineWidth := runewidth.StringWidth(currentLine.String())
		if lineWidth == 0 {
			currentLine.WriteString(word)
		} else if lineWidth+1+wordWidth <= width {
			currentLine.WriteString(" " + word)
		} else {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		}
	}
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}
	return strings.Join(lines, "\n")
}
