package primitives

import (
	"fmt"

	"github.com/augurysys/augury-node-tui/internal/styles"
	"github.com/charmbracelet/lipgloss"
)

// KeyHint displays a keyboard shortcut with description
type KeyHint struct {
	Key         string
	Description string
	Enabled     bool
}

// Render produces the styled key hint
func (h KeyHint) Render() string {
	palette := styles.DefaultPalette()
	typo := styles.DefaultTypography()

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(palette.AccentMauve)).
		Bold(true)

	var descStyle lipgloss.Style
	if h.Enabled {
		descStyle = typo.Body
	} else {
		descStyle = typo.Dim
	}

	keyPart := keyStyle.Render(fmt.Sprintf("[%s]", h.Key))
	descPart := descStyle.Render(h.Description)

	return fmt.Sprintf("%s %s", keyPart, descPart)
}
