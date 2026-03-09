package primitives

import (
	"github.com/augurysys/augury-node-tui/internal/styles"
	"github.com/charmbracelet/lipgloss"
)

// Status represents state for status-based color mapping
type Status int

const (
	StatusSuccess Status = iota
	StatusError
	StatusWarning
	StatusRunning
	StatusBlocked
	StatusUnavailable
)

// StatusBadge displays a colored status indicator
type StatusBadge struct {
	Label  string
	Status Status
}

// Render produces the styled status badge
func (b StatusBadge) Render() string {
	palette := styles.DefaultPalette()

	var icon string
	var color string

	switch b.Status {
	case StatusSuccess:
		icon = "✓"
		color = palette.Success
	case StatusError:
		icon = "✗"
		color = palette.Error
	case StatusWarning:
		icon = "⚠"
		color = palette.Warning
	case StatusRunning:
		icon = "●"
		color = palette.Info
	case StatusBlocked:
		icon = "⊘"
		color = palette.Overlay0
	case StatusUnavailable:
		icon = "◌"
		color = palette.Overlay0
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	return style.Render(icon + " " + b.Label)
}
