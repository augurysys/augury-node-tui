package styles

import (
	"strings"
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
		for i := 1; i < 7; i++ {
			c := color[i]
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				t.Errorf("invalid hex color: %s", color)
				break
			}
		}
	}
}

func TestTypography_AllStylesDefined(t *testing.T) {
	typo := DefaultTypography()

	for name, style := range map[string]lipgloss.Style{
		"Title": typo.Title, "Section": typo.Section, "Body": typo.Body,
		"Dim": typo.Dim, "Highlight": typo.Highlight,
	} {
		out := style.Render("test")
		if out == "" || !strings.Contains(out, "test") {
			t.Errorf("%s style produced invalid output: %q", name, out)
		}
	}
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
