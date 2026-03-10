package primitives

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestCard_RenderWithTitle(t *testing.T) {
	card := Card{
		Title:   "Test Card",
		Content: "This is content",
		Style:   CardNormal,
	}

	rendered := card.Render(40)

	if !strings.Contains(rendered, "Test Card") {
		t.Error("Card should contain title")
	}
	if !strings.Contains(rendered, "This is content") {
		t.Error("Card should contain content")
	}
	// Should have border characters
	if !strings.Contains(rendered, "─") && !strings.Contains(rendered, "┌") {
		t.Error("Card should have borders")
	}
}

func TestCard_WordWrapping(t *testing.T) {
	longContent := strings.Repeat("word ", 50)
	card := Card{
		Content: longContent,
		Style:   CardNormal,
	}

	rendered := card.Render(20)
	lines := strings.Split(rendered, "\n")

	for _, line := range lines {
		// Strip ANSI codes for accurate width check
		cleanLine := stripAnsi(line)
		if runewidth.StringWidth(cleanLine) > 22 { // 20 + border padding
			t.Errorf("Line exceeds max width: %d chars", runewidth.StringWidth(cleanLine))
		}
	}
}

func TestCard_StyleVariants(t *testing.T) {
	compact := Card{Content: "test", Style: CardCompact}
	normal := Card{Content: "test", Style: CardNormal}
	emphasized := Card{Content: "test", Style: CardEmphasized}

	compactLines := len(strings.Split(compact.Render(40), "\n"))
	normalLines := len(strings.Split(normal.Render(40), "\n"))

	// Compact (no padding) should not have more lines than normal
	if compactLines > normalLines {
		t.Error("Compact style should not have more lines than normal")
	}

	// Emphasized should be visually distinct (test by rendering)
	empRender := emphasized.Render(40)
	if empRender == normal.Render(40) {
		t.Error("Emphasized style should differ from normal")
	}
}

// Helper to strip ANSI codes
func stripAnsi(s string) string {
	// Simple ANSI strip: remove \x1b[...m sequences
	var result strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			i++
			continue
		}
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}
