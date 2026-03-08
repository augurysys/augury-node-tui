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
		t.Fatal("Enter should return a command")
	}
	if step.GetRootPath() != "/some/path" {
		t.Error("Path should be preserved")
	}

	// Run the command and verify the message
	msg := cmd()
	confirmMsg, ok := msg.(RootConfirmedMsg)
	if !ok {
		t.Fatalf("Command should return RootConfirmedMsg, got %T", msg)
	}
	if confirmMsg.Path != "/some/path" {
		t.Errorf("Message path should be '/some/path', got %q", confirmMsg.Path)
	}
}

func TestStepRoot_BackspaceHandlesUTF8(t *testing.T) {
	step := NewRootStep("")

	// Type "café"
	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("café")})
	if step.GetRootPath() != "café" {
		t.Errorf("Should have 'café', got %q", step.GetRootPath())
	}

	// Backspace should remove 'é', not corrupt the string
	step, _ = step.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if step.GetRootPath() != "caf" {
		t.Errorf("Backspace should remove 'é', got %q", step.GetRootPath())
	}
}
