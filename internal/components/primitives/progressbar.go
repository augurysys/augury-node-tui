package primitives

import (
	"fmt"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/styles"
	"github.com/charmbracelet/lipgloss"
)

// ProgressBar displays progress as filled/unfilled blocks
type ProgressBar struct {
	Current int
	Total   int
	Width   int
	Label   string
}

// Render produces the styled progress bar
func (p ProgressBar) Render() string {
	palette := styles.DefaultPalette()
	typo := styles.DefaultTypography()

	// Calculate percentage
	var pct float64
	if p.Total > 0 {
		pct = float64(p.Current) / float64(p.Total) * 100
	}

	// Calculate filled blocks (Width is the bar width in blocks)
	barWidth := p.Width
	if barWidth < 10 {
		barWidth = 10
	}

	filledBlocks := int(float64(barWidth) * pct / 100)
	if filledBlocks < 0 {
		filledBlocks = 0
	}
	if filledBlocks > barWidth {
		filledBlocks = barWidth
	}

	// Build bar visual
	filled := strings.Repeat("█", filledBlocks)
	unfilled := strings.Repeat("░", barWidth-filledBlocks)

	// Color based on progress (sequential mapping)
	var barColor string
	if pct < 50 {
		barColor = palette.Overlay0
	} else if pct < 80 {
		barColor = palette.Warning
	} else {
		barColor = palette.Success
	}

	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor))

	// Format: "Label: ████░░ 82% (820/1000)"
	labelPart := typo.Body.Render(p.Label + ": ")
	barPart := barStyle.Render(filled + unfilled)
	pctPart := typo.Body.Render(fmt.Sprintf(" %.0f%%", pct))
	fractionPart := typo.Dim.Render(fmt.Sprintf(" (%d/%d)", p.Current, p.Total))

	return labelPart + barPart + pctPart + fractionPart
}
