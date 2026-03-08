package setup

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStepSuccess_DisplaysCompletionMessage(t *testing.T) {
	step := NewSuccessStep()
	view := step.View()

	if !strings.Contains(view, "Setup Complete") || !strings.Contains(view, "✓") {
		t.Error("View should show success message")
	}
}

func TestStepSuccess_DisplaysNextSteps(t *testing.T) {
	step := NewSuccessStep()
	view := step.View()

	if !strings.Contains(view, "augury-node-tui") {
		t.Error("View should mention how to run the TUI")
	}
}

func TestStepSuccess_QuitOnEnter(t *testing.T) {
	step := NewSuccessStep()
	step, cmd := step.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Enter should return quit command")
	}
	// Note: can't easily test tea.Quit here, but verify cmd is not nil
}
