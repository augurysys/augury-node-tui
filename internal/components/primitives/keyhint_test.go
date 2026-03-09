package primitives

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestKeyHint_EnabledState(t *testing.T) {
	hint := KeyHint{
		Key:         "b",
		Description: "build",
		Enabled:     true,
	}

	rendered := hint.Render()

	if !strings.Contains(rendered, "b") {
		t.Error("Should contain key")
	}
	if !strings.Contains(rendered, "build") {
		t.Error("Should contain description")
	}
	// Enabled should not be dimmed (no "blocked" text)
	if strings.Contains(strings.ToLower(rendered), "blocked") {
		t.Error("Enabled hint should not show 'blocked'")
	}
}

func TestKeyHint_DisabledState(t *testing.T) {
	// Ensure color profile so Body vs Dim produce different ANSI output
	oldProfile := lipgloss.ColorProfile()
	t.Cleanup(func() {
		lipgloss.SetColorProfile(oldProfile)
	})
	lipgloss.SetColorProfile(termenv.TrueColor)

	hint := KeyHint{
		Key:         "b",
		Description: "build",
		Enabled:     false,
	}

	rendered := hint.Render()

	if !strings.Contains(rendered, "b") {
		t.Error("Should contain key")
	}
	if !strings.Contains(rendered, "build") {
		t.Error("Should contain description")
	}

	// Disabled hints should be visually distinct (dimmed color)
	enabledHint := KeyHint{Key: "b", Description: "build", Enabled: true}
	if rendered == enabledHint.Render() {
		t.Error("Disabled hint should render differently than enabled")
	}
}

func TestKeyHint_Format(t *testing.T) {
	hint := KeyHint{
		Key:         "ctrl+c",
		Description: "cancel",
		Enabled:     true,
	}

	rendered := hint.Render()

	// Should have bracket format [key]
	if !strings.Contains(rendered, "[") || !strings.Contains(rendered, "]") {
		t.Error("Key hint should use bracket format")
	}
}
