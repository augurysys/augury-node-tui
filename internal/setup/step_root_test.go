package setup

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStepRoot_AutoDetectDisplaysPath(t *testing.T) {
	step := NewRootStep("/detected/augury-node")
	view := step.View()

	if !strings.Contains(view, "/detected/augury-node") {
		t.Error("View should display detected path")
	}
}

func TestStepRoot_UserInputUpdatesPath(t *testing.T) {
	step := NewRootStep("")

	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/custom/path")})

	if step.GetRootPath() != "/custom/path" {
		t.Errorf("Path should be '/custom/path', got %q", step.GetRootPath())
	}
}

func TestStepRoot_EnterConfirms(t *testing.T) {
	step := NewRootStep("/some/path")

	step, cmd := step.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Error("Enter should return a command")
	}
	if step.GetRootPath() != "/some/path" {
		t.Error("Path should be preserved")
	}
}
