package setup

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStepSuccess_DisplaysCompletionMessage(t *testing.T) {
	step := NewSuccessStep([]string{})
	view := step.View()

	if !strings.Contains(view, "Setup Complete") || !strings.Contains(view, "✓") {
		t.Error("View should show success message")
	}
}

func TestStepSuccess_DisplaysNextSteps(t *testing.T) {
	step := NewSuccessStep([]string{})
	view := step.View()

	if !strings.Contains(view, "Run commands") && !strings.Contains(view, "Explore build") {
		t.Error("View should mention next steps (run commands, explore screens)")
	}
}

func TestStepSuccess_QuitOnEnter(t *testing.T) {
	step := NewSuccessStep([]string{})
	step, cmd := step.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Enter should return quit command")
	}
	// Note: can't easily test tea.Quit here, but verify cmd is not nil
}

func TestStepSuccess_QuitOnQ(t *testing.T) {
	step := NewSuccessStep([]string{})
	step, cmd := step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Fatal("q key should return quit command")
	}
}

func TestSuccessStep_UsesCardComponent(t *testing.T) {
	step := NewSuccessStep([]string{})
	view := step.View()

	// Should have card borders (Normal: ─┌ or ═╔, Thick: ━┏)
	if !strings.Contains(view, "─") && !strings.Contains(view, "┌") &&
		!strings.Contains(view, "━") && !strings.Contains(view, "┏") &&
		!strings.Contains(view, "═") && !strings.Contains(view, "╔") {
		t.Error("Success screen should use Card component with borders")
	}

	// Should contain success message
	if !strings.Contains(view, "Setup Complete") {
		t.Error("Should show success message")
	}
}
