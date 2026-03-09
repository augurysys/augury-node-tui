package styles

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestPalette_AllColorsDefined(t *testing.T) {
	p := DefaultPalette()

	colors := []string{
		p.Base, p.Surface0, p.Overlay0, p.Text,
		p.Success, p.Warning, p.Error, p.Info,
		p.AccentPink, p.AccentMauve, p.AccentPeach, p.AccentTeal,
	}

	for _, color := range colors {
		if color == "" {
			t.Errorf("color not defined")
		}
		// Validate hex format
		if len(color) != 7 || color[0] != '#' {
			t.Errorf("invalid hex color: %s", color)
		}
	}
}

func TestTypography_AllStylesDefined(t *testing.T) {
	typo := DefaultTypography()

	// Verify each style is usable
	_ = typo.Title.Render("test")
	_ = typo.Section.Render("test")
	_ = typo.Body.Render("test")
	_ = typo.Dim.Render("test")
	_ = typo.Highlight.Render("test")
}

func TestBorders_ThickThinNone(t *testing.T) {
	borders := DefaultBorders()
	empty := lipgloss.Border{}

	if borders.Thick == empty {
		t.Error("Thick border not defined")
	}
	if borders.Thin == empty {
		t.Error("Thin border not defined")
	}
	if borders.None != empty {
		t.Error("None border should be empty")
	}
}
